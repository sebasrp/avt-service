// Package handlers contains HTTP request handlers for the AVT service.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GreetingResponse represents the greeting response
type GreetingResponse struct {
	Message string `json:"message"`
}

// GreetingHandler handles personalized greeting requests
func GreetingHandler(c *gin.Context) {
	name := c.Param("name")

	c.JSON(http.StatusOK, GreetingResponse{
		Message: "Hello, " + name + "!",
	})
}
