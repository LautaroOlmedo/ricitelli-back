package salesadministrationpb

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"
)

func TestGeneratedMessagesRoundTripExactDecimalStrings(t *testing.T) {
	input := &SalesInvoice{Id: "invoice-1", TotalAmount: "0.3000"}
	data, err := proto.Marshal(protoadapt.MessageV2Of(input))
	if err != nil {
		t.Fatal(err)
	}
	var output SalesInvoice
	if err := proto.Unmarshal(data, protoadapt.MessageV2Of(&output)); err != nil {
		t.Fatal(err)
	}
	if output.TotalAmount != input.TotalAmount {
		t.Fatalf("total amount = %q, want %q", output.TotalAmount, input.TotalAmount)
	}
}
