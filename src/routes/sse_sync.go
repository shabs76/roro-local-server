package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shabs76/roro-local-server/controlers"
	"github.com/shabs76/roro-local-server/pkg/middleware"
)

func SetupSSESyncRoutes(router *gin.Engine) {
	sseGroup := router.Group("/sync")
	sseGroup.Use(middleware.AuthMiddleware())
	{
		sseGroup.GET("/auto", controlers.AutoDataSyncSSE)
		sseGroup.GET("/manifest/manual/:manifest-id", controlers.ManifestDataSyncSSE)
		sseGroup.GET("/trigger/publish/:manifestId", controlers.PublishVehiclesAndPackagesInspectionSSE)
		sseGroup.GET("/trigger/added-later/publish/:manifestId", controlers.PublishAddedLaterVehiclesAndPackagesSSE)
	}
}
