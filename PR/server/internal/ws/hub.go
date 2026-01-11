package ws

import (
	"sync"
	"pr-server/internal/models"
)

// Client represents a connected agent
type Client struct {
	ID     string
	Hub    *Hub
	Conn   interface{} // We'll use interface{} for now, we'll define the actual type later
	Send   chan []byte
	Agent  *models.Agent
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	clients map[string]*Client

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex to protect concurrent access to clients
	mutex sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.ID] = client
			h.mutex.Unlock()
			
		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
			}
			h.mutex.Unlock()
			
		case message := <-h.broadcast:
			h.mutex.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.ID)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// GetClient returns a client by ID
func (h *Hub) GetClient(id string) (*Client, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	client, ok := h.clients[id]
	return client, ok
}

// SendToClient sends a message to a specific client
func (h *Hub) SendToClient(clientID string, message []byte) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	client, ok := h.clients[clientID]
	if !ok {
		return false
	}
	
	select {
	case client.Send <- message:
		return true
	default:
		return false
	}
}

// Register registers a new client
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}