package dry_supply

import (
	"context"
	"errors"
)

func (s *Service) UpdateDrySupply(ctx context.Context, id, name string, reorderPoint int) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}
	return s.storage.UpdateDrySupply(ctx, id, name, reorderPoint)
}
