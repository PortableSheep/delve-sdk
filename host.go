package sdk

import (
	"encoding/json"
	"fmt"
	"time"
)

// HostRequest represents a generic request to the host (non-storage, non-command)
type HostRequest struct {
	Type      string         `json:"type"` // always host/request
	ID        string         `json:"id"`
	Operation string         `json:"operation"` // e.g. getConfig, subscribeConfig, requestPermission
	Plugin    string         `json:"plugin,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
}

// HostResponse is a generic host response
type HostResponse struct {
	Type    string `json:"type"` // host/response
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Result  any    `json:"result,omitempty"`
}

// sendHostRequest sends a generic host request and waits for response
func (p *Plugin) sendHostRequest(op string, data map[string]any) (*HostResponse, error) {
	if p == nil || p.conn == nil {
		return nil, fmt.Errorf("plugin connection not initialised")
	}
	id := p.generateRequestID()
	req := HostRequest{Type: "host/request", ID: id, Operation: op, Data: data}
	ch := make(chan HostResponse, 1)
	p.requestsMutex.Lock()
	p.hostPending[id] = ch
	p.requestsMutex.Unlock()
	defer func() {
		p.requestsMutex.Lock()
		delete(p.hostPending, id)
		p.requestsMutex.Unlock()
		close(ch)
	}()
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if err := p.conn.WriteMessage(1, b); err != nil { // 1 = TextMessage
		return nil, err
	}
	select {
	case resp := <-ch:
		if !resp.Success {
			return &resp, fmt.Errorf("%s", resp.Error)
		}
		return &resp, nil
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("host request timeout")
	}
}

// GetConfig asks host for current merged plugin configuration
func (p *Plugin) GetConfig() (map[string]any, error) {
	resp, err := p.sendHostRequest("getConfig", nil)
	if err != nil {
		return nil, err
	}
	if resp.Result == nil {
		return map[string]any{}, nil
	}
	if m, ok := resp.Result.(map[string]any); ok {
		return m, nil
	}
	// Attempt JSON roundtrip coercion
	b, _ := json.Marshal(resp.Result)
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	return out, nil
}

// RequestPermission asks the host to grant a permission; host may prompt or auto-grant.
func (p *Plugin) RequestPermission(name string) (bool, error) {
	resp, err := p.sendHostRequest("requestPermission", map[string]any{"permission": name})
	if err != nil {
		return false, err
	}
	if resp.Result == nil {
		return false, nil
	}
	if b, ok := resp.Result.(bool); ok {
		return b, nil
	}
	return false, nil
}

// SubscribeConfig asks host to push future configuration changes back as events
// Host side must emit {type:"event", event:"config:changed", data:{ "keys":[...], "config":{...} }}
// Returns true if subscription established.
func (p *Plugin) SubscribeConfig() (bool, error) {
	resp, err := p.sendHostRequest("subscribeConfig", nil)
	if err != nil {
		return false, err
	}
	if resp.Result == nil {
		return false, nil
	}
	if b, ok := resp.Result.(bool); ok {
		return b, nil
	}
	return false, nil
}
