package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"io"

	"fmt"
	"strings"

	"golang.org/x/crypto/sha3"
)

var (
    decrSalt   = ""
    TLSCert tls.Certificate
    X509cert *x509.Certificate
    X509pool *x509.CertPool
)

func CryptInit(certStrPemB64, caCertStrPemB64, keyStrPemB64, keyStrPemEncrB64, saltEncrB64, passphrase string) error {
	if len(certStrPemB64) == 0 {
		return fmt.Errorf("no certificate")
	}
	if len(caCertStrPemB64) == 0 {
		return fmt.Errorf("no CA certificate")
	}
	if len(keyStrPemB64) == 0 && len(keyStrPemEncrB64) == 0 {
		return fmt.Errorf("no encoded private key")
	}
	
	if saltEncrB64 != "" && passphrase != "" {
		if saltEncr, err := hex.DecodeString(saltEncrB64); err == nil {
			if decr, err := Decrypt(saltEncr, passphrase); err == nil {
				decrSalt = string(decr)
			} else {
				return fmt.Errorf("error decrypting salt: %s", err.Error())
			}
		} else {
			return fmt.Errorf("error decoding salt: %s", err.Error())
		}
	}
	
	var err     error
	var certBpem   []byte
	var caCertBpem []byte
    var keyBpem    []byte

	// The certs are public and not a secret - they are B64-encoded to be able to put them on one line (in 1Password, GH, etc)
	if certBpem, err = base64.StdEncoding.DecodeString(certStrPemB64); err != nil {
		return fmt.Errorf("cannot B64-decode server certificate: %s", err.Error())
	}
    if caCertBpem, err = base64.StdEncoding.DecodeString(caCertStrPemB64); err != nil {
		return fmt.Errorf("cannot B64-decode CA certificate: %s", err.Error())
	}

	// We should have a B64-encoded key, but the key could be encrypted, or not. The key in the binary would be encrypted, in servers probably not.
	// Try decoding the B64/non-encrypted PEM - it will fail if bad.
	if keyStrPemB64 != "" {
		if keyBpem, err = base64.StdEncoding.DecodeString(keyStrPemB64); err != nil {
			if keyStrPemEncrB64 == "" {
				return fmt.Errorf("cannot B64-decode private key: %s", err.Error())
			}
		}
	}
	if len(keyBpem) == 0 {
		if keyStrPemEncrB64 != "" {
			if keyBpem, err = base64.StdEncoding.DecodeString(keyStrPemEncrB64); err != nil {
				// Couldn't decode this one either - nothing more we can do. Fail.
				return fmt.Errorf("cannot B64-decode encrypted private key: %s", err.Error())
			} else if passphrase != "" {
				// We were able to B64-decode, and we have the passphrase to decrypt, so try to decrypt.
				if keyBpem, err = Decrypt(keyBpem, passphrase); err != nil {
					// Could not decrypt with the passphrase. Fail.
					return fmt.Errorf("cannot decrypt private key: %s", err.Error())
				}
				// If we get to here, we decrypted the PEM. Good.
			} else {
				return fmt.Errorf("no passphrase provided to decrypt private key")
			}
		}
	}
	if len(keyBpem) == 0 {
		return fmt.Errorf("could not create private key PEM") // Don't think we'd ever get here.
	}

	// If we get here, we either B-64 decoded the non-encrypted key, or B-64 decoded and decrypted the encrypted key. Either way, stored in keyBpem.

	X509pool = x509.NewCertPool()
	if !X509pool.AppendCertsFromPEM(caCertBpem) {
		return fmt.Errorf("cannot append CA cert to cert pool")
	}

	if TLSCert, err = tls.X509KeyPair(certBpem, keyBpem); err != nil {
		return fmt.Errorf("cannot load TLS key-pair: %s", err.Error())
	}
	if block, _ := pem.Decode(certBpem); block != nil {
		if X509cert, err = x509.ParseCertificate(block.Bytes); err != nil {
			return fmt.Errorf("cannot parse x509 certificate from server certificate PEM: %s", err.Error())
		}
	} else {
		return fmt.Errorf("cannot PEM decode server certificate from server certificate PEM")
	}
    return nil
}

func Hash(in string) (string, error) {
	if len(in) == 64 {
		return in, nil
	}
	if decrSalt == "" {
		return in, fmt.Errorf("cannot perform hash, no salt")
	}
	s := strings.TrimSpace(in)
	if s == "" {
		return "", fmt.Errorf("empty input string")
	}
	z := sha3.New256()
	z.Write([]byte(s + decrSalt))
	return hex.EncodeToString(z.Sum(nil)), nil
}


func EncryptB64(decr []byte, phrase string) (string, error) {
	if encr, err := Encrypt(decr, phrase); err == nil {
	  encStr := base64.StdEncoding.EncodeToString(encr)
	  return encStr, nil
	} else {
	  return "", fmt.Errorf("error encrypting data: %s", err.Error())
	}
  }
  
func DecryptB64(encr, phrase string) ([]byte, error) {
	if decStr, err := base64.StdEncoding.DecodeString(encr); err == nil {
		if decr, err := Decrypt([]byte(decStr), phrase); err == nil {
			return decr, nil
		} else {
			return nil, fmt.Errorf("error decrypting data: %s", err.Error())
		}
	} else {
		return nil, fmt.Errorf("error decoding data: %s", err.Error())
	}
}

func Encrypt(decr []byte, phrase string) ([]byte, error) {
    if block, err := aes.NewCipher([]byte(phrase)); err == nil {
        if gcm, err := cipher.NewGCM(block); err == nil {
            nonce := make([]byte, gcm.NonceSize())
            if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
                return nil, err
            }
            return gcm.Seal(nonce, nonce, decr, nil), nil   // Embed nonce into the start of the encrypted data.
        } else {
            return nil, err
        }
    } else {
        return nil, err
    }
}

func Decrypt(encr []byte, phrase string) ([]byte, error) {
    if block, err := aes.NewCipher([]byte(phrase)); err == nil {
        if gcm, err := cipher.NewGCM(block); err == nil {
            nonceSize := gcm.NonceSize()
            nonce, ciphertext := encr[:nonceSize], encr[nonceSize:]     // Nonce is embedded in the encrypted data.
            if plaintext, err := gcm.Open(nil, nonce, ciphertext, nil); err == nil {
                return plaintext, nil
            } else {
                return nil, fmt.Errorf("gcm.open: %s", err.Error())
            }
        } else {
            return nil, fmt.Errorf("NewGCM: %s", err.Error())
        }
    } else {
        return nil, fmt.Errorf("NewCipher: %s", err.Error())
    }
}

func HasOU(ou string) bool {
	if X509cert != nil {
		for _, certOU := range X509cert.Subject.OrganizationalUnit {
			if strings.EqualFold(certOU, ou) {
				return true
			}
		}
	}
	return false
}

func X509cname() string {
	if X509cert != nil {
		return X509cert.Subject.CommonName
	}
	return ""
}
func X509ou() string {
	if X509cert != nil {
		return strings.Join(X509cert.Subject.OrganizationalUnit, ",")
	}
	return ""
}
