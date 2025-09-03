# Delve Plugin SDK

A Go SDK for creating plugins that connect to the Delve host application with built-in heartbeat monitoring, state management, and graceful shutdown capabilities.

## Features

- **WebSocket Connection**: Reliable connection to Delve host
- **Heartbeat Monitoring**: Automatic health checks with graceful shutdown
- **State Management**: Plugin state persistence and recovery
- **Storage API**: Access to plugin-specific storage
- **Graceful Shutdown**: Clean shutdown with state saving
- **Connection Recovery**: Automatic reconnection handling

## Quick Start

### Basic Plugin

```go
package main

import (
    "log"
    sdk "github.com/portablesheep/plugin-sdk"
)

func main() {
    // Define plugin information
    pluginInfo := &sdk.RegisterRequest{
        Name:             "my-plugin",
        Description:      "My awesome plugin",
        Version:          "1.0.0",
        CustomElementTag: "my-plugin",
        UIComponentPath:  "index.js",
        Icon:             "🚀",
    }

    // Connect to host
    plugin, err := sdk.Start(pluginInfo)
    if err != nil {
        log.Fatal(err)
    }
    defer plugin.Close()

    // Handle messages
    plugin.Listen(func(messageType int, data []byte) {
        log.Printf("Received: %s", string(data))
    })
}
```

### Plugin with Heartbeat & State Management

```go
package main

import (
    "encoding/json"
    "log"
    "os"
    "time"
    sdk "github.com/portablesheep/plugin-sdk"
)

type MyPlugin struct {
    state map[string]interface{}
}

// Implement StateManager interface
func (p *MyPlugin) SaveState() error {
    data, _ := json.Marshal(p.state)
    return os.WriteFile("plugin-state.json", data, 0644)
}

func (p *MyPlugin) GetState() interface{} {
    return p.state
}

func main() {
    myPlugin := &MyPlugin{
        state: make(map[string]interface{}),
    }

    pluginInfo := &sdk.RegisterRequest{
        Name:             "stateful-plugin",
        Description:      "Plugin with state management",
        Version:          "1.0.0",
        CustomElementTag: "stateful-plugin",
        UIComponentPath:  "index.js",
    }

    plugin, err := sdk.Start(pluginInfo)
    if err != nil {
        log.Fatal(err)
    }
    defer plugin.Close()

    // Enable state management
    plugin.SetStateManager(myPlugin)

    // Start heartbeat (30s interval, 60s timeout)
    plugin.StartHeartbeat(30*time.Second, 60*time.Second)

    // Listen for messages
    plugin.Listen(func(messageType int, data []byte) {
        myPlugin.state["lastMessage"] = string(data)
        log.Printf("Message saved to state: %s", string(data))
    })
}
```

## API Reference

### Plugin Registration

#### `RegisterRequest`
```go
type RegisterRequest struct {
    Name             string `json:"name"`
    Description      string `json:"description"`
    Version          string `json:"version"`
    CustomElementTag string `json:"customElementTag"`
    UIComponentPath  string `json:"uiComponentPath"`
    Icon             string `json:"icon"`
}
```

- **Name**: Unique plugin identifier
- **Description**: Human-readable description
- **Version**: Plugin version (semver recommended)
- **CustomElementTag**: HTML custom element tag for frontend
- **UIComponentPath**: Path to frontend JavaScript file
- **Icon**: Emoji or icon for UI display

### Plugin Metadata (plugin.json)

For plugins distributed through the registry, create a `plugin.json` file with additional metadata:

```json
{
  "info": {
    "id": "my-plugin",
    "name": "My Plugin",
    "version": "v1.0.0",
    "description": "A comprehensive plugin description",
    "author": "Your Name",
    "license": "MIT",
    "homepage": "https://github.com/user/plugin-repo",
    "repository": "https://github.com/user/plugin-repo",
    "icon": "🚀",
    "screenshots": [
      "https://raw.githubusercontent.com/user/plugin-repo/main/screenshots/overview.png",
      "https://raw.githubusercontent.com/user/plugin-repo/main/screenshots/settings.png",
      "https://raw.githubusercontent.com/user/plugin-repo/main/screenshots/features.png"
    ],
    "tags": ["productivity", "development", "ui"],
    "category": "development-tools",
    "min_delve_version": "v0.1.0"
  },
    "runtime": {
        "executable": "my-plugin"
    },
    "frontend": {
        "entry": "frontend/component.js"
    }
}
```

#### Screenshots

Screenshots help users understand your plugin's functionality:

**Best Practices:**
- Use PNG format for best quality
- Minimum resolution: 1200x800 pixels
- Maximum file size: 500KB per image
- Show real data when possible (avoid empty states)
- Include 2-4 screenshots showing key features
- Use consistent browser/window sizing

**Screenshot URLs:**
- Host screenshots in your plugin repository
- Use GitHub raw URLs for reliability
- Example: `https://raw.githubusercontent.com/user/repo/main/screenshots/image.png`

**What to Screenshot:**
1. **Main interface** - Primary plugin view
2. **Key features** - Important functionality in action
3. **Settings/Configuration** - Plugin configuration options
4. **Different states** - Various plugin modes or views

Screenshots are displayed in:
- Plugin Manager (Delve application)
- Plugin Store interface
- Registry website
- Plugin documentation

### Core Methods

#### `Start(pluginInfo *RegisterRequest) (*Plugin, error)`
Establishes connection to Delve host and registers the plugin.

#### `Plugin.Listen(handler func(messageType int, data []byte))`
Starts message processing loop. **This method blocks** until connection closes.

