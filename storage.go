package sdk

import (
	"encoding/json"
	"fmt"
	"time"
)

// StorageRequest represents a storage operation request
type StorageRequest struct {
	ID        string         `json:"id"`
	Operation string         `json:"operation"`
	Type      string         `json:"type"`
	Key       string         `json:"key,omitempty"`
	Value     any            `json:"value,omitempty"`
	Version   string         `json:"version,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// StorageResponse represents a response to a storage request
type StorageResponse struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// StorageItem represents a stored item with metadata
type StorageItem struct {
	Key       string         `json:"key"`
	Value     any            `json:"value"`
	Type      string         `json:"type"`
	Version   string         `json:"version"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Store saves data with the specified key and storage type
func (p *Plugin) Store(storageType, key string, value any, version string) error {
	req := StorageRequest{
		Operation: "store",
		Type:      storageType,
		Key:       key,
		Value:     value,
		Version:   version,
	}

	_, err := p.sendStorageRequest(req)
	return err
}

// Load retrieves data for the specified key and storage type
func (p *Plugin) Load(storageType, key string) (*StorageItem, error) {
	req := StorageRequest{
		Operation: "load",
		Type:      storageType,
		Key:       key,
	}

	response, err := p.sendStorageRequest(req)
	if err != nil {
		return nil, err
	}

	// Convert response data to StorageItem
	itemData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal storage item: %w", err)
	}

	var item StorageItem
	if err := json.Unmarshal(itemData, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal storage item: %w", err)
	}

	return &item, nil
}

// Delete removes data for the specified key and storage type
func (p *Plugin) Delete(storageType, key string) error {
	req := StorageRequest{
		Operation: "delete",
		Type:      storageType,
		Key:       key,
	}

	_, err := p.sendStorageRequest(req)
	return err
}

// List returns all keys for the specified storage type
func (p *Plugin) List(storageType string) ([]string, error) {
	req := StorageRequest{
		Operation: "list",
		Type:      storageType,
	}

	response, err := p.sendStorageRequest(req)
	if err != nil {
		return nil, err
	}

	// Convert response data to string slice
	keysData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal keys: %w", err)
	}

	var keys []string
	if err := json.Unmarshal(keysData, &keys); err != nil {
		return nil, fmt.Errorf("failed to unmarshal keys: %w", err)
	}

	return keys, nil
}

// Clear removes all data for the specified storage type
func (p *Plugin) Clear(storageType string) error {
	req := StorageRequest{
		Operation: "clear",
		Type:      storageType,
	}

	_, err := p.sendStorageRequest(req)
	return err
}

// GetStats returns storage statistics
func (p *Plugin) GetStats() (map[string]int, error) {
	req := StorageRequest{
		Operation: "stats",
	}

	response, err := p.sendStorageRequest(req)
	if err != nil {
		return nil, err
	}

	// Convert response data to stats map
	statsData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stats: %w", err)
	}

	var stats map[string]int
	if err := json.Unmarshal(statsData, &stats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	return stats, nil
}

// Convenience methods for specific storage types

// StoreConfig stores configuration data
func (p *Plugin) StoreConfig(key string, value interface{}, version string) error {
	return p.Store("config", key, value, version)
}

// LoadConfig loads configuration data
func (p *Plugin) LoadConfig(key string) (*StorageItem, error) {
	return p.Load("config", key)
}

// StoreData stores application data
func (p *Plugin) StoreData(key string, value interface{}, version string) error {
	return p.Store("data", key, value, version)
}

// LoadData loads application data
func (p *Plugin) LoadData(key string) (*StorageItem, error) {
	return p.Load("data", key)
}

// StoreState stores current state
func (p *Plugin) StoreState(key string, value interface{}, version string) error {
	return p.Store("state", key, value, version)
}

// LoadState loads current state
func (p *Plugin) LoadState(key string) (*StorageItem, error) {
	return p.Load("state", key)
}

// StoreCache stores cache data
func (p *Plugin) StoreCache(key string, value interface{}, version string) error {
	return p.Store("cache", key, value, version)
}

// LoadCache loads cache data
func (p *Plugin) LoadCache(key string) (*StorageItem, error) {
	return p.Load("cache", key)
}
