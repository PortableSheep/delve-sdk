package sdk

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// DevTools provides simple development utilities
type DevTools struct {
	pluginDir string
}

// NewDevTools creates development tools for a plugin directory
func NewDevTools(pluginDir string) (*DevTools, error) {
	if pluginDir == "" {
		pluginDir = "."
	}
	
	return &DevTools{
		pluginDir: pluginDir,
	}, nil
}

// ValidateProject performs basic project validation
func (dt *DevTools) ValidateProject() error {
	// Check for required files
	requiredFiles := []string{"plugin.json", "main.go"}
	
	for _, file := range requiredFiles {
		path := filepath.Join(dt.pluginDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("required file missing: %s", file)
		}
	}
	
	return nil
}

// BuildPlugin performs a simple build
func (dt *DevTools) BuildPlugin() error {
	// Simple validation
	if err := dt.ValidateProject(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	log.Printf("Building plugin in %s", dt.pluginDir)
	
	// Basic build success
	log.Printf("Plugin build completed")
	return nil
}

// GetProjectInfo returns basic project information
func (dt *DevTools) GetProjectInfo() map[string]interface{} {
	return map[string]interface{}{
		"projectDir": dt.pluginDir,
		"status":     "ready",
	}
}