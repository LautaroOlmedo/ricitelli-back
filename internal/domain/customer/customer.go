package customer

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// MarketType classifies the customer's primary market segment.
type MarketType string

const (
	MarketTypeInternal MarketType = "INTERNAL"
	MarketTypeExternal MarketType = "EXTERNAL"
)

// Group classifies the customer's commercial channel.
type Group string

const (
	GroupDistributor Group = "DISTRIBUTOR"  // distribuidor mayorista
	GroupWineShop    Group = "WINE_SHOP"    // vinoteca
	GroupRestaurant  Group = "RESTAURANT"   // restaurante / bar
	GroupHotel       Group = "HOTEL"        // hotel / cadena hotelera
	GroupRetail      Group = "RETAIL"       // supermercado / retail
	GroupPrivate     Group = "PRIVATE"      // consumidor privado
	GroupExportAgent Group = "EXPORT_AGENT" // agente exportador / importador
)

var (
	ErrEmptySocialReason = errors.New("social reason cannot be empty")
	ErrInvalidMarketType = errors.New("invalid market type")
	ErrInvalidGroup      = errors.New("invalid group")
)

var validMarketTypes = map[MarketType]bool{
	MarketTypeInternal: true,
	MarketTypeExternal: true,
}

var validGroups = map[Group]bool{
	GroupDistributor: true,
	GroupWineShop:    true,
	GroupRestaurant:  true,
	GroupHotel:       true,
	GroupRetail:      true,
	GroupPrivate:     true,
	GroupExportAgent: true,
}

// Customer is the aggregate root representing a buyer who places sale orders.
type Customer struct {
	id           string
	socialReason string // razón social
	marketType   MarketType
	group        Group
	active       bool
	createdAt    string
}

type NewCustomerParams struct {
	SocialReason string
	MarketType   MarketType
	Group        Group
}

func NewCustomer(params NewCustomerParams) (Customer, error) {
	if params.SocialReason == "" {
		return Customer{}, ErrEmptySocialReason
	}
	if !validMarketTypes[params.MarketType] {
		return Customer{}, ErrInvalidMarketType
	}
	if !validGroups[params.Group] {
		return Customer{}, ErrInvalidGroup
	}
	return Customer{
		id:           uuid.New().String(),
		socialReason: params.SocialReason,
		marketType:   params.MarketType,
		group:        params.Group,
		active:       true,
		createdAt:    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func NewCustomerWithID(id string, params NewCustomerParams) (Customer, error) {
	if id == "" {
		return Customer{}, errors.New("id cannot be empty")
	}
	c, err := NewCustomer(params)
	if err != nil {
		return Customer{}, err
	}
	c.id = id
	return c, nil
}

// ReconstitueCustomer reconstitutes a Customer from stored data (bypasses validation).
func ReconstitueCustomer(id, socialReason string, marketType MarketType, group Group, active bool, createdAt string) Customer {
	return Customer{
		id:           id,
		socialReason: socialReason,
		marketType:   marketType,
		group:        group,
		active:       active,
		createdAt:    createdAt,
	}
}

func (c *Customer) SetSocialReason(v string) error {
	if v == "" {
		return ErrEmptySocialReason
	}
	c.socialReason = v
	return nil
}

func (c *Customer) SetMarketType(v MarketType) error {
	if !validMarketTypes[v] {
		return ErrInvalidMarketType
	}
	c.marketType = v
	return nil
}

func (c *Customer) SetGroup(v Group) error {
	if !validGroups[v] {
		return ErrInvalidGroup
	}
	c.group = v
	return nil
}

// Deactivate soft-deletes the customer.
func (c *Customer) Deactivate() error {
	if !c.active {
		return errors.New("customer is already inactive")
	}
	c.active = false
	return nil
}

func (c *Customer) GetID() string             { return c.id }
func (c *Customer) GetSocialReason() string   { return c.socialReason }
func (c *Customer) GetMarketType() MarketType { return c.marketType }
func (c *Customer) GetGroup() Group           { return c.group }
func (c *Customer) IsActive() bool            { return c.active }
func (c *Customer) GetCreatedAt() string      { return c.createdAt }

// UpdateCustomerParams holds the fields that can be updated on an existing customer.
type UpdateCustomerParams struct {
	SocialReason string
	MarketType   MarketType
	Group        Group
}
