package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/poixeai/proxify/infra/ctx"
	"github.com/poixeai/proxify/infra/watcher"
	"github.com/poixeai/proxify/util"
)

func Extractor() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		top, sub := util.ExtractRoute(path)
		query := c.Request.URL.RawQuery

		if query != "" {
			if sub == "" {
				sub = "?" + query
			} else {
				sub = sub + "?" + query
			}
		}

		// store top and sub path into context for later use
		c.Set(ctx.TopRoute, top)
		c.Set(ctx.SubPath, sub)

		// check if route exists in routes.json
		cfg := watcher.GetRoutes()
		found := false
		for i := range cfg.Routes {
			r := &cfg.Routes[i]

			if r.Path == "/"+top {
				found = true
				c.Set(ctx.TargetEndpoint, r.Target)

				// store matched route config
				c.Set(ctx.RouteConfig, r)

				break
			}
		}
		c.Set(ctx.Proxified, found)

		c.Next()
	}
}
