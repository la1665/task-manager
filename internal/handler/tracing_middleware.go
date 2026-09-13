package handler

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/la1665/task-manager/internal/tracing"
)

func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = tracing.NewRequestID()
		}

		ctx := tracing.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set("X-Request-ID", requestID)

		start := time.Now()
		c.Next()

		log.Printf("[%s] %s %s status=%d duration=%s",
			requestID, c.Request.Method, c.FullPath(), c.Writer.Status(), time.Since(start))
	}
}
