package middleware

import (
	"net/http"
	"strings"
	"ecommerce/pb"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(userClient pb.UserServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		token := parts[1]
		res, err := userClient.VerifySession(c.Request.Context(), &pb.VerifySessionRequest{Token: token})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
			c.Abort()
			return
		}

		c.Set("userID", res.UserId)
		c.Next()
	}
}