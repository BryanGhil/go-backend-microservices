package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures which frontends are allowed to talk to the API
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		// 1. The exact URL of your frontend
		AllowOrigins: []string{"http://127.0.0.1:5173"},

		// 2. The HTTP verbs your frontend is allowed to use
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},

		// 3. The Headers your frontend is allowed to send
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization", // Crucial so your frontend can send the Bearer Token!
		},

		// 4. MUST BE TRUE FOR COOKIES!
		AllowCredentials: true,

		// 5. How long the browser should cache the preflight response
		MaxAge: 12 * time.Hour,
	})
}