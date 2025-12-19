package config

import (
	"net"
	"os"
	"strings"
)

type AuthConfig struct {
	IPWhitelistRaw string
	IPNets         []*net.IPNet

	TokenHeader string
	TokenKey    string
}

func LoadAuthConfig() (*AuthConfig, error) {
	cfg := &AuthConfig{
		IPWhitelistRaw: strings.TrimSpace(os.Getenv("AUTH_IP_WHITELIST")),
		TokenHeader:    strings.TrimSpace(os.Getenv("AUTH_TOKEN_HEADER")),
		TokenKey:       strings.TrimSpace(os.Getenv("AUTH_TOKEN_KEY")),
	}

	// parse ip whitelist
	if cfg.IPWhitelistRaw != "" {
		items := strings.Split(cfg.IPWhitelistRaw, ",")
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}

			// single IP -> /32
			if !strings.Contains(item, "/") {
				item += "/32"
			}

			_, ipNet, err := net.ParseCIDR(item)
			if err != nil {
				return nil, err
			}
			cfg.IPNets = append(cfg.IPNets, ipNet)
		}
	}

	return cfg, nil
}
