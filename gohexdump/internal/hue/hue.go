// Package hue provides Philips Hue bridge integration for hexboard.
// It reads a TOML config at /var/lib/hexboard/hue.toml and exposes
// a TurnOn() method that fires a non-blocking PUT to the Hue v1 CLIP API.
package hue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

const configPath = "/var/lib/hexboard/hue.toml"

// Config holds the Hue connection parameters read from hue.toml.
// All three fields must be non-empty for Hue to be active.
type Config struct {
	BridgeIP string `toml:"bridge_ip"`
	APIKey   string `toml:"api_key"`
	DeviceID string `toml:"device_id"`
}

// LoadConfig reads /var/lib/hexboard/hue.toml.
// Returns (nil, nil) if the file is absent — Hue is silently disabled.
// Returns (nil, err) if the file exists but is malformed or incomplete.
// The caller should log the error and continue with Hue disabled.
func LoadConfig() (*Config, error) {
	var cfg Config
	_, err := toml.DecodeFile(configPath, &cfg)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // absent is not an error — just disabled
		}
		return nil, err
	}
	if cfg.BridgeIP == "" || cfg.APIKey == "" || cfg.DeviceID == "" {
		return nil, fmt.Errorf("hue.toml: bridge_ip, api_key, and device_id are all required")
	}
	return &cfg, nil
}

// hueClient has a 5-second timeout to prevent goroutine leak when bridge is unreachable.
var hueClient = &http.Client{Timeout: 5 * time.Second}

// TurnOn sends a PUT to the Hue bridge to turn the configured light on.
// Designed to be called as: go cfg.TurnOn()
// Logs and returns silently on any error — never blocks the caller.
func (c *Config) TurnOn() {
	body, err := json.Marshal(map[string]bool{"on": true})
	if err != nil {
		log.Printf("hue: marshal: %v", err)
		return
	}
	bridgeIP := strings.TrimRight(c.BridgeIP, "/")
	url := fmt.Sprintf("http://%s/api/%s/lights/%s/state", bridgeIP, c.APIKey, c.DeviceID)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		log.Printf("hue: request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := hueClient.Do(req)
	if err != nil {
		log.Printf("hue: put: %v", err)
		return
	}
	resp.Body.Close()
}
