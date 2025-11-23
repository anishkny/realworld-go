package src

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func CreateApp() *gin.Engine {
	app := gin.Default()
	api := app.Group("/api")
	{
		// Health
		api.GET("", func(c *gin.Context) {
			c.JSON(200, "OK")
		})
	}
	trapSignals()
	return app
}

// Trap OS signals for graceful shutdown (write coverage data etc)
func trapSignals() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		log.Println("Shutting down server...")
		// Perform any necessary cleanup here
		os.Exit(0)
	}()
}
