package product_inventory

import (
	"fmt"
	valueObject "ricitelli-back/internal/value-object"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAvailableDressed(t *testing.T) {

	productID := uuid.New().String()
	productionOrderID := uuid.New().String()

	movements := []valueObject.ProductMovement{
		{
			MovementType: valueObject.ProductProduced,
			Quantity:     100,
			Reference:    productionOrderID,
			Stage:        valueObject.Undressed,
			CreatedAt:    strconv.Itoa(time.Now().Minute() - 10),
		},
		{
			MovementType: valueObject.ProductStageOut,
			Quantity:     10,
			Reference:    productionOrderID,
			Stage:        valueObject.Undressed,
			CreatedAt:    strconv.Itoa(time.Now().Minute() - 9),
		},
		{
			MovementType: valueObject.ProductStageIn,
			Quantity:     10,
			Reference:    productionOrderID,
			Stage:        valueObject.Dressed,
			CreatedAt:    strconv.Itoa(time.Now().Minute() - 8),
		},

		/*{
			MovementType: valueObject.ProductReservedForSale, RESTA A VESTIDOS
			Quantity:     10,
			Reference:    productionOrderID,
			Stage:        valueObject.Dressed,
			CreatedAt:    strconv.Itoa(time.Now().Minute() - 7),
		},*/
	}

	prodInventory, err := NewProductInventory(productID, "PROD-SKU")
	if err != nil {
		t.Error(err)
	}

	prodInventory.movements = append(prodInventory.GetMovements(), movements...)
	fmt.Println("movements:", prodInventory.GetMovements())
	type testCase struct {
		testName         string
		expectedQuantity uint64
		expectedError    error
	}

	testCases := []testCase{
		{
			testName:         "Succes - Product Inventory Available",
			expectedQuantity: 10,
			expectedError:    nil,
		},
		/*{
			testName:         "Succes - Product Inventory Available",
			expectedQuantity: 20,
			expectedError:    nil,
		},*/
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			validateErr := prodInventory.ValidateMovements(prodInventory.GetProductID())

			if validateErr != nil {
				t.Error(validateErr)
			}

			stock, _ := prodInventory.AvailableDressed()

			fmt.Println(tc.expectedQuantity, stock)

			if stock != int64(tc.expectedQuantity) {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

/*func TestReserve(t *testing.T) {
productID := uuid.New().String()
productionOrderID := uuid.New().String()

movements := []valueObject.ProductMovement{
	{
		MovementType: valueObject.ProductProduced,
		Quantity:     100,
		Reference:    productionOrderID,
		Stage:        valueObject.Undressed,
		CreatedAt:    strconv.Itoa(time.Now().Minute() - 10),
	},
	{
		MovementType: valueObject.ProductStageOut,
		Quantity:     10,
		Reference:    productionOrderID,
		Stage:        valueObject.Undressed,
		CreatedAt:    strconv.Itoa(time.Now().Minute() - 9),
	},
	{
		MovementType: valueObject.ProductStageIn,
		Quantity:     10,
		Reference:    productionOrderID,
		Stage:        valueObject.Dressed,
		CreatedAt:    strconv.Itoa(time.Now().Minute() - 8),
	},

	/*{
		MovementType: valueObject.ProductReservedForSale, RESTA A VESTIDOS
		Quantity:     10,
		Reference:    productionOrderID,
		Stage:        valueObject.Dressed,
		CreatedAt:    strconv.Itoa(time.Now().Minute() - 7),
	},*/
