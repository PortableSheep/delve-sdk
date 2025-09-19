# Enhanced Delve Plugin SDK v2.0

The Enhanced Delve Plugin SDK provides comprehensive tools and features for building, securing, and managing plugins for the Delve application platform.

## 🚀 Features

### Core Enhancements
- **Enhanced Plugin Registry** - Advanced search, filtering, and metadata
- **Security & Sandboxing** - Comprehensive security policies and isolated execution
- **Schema Validation** - Type-safe plugin configuration validation  
- **Plugin Templates** - Code generation with customizable templates
- **Development Tools** - Build, test, and package plugins easily
- **Health Monitoring** - Real-time plugin health and performance metrics

### Architecture Improvements
- **Builder Pattern** - Fluent interface for plugin configuration
- **Middleware Support** - Request/response transformation pipeline
- **Lifecycle Hooks** - Custom hooks for plugin events
- **Context Awareness** - Timeout and cancellation support
- **Resource Monitoring** - Memory and CPU usage tracking

## 📦 Installation

```bash
go get github.com/PortableSheep/delve-sdk@v2.0.0
```

## 🔧 Quick Start

### Basic Plugin Creation

```go
package main

import (
    "context"
    "log"
    
    sdk "github.com/PortableSheep/delve-sdk"
)

func main() {
    // Create enhanced SDK
    enhancedSDK := sdk.NewEnhancedSDK(sdk.SDKConfig{
        SecurityLevel: sdk.SecurityLevelMedium,
        DevMode:       true,
        CacheEnabled:  true,
    })

    // Build plugin with enhanced features
    plugin, err := enhancedSDK.NewPluginBuilder().
        WithName("my-plugin").
        WithDescription("My enhanced plugin").
        WithVersion("1.0.0").
        WithSecurity(sdk.SecurityPolicy{
            Level: sdk.SecurityLevelMedium,
            AllowedPermissions: []sdk.PermissionType{
                sdk.PermissionUI,
                sdk.PermissionStorage,
            },
        }).
        WithValidation(&sdk.PluginSchema{
            Version: "1.0",
            Name:    "my-plugin-schema",
            Rules: []sdk.ValidationRule{
                {Field: "config.apiKey", Type: "string", Required: true},
            },
        }).
        WithMetrics(true).
        BuildEnhanced()

    if err != nil {
        log.Fatal(err)
    }

    // Initialize and start
    ctx := context.Background()
    if err := plugin.Initialize(ctx); err != nil {
        log.Fatal(err)
    }

    if err := plugin.Start(ctx); err != nil {
        log.Fatal(err)
    }

    log.Println("Enhanced plugin started successfully!")
}
```

### Plugin Generation from Template

```go
// Generate a new Go utility plugin
options := sdk.GenerationOptions{
    Name:        "weather-widget",
    Author:      "John Doe",
    Description: "A weather widget plugin",
    License:     "MIT",
    Version:     "1.0.0",
    OutputDir:   "./plugins",
    Template:    "go-utility",
    Variables: map[string]interface{}{
        "APIEndpoint": "https://api.weather.com",
    },
}

err := enhancedSDK.CreatePlugin(options)
if err != nil {
    log.Fatal(err)
}
```

## 🛡️ Security & Sandboxing

### Security Policies

```go
policy := sdk.SecurityPolicy{
    Level: sdk.SecurityLevelHigh,
    AllowedPermissions: []sdk.PermissionType{
        sdk.PermissionStorage,
        sdk.PermissionNetwork,
    },
    AllowedHosts: []string{"api.example.com"},
    MaxMemory:    50 * 1024 * 1024, // 50MB
    MaxCPU:      25.0,              // 25%
    Timeout:     30 * time.Second,
}

sandbox, err := enhancedSDK.CreateSandbox("my-plugin", &policy)
if err != nil {
    log.Fatal(err)
}

// Validate file access
err = sandbox.CheckFileAccess("/home/user/data.txt", "read")
if err != nil {
    log.Printf("File access denied: %v", err)
}

// Validate network access
err = sandbox.CheckNetworkAccess("api.example.com", 443)
if err != nil {
    log.Printf("Network access denied: %v", err)
}
```

## 🔍 Enhanced Registry

### Advanced Plugin Search

```go
filter := sdk.RegistrySearchFilter{
    Category:     "utility",
    Tags:         []string{"productivity", "tools"},
    Author:       "john-doe",
    UpdatedAfter: &time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    SortBy:       "rating",
    SortOrder:    "desc",
    Limit:        20,
}

results, err := enhancedSDK.SearchPlugins(filter)
if err != nil {
    log.Fatal(err)
}

for _, plugin := range results.Plugins {
    fmt.Printf("Plugin: %s (Rating: %.1f, Downloads: %d)\\n",
        plugin.Name, plugin.Metrics.Rating, plugin.Metrics.Downloads)
}
```

