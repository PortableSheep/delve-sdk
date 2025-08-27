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

// HeartbeatMessage represents a heartbeat message
type HeartbeatMessage struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// StateManager interface allows plugins to implement state saving
type StateManager interface {
	SaveState() error
	GetState() interface{}
}

// Plugin represents a connection to the host application.
type Plugin struct {
	conn            *websocket.Conn
	pendingRequests map[string]chan StorageResponse
	requestsMutex   sync.RWMutex
	requestCounter  int
	counterMutex    sync.Mutex
	stateManager    StateManager
	heartbeatStop   chan struct{}
	heartbeatDone   chan struct{}
	lastHeartbeat   time.Time
	heartbeatMutex  sync.RWMutex
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
		heartbeatStop:   make(chan struct{}),
		heartbeatDone:   make(chan struct{}),
		lastHeartbeat:   time.Now(),
	}

	return plugin, nil
}

// SetStateManager sets the state manager for graceful shutdown
func (p *Plugin) SetStateManager(sm StateManager) {
	p.stateManager = sm
}

// StartHeartbeat starts the heartbeat mechanism
func (p *Plugin) StartHeartbeat(interval time.Duration, timeout time.Duration) {
	go p.runHeartbeat(interval, timeout)
}

// runHeartbeat runs the heartbeat loop
func (p *Plugin) runHeartbeat(interval time.Duration, timeout time.Duration) {
	defer close(p.heartbeatDone)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.heartbeatStop:
			log.Println("Heartbeat stopped")
			return
		case <-ticker.C:
			// Send heartbeat
			heartbeat := HeartbeatMessage{
				Type:      "heartbeat",
				Timestamp: time.Now(),
			}

			data, err := json.Marshal(heartbeat)
			if err != nil {
				log.Printf("Failed to marshal heartbeat: %v", err)
				continue
			}

			// Set write deadline
			p.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := p.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
				p.gracefulShutdown("heartbeat send failed")
				return
			}

			// Check if we've received a response recently
			p.heartbeatMutex.RLock()
			lastHeartbeat := p.lastHeartbeat
			p.heartbeatMutex.RUnlock()

			if time.Since(lastHeartbeat) > timeout {
				log.Printf("Host heartbeat timeout exceeded (%v), shutting down gracefully", timeout)
				p.gracefulShutdown("heartbeat timeout")
				return
			}
		}
	}
}

// gracefulShutdown performs graceful shutdown with state saving
func (p *Plugin) gracefulShutdown(reason string) {
	log.Printf("Initiating graceful shutdown: %s", reason)

	if p.stateManager != nil {
		log.Println("Saving plugin state...")
		if err := p.stateManager.SaveState(); err != nil {
			log.Printf("Failed to save state: %v", err)
		} else {
			log.Println("Plugin state saved successfully")
		}
	}

	// Close the connection
	p.Close()

	// Exit the process
	log.Println("Plugin shutting down gracefully")
	time.Sleep(100 * time.Millisecond) // Give logs time to flush
	panic("Graceful shutdown")         // This will be caught by the plugin's main loop
}

// Listen runs a loop to process incoming messages from the host.
// It takes a handler function to be executed for each message.
func (p *Plugin) Listen(handler func(messageType int, data []byte)) {
	// Ensure the connection is closed when the listener exits.
	defer func() {
		p.stopHeartbeat()
		p.conn.Close()
	}()

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

		// Check if this is a heartbeat response
		if messageType == websocket.TextMessage {
			var heartbeat HeartbeatMessage
			if err := json.Unmarshal(data, &heartbeat); err == nil && heartbeat.Type == "heartbeat_ack" {
				p.heartbeatMutex.Lock()
				p.lastHeartbeat = time.Now()
				p.heartbeatMutex.Unlock()
				continue
			}
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

// stopHeartbeat stops the heartbeat mechanism
func (p *Plugin) stopHeartbeat() {
	select {
	case <-p.heartbeatStop:
		// Already stopped
		return
	default:
		close(p.heartbeatStop)
	}

	// Wait for heartbeat goroutine to finish
	select {
	case <-p.heartbeatDone:
	case <-time.After(5 * time.Second):
		log.Println("Heartbeat stop timeout")
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
	// Stop heartbeat first
	p.stopHeartbeat()

	if p.conn != nil {
		// Politely tell the server we're closing the connection.
		err := p.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Printf("Error sending close message: %v", err)
		}
		p.conn.Close()
	}
}
