package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/la1665/task-manager/internal/metrics"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}

		metrics.RequestsTotal.WithLabelValues(
			c.Request.Method, path, strconv.Itoa(c.Writer.Status()),
		).Inc()

		metrics.RequestLatencyHistogram.WithLabelValues(
			c.Request.Method, path,
		).Observe(time.Since(start).Seconds())
	}
}