### Plugin Metadata

```go
pluginInfo := &sdk.EnhancedPluginInfo{
    Name:        "awesome-plugin",
    Version:     "2.1.0",
    Description: "An awesome plugin with enhanced features",
    Author:      "jane-smith",
    Category:    "productivity",
    Tags:        []string{"automation", "workflow"},
    Dependencies: []sdk.PluginDependency{
        {Name: "base-utils", Version: "^1.0.0", Required: true},
    },
    Compatibility: sdk.PluginCompatibility{
        MinHostVersion: "1.5.0",
        Platforms:      []string{"linux", "darwin", "windows"},
        Architectures:  []string{"amd64", "arm64"},
    },
    Security: sdk.SecurityInfo{
        SecurityRating: "high",
        Permissions:    []string{"storage", "network"},
        Sandboxed:      true,
    },
}

err := enhancedSDK.Registry().AddPlugin(pluginInfo)
```

## ✅ Schema Validation

### Custom Validation Rules

```go
schema := &sdk.PluginSchema{
    Version:     "1.0",
    Name:        "api-client-schema",
    Description: "Schema for API client plugins",
    Rules: []sdk.ValidationRule{
        {
            Field:    "apiKey",
            Type:     "string",
            Required: true,
            MinLength: &[]int{10}[0],
            PatternStr: "^[a-zA-Z0-9]+$",
        },
        {
            Field: "timeout",
            Type:  "number",
            Min:   &[]float64{1}[0],
            Max:   &[]float64{300}[0],
        },
        {
            Field: "retries",
            Type:  "integer",
            CustomValidator: func(val interface{}) error {
                if retries, ok := val.(int); ok && retries < 0 {
                    return fmt.Errorf("retries must be non-negative")
                }
                return nil
            },
        },
    },
}

// Register schema
enhancedSDK.Validator().RegisterSchema("api-client", schema)

// Validate plugin config
config := map[string]interface{}{
    "apiKey":  "abc123xyz789",
    "timeout": 30.0,
    "retries": 3,
}

errors := enhancedSDK.ValidatePlugin("api-client", config)
if len(errors) > 0 {
    for _, err := range errors {
        fmt.Printf("Validation error: %s - %s\\n", err.Field, err.Message)
    }
}
```

## 🛠️ Development Tools

### Project Initialization

```bash
# Using the CLI (would need to be built separately)
delve-sdk init my-plugin --template=go-utility --author="John Doe"
```

Or programmatically:

```go
devTools, err := enhancedSDK.DevTools("./my-plugin")
if err != nil {
    log.Fatal(err)
}

// Initialize project
err = devTools.InitProject(sdk.GenerationOptions{
    Name:        "my-plugin",
    Author:      "John Doe",
    Description: "My awesome plugin",
    Template:    "go-utility",
    License:     "MIT",
    Version:     "1.0.0",
})
```

### Building and Testing

```go
// Build plugin
buildResult, err := devTools.Build("linux")
if err != nil {
    log.Fatal(err)
}

if buildResult.Success {
    fmt.Printf("Build successful in %v\\n", buildResult.Duration)
    fmt.Printf("Output files: %v\\n", buildResult.OutputFiles)
} else {
    fmt.Printf("Build failed: %v\\n", buildResult.Errors)
}

// Run tests
testResult, err := devTools.Test()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Tests: %d passed, %d failed (%.1f%% coverage)\\n",
    testResult.Passed, testResult.Failed, testResult.Coverage)

// Package for distribution
packageResult, err := devTools.Package()
if err != nil {
    log.Fatal(err)
}

if packageResult.Success {
    fmt.Printf("Package created: %s (%.2f MB)\\n",
        packageResult.FilePath, float64(packageResult.Size)/1024/1024)
}
```

## 📊 Health Monitoring

```go
// Plugin automatically reports health status
status := plugin.Status()
fmt.Printf("Plugin Status: %s\\n", status.State)
fmt.Printf("Health: %s\\n", status.Health)
fmt.Printf("Uptime: %v\\n", status.Uptime)
fmt.Printf("Memory Usage: %d MB\\n", status.Metrics["memoryUsage"])

// Custom health check
plugin.RegisterHook("health", func(ctx context.Context, data interface{}) error {
    // Custom health validation logic
    if !isServiceReachable() {
        return fmt.Errorf("external service unreachable")
    }
    return nil
})
```

