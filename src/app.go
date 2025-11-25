package src

import (
	"github.com/gin-gonic/gin"
)

func CreateApp() *gin.Engine {
	ConnectDatabase()
	app := gin.Default()
	api := app.Group("/api")
	{
		// Health
		api.GET("", func(c *gin.Context) { c.JSON(200, "OK") })

		// User
		api.POST("/users", RegisterUser)
	}
	TrapSignals()
	return app
}
