package sdk

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
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
	hostPending     map[string]chan HostResponse
	requestsMutex   sync.RWMutex
	requestCounter  int
	counterMutex    sync.Mutex
	stateManager    StateManager
	heartbeatStop   chan struct{}
	heartbeatDone   chan struct{}
	lastHeartbeat   time.Time
	heartbeatMutex  sync.RWMutex
	// command handling
	commandHandlers map[string]CommandHandler
	commandMutex    sync.RWMutex
	// event handling
	eventHandlers map[string][]func(any)
	eventMutex    sync.RWMutex
	// cached UI contributions (to allow partial, ergonomic updates)
	uiContrib UIContributions
	uiMutex   sync.Mutex
}

// UI contribution types (mirrors host expectation, but kept minimal and independent)
type CommandRef struct {
	ID   string        `json:"id"`
	Args []interface{} `json:"args,omitempty"`
}
type SidebarItem struct {
	ID             string      `json:"id"`
	Title          string      `json:"title"`
	Icon           string      `json:"icon,omitempty"`
	ElementTag     string      `json:"elementTag,omitempty"`
	ComponentPath  string      `json:"componentPath,omitempty"`
	Order          int         `json:"order,omitempty"`
	OnClickCommand *CommandRef `json:"onClickCommand,omitempty"`
}
type FooterWidget struct {
	ElementTag     string      `json:"elementTag,omitempty"`
	ComponentPath  string      `json:"componentPath,omitempty"`
	Order          int         `json:"order,omitempty"`
	Icon           string      `json:"icon,omitempty"`
	Title          string      `json:"title,omitempty"`
	BadgeCount     int         `json:"badgeCount,omitempty"`
	OnClickCommand *CommandRef `json:"onClickCommand,omitempty"`
}
type UIContributions struct {
	Sidebar []SidebarItem  `json:"sidebar,omitempty"`
	Footer  []FooterWidget `json:"footer,omitempty"`
}

// Emit sends an arbitrary event envelope to the host. The host can choose to
// broadcast this to frontend listeners for the plugin UI. The envelope shape:
// {"type":"event","event":"<eventName>","plugin":"<pluginName>","data":<payload>}
// Data must be JSON-serializable.
func (p *Plugin) Emit(event string, payload any) error {
	if p == nil || p.conn == nil {
		return fmt.Errorf("plugin connection not initialised")
	}
	env := map[string]any{
		"type":  "event",
		"event": event,
		"data":  payload,
	}
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return p.conn.WriteMessage(websocket.TextMessage, b)
}

// UpdateUIContributions sends an update payload to the host to modify UI contributions live.
// Host will normalize paths and re-emit aggregated UI contributions to frontend.
func (p *Plugin) UpdateUIContributions(update UIContributions) error {
	if p == nil || p.conn == nil {
		return fmt.Errorf("plugin connection not initialised")
	}
	env := map[string]any{
		"type":          "uiContributions/update",
		"contributions": update,
	}
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return p.conn.WriteMessage(websocket.TextMessage, b)
}

// SetUIContributions replaces the cached contributions and sends them to the host.
func (p *Plugin) SetUIContributions(c UIContributions) error {
	p.uiMutex.Lock()
	p.uiContrib = c
	p.uiMutex.Unlock()
	return p.UpdateUIContributions(c)
}

// UpsertFooterWidget inserts or replaces a footer widget in the cached contributions and sends update.
// Matching preference: ElementTag (if provided), else Title.
func (p *Plugin) UpsertFooterWidget(w FooterWidget) error {
	p.uiMutex.Lock()
	defer p.uiMutex.Unlock()
	replaced := false
	for i := range p.uiContrib.Footer {
		cur := p.uiContrib.Footer[i]
		if (w.ElementTag != "" && cur.ElementTag == w.ElementTag) || (w.ElementTag == "" && w.Title != "" && cur.Title == w.Title) {
			p.uiContrib.Footer[i] = w
			replaced = true
			break
		}
	}
	if !replaced {
		p.uiContrib.Footer = append(p.uiContrib.Footer, w)
	}
	return p.UpdateUIContributions(UIContributions{Footer: p.uiContrib.Footer})
}

// SetFooterBadgeCount finds a footer widget and updates its badge count.
// It identifies the widget by ElementTag when provided; otherwise by Title.
func (p *Plugin) SetFooterBadgeCount(identifier string, count int) error {
	p.uiMutex.Lock()
	defer p.uiMutex.Unlock()
	for i := range p.uiContrib.Footer {
		if p.uiContrib.Footer[i].ElementTag == identifier || (p.uiContrib.Footer[i].ElementTag == "" && p.uiContrib.Footer[i].Title == identifier) {
			p.uiContrib.Footer[i].BadgeCount = count
			return p.UpdateUIContributions(UIContributions{Footer: p.uiContrib.Footer})
		}
	}
	// if not found, create a minimal icon badge with title as identifier
	p.uiContrib.Footer = append(p.uiContrib.Footer, FooterWidget{Title: identifier, BadgeCount: count})
	return p.UpdateUIContributions(UIContributions{Footer: p.uiContrib.Footer})
}

