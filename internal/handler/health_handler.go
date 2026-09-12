// Package handler: Healthandler handles the health check endpoint
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealtHandler godoc
// @Summary      Check system health
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func HealtHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
