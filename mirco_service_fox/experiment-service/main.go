package main

import (
	"experiment-service/config"
	"experiment-service/routers"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	routers.RegisterRoutes(r)

	config.InitDB()
	// Make OSS optional for CI to avoid external dependencies
	if os.Getenv("ENABLE_OSS") == "true" {
		config.InitOSS()
	}

	r.Run(":8082")
}
