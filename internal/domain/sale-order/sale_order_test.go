package sale_order_test

import (
	sale_order "ricitelli-back/internal/domain/sale-order"
	valueObject "ricitelli-back/internal/value-object"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewSale(t *testing.T) {
	customerID := uuid.New().String()
	expectedItems := []valueObject.SaleOrderItem{
		{
			ProductID: "product UUID",
			Quantity:  1,
			UnitPrice: 1.0,
		},
		{},
	}

	type testCase struct {
		testName           string
		expectedCustomerID string
		expectedItems      []valueObject.SaleOrderItem
		expectedError      error
	}

	testCases := []testCase{
		{
			testName:           "Succes - SaleOrder created successfully",
			expectedCustomerID: customerID,
			expectedItems:      expectedItems,
			expectedError:      nil,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			newOrderSale, err := sale_order.NewSaleOrder(tc.expectedCustomerID, tc.expectedItems)

			if err != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Empty(t, newOrderSale)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}
