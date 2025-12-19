package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/poixeai/proxify/infra/config"
	"github.com/poixeai/proxify/infra/logger"
	"github.com/poixeai/proxify/infra/watcher"
	"github.com/poixeai/proxify/router"
	"github.com/poixeai/proxify/util"
)

func main() {
	// load .env
	_ = godotenv.Load()

	// init logger
	logger.InitLogger()

	// check .env
	authCfg, err := config.LoadAuthConfig()
	if err != nil {
		logger.Fatalf("Auth config error: %v", err)
	}

	// Token auth
	if authCfg.TokenKey != "" {
		if len(authCfg.TokenKey) < 16 {
			logger.Errorf("AUTH_TOKEN_KEY is too short (<16), refused to start")
			return
		}
		if authCfg.TokenHeader == "" {
			logger.Errorf("AUTH_TOKEN_HEADER is required when token auth enabled")
			return
		}
		logger.Infof("Token auth enabled, header=%s", authCfg.TokenHeader)
	}

	if len(authCfg.IPNets) > 0 {
		logger.Infof("IP whitelist enabled, rules=%d", len(authCfg.IPNets))
	}

	// init routes watcher
	if err := watcher.InitRoutesWatcher(); err != nil {
		logger.Errorf("Failed to load routes config: %v", err)
		return
	}

	// init gin
	r := gin.New()
	r.SetTrustedProxies(nil)

	// Inject authCfg into Gin context
	r.Use(func(c *gin.Context) {
		c.Set("auth_config", authCfg)
		c.Next()
	})

	// setup routes
	router.SetRoutes(r)

	// setup frontend static files
	MountFrontend(r)

	// start server
	port := util.GetEnvPort()
	if err := r.Run(":" + port); err != nil {
		logger.Errorf("Failed to start server: %v", err)
		return
	}

	logger.Infof("Server running on port %s", port)
}
