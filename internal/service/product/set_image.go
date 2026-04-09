package product

import "context"

func (s *ProductService) SetProductImage(ctx context.Context, id, imageURL string) error {
	return s.Storage.SetProductImage(ctx, id, imageURL)
}

func (s *ProductService) GetProductImage(ctx context.Context, id string) (string, error) {
	return s.Storage.GetProductImage(ctx, id)
}
