package config

import (
	"encoding/json"
	"os"
)

type Route struct {
	Path        string `json:"path"`
	Target      string `json:"target"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// model mapping (optional)
	ModelMap map[string]string `json:"model_map,omitempty"`
}

type RoutesConfig struct {
	Routes []Route `json:"routes"`
}

func LoadRoutesConfig(path string) (*RoutesConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg RoutesConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
