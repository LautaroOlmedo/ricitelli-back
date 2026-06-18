package value_object_test

import (
	"testing"

	valueObject "ricitelli-back/internal/value-object"
)

func TestDecimalArithmeticIsExact(t *testing.T) {
	sum, err := valueObject.AddDecimal("0.10", "0.20")
	if err != nil || sum != "0.3" {
		t.Fatalf("AddDecimal() = %q, %v; want 0.3, nil", sum, err)
	}
	difference, err := valueObject.SubtractDecimal("100.0000", "33.3333")
	if err != nil || difference != "66.6667" {
		t.Fatalf("SubtractDecimal() = %q, %v; want 66.6667, nil", difference, err)
	}
}

func TestDecimalToUint64RejectsFraction(t *testing.T) {
	if _, err := valueObject.DecimalToUint64("1.5"); err == nil {
		t.Fatal("DecimalToUint64() expected error")
	}
	got, err := valueObject.DecimalToUint64("12.0000")
	if err != nil || got != 12 {
		t.Fatalf("DecimalToUint64() = %d, %v; want 12, nil", got, err)
	}
}
