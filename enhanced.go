package sdk

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// Enhanced Plugin SDK features for improved extensibility

// ContextAPI provides when-clause evaluation and context management
type ContextAPI struct {
	plugin *Plugin
}

// SetContext sets a context variable that can be used in when clauses
func (c *ContextAPI) SetContext(key string, value interface{}) error {
	if c.plugin == nil || c.plugin.conn == nil {
		return fmt.Errorf("plugin connection not initialised")
	}

	_, err := c.plugin.sendHostRequest("setContext", map[string]any{
		"key":   key,
		"value": value,
	})
	return err
}

// GetContext retrieves all current context variables
func (c *ContextAPI) GetContext() (map[string]interface{}, error) {
	if c.plugin == nil || c.plugin.conn == nil {
		return nil, fmt.Errorf("plugin connection not initialised")
	}

	resp, err := c.plugin.sendHostRequest("getContext", nil)
	if err != nil {
		return nil, err
	}

	if m, ok := resp.Result.(map[string]interface{}); ok {
		return m, nil
	}
	return make(map[string]interface{}), nil
}

// SettingsAPI provides plugin configuration management
type SettingsAPI struct {
	plugin *Plugin
}

// GetSetting retrieves a plugin setting value
func (s *SettingsAPI) GetSetting(key string) (interface{}, error) {
	if s.plugin == nil || s.plugin.conn == nil {
		return nil, fmt.Errorf("plugin connection not initialised")
	}

	resp, err := s.plugin.sendHostRequest("getSetting", map[string]any{
		"key": key,
	})
	if err != nil {
		return nil, err
	}
	return resp.Result, nil
}

// SetSetting updates a plugin setting value
func (s *SettingsAPI) SetSetting(key string, value interface{}) error {
	if s.plugin == nil || s.plugin.conn == nil {
		return fmt.Errorf("plugin connection not initialised")
	}

	_, err := s.plugin.sendHostRequest("setSetting", map[string]any{
		"key":   key,
		"value": value,
	})
	return err
}

// OnSettingChanged registers a callback for setting changes
func (s *SettingsAPI) OnSettingChanged(key string, handler func(interface{})) {
	if s.plugin == nil {
		return
	}

	s.plugin.OnEvent("setting:changed:"+key, func(data any) {
		handler(data)
	})
}

// ViewAPI provides view contribution management
type ViewAPI struct {
	plugin *Plugin
}

// ViewProvider interface for tree view data providers
type ViewProvider interface {
	GetTreeData() ([]TreeItem, error)
	GetChildren(element *TreeItem) ([]TreeItem, error)
	OnDidChangeTreeData(handler func())
}

// TreeItem represents an item in a tree view
type TreeItem struct {
	ID          string      `json:"id"`
	Label       string      `json:"label"`
	Icon        string      `json:"icon,omitempty"`
	Command     *CommandRef `json:"command,omitempty"`
	Children    []TreeItem  `json:"children,omitempty"`
	Collapsible bool        `json:"collapsible"`
	Tooltip     string      `json:"tooltip,omitempty"`
	Context     string      `json:"contextValue,omitempty"`
}

// RegisterViewProvider registers a tree data provider for a view
func (v *ViewAPI) RegisterViewProvider(viewId string, provider ViewProvider) error {
	if v.plugin == nil || v.plugin.conn == nil {
		return fmt.Errorf("plugin connection not initialised")
	}

	// Store the provider for handling data requests
	// This would require additional infrastructure in the plugin SDK
	return fmt.Errorf("view providers not yet implemented")
}

// RefreshView triggers a refresh of a specific view
func (v *ViewAPI) RefreshView(viewId string) error {
	if v.plugin == nil || v.plugin.conn == nil {
		return fmt.Errorf("plugin connection not initialised")
	}

	return v.plugin.Emit("view:refresh", map[string]any{
		"viewId": viewId,
	})
}

