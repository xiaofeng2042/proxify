package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/poixeai/proxify/infra/config"
	"github.com/poixeai/proxify/infra/ctx"
	"github.com/poixeai/proxify/infra/logger"
	"github.com/poixeai/proxify/util"
)

func GinRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID := util.GenerateRequestID()
		c.Set(ctx.RequestID, reqID)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method

		// using init url path
		path := c.Request.URL.Path

		clientIP := c.ClientIP()
		targetURL := c.GetString(ctx.TargetURL)

		topRoute := c.GetString(ctx.TopRoute)
		if config.ReservedTopRoutes[topRoute] {
			targetURL = "-"
		}

		// neglect empty path
		if path == "" {
			return
		}

		// static resource not logged
		if isStaticAsset(path) {
			return
		}

		logger.Infof(
			"%s | %d | %s | %s -> %s | %v | %s",
			reqID, status, method, path, targetURL, latency, clientIP,
		)
	}
}

func isStaticAsset(path string) bool {
	if path == "/" {
		return false
	}

	// 常见静态文件路径或后缀
	staticPrefixes := []string{
		"/assets/", "/static/", "/js/", "/css/", "/img/", "/fonts/",
	}

	staticExts := []string{
		".js", ".css", ".map", ".ico", ".png", ".jpg", ".jpeg", ".svg",
		".webp", ".woff", ".woff2", ".ttf", ".json", ".wasm",
	}

	for _, prefix := range staticPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	for _, ext := range staticExts {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
