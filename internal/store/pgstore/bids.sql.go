// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: bids.sql

package pgstore

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createBid = `-- name: CreateBid :one
INSERT INTO bids (
    product_id, bidder_id, bid_amount
) VALUES ($1, $2, $3)
RETURNING id, created_at
`

type CreateBidParams struct {
	ProductID uuid.UUID `json:"product_id"`
	BidderID  uuid.UUID `json:"bidder_id"`
	BidAmount float64   `json:"bid_amount"`
}

type CreateBidRow struct {
	ID        uuid.UUID          `json:"id"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}

func (q *Queries) CreateBid(ctx context.Context, arg CreateBidParams) (CreateBidRow, error) {
	row := q.db.QueryRow(ctx, createBid, arg.ProductID, arg.BidderID, arg.BidAmount)
	var i CreateBidRow
	err := row.Scan(&i.ID, &i.CreatedAt)
	return i, err
}

const getBidsByProductId = `-- name: GetBidsByProductId :many
SELECT id, product_id, bidder_id, bid_amount, created_at FROM bids
WHERE product_id = $1
ORDER BY bid_amount DESC
`

func (q *Queries) GetBidsByProductId(ctx context.Context, productID uuid.UUID) ([]Bid, error) {
	rows, err := q.db.Query(ctx, getBidsByProductId, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Bid
	for rows.Next() {
		var i Bid
		if err := rows.Scan(
			&i.ID,
			&i.ProductID,
			&i.BidderID,
			&i.BidAmount,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
