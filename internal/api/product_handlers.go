package api

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rocketseat/go-first-auth/internal/services"
	"github.com/rocketseat/go-first-auth/internal/usecase/product"
)

func (api *Api) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	data, problems, err := decodeValidJson[product.CreateProductReq](r)
	if err != nil {
		_ = encodeJson(w, r, http.StatusBadRequest, problems)
		return
	}

	userID, ok := api.Session.Get(r.Context(), "authenticatedUserId").(uuid.UUID)
	if !ok {
		_ = encodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error try again later",
		})
		return
	}

	id, err := api.ProductService.CreateProduct(
		r.Context(),
		userID,
		data.ProductName,
		data.BasePrice,
		pgtype.Timestamptz{Time: data.AuctionStart, Valid: true},
		pgtype.Timestamptz{Time: data.AuctionEnd, Valid: true},
	)
	if err != nil {
		_ = encodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "failed to create product auction",
		})
	}

	// We're using context.Background() because if we use r.Context() the context will be cancelled whenever the request is finished, killing our go-routine before any user is able to join in it.
	ctx, _ := context.WithDeadline(context.Background(), data.AuctionEnd)

	newAuctionRoom := services.NewAuctionRoom(ctx, id.String())
	go newAuctionRoom.Run()

	api.AuctionLobby.Lock()
	api.AuctionLobby.Rooms[id] = newAuctionRoom
	api.AuctionLobby.Unlock()

	_ = encodeJson(w, r, http.StatusCreated, map[string]any{
		"product_id": id,
	})
}

// GET
func (api *Api) handleListProducts(w http.ResponseWriter, r *http.Request) {
	encodeJson(w, r, http.StatusOK, map[string]any{
		"data": "hey!",
	})
}

// GET
func (api *Api) handleListProductById(w http.ResponseWriter, r *http.Request) {
	panic("todo")
}
