package sdk

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Plugin represents a connection to the host application.
type Plugin struct {
	conn            *websocket.Conn
	pendingRequests map[string]chan StorageResponse
	requestsMutex   sync.RWMutex
	requestCounter  int
	counterMutex    sync.Mutex
}

// Start establishes a WebSocket connection with the host and registers the plugin.
func Start(pluginInfo *RegisterRequest) (*Plugin, error) {
	wsPort := flag.Int("ws-port", 0, "WebSocket server port of the main application")
	flag.Parse()

	if *wsPort == 0 {
		return nil, fmt.Errorf("host WebSocket port not provided via --ws-port flag")
	}

	// Construct the WebSocket URL.
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("127.0.0.1:%d", *wsPort), Path: "/ws"}
	log.Printf("Connecting to host at %s", u.String())

	// Dial the host.
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to host WebSocket: %w", err)
	}

	// Marshal the registration info.
	payload, err := json.Marshal(pluginInfo)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to marshal plugin info: %w", err)
	}

	// Send the registration message.
	if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send registration message: %w", err)
	}

	log.Println("Plugin successfully sent registration request.")

	plugin := &Plugin{
		conn:            conn,
		pendingRequests: make(map[string]chan StorageResponse),
	}

	return plugin, nil
}

// Listen runs a loop to process incoming messages from the host.
// It takes a handler function to be executed for each message.
func (p *Plugin) Listen(handler func(messageType int, data []byte)) {
	// Ensure the connection is closed when the listener exits.
	defer p.conn.Close()

	for {
		messageType, data, err := p.conn.ReadMessage()
		if err != nil {
			// Check for a clean close from the server.
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Println("Host closed the connection cleanly.")
			} else {
				log.Printf("Error reading message from host: %v", err)
			}
			break // Exit loop on any error.
		}
		// Check if this is a storage response
		if messageType == websocket.TextMessage {
			var response StorageResponse
			if err := json.Unmarshal(data, &response); err == nil && response.ID != "" {
				// This is a storage response
				p.requestsMutex.RLock()
				responseChan, exists := p.pendingRequests[response.ID]
				p.requestsMutex.RUnlock()

				if exists {
					select {
					case responseChan <- response:
					case <-time.After(time.Second):
						log.Printf("Warning: Storage response channel timeout for request %s", response.ID)
					}
					continue
				}
			}
		}

		// Execute the plugin-defined handler for the message.
		handler(messageType, data)
	}
}

// generateRequestID generates a unique request ID
func (p *Plugin) generateRequestID() string {
	p.counterMutex.Lock()
	defer p.counterMutex.Unlock()
	p.requestCounter++
	return fmt.Sprintf("req_%d_%d", time.Now().Unix(), p.requestCounter)
}

// sendStorageRequest sends a storage request and waits for the response
func (p *Plugin) sendStorageRequest(req StorageRequest) (*StorageResponse, error) {
	req.ID = p.generateRequestID()

	// Create response channel
	responseChan := make(chan StorageResponse, 1)
	p.requestsMutex.Lock()
	p.pendingRequests[req.ID] = responseChan
	p.requestsMutex.Unlock()

	// Cleanup on exit
	defer func() {
		p.requestsMutex.Lock()
		delete(p.pendingRequests, req.ID)
		p.requestsMutex.Unlock()
		close(responseChan)
	}()

	// Send request
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal storage request: %w", err)
	}

	if err := p.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return nil, fmt.Errorf("failed to send storage request: %w", err)
	}

	// Wait for response with timeout
	select {
	case response := <-responseChan:
		if !response.Success {
			return nil, fmt.Errorf("storage operation failed: %s", response.Error)
		}
		return &response, nil
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("storage request timeout")
	}
}

// Close sends a close message and shuts down the connection.
func (p *Plugin) Close() {
	if p.conn != nil {
		// Politely tell the server we're closing the connection.
		err := p.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Printf("Error sending close message: %v", err)
		}
		p.conn.Close()
	}
}
