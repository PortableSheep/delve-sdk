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

// RegisterRequest defines the structure for the registration request.
// The JSON tags MUST match what the host application expects.
type RegisterRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	UiComponentPath  string `json:"uiComponentPath"`
	CustomElementTag string `json:"customElementTag"`
}

// StorageRequest represents a storage operation request
type StorageRequest struct {
	ID        string                 `json:"id"`
	Operation string                 `json:"operation"`
	Type      string                 `json:"type"`
	Key       string                 `json:"key,omitempty"`
	Value     interface{}            `json:"value,omitempty"`
	Version   string                 `json:"version,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// StorageResponse represents a response to a storage request
type StorageResponse struct {
	ID      string      `json:"id"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// StorageItem represents a stored item with metadata
type StorageItem struct {
	Key       string                 `json:"key"`
	Value     interface{}            `json:"value"`
	Type      string                 `json:"type"`
	Version   string                 `json:"version"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
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

// Store saves data with the specified key and storage type
func (p *Plugin) Store(storageType, key string, value interface{}, version string) error {
	req := StorageRequest{
		Operation: "store",
		Type:      storageType,
		Key:       key,
		Value:     value,
		Version:   version,
	}

	_, err := p.sendStorageRequest(req)
	return err
}

// Load retrieves data for the specified key and storage type
func (p *Plugin) Load(storageType, key string) (*StorageItem, error) {
	req := StorageRequest{
		Operation: "load",
		Type:      storageType,
		Key:       key,
	}

	response, err := p.sendStorageRequest(req)
	if err != nil {
		return nil, err
	}

	// Convert response data to StorageItem
	itemData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal storage item: %w", err)
	}

	var item StorageItem
	if err := json.Unmarshal(itemData, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal storage item: %w", err)
	}

	return &item, nil
}

// Delete removes data for the specified key and storage type
func (p *Plugin) Delete(storageType, key string) error {
	req := StorageRequest{
		Operation: "delete",
		Type:      storageType,
		Key:       key,
	}

	_, err := p.sendStorageRequest(req)
	return err
}

// List returns all keys for the specified storage type
func (p *Plugin) List(storageType string) ([]string, error) {
	req := StorageRequest{
		Operation: "list",
		Type:      storageType,
	}

	response, err := p.sendStorageRequest(req)
	if err != nil {
		return nil, err
	}

	// Convert response data to string slice
	keysData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal keys: %w", err)
	}

	var keys []string
	if err := json.Unmarshal(keysData, &keys); err != nil {
		return nil, fmt.Errorf("failed to unmarshal keys: %w", err)
	}

	return keys, nil
}

// Clear removes all data for the specified storage type
func (p *Plugin) Clear(storageType string) error {
	req := StorageRequest{
		Operation: "clear",
		Type:      storageType,
	}

	_, err := p.sendStorageRequest(req)
	return err
}

// GetStats returns storage statistics
func (p *Plugin) GetStats() (map[string]int, error) {
	req := StorageRequest{
		Operation: "stats",
	}

	response, err := p.sendStorageRequest(req)
	if err != nil {
		return nil, err
	}

	// Convert response data to stats map
	statsData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stats: %w", err)
	}

	var stats map[string]int
	if err := json.Unmarshal(statsData, &stats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	return stats, nil
}

// Convenience methods for specific storage types

// StoreConfig stores configuration data
func (p *Plugin) StoreConfig(key string, value interface{}, version string) error {
	return p.Store("config", key, value, version)
}

// LoadConfig loads configuration data
func (p *Plugin) LoadConfig(key string) (*StorageItem, error) {
	return p.Load("config", key)
}

// StoreData stores application data
func (p *Plugin) StoreData(key string, value interface{}, version string) error {
	return p.Store("data", key, value, version)
}

// LoadData loads application data
func (p *Plugin) LoadData(key string) (*StorageItem, error) {
	return p.Load("data", key)
}

// StoreState stores current state
func (p *Plugin) StoreState(key string, value interface{}, version string) error {
	return p.Store("state", key, value, version)
}

// LoadState loads current state
func (p *Plugin) LoadState(key string) (*StorageItem, error) {
	return p.Load("state", key)
}

// StoreCache stores cache data
func (p *Plugin) StoreCache(key string, value interface{}, version string) error {
	return p.Store("cache", key, value, version)
}

// LoadCache loads cache data
func (p *Plugin) LoadCache(key string) (*StorageItem, error) {
	return p.Load("cache", key)
}

// handleResponses processes storage responses in the background
func (p *Plugin) handleResponses() {
	// This method is no longer needed as response handling
	// is done directly in the Listen method
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
