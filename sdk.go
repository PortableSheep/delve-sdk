package sdk

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/PortableSheep/delve_sdk/plugin_comms"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Plugin represents a third-party plugin.
type Plugin struct {
	client plugin_comms.PluginManagerClient
	conn   *grpc.ClientConn
}

// Start connects to the host application, registers the plugin, and returns a client for further communication.
// It handles parsing the required '--grpc-port' flag from the host.
func Start(pluginInfo *plugin_comms.RegisterRequest) (*Plugin, error) {
	grpcPort := flag.Int("grpc-port", 0, "gRPC server port of the main application")
	flag.Parse()

	if *grpcPort == 0 {
		return nil, fmt.Errorf("host gRPC port not provided via --grpc-port flag")
	}

	// Create a context with a timeout for the dial operation.
	dialCtx, dialCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dialCancel()

	// Establish connection with a timeout and blocking.
	conn, err := grpc.DialContext(
		dialCtx,
		fmt.Sprintf("127.0.0.1:%d", *grpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to host within timeout: %w", err)
	}

	client := plugin_comms.NewPluginManagerClient(conn)
	// The registration call itself can have a shorter timeout.
	registerCtx, registerCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer registerCancel()

	// Register with the host
	res, err := client.Register(registerCtx, pluginInfo)
	if err != nil {
		_ = conn.Close() // Clean up connection on failure
		return nil, fmt.Errorf("failed to register with host: %w", err)
	}

	if !res.Success {
		_ = conn.Close() // Clean up connection on failure
		return nil, fmt.Errorf("registration rejected by host: %s", res.Message)
	}

	log.Printf("Plugin successfully registered with host: %s", res.Message)

	return &Plugin{
		client: client,
		conn:   conn,
	}, nil
}

func (p *Plugin) Close() {
	if p.conn != nil {
		_ = p.conn.Close()
	}
}
