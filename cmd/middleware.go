package cmd

import (
	"downtimetracker/internals/database/redis"
	"downtimetracker/internals/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("access_token")
		if err == nil && accessToken != "" {
			email, jerr := utils.ValidateJWT(accessToken)
			if jerr == nil {
				c.Set("email", email)
				c.Next()
				return
			}
		}

		refreshToken, err := c.Cookie("refresh_token")
		if err != nil || refreshToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired, please login again"})
			c.Abort()
			return
		}

		email, err := utils.ValidateJWT(refreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		storedToken, err := redis.RDB.Get(ctx, "session:"+email).Result()
		if err != nil || storedToken != refreshToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session revoked or expired"})
			c.Abort()
			return
		}

		newAccess, newRefresh, err := utils.GenerateTokens(email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Sever minting failed"})
			c.Abort()
			return
		}

		redis.RDB.Set(ctx, "session:"+email, newRefresh, 7*24*time.Hour)

		c.SetCookie("access_token", newAccess, 15*60, "/", "", false, true)
		c.SetCookie("refresh_token", newRefresh, 7*24*60*60, "/", "", false, true)

		c.Set("email", email)
		c.Next()
	}
}

