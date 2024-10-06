package product

import (
	"context"
	"time"

	"github.com/rocketseat/go-first-auth/internal/validator"
)

type CreateProductReq struct {
	ProductName  string    `json:"product_name"`
	Description  string    `json:"description"`
	BasePrice    float64   `json:"base_price"`
	AuctionStart time.Time `json:"auction_start"`
	AuctionEnd   time.Time `json:"auction_end"`
}

const minAuctionDuration = 2 * time.Hour

func (req CreateProductReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.ProductName), "product_name", "this field cannot be blank")
	eval.CheckField(validator.NotBlank(req.Description), "description", "this field cannot be blank")
	eval.CheckField(
		validator.MinChars(req.Description, 10) &&
			validator.MaxChars(req.Description, 255),
		"description",
		"this field must have a length between 10 and 255")
	eval.CheckField(req.BasePrice >= 0, "base_price", "base price must be or equal to zero")

	eval.CheckField(req.AuctionStart.After(time.Now()), "auction_start", "auction start must be in the future")

	eval.CheckField(req.AuctionEnd.Sub(req.AuctionStart) >= minAuctionDuration, "auction_end", "auction end must be at least 2 hours after auction start")

	return eval
}
