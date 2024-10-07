package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rocketseat/go-first-auth/internal/services"
)

// ws/subscribe/{product_id}
func (api *Api) handleSubcribeUserToAuction(w http.ResponseWriter, r *http.Request) {
	rawProductId := chi.URLParam(r, "product_id")

	productId, err := uuid.Parse(rawProductId)
	if err != nil {
		encodeJson(w, r, http.StatusNotFound, map[string]any{
			"message": "failed to parse uuid - must be a valid uuid",
		})
		return
	}

	_, err = api.ProductService.GetProductById(r.Context(), productId)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			encodeJson(w, r, http.StatusNotFound, map[string]any{
				"message": "product with given id not found",
			})
			return
		}
		encodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"message": "unexpected error, try again later.",
		})
		return
	}

	// Get the room info before
	userId, ok := api.Session.Get(r.Context(), "authenticatedUserId").(uuid.UUID)
	if !ok {
		encodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"message": "failed to authenticate your session",
		})
		return
	}

	api.AuctionLobby.Lock()
	room, ok := api.AuctionLobby.Rooms[productId]
	api.AuctionLobby.Unlock()

	if !ok {
		encodeJson(w, r, http.StatusBadRequest, map[string]any{
			"message": "The Auction has been endded or does not exist.",
		})
		return
	}

	// if it exist **THEN** try to change protocols
	conn, err := api.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		encodeJson(w, r, http.StatusBadRequest, map[string]any{
			"message": "could not upgrade connection to a websocket connection",
		})
		return
	}
	client := services.NewClient(room, conn, userId)

	go client.ReadEventLoop()
	go client.WriteEventLoop()
	room.Register <- client
}
