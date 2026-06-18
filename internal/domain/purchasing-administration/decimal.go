package purchasing_administration

import (
	"errors"
	"math/big"
	"regexp"
	"strings"
)

var decimalPattern = regexp.MustCompile(`^[+-]?(0|[1-9][0-9]*)(\.[0-9]+)?$`)

var (
	ErrInvalidDecimal     = errors.New("invalid decimal string")
	ErrNegativeDecimal    = errors.New("decimal must not be negative")
	ErrNonPositiveDecimal = errors.New("decimal must be greater than zero")
)

// NormalizeDecimal returns a canonical, non-exponent decimal string.
func NormalizeDecimal(value string) (string, error) {
	if !decimalPattern.MatchString(value) {
		return "", ErrInvalidDecimal
	}
	negative := strings.HasPrefix(value, "-")
	value = strings.TrimPrefix(strings.TrimPrefix(value, "+"), "-")
	parts := strings.SplitN(value, ".", 2)
	integer := strings.TrimLeft(parts[0], "0")
	if integer == "" {
		integer = "0"
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = strings.TrimRight(parts[1], "0")
	}
	if integer == "0" && fraction == "" {
		return "0", nil
	}
	result := integer
	if fraction != "" {
		result += "." + fraction
	}
	if negative {
		result = "-" + result
	}
	return result, nil
}

func ValidateNonNegativeDecimal(value string) error {
	n, err := NormalizeDecimal(value)
	if err != nil {
		return err
	}
	if strings.HasPrefix(n, "-") {
		return ErrNegativeDecimal
	}
	return nil
}

func ValidatePositiveDecimal(value string) error {
	n, err := NormalizeDecimal(value)
	if err != nil {
		return err
	}
	if n == "0" || strings.HasPrefix(n, "-") {
		return ErrNonPositiveDecimal
	}
	return nil
}

func AddDecimals(values ...string) (string, error) {
	total := new(big.Rat)
	for _, value := range values {
		r, err := decimalRat(value)
		if err != nil {
			return "", err
		}
		total.Add(total, r)
	}
	return ratToDecimal(total), nil
}

func SubtractDecimals(left, right string) (string, error) {
	a, err := decimalRat(left)
	if err != nil {
		return "", err
	}
	b, err := decimalRat(right)
	if err != nil {
		return "", err
	}
	return ratToDecimal(new(big.Rat).Sub(a, b)), nil
}

func CompareDecimals(left, right string) (int, error) {
	a, err := decimalRat(left)
	if err != nil {
		return 0, err
	}
	b, err := decimalRat(right)
	if err != nil {
		return 0, err
	}
	return a.Cmp(b), nil
}

func decimalRat(value string) (*big.Rat, error) {
	n, err := NormalizeDecimal(value)
	if err != nil {
		return nil, err
	}
	r, ok := new(big.Rat).SetString(n)
	if !ok {
		return nil, ErrInvalidDecimal
	}
	return r, nil
}

func ratToDecimal(r *big.Rat) string {
	if r.Sign() == 0 {
		return "0"
	}
	den := new(big.Int).Set(r.Denom())
	two, five := 0, 0
	for new(big.Int).Mod(den, big.NewInt(2)).Sign() == 0 {
		den.Div(den, big.NewInt(2))
		two++
	}
	for new(big.Int).Mod(den, big.NewInt(5)).Sign() == 0 {
		den.Div(den, big.NewInt(5))
		five++
	}
	if den.Cmp(big.NewInt(1)) != 0 {
		return r.FloatString(18)
	}
	scale := two
	if five > scale {
		scale = five
	}
	n, _ := NormalizeDecimal(r.FloatString(scale))
	return n
}
