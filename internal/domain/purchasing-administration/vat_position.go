package purchasing_administration

import (
	"errors"
	"regexp"
)

var monthPattern = regexp.MustCompile(`^[0-9]{4}-(0[1-9]|1[0-2])$`)

type VATPositionLine struct {
	TaxRate, SalesDebit, PurchaseCredit, NetVAT string
}

type MonthlyVATPosition struct {
	Month                                        string
	Lines                                        []VATPositionLine
	SalesDebitTotal, PurchaseCreditTotal, NetVAT string
}

func NewMonthlyVATPosition(month string, lines []VATPositionLine) (MonthlyVATPosition, error) {
	if !monthPattern.MatchString(month) {
		return MonthlyVATPosition{}, errors.New("month must use YYYY-MM")
	}
	result := MonthlyVATPosition{Month: month, Lines: append([]VATPositionLine(nil), lines...),
		SalesDebitTotal: "0", PurchaseCreditTotal: "0", NetVAT: "0"}
	for i := range result.Lines {
		line := &result.Lines[i]
		var err error
		line.TaxRate, err = normalizedTaxRate(line.TaxRate)
		if err != nil {
			return MonthlyVATPosition{}, errors.New("VAT line contains an invalid tax_rate decimal string")
		}
		line.SalesDebit, err = normalizedSignedMoney(line.SalesDebit)
		if err != nil {
			return MonthlyVATPosition{}, errors.New("VAT line contains an invalid sales_debit decimal string")
		}
		line.PurchaseCredit, err = normalizedSignedMoney(line.PurchaseCredit)
		if err != nil {
			return MonthlyVATPosition{}, errors.New("VAT line contains an invalid purchase_credit decimal string")
		}
		line.NetVAT, _ = SubtractDecimals(line.SalesDebit, line.PurchaseCredit)
		result.SalesDebitTotal, _ = AddDecimals(result.SalesDebitTotal, line.SalesDebit)
		result.PurchaseCreditTotal, _ = AddDecimals(result.PurchaseCreditTotal, line.PurchaseCredit)
	}
	result.NetVAT, _ = SubtractDecimals(result.SalesDebitTotal, result.PurchaseCreditTotal)
	return result, nil
}
