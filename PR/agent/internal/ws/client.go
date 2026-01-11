package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	URL           string
	Token         string
	Connection    *websocket.Conn
	Send          chan []byte
	Receive       chan []byte
	Connected     bool
	OnConnect     func()
	OnDisconnect  func()
	OnCommand     func(map[string]interface{})
}

func NewClient(serverURL, token string) *Client {
	// Convert HTTP URL to WebSocket URL
	wsURL := serverURL
	if serverURL[:4] == "http" {
		wsURL = "ws" + serverURL[4:]
	}
	
	return &Client{
		URL:       wsURL,
		Token:     token,
		Send:      make(chan []byte, 256),
		Receive:   make(chan []byte, 256),
		Connected: false,
	}
}

func (c *Client) Connect() error {
	// Parse the URL
	u, err := url.Parse(c.URL + "/ws")
	if err != nil {
		return fmt.Errorf("error parsing URL: %v", err)
	}
	
	// Add token as query parameter
	q := u.Query()
	q.Set("token", c.Token)
	u.RawQuery = q.Encode()
	
	// Set up headers
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+c.Token)
	
	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		return fmt.Errorf("error connecting to WebSocket: %v", err)
	}
	
	c.Connection = conn
	c.Connected = true
	
	// Start message handling goroutines
	go c.readPump()
	go c.writePump()
	
	if c.OnConnect != nil {
		c.OnConnect()
	}
	
	return nil
}

func (c *Client) Disconnect() {
	c.Connected = false
	if c.Connection != nil {
		c.Connection.Close()
	}
	
	if c.OnDisconnect != nil {
		c.OnDisconnect()
	}
}

func (c *Client) readPump() {
	defer func() {
		c.Connection.Close()
		c.Connected = false
	}()
	
	for {
		if !c.Connected {
			break
		}
		
		_, message, err := c.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		// Parse the received message
		var msgMap map[string]interface{}
		err = json.Unmarshal(message, &msgMap)
		if err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}
		
		// Send to receive channel
		select {
		case c.Receive <- message:
		default:
			// Channel is full, skip message
		}
		
		// Handle command if present
		if _, ok := msgMap["action"]; ok {
			if c.OnCommand != nil {
				c.OnCommand(msgMap)
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	
	for {
		if !c.Connected {
			break
		}
		
		select {
		case message, ok := <-c.Send:
			if !ok {
				// Channel closed, close connection
				c.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			if err := c.Connection.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Write error: %v", err)
				return
			}
		case <-ticker.C:
			// Send ping
			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping error: %v", err)
				return
			}
		}
	}
}

func (c *Client) SendMetrics(hashrate float64, cpuUsage float64, ramUsage float64, tempCPU int, status string) error {
	metrics := map[string]interface{}{
		"hashrate":  hashrate,
		"cpu_usage": cpuUsage,
		"ram_usage": ramUsage,
		"temp_cpu":  tempCPU,
		"status":    status,
		"platform":  "unknown", // This would be determined dynamically
		"version":   "1.0.0", // This would come from config or build
	}
	
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	
	select {
	case c.Send <- data:
	default:
		// Channel is full
		return fmt.Errorf("send channel is full")
	}
	
	return nil
}

func (c *Client) Reconnect() error {
	c.Disconnect()
	time.Sleep(5 * time.Second) // Wait before reconnecting
	return c.Connect()
}