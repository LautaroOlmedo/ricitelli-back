package value_object

import (
	"errors"
	"math/big"
	"strings"
)

var ErrInvalidDecimal = errors.New("invalid decimal string")

// ValidateDecimal verifies an exact base-10 decimal without converting it to a
// floating-point value. maxScale limits digits after the decimal separator.
func ValidateDecimal(value string, maxScale int, positive bool) error {
	coefficient, scale, err := parseDecimal(value)
	if err != nil || coefficient.Sign() < 0 || scale > maxScale {
		return ErrInvalidDecimal
	}
	if positive && coefficient.Sign() <= 0 {
		return ErrInvalidDecimal
	}
	return nil
}

// AddDecimal adds exact base-10 decimal strings and returns a canonical string.
func AddDecimal(left, right string) (string, error) {
	lc, ls, err := parseDecimal(left)
	if err != nil {
		return "", err
	}
	rc, rs, err := parseDecimal(right)
	if err != nil {
		return "", err
	}
	lc, rc, scale := alignDecimals(lc, ls, rc, rs)
	return formatDecimal(new(big.Int).Add(lc, rc), scale), nil
}

// SubtractDecimal subtracts exact base-10 decimal strings.
func SubtractDecimal(left, right string) (string, error) {
	lc, ls, err := parseDecimal(left)
	if err != nil {
		return "", err
	}
	rc, rs, err := parseDecimal(right)
	if err != nil {
		return "", err
	}
	lc, rc, scale := alignDecimals(lc, ls, rc, rs)
	return formatDecimal(new(big.Int).Sub(lc, rc), scale), nil
}

// CompareDecimal compares exact base-10 decimal strings.
func CompareDecimal(left, right string) (int, error) {
	lc, ls, err := parseDecimal(left)
	if err != nil {
		return 0, err
	}
	rc, rs, err := parseDecimal(right)
	if err != nil {
		return 0, err
	}
	lc, rc, _ = alignDecimals(lc, ls, rc, rs)
	return lc.Cmp(rc), nil
}

// DecimalToUint64 converts a positive whole decimal such as "12" or
// "12.0000" to uint64 without using floating-point arithmetic.
func DecimalToUint64(value string) (uint64, error) {
	coefficient, scale, err := parseDecimal(value)
	if err != nil || coefficient.Sign() <= 0 {
		return 0, ErrInvalidDecimal
	}
	divisor := pow10(scale)
	remainder := new(big.Int)
	whole := new(big.Int)
	whole.QuoRem(coefficient, divisor, remainder)
	if remainder.Sign() != 0 || !whole.IsUint64() {
		return 0, ErrInvalidDecimal
	}
	return whole.Uint64(), nil
}

func parseDecimal(value string) (*big.Int, int, error) {
	if value == "" || strings.TrimSpace(value) != value {
		return nil, 0, ErrInvalidDecimal
	}
	negative := strings.HasPrefix(value, "-")
	unsigned := value
	if negative {
		unsigned = strings.TrimPrefix(value, "-")
	}
	if unsigned == "" || strings.HasPrefix(unsigned, "+") {
		return nil, 0, ErrInvalidDecimal
	}
	parts := strings.Split(unsigned, ".")
	if len(parts) > 2 || parts[0] == "" || (len(parts) == 2 && parts[1] == "") {
		return nil, 0, ErrInvalidDecimal
	}
	for _, part := range parts {
		for _, r := range part {
			if r < '0' || r > '9' {
				return nil, 0, ErrInvalidDecimal
			}
		}
	}
	scale := 0
	digits := parts[0]
	if len(parts) == 2 {
		scale = len(parts[1])
		digits += parts[1]
	}
	coefficient, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return nil, 0, ErrInvalidDecimal
	}
	if negative {
		coefficient.Neg(coefficient)
	}
	return coefficient, scale, nil
}

func alignDecimals(left *big.Int, leftScale int, right *big.Int, rightScale int) (*big.Int, *big.Int, int) {
	scale := leftScale
	if rightScale > scale {
		scale = rightScale
	}
	left = new(big.Int).Mul(left, pow10(scale-leftScale))
	right = new(big.Int).Mul(right, pow10(scale-rightScale))
	return left, right, scale
}

func pow10(power int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(power)), nil)
}

func formatDecimal(coefficient *big.Int, scale int) string {
	negative := coefficient.Sign() < 0
	digits := new(big.Int).Abs(coefficient).String()
	if scale > 0 {
		for len(digits) <= scale {
			digits = "0" + digits
		}
		digits = digits[:len(digits)-scale] + "." + digits[len(digits)-scale:]
		digits = strings.TrimRight(strings.TrimRight(digits, "0"), ".")
	}
	digits = strings.TrimLeft(digits, "0")
	if digits == "" || strings.HasPrefix(digits, ".") {
		digits = "0" + digits
	}
	if digits == "0" {
		return digits
	}
	if negative {
		return "-" + digits
	}
	return digits
}
