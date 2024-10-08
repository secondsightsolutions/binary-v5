package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var rgxNdc542 *regexp.Regexp
var rgxNdc11  *regexp.Regexp
var rgxNdc541 *regexp.Regexp
var rgxNdc532 *regexp.Regexp
var rgxNdc442 *regexp.Regexp
var rgxNdc641 *regexp.Regexp
var rgxNdc632 *regexp.Regexp

var rgxDigits *regexp.Regexp
var rgxWhlInv *regexp.Regexp
var rgxPresQ  *regexp.Regexp
var rgxSvPrQ  *regexp.Regexp
var rgxI340   *regexp.Regexp
var rgxI340Pr *regexp.Regexp
var rgxAlpha  *regexp.Regexp
var rgxNonDig *regexp.Regexp
var rgxDea    *regexp.Regexp
var rgxHCPC   *regexp.Regexp
var rgxPOS    *regexp.Regexp

func init() {
	rgxNdc542 = regexp.MustCompile(`^\d{5}-\d{4}-\d{2}$`)
	rgxNdc11  = regexp.MustCompile(`^\d{11}$`)
	rgxNdc541 = regexp.MustCompile(`^\d{5}-\d{4}-\d{1}$`)
	rgxNdc532 = regexp.MustCompile(`^\d{5}-\d{3}-\d{2}$`)
	rgxNdc442 = regexp.MustCompile(`^\d{4}-\d{4}-\d{2}$`)
	rgxNdc632 = regexp.MustCompile(`^\d{6}-\d{3}-\d{2}$`)
	rgxNdc641 = regexp.MustCompile(`^\d{6}-\d{4}-\d{1}$`)
	rgxDigits = regexp.MustCompile(`^\d+$`)
	rgxWhlInv = regexp.MustCompile(`^[0-9-,.#]+$`)
	rgxPresQ  = regexp.MustCompile(`^(01|1|12)$`)
	rgxSvPrQ  = regexp.MustCompile(`^(01|1|05|5|07|7|10|12)$`)
	rgxI340   = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	rgxI340Pr = regexp.MustCompile(`^CAN|RRC|CAH|PED|FQ|SCH|BL|DSH|CH|TB|FP|RW4|HV|STD|NH|UI|URB|HM|RWI`)
	rgxAlpha  = regexp.MustCompile(`^[a-zA-Z]+$`)
	rgxNonDig = regexp.MustCompile(`[^0-9]+`)
	rgxDea    = regexp.MustCompile(`^(A|B|C|D|E|F|G|H|J|K|L|M|P|R|S|T|U|X)[A-Za-z0-9][\d]+$`)
	rgxHCPC   = regexp.MustCompile(`^[a-zA-Z][0-9]+$`)
	rgxPOS    = regexp.MustCompile(`^[a-zA-Z0-9]{2}$`)
}

var hospitalTypeCodes = []string{"RRC", "SCH", "CAH", "DSH", "PED", "CAN"}

var hospitalACATypeCodes = []string{"SCH", "RRC", "CAH", "CAN"}

func IsHospital(id340b string) bool {
	for _, prfx := range hospitalTypeCodes {
		if strings.HasPrefix(id340b, prfx) {
			return true
		}
	}
	return false
}

func IsHospitalACA(id340b string) bool {
	id340b = strings.ToUpper(id340b)
	for _, prfx := range hospitalACATypeCodes {
		if strings.HasPrefix(id340b, prfx) {
			return true
		}
	}
	return false
}

func IsGrantee(id340b string) bool {
	upr := strings.ToUpper(id340b)

	return !IsHospital(upr)
}

func Is64bitHash(val string) bool {
	return len(val) == 64
}

// func IsValidDate(val string) bool {
// 	for _, _fmt := range Fmts {
// 		if _, err := time.Parse(_fmt, val); err == nil {
// 			return true
// 		}
// 	}
// 	return false
// }

func IsValidSPID(spid, qual string) bool {
	return ValidateSPID(spid, qual) == nil
}
func ValidateSPID(spid, qual string) error {
	switch qual {
	case "1", "01":
		if !IsValidNPI(spid) {
			return fmt.Errorf("%s is not a valid NPI", spid)
		}
	case "5", "05":
		if !IsValidDEA(spid) {
			return fmt.Errorf("%s is not a valid DEA", spid)
		}
	case "7", "07":
		if !IsValidNCPDP(spid) {
			return fmt.Errorf("%s is not a valid NCPDP", spid)
		}
	case "10":
	case "12":
	default:
		return fmt.Errorf("qualifier %s not valid/recognized", qual)
	}
	return nil
}

