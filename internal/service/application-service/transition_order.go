package application_service

import (
	"context"
	"fmt"

	sale_order "ricitelli-back/internal/domain/sale-order"
)

// TransitionSaleOrderStatus advances the sale order status and applies
// inventory side-effects for each transition.
//
//   - → INVOICED : dispatches (consumes) the reserved dressed stock for each item.
//   - → CANCELLED: releases any reserved dressed stock back to available.
//   - All others : pure status update, no inventory changes.
func (s *Service) TransitionSaleOrderStatus(ctx context.Context, orderID string, newStatus sale_order.Status) (*sale_order.SaleOrder, error) {
	order, err := s.SaleOrderService.GetSaleOrderByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("transition: get sale order: %w", err)
	}

	switch newStatus {
	case sale_order.StatusInvoiced:
		if err := s.dispatchStock(ctx, order); err != nil {
			return nil, err
		}
	case sale_order.StatusCancelled:
		if err := s.releaseStock(ctx, order); err != nil {
			return nil, err
		}
	}

	updated, err := s.SaleOrderService.UpdateSaleOrderStatus(ctx, orderID, newStatus)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// dispatchStock reduces physical and committed dressed stock for each order item.
func (s *Service) dispatchStock(ctx context.Context, order *sale_order.SaleOrder) error {
	userID := userIDFromCtx(ctx)
	for _, item := range order.GetItems() {
		inv, err := s.ProductInventoryService.GetProductInventory(item.ProductID)
		if err != nil {
			return fmt.Errorf("dispatch: get inventory %s: %w", item.ProductID, err)
		}
		if err := inv.Dispatch(order.GetID(), item.Quantity, userID); err != nil {
			return fmt.Errorf("dispatch: dispatch stock for %s: %w", item.ProductID, err)
		}
		if err := s.ProductInventoryService.SaveProductInventory(*inv); err != nil {
			return fmt.Errorf("dispatch: save inventory %s: %w", item.ProductID, err)
		}
	}
	return nil
}

// releaseStock cancels reserved dressed stock when an order is cancelled.
func (s *Service) releaseStock(ctx context.Context, order *sale_order.SaleOrder) error {
	userID := userIDFromCtx(ctx)
	for _, item := range order.GetItems() {
		inv, err := s.ProductInventoryService.GetProductInventory(item.ProductID)
		if err != nil {
			return fmt.Errorf("release: get inventory %s: %w", item.ProductID, err)
		}
		if err := inv.ReleaseReservation(order.GetID(), item.Quantity, userID); err != nil {
			return fmt.Errorf("release: release stock for %s: %w", item.ProductID, err)
		}
		if err := s.ProductInventoryService.SaveProductInventory(*inv); err != nil {
			return fmt.Errorf("release: save inventory %s: %w", item.ProductID, err)
		}
	}
	return nil
}
