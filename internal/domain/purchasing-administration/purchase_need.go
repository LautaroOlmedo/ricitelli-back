package purchasing_administration

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PurchaseNeedStatus string

const (
	PurchaseNeedDraft     PurchaseNeedStatus = "DRAFT"
	PurchaseNeedOpen      PurchaseNeedStatus = "OPEN"
	PurchaseNeedQuoted    PurchaseNeedStatus = "QUOTED"
	PurchaseNeedOrdered   PurchaseNeedStatus = "ORDERED"
	PurchaseNeedClosed    PurchaseNeedStatus = "CLOSED"
	PurchaseNeedCancelled PurchaseNeedStatus = "CANCELLED"
)

type PurchaseNeedItem struct {
	ID, PurchaseNeedID, DrySupplyID, Description, Quantity, Unit, CreatedAt, UpdatedAt string
	LineNumber                                                                         int32
}

type PurchaseNeed struct {
	ID, RequestedDate, RequiredByDate, Notes, IdempotencyKey, CreatedAt, UpdatedAt string
	NeedNumber                                                                     int64
	Status                                                                         PurchaseNeedStatus
	Items                                                                          []PurchaseNeedItem
}

type NewPurchaseNeedParams struct {
	NeedNumber                                           int64
	RequestedDate, RequiredByDate, Notes, IdempotencyKey string
	Items                                                []PurchaseNeedItem
}

func NewPurchaseNeed(params NewPurchaseNeedParams) (PurchaseNeed, error) {
	if params.NeedNumber <= 0 {
		return PurchaseNeed{}, errors.New("need_number must be greater than zero")
	}
	if err := validateDate(params.RequestedDate, "requested_date"); err != nil {
		return PurchaseNeed{}, err
	}
	if err := validateOptionalDate(params.RequiredByDate, "required_by_date"); err != nil {
		return PurchaseNeed{}, err
	}
	if err := validateDateOrder(params.RequestedDate, params.RequiredByDate, "requested_date", "required_by_date"); err != nil {
		return PurchaseNeed{}, err
	}
	if strings.TrimSpace(params.IdempotencyKey) == "" {
		return PurchaseNeed{}, errors.New("idempotency_key is required")
	}
	if len(params.Items) == 0 {
		return PurchaseNeed{}, errors.New("purchase need must have at least one item")
	}
	id, now := uuid.NewString(), time.Now().UTC().Format(time.RFC3339Nano)
	items := make([]PurchaseNeedItem, len(params.Items))
	seen := map[int32]bool{}
	for i, item := range params.Items {
		if item.LineNumber <= 0 || seen[item.LineNumber] {
			return PurchaseNeed{}, errors.New("item line_number must be positive and unique")
		}
		if strings.TrimSpace(item.Description) == "" || strings.TrimSpace(item.Unit) == "" {
			return PurchaseNeed{}, errors.New("item description and unit are required")
		}
		qty, err := normalizedMoney(item.Quantity, true)
		if err != nil {
			return PurchaseNeed{}, errors.New("item quantity must be a positive decimal string")
		}
		seen[item.LineNumber] = true
		item.ID, item.PurchaseNeedID, item.Quantity = uuid.NewString(), id, qty
		item.CreatedAt, item.UpdatedAt = now, now
		items[i] = item
	}
	return PurchaseNeed{ID: id, NeedNumber: params.NeedNumber, RequestedDate: params.RequestedDate,
		RequiredByDate: params.RequiredByDate, Status: PurchaseNeedDraft, Notes: params.Notes,
		IdempotencyKey: params.IdempotencyKey, CreatedAt: now, UpdatedAt: now, Items: items}, nil
}

var purchaseNeedTransitions = map[PurchaseNeedStatus]map[PurchaseNeedStatus]bool{
	PurchaseNeedDraft:   {PurchaseNeedOpen: true, PurchaseNeedCancelled: true},
	PurchaseNeedOpen:    {PurchaseNeedQuoted: true, PurchaseNeedCancelled: true},
	PurchaseNeedQuoted:  {PurchaseNeedOrdered: true, PurchaseNeedOpen: true, PurchaseNeedCancelled: true},
	PurchaseNeedOrdered: {PurchaseNeedClosed: true, PurchaseNeedCancelled: true},
}

func (n *PurchaseNeed) UpdateStatus(status PurchaseNeedStatus) error {
	if !purchaseNeedTransitions[n.Status][status] {
		return errors.New("invalid purchase need status transition")
	}
	n.Status, n.UpdatedAt = status, time.Now().UTC().Format(time.RFC3339Nano)
	return nil
}
