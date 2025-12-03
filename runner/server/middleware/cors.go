package middleware

import (
	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	"github.com/gin-gonic/gin"
)

func CORS(c *gin.Context) {
	h := c.Writer.Header()
	h.Set("Access-Control-Allow-Origin", config.Get().Origins)
	h.Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST")
	h.Set("Access-Control-Allow-Headers", "Content-Type, Nexa-KeepCache")
	h.Set("Access-Control-Allow-Credentials", "true")
	h.Set("Access-Control-Max-Age", "86400")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}

	c.Next()
}