## 🎨 Plugin Templates

### Available Templates

- **go-utility**: Full-featured Go plugin with frontend
- **theme**: CSS theme plugin
- **integration**: API integration plugin
- **dashboard**: Data visualization plugin

### Custom Templates

```go
customTemplate := &sdk.PluginTemplate{
    Name:        "custom-api",
    Description: "Custom API integration template",
    Type:        "integration",
    Language:    "go",
    Files: []sdk.TemplateFile{
        {
            Path:     "main.go",
            Template: true,
            Content:  "// Custom template content with {{.Variables}}",
        },
    },
    Variables: map[string]interface{}{
        "APIVersion": "v1",
    },
}

enhancedSDK.Generator().RegisterTemplate(customTemplate)
```


## 🔧 Configuration

### SDK Configuration

```go
config := sdk.SDKConfig{
    RegistryURL:   "https://registry.delve.sh",
    SecurityLevel: sdk.SecurityLevelMedium,
    CacheEnabled:  true,
    DevMode:       false,
    LogLevel:      sdk.LogLevelInfo,
    CustomTemplates: map[string]*sdk.PluginTemplate{
        "my-template": customTemplate,
    },
}

sdk := sdk.NewEnhancedSDK(config)
```

## 📚 API Reference

### Core Types

- `EnhancedSDK` - Main SDK interface
- `EnhancedPlugin` - Enhanced plugin interface with lifecycle methods
- `SecurityManager` - Handles security policies and sandboxing
- `EnhancedRegistry` - Advanced plugin registry with search and metadata
- `SchemaValidator` - Plugin configuration validation
- `PluginGenerator` - Template-based plugin generation
- `DevTools` - Development and build tools

### Security Types

- `SecurityLevel` - Security level enumeration
- `SecurityPolicy` - Security policy configuration
- `Permission` - Permission request structure
- `Sandbox` - Isolated execution environment

### Registry Types

- `EnhancedPluginInfo` - Complete plugin metadata
- `RegistrySearchFilter` - Advanced search filtering
- `PluginMetrics` - Usage and performance metrics

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- Documentation: [https://docs.delve.sh/sdk](https://docs.delve.sh/sdk)
- Issues: [GitHub Issues](https://github.com/PortableSheep/delve-sdk/issues)
- Community: [Discord Server](https://discord.gg/delve)

## 🗺️ Roadmap

- [ ] Visual plugin builder interface
- [ ] Plugin marketplace integration
- [ ] Advanced analytics and monitoring
- [ ] Multi-language SDK support (Python, JavaScript, Rust)
- [ ] Plugin dependency management system
- [ ] Automated security scanning
- [ ] Performance profiling tools
- [ ] Plugin migration utilities

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

## UI Contributions (Sidebar & Footer)

Plugins can extend the Delve UI with sidebar items and footer widgets, and update them live at runtime.

Helpers on Plugin:
- `SetUIContributions(c UIContributions)` – replace all sidebar/footer entries
- `UpsertSidebarItem(item SidebarItem)` – insert or replace a sidebar item and emit an update
- `UpsertFooterWidget(w FooterWidget)` – insert or replace a footer widget and emit an update
- `SetFooterBadgeCount(identifier string, count int)` – update a footer widget badge by ElementTag or Title

Types:
```go
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
```

Minimal usage:
```go
// During registration
pluginInfo := &sdk.RegisterRequest{
    Name:             "my-plugin",
    Description:      "Demo",
    CustomElementTag: "my-plugin",
    UiComponentPath:  "frontend/component.js",
}
plugin, _ := sdk.Start(pluginInfo)

// Contribute a sidebar item and a footer icon
_ = plugin.SetUIContributions(sdk.UIContributions{
    Sidebar: []sdk.SidebarItem{
        { Title: "My Plugin", ElementTag: "my-plugin", ComponentPath: "frontend/component.js", OnClickCommand: &sdk.CommandRef{ID: "my-plugin.open"} },
    },
    Footer: []sdk.FooterWidget{
        { Title: "My Plugin", Icon: "mdi:rocket", OnClickCommand: &sdk.CommandRef{ID: "my-plugin.open"} },
    },
})

// Update badge later (e.g., unread count)
_ = plugin.SetFooterBadgeCount("My Plugin", 5)

// Default click handler returns an activation hint for the frontend
plugin.OnCommand("my-plugin.open", func(ctx context.Context, args []any) (any, error) {
    return map[string]any{"activateUI": map[string]any{"plugin": "my-plugin", "tag": "my-plugin"}}, nil
})
```

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
