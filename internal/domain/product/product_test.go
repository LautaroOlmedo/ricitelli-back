package product_test

import (
	"ricitelli-back/internal/domain/product"
	valueObject "ricitelli-back/internal/value-object"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProduct(t *testing.T) {

	billOfDrySupplyItem := []valueObject.BillOfDrySupply{
		{
			DrySupplyID:     " UUID Etiqueta Riccitelli and Father, Malbec Cab Franc",
			QuantityPerUnit: 1,
		},
		{
			DrySupplyID:     "UUID Contraetiqueta Riccitelli and Father, Malbec Cab Franc",
			QuantityPerUnit: 1,
		},
		{
			DrySupplyID:     "UUID Caja x 12 750 cc Riccitelli and Father, Malbec Cab Franc",
			QuantityPerUnit: 1,
		},
	}

	type testCase struct {
		testName                string
		expectedName            string
		expectedBillOfDrySupply []valueObject.BillOfDrySupply
		expectedError           error
	}

	testCases := []testCase{
		{
			testName:                "Success - Product created correctly",
			expectedName:            "Riccitelli and Father, Malbec Cab Franc",
			expectedBillOfDrySupply: billOfDrySupplyItem,
			expectedError:           nil,
		},
		{
			testName:                "Error - invalid product name",
			expectedName:            "",
			expectedBillOfDrySupply: billOfDrySupplyItem,
			expectedError:           product.ErrInvalidName,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			newProd, err := product.NewProduct(tc.expectedName, tc.expectedBillOfDrySupply)
			if err != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Empty(t, newProd)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}
