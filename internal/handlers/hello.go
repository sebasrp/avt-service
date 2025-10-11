package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HelloHandler handles the root endpoint
func HelloHandler(c *gin.Context) {
	c.String(http.StatusOK, "Hello, World!")
}
