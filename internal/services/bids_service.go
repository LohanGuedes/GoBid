package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lohanguedes/gobid/internal/store/pgstore"
)

type BidsStore interface {
	PlaceBid(ctx context.Context, params pgstore.CreateBidParams) error
	GetBid(ctx context.Context, id int64) (pgstore.Bid, error)
}

type BidsService struct {
	pool *pgxpool.Pool
	// TODO: Make this an interface for better idiomatic code:
	db *pgstore.Queries
}

func NewBidsService(pool *pgxpool.Pool) BidsService {
	return BidsService{
		pool: pool,
		db:   pgstore.New(pool),
	}
}

var ErrBidIsTooLow = errors.New("the bid value is too low or a higher bid was already placed")

func (s BidsService) PlaceBid(ctx context.Context, product_id, bidder_id uuid.UUID, amount float64) (pgstore.Bid, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return pgstore.Bid{}, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}

		err = tx.Commit(ctx)
	}()

	// Use qtx (queriesTx) instead
	qtx := s.db.WithTx(tx)
	product, err := qtx.GetProductById(ctx, product_id)
	if err != nil {
		return pgstore.Bid{}, err
	}

	highestBid, err := qtx.GetHighestBidByProductId(ctx, product_id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, err
		}
	}

	if product.BasePrice >= amount || highestBid.BidAmount >= amount {
		return pgstore.Bid{}, ErrBidIsTooLow
	}

	highestBid, err = qtx.CreateBid(ctx, pgstore.CreateBidParams{
		ProductID: product_id,
		BidderID:  bidder_id,
		BidAmount: amount,
	})
	if err != nil {
		return pgstore.Bid{}, err
	}

	return highestBid, err
}
