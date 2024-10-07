package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rocketseat/go-first-auth/internal/store/pgstore"
)

var ErrProductNotFound = errors.New("product not found in database")

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

type ProductData struct {
	ID           uuid.UUID `json:"id"`
	SellerID     uuid.UUID `json:"seller_id"`
	ProductName  string    `json:"product_name"`
	Description  string    `json:"description"`
	BasePrice    float64   `json:"base_price"`
	AuctionStart time.Time `json:"auction_start"`
	AuctionEnd   time.Time `json:"auction_end"`
	IsSold       bool      `json:"is_sold"`
}

func (s *ProductService) GetProductById(ctx context.Context, id uuid.UUID) (ProductData, error) {
	product, err := s.db.GetProductById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ProductData{}, ErrProductNotFound
		}
		return ProductData{}, err
	}

	return ProductData{
		product.ID,
		product.SellerID,
		product.ProductName,
		product.Description,
		product.BasePrice,
		product.AuctionStart.Time,
		product.AuctionEnd.Time,
		product.IsSold,
	}, nil
}
