package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rocketseat/go-first-auth/internal/store/pgstore"
)

type ProductService struct {
	// TODO: make this a interface in order to be more idiomatic
	db *pgstore.Queries
}

// This should recieve a Interface that satisfies the types
func NewProductService(dbtx pgstore.DBTX) ProductService {
	return ProductService{
		db: pgstore.New(dbtx),
	}
}

func (s *ProductService) CreateProduct(
	ctx context.Context,
	sellerID uuid.UUID,
	productName string,
	basePrice float64,
	auctionStart, auctionEnd pgtype.Timestamptz,
) (uuid.UUID, error) {
	id, err := s.db.CreateProduct(ctx, pgstore.CreateProductParams{
		SellerID:     sellerID,
		ProductName:  productName,
		BasePrice:    basePrice,
		AuctionStart: auctionStart,
		AuctionEnd:   auctionEnd,
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}