// UpsertSidebarItem inserts or replaces a sidebar item in the cached contributions and sends update.
// Matching preference: ID (if provided), else ElementTag, else Title.
func (p *Plugin) UpsertSidebarItem(item SidebarItem) error {
	p.uiMutex.Lock()
	defer p.uiMutex.Unlock()
	replaced := false
	for i := range p.uiContrib.Sidebar {
		cur := p.uiContrib.Sidebar[i]
		if (item.ID != "" && cur.ID == item.ID) || (item.ID == "" && item.ElementTag != "" && cur.ElementTag == item.ElementTag) || (item.ID == "" && item.ElementTag == "" && item.Title != "" && cur.Title == item.Title) {
			p.uiContrib.Sidebar[i] = item
			replaced = true
			break
		}
	}
	if !replaced {
		p.uiContrib.Sidebar = append(p.uiContrib.Sidebar, item)
	}
	return p.UpdateUIContributions(UIContributions{Sidebar: p.uiContrib.Sidebar})
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
		hostPending:     make(map[string]chan HostResponse),
		heartbeatStop:   make(chan struct{}),
		heartbeatDone:   make(chan struct{}),
		lastHeartbeat:   time.Now(),
		commandHandlers: make(map[string]CommandHandler),
		eventHandlers:   make(map[string][]func(any)),
	}

	return plugin, nil
}

// CommandHandler is a function that handles a command from the host.
// args are deserialized JSON values. Return value must be JSON-serializable.
type CommandHandler func(ctx context.Context, args []any) (any, error)

// CommandMap convenience for bulk registration

// OnCommand registers a handler for a given command ID.
func (p *Plugin) OnCommand(command string, handler CommandHandler) {
	p.commandMutex.Lock()
	defer p.commandMutex.Unlock()
	p.commandHandlers[command] = handler
}

// OnEvent registers a callback for a named event emitted by the plugin (loopback)
// or broadcast by host to the plugin (future use). Multiple handlers may be registered.
func (p *Plugin) OnEvent(event string, handler func(any)) {
	p.eventMutex.Lock()
	defer p.eventMutex.Unlock()
	p.eventHandlers[event] = append(p.eventHandlers[event], handler)
}

// RegisterCommands registers multiple commands at once.

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

	// Exit the process gracefully
	log.Println("Plugin shutting down gracefully")
	time.Sleep(100 * time.Millisecond) // Give logs time to flush
	os.Exit(0)                         // Clean exit instead of panic
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

		// Intercept structured envelopes (commands, storage, etc.)
		if messageType == websocket.TextMessage {
			// 1) storage response
			var response StorageResponse
			if err := json.Unmarshal(data, &response); err == nil && response.ID != "" {
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

			// 1b) host response
			var hResp HostResponse
			if err := json.Unmarshal(data, &hResp); err == nil && hResp.ID != "" && hResp.Type == "host/response" {
				p.requestsMutex.RLock()
				hch, exists := p.hostPending[hResp.ID]
				p.requestsMutex.RUnlock()
				if exists {
					select {
					case hch <- hResp:
					case <-time.After(time.Second):
						log.Printf("Warning: Host response channel timeout for request %s", hResp.ID)
					}
					continue
				}
			}

			// 2) event envelope (loopback or broadcast)
			var evEnv struct {
				Type  string `json:"type"`
				Event string `json:"event"`
				Data  any    `json:"data"`
			}
			if err := json.Unmarshal(data, &evEnv); err == nil && evEnv.Type == "event" && evEnv.Event != "" {
				p.eventMutex.RLock()
				handlers := p.eventHandlers[evEnv.Event]
				p.eventMutex.RUnlock()
				for _, h := range handlers {
					go h(evEnv.Data)
				}
				// continue to next frame (no further processing required)
				continue
			}

			// 3) command execute envelope
			var env map[string]any
			if err := json.Unmarshal(data, &env); err == nil {
				if t, ok := env["type"].(string); ok && t == "command/execute" {
					id, _ := env["id"].(string)
					cmdID, _ := env["command"].(string)
					rawArgs, _ := env["args"].([]any)

					// Find handler
					p.commandMutex.RLock()
					h, exists := p.commandHandlers[cmdID]
					p.commandMutex.RUnlock()

					var result any
					var errStr string
					success := true
					if !exists {
						success = false
						errStr = "unknown command"
					} else {
						// Execute handler with a background context
						ctx := context.Background()
						res, err := h(ctx, rawArgs)
						if err != nil {
							success = false
							errStr = err.Error()
						} else {
							result = res
						}
					}

					reply := map[string]any{
						"type":    "command/result",
						"id":      id,
						"success": success,
					}
					if errStr != "" {
						reply["error"] = errStr
					}
					if result != nil {
						reply["result"] = result
					}
					if b, err := json.Marshal(reply); err == nil {
						_ = p.conn.WriteMessage(websocket.TextMessage, b)
					}
					continue
				}
			}
		}

		// Execute the plugin-defined handler for the message.
		if handler != nil {
			handler(messageType, data)
		}
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
