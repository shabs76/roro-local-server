package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shabs76/roro-local-server/controlers"
)

func SetupUserRoutes(router *gin.Engine) {
	userGroup := router.Group("/users")
	{
		userGroup.POST("/login", controlers.LoginUser)
	}
}
