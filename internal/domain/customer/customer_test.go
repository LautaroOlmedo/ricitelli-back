package customer_test

import (
	"testing"

	"ricitelli-back/internal/domain/customer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCustomer_Success(t *testing.T) {
	c, err := customer.NewCustomer(customer.NewCustomerParams{
		SocialReason: "Vinos del Sur S.A.",
		MarketType:   customer.MarketTypeInternal,
		Group:        customer.GroupDistributor,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, c.GetID())
	assert.Equal(t, "Vinos del Sur S.A.", c.GetSocialReason())
	assert.Equal(t, customer.MarketTypeInternal, c.GetMarketType())
	assert.Equal(t, customer.GroupDistributor, c.GetGroup())
	assert.True(t, c.IsActive())
	assert.NotEmpty(t, c.GetCreatedAt())
}

func TestNewCustomer_ExternalMarket(t *testing.T) {
	c, err := customer.NewCustomer(customer.NewCustomerParams{
		SocialReason: "Japan Imports Ltd.",
		MarketType:   customer.MarketTypeExternal,
		Group:        customer.GroupExportAgent,
	})
	require.NoError(t, err)
	assert.Equal(t, customer.MarketTypeExternal, c.GetMarketType())
	assert.Equal(t, customer.GroupExportAgent, c.GetGroup())
}

func TestNewCustomer_EmptySocialReason(t *testing.T) {
	_, err := customer.NewCustomer(customer.NewCustomerParams{
		SocialReason: "",
		MarketType:   customer.MarketTypeInternal,
		Group:        customer.GroupRetail,
	})
	assert.ErrorIs(t, err, customer.ErrEmptySocialReason)
}

func TestNewCustomer_InvalidMarketType(t *testing.T) {
	_, err := customer.NewCustomer(customer.NewCustomerParams{
		SocialReason: "Bodega Test",
		MarketType:   customer.MarketType("INVALID"),
		Group:        customer.GroupRetail,
	})
	assert.ErrorIs(t, err, customer.ErrInvalidMarketType)
}

func TestNewCustomer_InvalidGroup(t *testing.T) {
	_, err := customer.NewCustomer(customer.NewCustomerParams{
		SocialReason: "Bodega Test",
		MarketType:   customer.MarketTypeInternal,
		Group:        customer.Group("UNKNOWN"),
	})
	assert.ErrorIs(t, err, customer.ErrInvalidGroup)
}

func TestNewCustomerWithID(t *testing.T) {
	c, err := customer.NewCustomerWithID("cust-001", customer.NewCustomerParams{
		SocialReason: "Restaurante La Vid",
		MarketType:   customer.MarketTypeInternal,
		Group:        customer.GroupRestaurant,
	})
	require.NoError(t, err)
	assert.Equal(t, "cust-001", c.GetID())
}

func TestNewCustomerWithID_EmptyID(t *testing.T) {
	_, err := customer.NewCustomerWithID("", customer.NewCustomerParams{
		SocialReason: "Bodega Test",
		MarketType:   customer.MarketTypeInternal,
		Group:        customer.GroupRetail,
	})
	assert.Error(t, err)
}

func TestCustomer_Deactivate(t *testing.T) {
	c, err := customer.NewCustomer(customer.NewCustomerParams{
		SocialReason: "Hotel Patagonia",
		MarketType:   customer.MarketTypeInternal,
		Group:        customer.GroupHotel,
	})
	require.NoError(t, err)
	assert.True(t, c.IsActive())

	err = c.Deactivate()
	require.NoError(t, err)
	assert.False(t, c.IsActive())
}

func TestCustomer_Deactivate_AlreadyInactive(t *testing.T) {
	c, _ := customer.NewCustomer(customer.NewCustomerParams{
		SocialReason: "Hotel Patagonia",
		MarketType:   customer.MarketTypeInternal,
		Group:        customer.GroupHotel,
	})
	_ = c.Deactivate()

	err := c.Deactivate()
	assert.Error(t, err)
}