func IsValidPOS(val string) bool {
	return ValidatePOS(val) == nil
}
func ValidatePOS(val string) error {
	if !rgxPOS.MatchString(val) {
		return fmt.Errorf("invalid format")
	}
	return nil
}

func IsValidHCPC(val string) bool {
	return ValidateHCPC(val) == nil
}
func ValidateHCPC(val string) error {
	if len(val) == 0 || len(val) > 5 {
		return fmt.Errorf("invalid length")
	}
	if !rgxHCPC.MatchString(val) {
		return fmt.Errorf("invalid format")
	}
	return nil
}

func IsValidRxNumber(val string) bool {
	return ValidateRxNumber(val) == nil
}
func ValidateRxNumber(val string) error {
	// Either RX-CNT or RX. RX length >=5 and <=12. Count 0 - 99 (probably should be 1-99). Both RX and CNT must be integers.
	rx  := val
	cnt := ""
	tks := strings.Split(rx, "-")
	if len(tks) > 2 {
		return fmt.Errorf("token count must be 1 or 2 (is %d)", len(tks))
	} else if len(tks) == 2 {
		rx = tks[0]
		cnt = tks[1]
		if len(cnt) < 1 || len(cnt) > 2 {
			return fmt.Errorf("fill count length must be 1 or 2 (is %d)", len(cnt))
		}
		if num, err := strconv.ParseInt(cnt, 10, 32); err != nil {
			return err
		} else if num < 0 || num > 99 {
			return fmt.Errorf("fill count must be >= 0 and <= 99 (is %d)", num)
		}
	}
	if len(rx) < 5 || len(rx) > 12 {
		return fmt.Errorf("base length must be >= 5 and <= 12 (is %d)", len(rx))
	}
	return nil
}

func IsValidNDC(val string) bool {
	return ValidateNDC(val) == nil
}
func ValidateNDC(val string) error {
	if rgxNdc11.MatchString(val) || rgxNdc542.MatchString(val) {
		return nil
	}
	return fmt.Errorf("not 5-4-2 nor 11 digit (is %s)", val)
}

func IsValidQuantity(val string) bool {
	return ValidateQuantity(val) == nil
}
func ValidateQuantity(val string) error {
	if _, err := strconv.ParseFloat(val, 64); err == nil {
		return fmt.Errorf("value not float (is %s)", val)
	}
	return nil
}

func IsValidWholesalerInvoiceNumber(val string) bool {
	return ValidateWholesalerInvoiceNumber(val) == nil
}
func ValidateWholesalerInvoiceNumber(val string) error {
	if !rgxWhlInv.MatchString(val) {
		return fmt.Errorf("value not string (is %s)", val)
	}
	return nil
}

func IsValidPrescriberIDQualifier(val string) bool {
	return ValidatePrescriberIDQualifier(val) == nil
}
func ValidatePrescriberIDQualifier(val string) error {
	if !rgxPresQ.MatchString(val) {
		return fmt.Errorf("value not valid prescriber id qualifier (is %s)", val)
	}
	return nil
}

func IsValidPhysicianID(val string) bool {
	return ValidatePhysicianID(val) == nil
}
func ValidatePhysicianID(val string) error {
	if !IsValidNPI(val) && !IsValidNCPDP(val) && !IsValidDEA(val) {
		return fmt.Errorf("value not NPI, NCPDP, nor DEA (is %s)", val)
	}
	return nil
}

func IsValidPrescriberID(val string) bool {
	return ValidatePrescriberID(val) == nil
}
func ValidatePrescriberID(val string) error {
	if !IsValidNPI(val) && !IsValidNCPDP(val) && !IsValidDEA(val) {
		return fmt.Errorf("value not NPI, NCPDP, nor DEA (is %s)", val)
	}
	return nil
}

func IsValidServiceProviderIDQualifier(val string) bool {
	return ValidateServiceProviderIDQualifier(val) == nil
}
func ValidateServiceProviderIDQualifier(val string) error {
	if !rgxSvPrQ.MatchString(val) {
		return fmt.Errorf("value not valid service provider id qualifier (is %s)", val)
	}
	return nil
}

func IsValidServiceProviderID(val string) bool {
	return ValidateServiceProviderID(val) == nil
}
func ValidateServiceProviderID(val string) error {
	if !IsValidNPI(val) && !IsValidNCPDP(val) && !IsValidDEA(val) {
		return fmt.Errorf("value not NPI, NCPDP, nor DEA (is %s)", val)
	}
	return nil
}

