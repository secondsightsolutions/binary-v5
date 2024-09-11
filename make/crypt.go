package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"

	"github.com/secondsightsolutions/binary-v4/util"
)

func main() {
	_encr  := flag.String("encrypt", "", "File to encrypt")
	_decr  := flag.String("decrypt", "", "File to decrypt")
	_encd  := flag.String("encode",  "", "File to b64-encode")
	_decd  := flag.String("decode",  "", "File to b64-decode")
	_file  := flag.String("output",  "", "Output file")
  _phrs := flag.String("phrase",  "", "Passphrase")
	flag.Parse()

	if *_encr != "" {
		if err := encryptFile(*_encr, *_file, *_phrs); err != nil {
			fmt.Println(err.Error())
		}
	} else if *_decr != "" {
		if err := decryptFile(*_decr, *_file, *_phrs); err != nil {
			fmt.Println(err.Error())
		}
	} else if *_encd != "" {
		if err := encodeFile(*_encd, *_file); err != nil {
			fmt.Println(err.Error())
		}
	} else if *_decd != "" {
		if err := decodeFile(*_decd, *_file); err != nil {
			fmt.Println(err.Error())
		}
	} else {
		fmt.Println("You must pick one of the options!")
	}
}

func encryptFile(fm, to, phrase string) error {
	if clear, err := os.ReadFile(fm); err == nil {
		if encrStr, err := util.EncryptB64(clear, phrase); err == nil {
			if err := os.WriteFile(to, []byte(encrStr), 0400); err != nil {
				return fmt.Errorf("error writing file %s: %s", to, err.Error())
			}
		} else {
			return err
		}
	} else {
		return err
	}
	return nil
}

func decryptFile(fm, to, phrase string) error {
	if encrypted, err := os.ReadFile(fm); err == nil {
		if decrStr, err := util.DecryptB64(string(encrypted), phrase); err == nil {
			if err := os.WriteFile(to, []byte(decrStr), 0400); err != nil {
				return fmt.Errorf("error writing file %s: %s", to, err.Error())
			}
		} else {
			return err
		}
	} else {
		return err
	}
	return nil
}

func encodeFile(fm, to string) error {
	if clear, err := os.ReadFile(fm); err == nil {
		encStr := base64.StdEncoding.EncodeToString(clear)
		if err := os.WriteFile(to, []byte(encStr), 0400); err != nil {
			return fmt.Errorf("error writing file %s: %s", to, err.Error())
		}
		return nil
	} else {
		return err
	}
}

func decodeFile(fm, to string) error {
	if encoded, err := os.ReadFile(fm); err == nil {
		decoded := make([]byte, len(encoded))
		if _, err := base64.StdEncoding.Decode(decoded, encoded); err == nil {
			if err := os.WriteFile(to, decoded, 0400); err != nil {
				return fmt.Errorf("error writing file %s: %s", to, err.Error())
			}
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}
