package application_service_test

import (
	sale_order "ricitelli-back/internal/domain/sale-order"
	"ricitelli-back/internal/service/application-service/mocks"
	valueObject "ricitelli-back/internal/value-object"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

func TestCreateOrder(t *testing.T) {

	customerID := uuid.New().String()
	items := []valueObject.SaleOrderItem{
		{
			ProductID: uuid.New().String(),
			Quantity:  10,
			UnitPrice: 100.69,
		},
		{
			ProductID: uuid.New().String(),
			Quantity:  10,
			UnitPrice: 217.98,
		},
	}

	expectedSaleOrder, err := sale_order.NewSaleOrder(customerID, items)
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		testName           string
		expectedCustomerID string
		setupMock          func(mockSaleOrderService *mocks.MockSaleOrderService, mockOrderProductionService *mocks.MockProductionOrderService, mockProductService *mocks.MockProductService)
		expectedItems      []valueObject.SaleOrderItem
		expectedError      error
	}

	testCases := []testCase{
		{
			testName:           "Succes - Order process correctly",
			expectedCustomerID: customerID,
			setupMock: func(mockSaleOrderService *mocks.MockSaleOrderService, mockOrderProductionService *mocks.MockProductionOrderService, mockProductService *mocks.MockProductService) {
				mockSaleOrderService.
					EXPECT().
					CreateSaleOrder(gomock.Any(), customerID, items).
					Return(expectedSaleOrder, nil)

			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()
		})
	}

}