func IsValidBillingProviderIDQualifier(val string) bool {
	return ValidateBillingProviderIDQualifier(val) == nil
}
func ValidateBillingProviderIDQualifier(val string) error {
	if !rgxSvPrQ.MatchString(val) {
		return fmt.Errorf("value not valid service provider id qualifier (is %s)", val)
	}
	return nil
}

func IsValidBillingProviderID(val string) bool {
	return ValidateBillingProviderID(val) == nil
}
func ValidateBillingProviderID(val string) error {
	if !IsValidNPI(val) && !IsValidNCPDP(val) && !IsValidDEA(val) {
		return fmt.Errorf("value not NPI, NCPDP, nor DEA (is %s)", val)
	}
	return nil
}

func IsValidContractedEntityID(val string) bool {
	return ValidateContractedEntityID(val) == nil
}
func ValidateContractedEntityID(val string) error {
	if strings.HasSuffix(val, "-") {
		return fmt.Errorf("value has - at end (is %s)", val)
	}
	if !rgxI340.MatchString(val) {
		return fmt.Errorf("value not valid 340B ID value (is %s)", val)
	}

	if !rgxI340Pr.MatchString(val) {
		return fmt.Errorf("value does not have valid prefix (is %s)", val)
	}
	if !rgxAlpha.MatchString(val[0:2]) {
		return fmt.Errorf("value does not start with two characters (is %s)", val)
	}
	return nil
}

func IsValidNPI(val string) bool {
	if !rgxDigits.MatchString(val) {
		return false
	}
	if strings.HasPrefix(val, "0") {
		return false
	}
	if len(val) != 10 {
		return false
	}
	fixed  := rgxNonDig.ReplaceAllString(val, "")
	nCheck := 24
	bEven  := false

	for n := len(fixed) - 1; n >= 0; n-- {
		strArr := []rune(fixed)
		nDigit, _ := strconv.Atoi(string(strArr[n]))

		if bEven {
			nDigit = nDigit * 2
		}
		if bEven && nDigit > 9 {
			nDigit -= 9
		}
		nCheck += nDigit
		bEven = !bEven
	}
	return nCheck%10 == 0
}

func IsValidNCPDP(val string) bool {
	if !rgxDigits.MatchString(val) {
		return false
	}
	if strings.HasPrefix(val, "00") {
		return false
	}
	if len(val) != 7 {
		return false
	}
	chars := strings.Split(val, "")

	x, err := strconv.Atoi(chars[0])
	if err != nil {
		return false
	}
	y, err := strconv.Atoi(chars[2])
	if err != nil {
		return false
	}
	z, err := strconv.Atoi(chars[4])
	if err != nil {
		return false
	}

	a := x + y + z

	x, err = strconv.Atoi(chars[1])
	if err != nil {
		return false
	}
	y, err = strconv.Atoi(chars[3])
	if err != nil {
		return false
	}
	z, err = strconv.Atoi(chars[5])
	if err != nil {
		return false
	}

	b := 2 * (x + y + z)

	x, err = strconv.Atoi(chars[6])
	if err != nil {
		return false
	}
	return (a+b)%10 == x
}

func IsValidDEA(val string) bool {
	if len(val) != 9 {
		return false
	}
	if !rgxDea.MatchString(val) {
		return false
	}
	chars := strings.Split(val, "")
	uid := chars[2:8]
	checkSum := chars[len(chars)-1]

	a, err := strconv.Atoi(uid[0])
	if err != nil {
		return false
	}
	b, err := strconv.Atoi(uid[2])
	if err != nil {
		return false
	}
	c, err := strconv.Atoi(uid[4])
	if err != nil {
		return false
	}
	d, err := strconv.Atoi(uid[1])
	if err != nil {
		return false
	}
	e, err := strconv.Atoi(uid[3])
	if err != nil {
		return false
	}
	f, err := strconv.Atoi(uid[5])
	if err != nil {
		return false
	}
	value := (a + b + c) + 2*(d+e+f)
	valueStr := fmt.Sprintf("%d", value)
	valueChars := strings.Split(valueStr, "")

	return valueChars[len(valueChars)-1] == checkSum
}

func IsValidClaimsConformFlag(val string) bool {
	cleaned := strings.ToLower(val)
	return cleaned == "true" || cleaned == "false"
}
