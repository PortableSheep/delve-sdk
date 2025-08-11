package sdk

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

// Plugin represents a connection to the host application.
type Plugin struct {
	conn *websocket.Conn
}

// RegisterRequest defines the structure for the registration request.
// The JSON tags MUST match what the host application expects.
type RegisterRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	UiComponentPath  string `json:"uiComponentPath"`
	CustomElementTag string `json:"customElementTag"`
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

	return &Plugin{conn: conn}, nil
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
		// Execute the plugin-defined handler for the message.
		handler(messageType, data)
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
