package src

import (
	"github.com/gin-gonic/gin"
)

func CreateApp() *gin.Engine {
	ConnectDatabase()
	app := gin.Default()
	api := app.Group("/api", Authenticate)
	{
		// Health
		api.GET("", func(c *gin.Context) { c.JSON(200, "OK") })

		// User
		api.POST("/users", RegisterUser)
		api.POST("/users/login", LoginUser)
		api.GET("/user", GetUser)
		api.PUT("/user", UpdateUser)

		// Profile
		api.GET("/profiles/:username", GetProfile)
		api.POST("/profiles/:username/follow", FollowUser)
	}
	TrapSignals()
	return app
}