// Enhanced plugin structure with new APIs
func (p *Plugin) Context() *ContextAPI {
	return &ContextAPI{plugin: p}
}

func (p *Plugin) Settings() *SettingsAPI {
	return &SettingsAPI{plugin: p}
}

func (p *Plugin) Views() *ViewAPI {
	return &ViewAPI{plugin: p}
}

// Enhanced registration with additional contribution points
type EnhancedRegisterRequest struct {
	Name             string                   `json:"name"`
	Description      string                   `json:"description"`
	UIComponentPath  string                   `json:"uiComponentPath"`
	CustomElementTag string                   `json:"customElementTag"`
	Contributions    *EnhancedUIContributions `json:"contributions,omitempty"`

	// Additional metadata
	Version    string   `json:"version,omitempty"`
	Author     string   `json:"author,omitempty"`
	Homepage   string   `json:"homepage,omitempty"`
	Keywords   []string `json:"keywords,omitempty"`
	License    string   `json:"license,omitempty"`
	Repository string   `json:"repository,omitempty"`
}

// Enhanced UI contributions with more extension points
type EnhancedUIContributions struct {
	Sidebar   []SidebarItem      `json:"sidebar,omitempty"`
	Footer    []FooterWidget     `json:"footer,omitempty"`
	Views     []ViewContrib      `json:"views,omitempty"`
	Menus     []MenuContrib      `json:"menus,omitempty"`
	StatusBar []StatusBarContrib `json:"statusBar,omitempty"`
}

type ViewContrib struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Container string `json:"container"` // sidebar, panel, explorer
	When      string `json:"when,omitempty"`
	Icon      string `json:"icon,omitempty"`
}

type MenuContrib struct {
	Location string `json:"location"` // commandPalette, explorer/context, etc.
	Command  string `json:"command"`
	When     string `json:"when,omitempty"`
	Group    string `json:"group,omitempty"`
	Title    string `json:"title,omitempty"`
}

type StatusBarContrib struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Command   string `json:"command,omitempty"`
	Tooltip   string `json:"tooltip,omitempty"`
	Color     string `json:"color,omitempty"`
	Priority  int    `json:"priority,omitempty"`
	Alignment string `json:"alignment,omitempty"` // left, right
}

// Enhanced plugin startup with additional features
func StartEnhanced(pluginInfo *EnhancedRegisterRequest) (*Plugin, error) {
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

	log.Printf("Plugin successfully sent enhanced registration request: %s", pluginInfo.Name)

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

// Plugin capability detection
func (p *Plugin) SupportsEnhancedFeatures() bool {
	// Check if host supports enhanced features
	// This would be determined during registration
	return false // For now, return false until host implements enhanced features
}

// Feature flags for progressive enhancement
type PluginCapabilities struct {
	Views         bool `json:"views"`
	Menus         bool `json:"menus"`
	StatusBar     bool `json:"statusBar"`
	Configuration bool `json:"configuration"`
	Context       bool `json:"context"`
	Themes        bool `json:"themes"`
	Commands      bool `json:"commands"`
}

func (p *Plugin) GetCapabilities() (*PluginCapabilities, error) {
	// Query host for supported capabilities
	resp, err := p.sendHostRequest("getCapabilities", nil)
	if err != nil {
		return nil, err
	}

	caps := &PluginCapabilities{}
	if resp.Result != nil {
		// Parse capabilities from response
		if m, ok := resp.Result.(map[string]interface{}); ok {
			caps.Views, _ = m["views"].(bool)
			caps.Menus, _ = m["menus"].(bool)
			caps.StatusBar, _ = m["statusBar"].(bool)
			caps.Configuration, _ = m["configuration"].(bool)
			caps.Context, _ = m["context"].(bool)
			caps.Themes, _ = m["themes"].(bool)
			caps.Commands, _ = m["commands"].(bool)
		}
	}

	return caps, nil
}
