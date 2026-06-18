package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func ptr[T any](value T) *T { return &value }

func field(name string, number int32, kind descriptorpb.FieldDescriptorProto_Type, repeated bool, typeName string) *descriptorpb.FieldDescriptorProto {
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	if repeated {
		label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	}
	f := &descriptorpb.FieldDescriptorProto{Name: ptr(name), Number: ptr(number), Label: &label, Type: &kind}
	if typeName != "" {
		f.TypeName = ptr(typeName)
	}
	return f
}

func message(name string, fields ...*descriptorpb.FieldDescriptorProto) *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{Name: ptr(name), Field: fields}
}

func method(name, input, output string) *descriptorpb.MethodDescriptorProto {
	return &descriptorpb.MethodDescriptorProto{Name: ptr(name), InputType: ptr(input), OutputType: ptr(output)}
}

func main() {
	s, i32, i64, msg := descriptorpb.FieldDescriptorProto_TYPE_STRING, descriptorpb.FieldDescriptorProto_TYPE_INT32, descriptorpb.FieldDescriptorProto_TYPE_INT64, descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
	p := ".sales_administration."
	fd := &descriptorpb.FileDescriptorProto{
		Name: ptr("sales_administration.proto"), Package: ptr("sales_administration"), Syntax: ptr("proto3"),
		Options: &descriptorpb.FileOptions{GoPackage: ptr("ricitelli-back/cmd/http/gen/sales_administration;salesadministrationpb")},
		MessageType: []*descriptorpb.DescriptorProto{
			message("SalesInvoiceItem", field("line_number", 1, i32, false, ""), field("product_id", 2, s, false, ""), field("description", 3, s, false, ""), field("quantity", 4, s, false, ""), field("unit_price", 5, s, false, ""), field("tax_rate", 6, s, false, ""), field("net_amount", 7, s, false, ""), field("tax_amount", 8, s, false, ""), field("total_amount", 9, s, false, "")),
			message("SalesInvoice", field("id", 1, s, false, ""), field("customer_id", 2, s, false, ""), field("sale_order_id", 3, s, false, ""), field("document_type", 4, s, false, ""), field("point_of_sale", 5, i32, false, ""), field("document_number", 6, i64, false, ""), field("issue_date", 7, s, false, ""), field("due_date", 8, s, false, ""), field("currency", 9, s, false, ""), field("exchange_rate", 10, s, false, ""), field("subtotal", 11, s, false, ""), field("tax_total", 12, s, false, ""), field("total_amount", 13, s, false, ""), field("status", 14, s, false, ""), field("idempotency_key", 15, s, false, ""), field("items", 16, msg, true, p+"SalesInvoiceItem"), field("created_at", 17, s, false, ""), field("updated_at", 18, s, false, "")),
			message("CreateSalesInvoiceRequest", field("customer_id", 1, s, false, ""), field("sale_order_id", 2, s, false, ""), field("document_type", 3, s, false, ""), field("point_of_sale", 4, i32, false, ""), field("document_number", 5, i64, false, ""), field("issue_date", 6, s, false, ""), field("due_date", 7, s, false, ""), field("currency", 8, s, false, ""), field("exchange_rate", 9, s, false, ""), field("subtotal", 10, s, false, ""), field("tax_total", 11, s, false, ""), field("total_amount", 12, s, false, ""), field("idempotency_key", 13, s, false, ""), field("items", 14, msg, true, p+"SalesInvoiceItem")),
			message("RemittanceItem", field("line_number", 1, i32, false, ""), field("product_id", 2, s, false, ""), field("description", 3, s, false, ""), field("quantity", 4, s, false, ""), field("lot_number", 5, s, false, "")),
			message("Remittance", field("id", 1, s, false, ""), field("customer_id", 2, s, false, ""), field("sale_order_id", 3, s, false, ""), field("sales_invoice_id", 4, s, false, ""), field("point_of_sale", 5, i32, false, ""), field("document_number", 6, i64, false, ""), field("issue_date", 7, s, false, ""), field("delivery_date", 8, s, false, ""), field("status", 9, s, false, ""), field("idempotency_key", 10, s, false, ""), field("items", 11, msg, true, p+"RemittanceItem"), field("created_at", 12, s, false, ""), field("updated_at", 13, s, false, "")),
			message("CreateRemittanceRequest", field("customer_id", 1, s, false, ""), field("sale_order_id", 2, s, false, ""), field("sales_invoice_id", 3, s, false, ""), field("point_of_sale", 4, i32, false, ""), field("document_number", 5, i64, false, ""), field("issue_date", 6, s, false, ""), field("delivery_date", 7, s, false, ""), field("idempotency_key", 8, s, false, ""), field("items", 9, msg, true, p+"RemittanceItem")),
			message("CustomerReceiptAllocation", field("sales_invoice_id", 1, s, false, ""), field("allocated_amount", 2, s, false, "")),
			message("CustomerReceipt", field("id", 1, s, false, ""), field("customer_id", 2, s, false, ""), field("receipt_number", 3, i64, false, ""), field("receipt_date", 4, s, false, ""), field("currency", 5, s, false, ""), field("exchange_rate", 6, s, false, ""), field("amount", 7, s, false, ""), field("payment_method", 8, s, false, ""), field("payment_reference", 9, s, false, ""), field("status", 10, s, false, ""), field("idempotency_key", 11, s, false, ""), field("allocations", 12, msg, true, p+"CustomerReceiptAllocation"), field("created_at", 13, s, false, ""), field("updated_at", 14, s, false, "")),
			message("CreateCustomerReceiptRequest", field("customer_id", 1, s, false, ""), field("receipt_number", 2, i64, false, ""), field("receipt_date", 3, s, false, ""), field("currency", 4, s, false, ""), field("exchange_rate", 5, s, false, ""), field("amount", 6, s, false, ""), field("payment_method", 7, s, false, ""), field("payment_reference", 8, s, false, ""), field("idempotency_key", 9, s, false, ""), field("allocations", 10, msg, true, p+"CustomerReceiptAllocation")),
			message("IDRequest", field("id", 1, s, false, "")),
			message("InvoiceOutstandingBalance", field("sales_invoice_id", 1, s, false, ""), field("outstanding_balance", 2, s, false, "")),
		},
		Service: []*descriptorpb.ServiceDescriptorProto{{Name: ptr("SalesAdministrationService"), Method: []*descriptorpb.MethodDescriptorProto{
			method("CreateSalesInvoice", p+"CreateSalesInvoiceRequest", p+"SalesInvoice"), method("GetSalesInvoiceByID", p+"IDRequest", p+"SalesInvoice"), method("IssueSalesInvoice", p+"IDRequest", p+"SalesInvoice"), method("GetSalesInvoiceOutstandingBalance", p+"IDRequest", p+"InvoiceOutstandingBalance"),
			method("CreateRemittance", p+"CreateRemittanceRequest", p+"Remittance"), method("GetRemittanceByID", p+"IDRequest", p+"Remittance"), method("ConfirmRemittance", p+"IDRequest", p+"Remittance"),
			method("CreateCustomerReceipt", p+"CreateCustomerReceiptRequest", p+"CustomerReceipt"), method("GetCustomerReceiptByID", p+"IDRequest", p+"CustomerReceipt"), method("PostCustomerReceipt", p+"IDRequest", p+"CustomerReceipt"),
		}}},
	}
	request := &pluginpb.CodeGeneratorRequest{FileToGenerate: []string{fd.GetName()}, ProtoFile: []*descriptorpb.FileDescriptorProto{fd}, Parameter: ptr("module=ricitelli-back")}
	input, err := proto.Marshal(request)
	if err != nil {
		panic(err)
	}
	cmd := exec.Command(filepath.Join(".tools", "bin", "protoc-gen-go.exe"))
	cmd.Stdin = bytes.NewReader(input)
	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	response := new(pluginpb.CodeGeneratorResponse)
	if err := proto.Unmarshal(output, response); err != nil {
		panic(err)
	}
	if response.Error != nil {
		panic(*response.Error)
	}
	for _, file := range response.File {
		if err := os.MkdirAll(filepath.Dir(file.GetName()), 0755); err != nil {
			panic(err)
		}
		if err := os.WriteFile(file.GetName(), []byte(file.GetContent()), 0644); err != nil {
			panic(err)
		}
		fmt.Println(file.GetName())
	}
}
