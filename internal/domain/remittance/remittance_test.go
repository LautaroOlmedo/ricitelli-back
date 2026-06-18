package remittance_test

import (
	"testing"

	"ricitelli-back/internal/domain/remittance"
)

func validParams() remittance.NewRemittanceParams {
	return remittance.NewRemittanceParams{
		CustomerID: "customer-1", SaleOrderID: "order-1", SalesInvoiceID: "invoice-1",
		PointOfSale: 1, DocumentNumber: 22, IssueDate: "2026-06-09", IdempotencyKey: "remittance-key",
		Items: []remittance.Item{{LineNumber: 1, ProductID: "product-1", Description: "Partial shipment", Quantity: "2.0000"}},
	}
}

func TestRemittanceRequiresSaleOrderButInvoiceIsOptional(t *testing.T) {
	params := validParams()
	params.SalesInvoiceID = ""
	if _, err := remittance.NewRemittance(params); err != nil {
		t.Fatalf("NewRemittance() optional invoice error = %v", err)
	}
	params.SaleOrderID = ""
	if _, err := remittance.NewRemittance(params); err == nil {
		t.Fatal("NewRemittance() expected mandatory sale order error")
	}
}

func TestRemittanceConfirmationIsIdempotent(t *testing.T) {
	r, err := remittance.NewRemittance(validParams())
	if err != nil {
		t.Fatal(err)
	}
	dispatch, err := r.Confirm()
	if err != nil || !dispatch {
		t.Fatalf("first Confirm() = %v, %v; want true, nil", dispatch, err)
	}
	dispatch, err = r.Confirm()
	if err != nil || dispatch {
		t.Fatalf("second Confirm() = %v, %v; want false, nil", dispatch, err)
	}
	if r.IsEditable() {
		t.Fatal("confirmed remittance must not be editable")
	}
}
