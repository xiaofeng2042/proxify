package middleware

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/poixeai/proxify/infra/config"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, exists := c.Get("auth_config")
		if !exists {
			c.Next()
			return
		}

		cfg := v.(*config.AuthConfig)

		// ===== IP Whitelist =====
		if len(cfg.IPNets) > 0 {
			ipStr := c.ClientIP()
			ip := net.ParseIP(ipStr)
			allowed := false

			for _, netw := range cfg.IPNets {
				if netw.Contains(ip) {
					allowed = true
					break
				}
			}

			if !allowed {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "IP not allowed",
				})
				return
			}
		}

		// ===== Token Auth =====
		if cfg.TokenKey != "" {
			token := c.GetHeader(cfg.TokenHeader)
			if token == "" || token != cfg.TokenKey {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid token",
				})
				return
			}
		}

		c.Next()
	}
}
