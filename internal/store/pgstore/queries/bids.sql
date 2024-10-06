-- name: CreateBid :one
INSERT INTO bids (
    product_id, bidder_id, bid_amount
) VALUES ($1, $2, $3)
RETURNING id, created_at;

-- name: GetBidsByProductId :many
SELECT * FROM bids
WHERE product_id = $1
ORDER BY bid_amount DESC;

-- GetHighestBidByProductId :one
SELECT * FROM bids
WHERE product_id = $1
ORDER BY bid_amount DESC
LIMIT 1;