#### `Plugin.Close()`
Gracefully closes the connection to the host.

### Heartbeat System

#### `Plugin.SetStateManager(sm StateManager)`
Sets a state manager for graceful shutdown with state saving.

#### `Plugin.StartHeartbeat(interval, timeout time.Duration)`
Starts heartbeat monitoring:
- **interval**: How often to send heartbeats to host
- **timeout**: How long to wait for host response before shutdown

**Recommended settings:**
- Interval: 30 seconds
- Timeout: 60 seconds

### State Management

Implement the `StateManager` interface for automatic state persistence:

```go
type StateManager interface {
    SaveState() error        // Save current plugin state
    GetState() interface{}   // Return current state
}
```

**When state is saved:**
- Before graceful shutdown
- When heartbeat timeout occurs
- When host connection is lost

### Storage API

#### `Plugin.SetItem(key, value string) error`
Store a key-value pair in plugin storage.

#### `Plugin.GetItem(key string) (string, error)`
Retrieve a value from plugin storage.

#### `Plugin.DeleteItem(key string) error`
Delete a key from plugin storage.

#### `Plugin.ListItems() ([]string, error)`
List all keys in plugin storage.

#### `Plugin.ClearStorage() error`
Clear all plugin storage.

## Heartbeat Behavior

The heartbeat system provides robust connection monitoring:

### Normal Operation
1. Plugin sends heartbeat every `interval` seconds
2. Host responds with heartbeat acknowledgment
3. Plugin tracks last successful heartbeat response

### Failure Detection
1. If no heartbeat response received within `timeout`
2. Plugin logs timeout warning
3. Calls `StateManager.SaveState()` if configured
4. Plugin shuts down gracefully

### Benefits
- **Prevents zombie processes** when host crashes
- **Saves plugin state** before unexpected shutdown
- **Reduces resource usage** by cleaning up orphaned plugins
- **Improves system stability**

## Error Handling

### Connection Errors
```go
plugin, err := sdk.Start(pluginInfo)
if err != nil {
    log.Printf("Failed to connect: %v", err)
    // Handle connection failure
}
```

### Heartbeat Timeout
The plugin will automatically:
1. Log timeout warning
2. Save state (if StateManager configured)
3. Exit gracefully

### Storage Errors
```go
err := plugin.SetItem("key", "value")
if err != nil {
    log.Printf("Storage error: %v", err)
    // Handle storage failure
}
```

## Best Practices

### 1. Always Use Heartbeat
```go
// Recommended for all production plugins
plugin.StartHeartbeat(30*time.Second, 60*time.Second)
```

### 2. Implement State Management
```go
type MyPlugin struct {
    importantData map[string]interface{}
    configPath    string
}

func (p *MyPlugin) SaveState() error {
    // Save critical state that should survive restarts
    return saveToFile(p.configPath, p.importantData)
}
```

### 3. Handle Graceful Shutdown
```go
import (
    "os"
    "os/signal"
    "syscall"
)

// Listen for OS signals
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-sigChan
    log.Println("Shutting down gracefully...")
    plugin.Close()
    os.Exit(0)
}()
```

### 4. Use Context for Background Work
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Do background work
        }
    }
}()
```

### 5. Log Important Events
```go
plugin.Listen(func(messageType int, data []byte) {
    log.Printf("Received message type %d: %s", messageType, string(data))
    // Process message
})
```

## Plugin Lifecycle

1. **Startup**: Plugin connects and registers with host
2. **Running**: Plugin processes messages and performs work
3. **Heartbeat**: Regular health checks with host
4. **Shutdown**: Graceful cleanup when connection lost or timeout
5. **State Save**: Critical data persisted before exit

## Publishing to Registry

To make your plugin available through the Delve Plugin Registry:

1. **Create plugin.json** with complete metadata including screenshots
2. **Add screenshots** to a `screenshots/` directory in your repository
3. **Submit to registry** via pull request or registry submission process
4. **Include releases** with compiled binaries for supported platforms

**Directory Structure:**
```
my-plugin/
├── plugin.json           # Plugin metadata
├── main.go              # Plugin source code
├── frontend/
│   └── component.js     # Frontend component
├── screenshots/         # Plugin screenshots
│   ├── overview.png
│   ├── settings.png
│   └── features.png
└── releases/           # Compiled binaries
    └── v1.0.0/
        ├── my-plugin-linux-amd64
        ├── my-plugin-darwin-amd64
        └── my-plugin-windows-amd64.exe
```

## Frontend Integration

Your plugin's frontend component should match the `CustomElementTag`:

```javascript
// If CustomElementTag is "my-plugin"
class MyPlugin extends HTMLElement {
    connectedCallback() {
        this.innerHTML = '<h1>My Plugin UI</h1>';
    }
}

customElements.define('my-plugin', MyPlugin);
```

## Troubleshooting

### Plugin Not Appearing in Sidebar
- Check host-config.yml has plugin enabled
- Verify plugin executable exists and is executable
- Check plugin registration logs

### Connection Issues
- Ensure host is running on correct WebSocket port
- Check firewall settings
- Verify --ws-port flag is passed correctly

### Heartbeat Timeouts
- Check host system resources
- Adjust timeout values for slower systems
- Monitor network connectivity

### State Not Saving
- Verify StateManager interface implementation
- Check file permissions for state files
- Review error logs during shutdown

## Examples

See `example.go` for a complete plugin implementation with:
- Heartbeat monitoring
- State management
- Background work
- Graceful shutdown
- Signal handling

## Requirements

- Go 1.19 or later
- Delve host application
- WebSocket support
