// Package handler: Healthandler handles the health check endpoint
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Healthandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
