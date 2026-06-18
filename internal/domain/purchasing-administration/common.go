package purchasing_administration

import (
	"errors"
	"strings"
	"time"
)

type Currency string

const (
	CurrencyARS Currency = "ARS"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyCAD Currency = "CAD"
)

var validCurrencies = map[Currency]bool{
	CurrencyARS: true, CurrencyUSD: true, CurrencyEUR: true, CurrencyCAD: true,
}

func validateCurrency(currency Currency) error {
	if currency == "" {
		return nil
	}
	if !validCurrencies[currency] {
		return errors.New("invalid currency")
	}
	return nil
}

func validateDate(value, name string) error {
	if value == "" {
		return errors.New(name + " is required")
	}
	if _, err := time.Parse(time.DateOnly, value); err != nil {
		return errors.New(name + " must use YYYY-MM-DD")
	}
	return nil
}

func validateOptionalDate(value, name string) error {
	if value == "" {
		return nil
	}
	return validateDate(value, name)
}

func validateDateOrder(from, to, fromName, toName string) error {
	if to == "" {
		return nil
	}
	a, _ := time.Parse(time.DateOnly, from)
	b, _ := time.Parse(time.DateOnly, to)
	if b.Before(a) {
		return errors.New(toName + " must not be before " + fromName)
	}
	return nil
}

func normalizedMoney(value string, positive bool) (string, error) {
	return normalizedDecimal(value, positive, 14, 4)
}

func normalizedExchangeRate(value string) (string, error) {
	return normalizedDecimal(value, true, 12, 6)
}

func normalizedTaxRate(value string) (string, error) {
	return normalizedDecimal(value, false, 3, 4)
}

func normalizedSignedMoney(value string) (string, error) {
	if _, err := decimalRat(value); err != nil {
		return "", err
	}
	normalized, err := NormalizeDecimal(value)
	if err != nil {
		return "", err
	}
	return validateDecimalShape(normalized, 14, 4)
}

func normalizedDecimal(value string, positive bool, maxIntegerDigits, maxScale int) (string, error) {
	var err error
	if positive {
		err = ValidatePositiveDecimal(value)
	} else {
		err = ValidateNonNegativeDecimal(value)
	}
	if err != nil {
		return "", err
	}
	normalized, err := NormalizeDecimal(value)
	if err != nil {
		return "", err
	}
	return validateDecimalShape(normalized, maxIntegerDigits, maxScale)
}

func validateDecimalShape(normalized string, maxIntegerDigits, maxScale int) (string, error) {
	unsigned := strings.TrimPrefix(normalized, "-")
	parts := strings.SplitN(unsigned, ".", 2)
	if len(parts[0]) > maxIntegerDigits {
		return "", ErrInvalidDecimal
	}
	if len(parts) == 2 && len(parts[1]) > maxScale {
		return "", ErrInvalidDecimal
	}
	return normalized, nil
}
