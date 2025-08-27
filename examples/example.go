package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	sdk "github.com/PortableSheep/delve-sdk"
)

// ExamplePlugin demonstrates how to create a plugin with heartbeat and state management
type ExamplePlugin struct {
	plugin     *sdk.Plugin
	state      map[string]interface{}
	stateMutex sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	stateFile  string
}

// NewExamplePlugin creates a new example plugin
func NewExamplePlugin() *ExamplePlugin {
	ctx, cancel := context.WithCancel(context.Background())
	return &ExamplePlugin{
		state:     make(map[string]interface{}),
		ctx:       ctx,
		cancel:    cancel,
		stateFile: "example-plugin-state.json",
	}
}

// SaveState implements the StateManager interface
func (ep *ExamplePlugin) SaveState() error {
	ep.stateMutex.RLock()
	defer ep.stateMutex.RUnlock()

	log.Println("Saving plugin state...")

	// Add timestamp to state
	ep.state["lastSaved"] = time.Now()
	ep.state["version"] = "1.0.0"

	data, err := json.MarshalIndent(ep.state, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(ep.stateFile, data, 0644); err != nil {
		return err
	}

	log.Printf("State saved to %s", ep.stateFile)
	return nil
}

// GetState implements the StateManager interface
func (ep *ExamplePlugin) GetState() interface{} {
	ep.stateMutex.RLock()
	defer ep.stateMutex.RUnlock()

	// Return a copy of the state
	stateCopy := make(map[string]interface{})
	for k, v := range ep.state {
		stateCopy[k] = v
	}
	return stateCopy
}

// LoadState loads previously saved state
func (ep *ExamplePlugin) LoadState() error {
	if _, err := os.Stat(ep.stateFile); os.IsNotExist(err) {
		log.Println("No previous state file found, starting fresh")
		return nil
	}

	data, err := os.ReadFile(ep.stateFile)
	if err != nil {
		return err
	}

	ep.stateMutex.Lock()
	defer ep.stateMutex.Unlock()

	if err := json.Unmarshal(data, &ep.state); err != nil {
		return err
	}

	log.Printf("Loaded previous state from %s", ep.stateFile)
	return nil
}

// UpdateState safely updates plugin state
func (ep *ExamplePlugin) UpdateState(key string, value interface{}) {
	ep.stateMutex.Lock()
	defer ep.stateMutex.Unlock()
	ep.state[key] = value
}

// GetStateValue safely gets a value from plugin state
func (ep *ExamplePlugin) GetStateValue(key string) interface{} {
	ep.stateMutex.RLock()
	defer ep.stateMutex.RUnlock()
	return ep.state[key]
}

// handleMessage processes messages from the Delve host
func (ep *ExamplePlugin) handleMessage(messageType int, data []byte) {
	// Handle different message types here
	log.Printf("Received message type %d: %s", messageType, string(data))

	// Example: Update state based on received messages
	ep.UpdateState("lastMessageTime", time.Now())
	ep.UpdateState("messageCount", ep.GetStateValue("messageCount").(int)+1)
}

// run starts the main plugin loop
func (ep *ExamplePlugin) run() error {
	// Load any previous state
	if err := ep.LoadState(); err != nil {
		log.Printf("Failed to load state: %v", err)
	}

	// Initialize state if needed
	if ep.GetStateValue("messageCount") == nil {
		ep.UpdateState("messageCount", 0)
	}

	// Create plugin registration info
	pluginInfo := &sdk.RegisterRequest{
		Name:             "example-plugin",
		Description:      "Example plugin with heartbeat and state management",
		CustomElementTag: "example-plugin",
		UiComponentPath:  "index.js",
	}

	// Connect to Delve host
	plugin, err := sdk.Start(pluginInfo)
	if err != nil {
		return err
	}
	ep.plugin = plugin

	log.Println("Plugin connected to host successfully")

	// Set up state management
	plugin.SetStateManager(ep)

	// Start heartbeat with 30 second interval and 60 second timeout
	plugin.StartHeartbeat(30*time.Second, 60*time.Second)
	log.Println("Heartbeat started (30s interval, 60s timeout)")

	// Set up graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, shutting down gracefully...")
		ep.cancel()
	}()

	// Start example background work
	go ep.backgroundWork()

	// Listen for messages from host (this blocks until connection closes)
	plugin.Listen(ep.handleMessage)

	return nil
}

// backgroundWork simulates plugin work
func (ep *ExamplePlugin) backgroundWork() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ep.ctx.Done():
			log.Println("Background work stopping...")
			return
		case <-ticker.C:
			// Simulate some work
			ep.UpdateState("lastWorkTime", time.Now())
			ep.UpdateState("workCount", ep.GetStateValue("workCount").(int)+1)
			log.Println("Background work completed")
		}
	}
}

func main() {
	log.Println("Starting Example Plugin...")

	// Create plugin instance
	plugin := NewExamplePlugin()

	// Initialize work count if needed
	if plugin.GetStateValue("workCount") == nil {
		plugin.UpdateState("workCount", 0)
	}

	// Run the plugin
	if err := plugin.run(); err != nil {
		log.Fatalf("Plugin failed: %v", err)
	}

	log.Println("Example Plugin shutting down")
}
