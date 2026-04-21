package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// 1. Define the Histogram
// A Histogram groups response times into "buckets" (e.g., requests that took <50ms, <100ms, etc.)
var httpDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds.",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}, // Buckets from 50ms to 10s
	},
	[]string{"method", "path", "status"}, // Labels so we can filter in Grafana
)

// MetricsMiddleware acts as our stopwatch
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start the stopwatch
		start := time.Now()

		// Process the actual request (e.g., hit the Order Service)
		c.Next()

		// Stop the stopwatch
		duration := time.Since(start).Seconds()

		// Grab the details of the request
		method := c.Request.Method
		path := c.FullPath() // Use FullPath (e.g., /api/products/:id) to avoid unique URLs ruining the grouping
		status := strconv.Itoa(c.Writer.Status())

		// If path is empty (e.g., a 404 error), just record it as "unknown"
		if path == "" {
			path = "unknown"
		}

		// Record the data into Prometheus!
		httpDuration.WithLabelValues(method, path, status).Observe(duration)
	}
}