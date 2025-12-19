package middleware

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/poixeai/proxify/infra/ctx"
	"github.com/poixeai/proxify/infra/logger"
	"github.com/poixeai/proxify/util"
)

// ModelRewrite rewrites the `model` field in Chat Completions style requests
// based on route-level modelMap configuration.
func ModelRewrite() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get current route config from context
		route := ctx.GetRoute(c)
		if route == nil || len(route.ModelMap) == 0 {
			c.Next()
			return
		}

		// no body, nothing to do
		if c.Request.Body == nil {
			c.Next()
			return
		}

		// read original body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Warnf("ModelRewrite: failed to read request body: %v", err)
			c.Next()
			return
		}

		// attempt to rewrite model
		newBody, rewritten, err := util.RewriteChatCompletionModel(
			bodyBytes,
			route.ModelMap,
		)
		if err != nil {
			logger.Warnf("ModelRewrite: rewrite failed: %v", err)
			// restore original body
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			c.Next()
			return
		}

		if rewritten {
			logger.Infof(
				"ModelRewrite: route=%s model rewritten",
				route.Name,
			)
			bodyBytes = newBody
		}

		// IMPORTANT: restore body for downstream handlers
		c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		c.Request.ContentLength = int64(len(bodyBytes))

		c.Next()
	}
}
