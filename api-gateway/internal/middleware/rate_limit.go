package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
)

// RateLimitMiddleware now takes the Redis Limiter as an argument
func RateLimitMiddleware(limiter *redis_rate.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		
		var limiterKey string

		// 1. Determine the Key (User ID or IP)
		userID, exists := c.Get("userID")
		if exists {
			limiterKey = fmt.Sprintf("rate_limit:user_%v", userID)
		} else {
			limiterKey = fmt.Sprintf("rate_limit:ip_%s", c.ClientIP())
		}

		// 2. Ask Redis if this request is allowed
		// We allow 2 requests per second, with a maximum burst of 5.
		limit := redis_rate.PerSecond(2)
		limit.Burst = 10

		res, err := limiter.Allow(c.Request.Context(), limiterKey, limit)
		if err != nil {
			// If Redis is temporarily down, it's safer to let the request through 
			// rather than breaking the whole API, but we log the error.
			fmt.Println("Redis Rate Limiter Error:", err)
			c.Next()
			return
		}

		// 3. Check the result from Redis
		if res.Allowed == 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded. Please wait a moment.",
				"retry_after": res.RetryAfter.Seconds(), // Tells the frontend exactly how long to wait!
			})
			c.Abort()
			return
		}

		// 4. Inject rate limit headers (Standard API practice)
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", res.Remaining))

		c.Next()
	}
}