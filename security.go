package sdk

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// Simplified security model focused on transparency and user consent

// PermissionType represents different types of permissions for transparency
type PermissionType string

const (
	PermissionFileSystem    PermissionType = "filesystem"
	PermissionNetwork       PermissionType = "network"
	PermissionStorage       PermissionType = "storage"
	PermissionClipboard     PermissionType = "clipboard"
	PermissionNotifications PermissionType = "notifications"
)

// Permission represents a permission declaration (for transparency, not enforcement)
type Permission struct {
	Type        PermissionType `json:"type"`
	Description string         `json:"description"`
	Reason      string         `json:"reason"` // Why the plugin needs this permission
}

// PluginSecurity provides simple security utilities
type PluginSecurity struct {
	PluginName  string
	Permissions []Permission
	Violations  []SecurityViolation
}

// SecurityViolation represents a security issue (for logging/monitoring)
type SecurityViolation struct {
	PluginName string    `json:"pluginName"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
	Severity   string    `json:"severity"` // "low", "medium", "high"
}

// NewPluginSecurity creates a new security context for a plugin
func NewPluginSecurity(pluginName string) *PluginSecurity {
	return &PluginSecurity{
		PluginName:  pluginName,
		Permissions: make([]Permission, 0),
		Violations:  make([]SecurityViolation, 0),
	}
}

// DeclarePermission adds a permission declaration (for transparency)
func (ps *PluginSecurity) DeclarePermission(ptype PermissionType, description, reason string) {
	permission := Permission{
		Type:        ptype,
		Description: description,
		Reason:      reason,
	}
	ps.Permissions = append(ps.Permissions, permission)
}

// LogViolation records a security concern
func (ps *PluginSecurity) LogViolation(message, severity string) {
	violation := SecurityViolation{
		PluginName: ps.PluginName,
		Message:    message,
		Timestamp:  time.Now(),
		Severity:   severity,
	}
	ps.Violations = append(ps.Violations, violation)
}

// GetPermissions returns declared permissions
func (ps *PluginSecurity) GetPermissions() []Permission {
	return ps.Permissions
}

// GetViolations returns logged violations
func (ps *PluginSecurity) GetViolations() []SecurityViolation {
	return ps.Violations
}

// ValidateInput provides basic input sanitization
func ValidateInput(input string) error {
	if len(input) > 10000 {
		return fmt.Errorf("input too long: %d characters", len(input))
	}

	// Check for common injection patterns
	dangerous := []string{"../", "<script", "javascript:", "eval(", "exec("}
	for _, pattern := range dangerous {
		if contains(input, pattern) {
			return fmt.Errorf("potentially dangerous input detected: %s", pattern)
		}
	}

	return nil
}

// GetSystemInfo returns basic system information for compatibility
func GetSystemInfo() map[string]interface{} {
	hostname, _ := os.Hostname()
	user := getCurrentUser()

	return map[string]interface{}{
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"go_version": runtime.Version(),
		"hostname":   hostname,
		"user":       user,
	}
}

// Helper functions
func getCurrentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "unknown"
}
