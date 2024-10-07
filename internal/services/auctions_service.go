package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageKind int

const (
	// Errors
	AuctionFinshed MessageKind = iota
	// NOTE: Maybe this should be an error....
	InvalidJSON
	FailedToPlaceBid

	// User Actiosn
	PlaceBid
)

// Will hold all the auctionRooms
type AuctionLobby struct {
	Rooms      map[uuid.UUID]*AuctionRoom
	sync.Mutex // Maybe try not using a mutex but handle this with channels.
}

type Message struct {
	Message  string      `json:"message"`
	Kind     MessageKind `json:"kind"`
	BidValue float64     `json:"bid_value"`
}

// A WS "chat" for a specific product.
type AuctionRoom struct {
	// Holds the deadline for the auction
	Context context.Context

	// Sync method for every message that needs to be Broadcast
	Broadcast chan Message

	// Users that need to be added or removed from the auction room
	Register   chan *Client
	Unregister chan *Client

	Clients map[*Client]bool

	roomID string
}

func NewAuctionRoom(ctx context.Context, id string) *AuctionRoom {
	return &AuctionRoom{
		roomID:     id,
		Broadcast:  make(chan Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Context:    ctx,
	}
}

// Should run in a go routine
func (r *AuctionRoom) Run() {
	for {
		select {
		case client := <-r.Register:
			r.Clients[client] = true
		case client := <-r.Unregister:
			delete(r.Clients, client)
			close(client.Send)
		case message := <-r.Broadcast:
			for client := range r.Clients {
				// This will also send back to the user
				// the message he also sent, so lets filter this and only send him the necessary "success message"
				select {
				case client.Send <- message:
					// do nothing
				default:
					close(client.Send)
					delete(r.Clients, client)
				}
			}
		case <-r.Context.Done():
			for client := range r.Clients {
				client.Send <- Message{Kind: AuctionFinshed, Message: "auction has been finished"}
			}
			return
		}
	}
}

// Review if this is really needed...
const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 512
)

type Client struct {
	Room   *AuctionRoom
	Conn   *websocket.Conn
	Send   chan Message
	UserId uuid.UUID
}

func NewClient(room *AuctionRoom, conn *websocket.Conn, userId uuid.UUID) *Client {
	return &Client{
		Room:   room,
		Conn:   conn,
		Send:   make(chan Message, 512),
		UserId: userId,
	}
}

func (c *Client) WriteEventLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	// WriteEventLoop
	for {
		select {
		case message, ok := <-c.Send:
			// If a deadline is meet the underlying c.Conn is corrupt and all writes will return an error.
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Make this a valid type...
				c.Conn.WriteJSON(Message{
					Kind:    websocket.CloseMessage,
					Message: "closing websocket conn",
				})
				return
			}

			err := c.Conn.WriteJSON(message)
			// if err != nil || message.Kind == AuctionFinshed {
			if err != nil {
				c.Room.Unregister <- c
				c.Conn.Close()
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Read messages from the client and sends them in order to be broadcast into the AuctionRoom
// one per conn
func (c *Client) ReadEventLoop() {
	defer func() {
		fmt.Println("its breakig the read loop ")
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	// Maybe remove those deadline stuff...
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var m Message
		err := c.Conn.ReadJSON(&m)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("Unexpected Close Error", "error", err)
				return
			}
			c.Conn.WriteJSON(map[string]any{
				"error": "invalid json",
			})
		}

		// TODO: Handle multiple kinds of message
		switch m.Kind {
		case PlaceBid:
			fmt.Printf("A new bid with the value %0.2f was placed", m.BidValue)
		}

		c.Room.Broadcast <- m
	}
}
