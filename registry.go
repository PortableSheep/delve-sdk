package sdk

import (
	"encoding/json"
	"fmt"
	"time"
)

// PluginTemplate represents a basic plugin template
type PluginTemplate struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Files       map[string]string      `json:"files"`
	Variables   map[string]interface{} `json:"variables"`
}

// RegistryPlugin represents a plugin in the registry
type RegistryPlugin struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// EnhancedPluginInfo represents extended plugin information
type EnhancedPluginInfo struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Tags        []string  `json:"tags"`
	Category    string    `json:"category"`
	Downloads   int       `json:"downloads"`
	Rating      float64   `json:"rating"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// EnhancedRegistry provides simple registry operations
type EnhancedRegistry struct {
	plugins map[string]*EnhancedPluginInfo
}

// NewEnhancedRegistry creates a new enhanced registry
func NewEnhancedRegistry() *EnhancedRegistry {
	return &EnhancedRegistry{
		plugins: make(map[string]*EnhancedPluginInfo),
	}
}

// AddPlugin adds a plugin to the registry
func (r *EnhancedRegistry) AddPlugin(plugin *EnhancedPluginInfo) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	r.plugins[plugin.Name] = plugin
	return nil
}

// GetPlugin retrieves a plugin by name
func (r *EnhancedRegistry) GetPlugin(name string) (*EnhancedPluginInfo, error) {
	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	return plugin, nil
}

// ListPlugins returns all plugins in the registry
func (r *EnhancedRegistry) ListPlugins() []*EnhancedPluginInfo {
	plugins := make([]*EnhancedPluginInfo, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// SearchPlugins searches for plugins by query
func (r *EnhancedRegistry) SearchPlugins(query string) []*EnhancedPluginInfo {
	var results []*EnhancedPluginInfo

	for _, plugin := range r.plugins {
		if contains(plugin.Name, query) || contains(plugin.Description, query) {
			results = append(results, plugin)
		}
	}

	return results
}

// ExportToJSON exports the registry to JSON
func (r *EnhancedRegistry) ExportToJSON() ([]byte, error) {
	return json.MarshalIndent(r.plugins, "", "  ")
}

// ImportFromJSON imports plugins from JSON
func (r *EnhancedRegistry) ImportFromJSON(data []byte) error {
	var plugins map[string]*EnhancedPluginInfo
	if err := json.Unmarshal(data, &plugins); err != nil {
		return err
	}

	r.plugins = plugins
	return nil
}

// Helper function for string searching
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
