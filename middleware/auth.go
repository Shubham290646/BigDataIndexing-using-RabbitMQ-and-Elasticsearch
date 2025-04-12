package middleware

import (
	"net/http"
	"os"
	"strings"

	"log"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
)

func OAuth2Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := os.Getenv("CLIENT_ID")
		if clientID == "" {
			log.Fatal("CLIENT_ID not set in environment")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Print(authHeader)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing or empty"})
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		_, err := idtoken.Validate(c, idToken, clientID)
		if err != nil {
			log.Print(err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		c.Next()
	}
}
