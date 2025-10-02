package sdk

// PluginStartupStatus represents the lifecycle phase reported by the host for a plugin process.
type PluginStartupStatus string

const (
	PluginStatusUnknown      PluginStartupStatus = "unknown"
	PluginStatusStarting     PluginStartupStatus = "starting"
	PluginStatusRunning      PluginStartupStatus = "running"
	PluginStatusFailed       PluginStartupStatus = "failed"
	PluginStatusStopped      PluginStartupStatus = "stopped"
	PluginStatusReconnecting PluginStartupStatus = "reconnecting"
	PluginStatusUnhealthy    PluginStartupStatus = "unhealthy"
)

// CommandRef points to a plugin command for simple UI-triggered actions.
type CommandRef struct {
	ID   string        `json:"id"`
	Args []interface{} `json:"args,omitempty"`
}

// SidebarItem is a plugin-contributed sidebar navigation entry.
type SidebarItem struct {
	Plugin         string      `json:"plugin,omitempty"`
	ID             string      `json:"id,omitempty"`
	Title          string      `json:"title"`
	Icon           string      `json:"icon,omitempty"`
	ElementTag     string      `json:"elementTag,omitempty"`
	ComponentPath  string      `json:"componentPath,omitempty"`
	Order          int         `json:"order,omitempty"`
	OnClickCommand *CommandRef `json:"onClickCommand,omitempty"`
}

// FooterWidget is a plugin-contributed footer widget.
type FooterWidget struct {
	Plugin         string      `json:"plugin,omitempty"`
	ElementTag     string      `json:"elementTag,omitempty"`
	ComponentPath  string      `json:"componentPath,omitempty"`
	Order          int         `json:"order,omitempty"`
	Icon           string      `json:"icon,omitempty"`
	Title          string      `json:"title,omitempty"`
	BadgeCount     int         `json:"badgeCount,omitempty"`
	OnClickCommand *CommandRef `json:"onClickCommand,omitempty"`
}

// UIContributions lists optional UI extension points a plugin contributes.
type UIContributions struct {
	Sidebar []SidebarItem  `json:"sidebar,omitempty"`
	Footer  []FooterWidget `json:"footer,omitempty"`
}

// PluginRegistration holds registration data shared between host and plugins.
type PluginRegistration struct {
	Name             string              `json:"name"`
	Description      string              `json:"description"`
	UIComponentPath  string              `json:"uiComponentPath"`
	CustomElementTag string              `json:"customElementTag"`
	Version          string              `json:"version,omitempty"`
	Icon             string              `json:"icon,omitempty"`
	StartupStatus    PluginStartupStatus `json:"startupStatus,omitempty"`
	Type             string              `json:"type,omitempty"`
	HasUI            bool                `json:"hasUI"`
	Permissions      []string            `json:"permissions,omitempty"`
	Contributions    *UIContributions    `json:"contributions,omitempty"`
}
