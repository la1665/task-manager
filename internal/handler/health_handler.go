// Package handler: Healthandler handles the health check endpoint
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealtHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
