package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/tkircsi/azure-worm-storage/handler"
)

// HealthCheck is used for container HEALTHCHECK
// GET /healthcheck
// Response: "OK"
func HealthCheck(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

// HomeAPI returns the Demo API version number
// GET /
func HomeAPI(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "version: 0.1",
	})
}

// CheckAppAPIKey checks API_KEY
// GET /api
func CheckAppAPIKey(c *gin.Context) {
	const demoKey string = "11AA22BB"
	apiKey := c.GetHeader("API_KEY")
	if len(apiKey) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "API_KEY is required!",
		})
		return
	}

	if apiKey != demoKey {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid API_KEY",
		})
		return
	}

	c.Next()
}


func main() {
	
	r := gin.Default()
	r.GET("/", HomeAPI)
	r.GET("/healthcheck", HealthCheck)
	api := r.Group("/api")
	api.Use(CheckAppAPIKey)
	{
		api.GET("/claims", handler.GetByPrefix())
		// api.GET("/time", GetTime("Europe/Budapest"))
		api.POST("/claims", handler.Add())
	}
	r.Run(":5000")
}
