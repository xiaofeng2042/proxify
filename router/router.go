// for static routes
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/poixeai/proxify/controller"
	"github.com/poixeai/proxify/middleware"
)

func SetRoutes(r *gin.Engine) {
	// basic middleware
	r.Use(middleware.Recover())
	r.Use(middleware.CORS())
	r.Use(middleware.GinRequestLogger())
	r.Use(middleware.Extractor())
	r.Use(middleware.Auth())
	r.Use(middleware.ResponsesToChat())  // Convert Responses API to Chat Completions (request)
	r.Use(middleware.ModelRewrite())
	r.Use(middleware.ResponseTransform()) // Convert Chat Completions to Responses API (response)

	// ==== routes.json ====
	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/", controller.ShowPathHandler)
		apiGroup.GET("/routes", controller.RoutesHandler)
	}
}
