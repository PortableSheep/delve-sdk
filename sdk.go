package sdk

// Package sdk provides a WebSocket-based plugin system for Go applications.
//
// This package allows plugins to connect to a host application via WebSocket,
// register themselves, and communicate with the host for storage operations
// and message handling.
//
// Basic usage:
//
//	// Create registration request
//	pluginInfo := &sdk.RegisterRequest{
//		Name:             "My Plugin",
//		Description:      "A sample plugin",
//		UiComponentPath:  "/path/to/component.js",
//		CustomElementTag: "my-plugin",
//	}
//
//	// Start the plugin
//	plugin, err := sdk.Start(pluginInfo)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer plugin.Close()
//
//	// Listen for messages
//	plugin.Listen(func(messageType int, data []byte) {
//		// Handle incoming messages
//	})
