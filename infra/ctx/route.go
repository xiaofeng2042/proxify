package ctx

import (
	"github.com/gin-gonic/gin"
	"github.com/poixeai/proxify/infra/config"
)

func GetRoute(c *gin.Context) *config.Route {
	if v, ok := c.Get(RouteConfig); ok {
		if route, ok := v.(*config.Route); ok {
			return route
		}
	}
	return nil
}
