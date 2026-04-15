package reporting

import (
	"errors"
	"time"
)

// ReportType identifies the kind of report.
type ReportType string

const (
	TypeSales      ReportType = "SALES"
	TypeProduction ReportType = "PRODUCTION"
	TypeGeneral    ReportType = "GENERAL"
	TypeLowStock   ReportType = "LOW_STOCK"
	TypeLot        ReportType = "LOT"
	TypeCustomer   ReportType = "CUSTOMER"
)

// ReportFilter holds parameters for generating a report.
type ReportFilter struct {
	FromDate   string // RFC3339
	ToDate     string // RFC3339
	CustomerID string
	Market     string
	Currency   string
	ProductID  string
	LotNumber  string
}

// Validate checks date range validity when applicable.
func (f ReportFilter) Validate(requireRange bool) error {
	if !requireRange && f.FromDate == "" && f.ToDate == "" {
		return nil
	}
	if f.FromDate == "" || f.ToDate == "" {
		return errors.New("from_date and to_date are required")
	}
	from, err := time.Parse(time.RFC3339, f.FromDate)
	if err != nil {
		return errors.New("from_date must be RFC3339")
	}
	to, err := time.Parse(time.RFC3339, f.ToDate)
	if err != nil {
		return errors.New("to_date must be RFC3339")
	}
	if from.After(to) {
		return errors.New("from_date must be <= to_date")
	}
	return nil
}

// ParseFrom returns the parsed from_date or zero time when empty.
func (f ReportFilter) ParseFrom() time.Time {
	t, _ := time.Parse(time.RFC3339, f.FromDate)
	return t
}

// ParseTo returns the parsed to_date or zero time when empty.
func (f ReportFilter) ParseTo() time.Time {
	t, _ := time.Parse(time.RFC3339, f.ToDate)
	return t
}
