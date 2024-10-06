-- name: CreateProduct :one
INSERT INTO products (
    seller_id, product_name, description,
    base_price, auction_start, auction_end
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: DeleteProduct :exec
DELETE FROM products
WHERE id = $1;

-- name: GetProductById :one
SELECT 
    id,
    seller_id,
    product_name,
    description,
    base_price,
    auction_start,
    auction_end,
    is_sold
FROM products
WHERE id = $1;

-- name: ListLiveProductAuctions :many
SELECT
    id,
    seller_id,
    product_name,
    description,
    base_price,
    auction_start,
    auction_end,
    is_sold
FROM products
WHERE auction_end > now() AND auction_start < now() AND is_sold = false;

-- name: ListActiveAndUpcomingAuctions :many
SELECT
    id,
    seller_id,
    product_name,
    description,
    base_price,
    auction_start,
    auction_end,
    is_sold
FROM products
WHERE auction_end > now() AND is_sold = false;

-- name: GetProductsByUser :many
SELECT
    id,
    seller_id,
    product_name,
    description,
    base_price,
    auction_start,
    auction_end,
    is_sold
FROM products
WHERE seller_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;


-- name: ListAllProducts :many
SELECT
    id,
    seller_id,
    product_name,
    description,
    base_price,
    auction_start,
    auction_end,
    is_sold
FROM products
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

