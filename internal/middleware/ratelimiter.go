package middleware

import (
	"net/http"
	"time"

	"github.com/flaviojohansson/goexpert-rate-limiter/internal/limiter"

	"github.com/gin-gonic/gin"
)

func RateLimiter(l *limiter.Limiter, ipLimit, tokenLimit int, window, blockTime time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		var key string
		var limit int

		// Prioriza token sobre IP
		token := c.GetHeader("API_KEY")
		if token != "" {
			key = "token:" + token
			limit = tokenLimit
		} else {
			key = "ip:" + c.ClientIP()
			limit = ipLimit
		}

		allowed, err := l.Check(key, limit, window, blockTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"message": "you have reached the maximum number of requests or actions allowed within a certain time frame",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
