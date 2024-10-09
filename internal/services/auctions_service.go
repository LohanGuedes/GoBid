package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageKind int

const (
	// Responses
	AuctionFinshed MessageKind = iota
	InvalidJSON
	FailedToPlaceBid
	NewHigherBid
	SuccessfullyPlacedBid

	// Requests
	PlaceBid

	// Internal
	Disconnect
)

// Will hold all the auctionRooms
type AuctionLobby struct {
	Rooms map[uuid.UUID]*AuctionRoom
	sync.Mutex
}

type Message struct {
	Message  string      `json:"message,omitempty"`
	Kind     MessageKind `json:"kind"`
	BidValue float64     `json:"bid_value,omitempty"`
	UserID   uuid.UUID   `json:"user_id,omitempty"`
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

	Clients map[uuid.UUID]*Client

	ProductService *ProductService
	BidsService    *BidsService
	ID             uuid.UUID
}

func NewAuctionRoom(ctx context.Context, id uuid.UUID, productService *ProductService, bidsService *BidsService) *AuctionRoom {
	return &AuctionRoom{
		ID:             id,
		Broadcast:      make(chan Message),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		Clients:        make(map[uuid.UUID]*Client),
		Context:        ctx,
		ProductService: productService,
		BidsService:    bidsService,
	}
}

func (r *AuctionRoom) broadCastMessage(message Message) {
	slog.Info("Message Recieved", "RoomId", r.ID, "message", message, "user_id", message.UserID)
	switch message.Kind {
	case PlaceBid:
		bid, err := r.BidsService.PlaceBid(r.Context, r.ID, message.UserID, message.BidValue)
		if err != nil {
			if errors.Is(err, ErrBidIsTooLow) {
				// Write back to the user that the bid is too low
				if client, ok := r.Clients[message.UserID]; ok {
					client.Send <- Message{Kind: FailedToPlaceBid, Message: ErrBidIsTooLow.Error(), UserID: message.UserID}
					return
				}
			}
		}

		if client, ok := r.Clients[message.UserID]; ok {
			client.Send <- Message{Kind: SuccessfullyPlacedBid, Message: "Your bid was successfully placed."}
		}

		for id, client := range r.Clients {
			newBidMessage := Message{Kind: NewHigherBid, Message: "A new bid was placed", BidValue: bid.BidAmount}
			if id == message.UserID { // Do not send this to the user.
				continue
			}
			select {
			case client.Send <- newBidMessage:
			default:
				close(client.Send)
				delete(r.Clients, id)
			}
		}
	case InvalidJSON:
		client, ok := r.Clients[message.UserID]
		if !ok {
			slog.Info("Client not found in hashmap")
			return
		}
		client.Send <- message
	}
}

func (r *AuctionRoom) unregisterClient(client *Client) {
	slog.Info("New user disconnected", "userID", client.UserId)
	delete(r.Clients, client.UserId)
	close(client.Send)
}

func (r *AuctionRoom) registerClient(client *Client) {
	slog.Info("New user connected", "client", client)
	r.Clients[client.UserId] = client
}

// Should run in a go routine
func (r *AuctionRoom) Run() {
	defer func() {
		close(r.Broadcast)
		close(r.Register)
		close(r.Unregister)
	}()

	for {
		select {
		case client := <-r.Register:
			r.registerClient(client)

		case client := <-r.Unregister:
			r.unregisterClient(client)

		case message := <-r.Broadcast:
			r.broadCastMessage(message)

		case <-r.Context.Done():
			slog.Info("Auction ending", "auctionID", r.ID)
			for _, client := range r.Clients {
				client.Send <- Message{Kind: AuctionFinshed, Message: "auction has been finished"}
			}
			return
		}
	}
}

// TODO: Review if this is really needed...
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
			if message.Kind == AuctionFinshed {
				close(c.Send)
				return
			}

			// NOTE: If a deadline is meet the underlying c.Conn is corrupt and all writes will return an error.
			// NOTE: Should I really add this to the videos???
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
			if err != nil {
				c.Room.Unregister <- c
				return
			}

		case <-ticker.C:
			// NOTE: Should I really add this to the videos???
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
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	// NOTE: Maybe remove those deadline stuff...
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var m Message
		// NOTE: inform the user that sent this message to the room
		m.UserID = c.UserId
		err := c.Conn.ReadJSON(&m)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("Unexpected Close Error", "error", err)
				return
			}

			c.Room.Broadcast <- Message{
				Kind:    InvalidJSON,
				Message: "invalid json",
				UserID:  m.UserID,
			}
		}
		c.Room.Broadcast <- m
	}
}
