package api

import (
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/lohanguedes/gobid/internal/services"
)

// This file is only used for documentind the api constraints. and injecting dependencies.
type Api struct {
	Router         *chi.Mux
	Session        *scs.SessionManager
	UserService    services.UserService
	ProductService services.ProductService
	BidsService    services.BidsService
	Upgrader       websocket.Upgrader
	AuctionLobby   services.AuctionLobby
}
