package main

// App, this will be an showcase of an app with:
// - Login (auth & authorization)
//		The goal with login stuff, is to present good practices in:
//			- DONE: PasswordHashing
//			- DONE: Security configs -> CSRF
//			- DONE: Database setups
//			- TODO: Websocket AuctionStuff
//			- TODO: Swagger with Swaggo

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/lohanguedes/gobid/internal/api"
	"github.com/lohanguedes/gobid/internal/services"
)

// This function is special, you can creaty any ammount of them inside a
// package. and it will be ran when the package is called the first
// time, only one time.
func init() {
	gob.Register(uuid.UUID{})
}

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv("GOBID_DATABASE_USER"),
		os.Getenv("GOBID_DATABASE_PASSWORD"),
		os.Getenv("GOBID_DATABASE_HOST"),
		os.Getenv("GOBID_DATABASE_PORT"),
		os.Getenv("GOBID_DATABASE_NAME"),
	))
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	// Configuration
	s := scs.New()
	s.Store = pgxstore.New(pool)
	s.Lifetime = 24 * time.Hour
	s.Cookie.HttpOnly = true
	s.Cookie.SameSite = http.SameSiteLaxMode

	api := api.Api{
		Router:         chi.NewMux(),
		Session:        s,
		UserService:    services.NewUserService(pool),
		ProductService: services.NewProductService(pool),
		BidsService:    services.NewBidsService(pool),
		Upgrader: websocket.Upgrader{
			// For tests and development only
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		AuctionLobby: services.AuctionLobby{
			Rooms: make(map[uuid.UUID]*services.AuctionRoom),
		},
	}

	api.BindRoutes()

	fmt.Println("Starting server on port :3080")
	if err := http.ListenAndServe("localhost:3080", api.Router); err != nil {
		panic(err)
	}
}
