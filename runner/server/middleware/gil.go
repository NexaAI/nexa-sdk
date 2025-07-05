package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

var lock sync.Mutex

func GIL(c *gin.Context) {
	locked := lock.TryLock()

	if !locked {
		c.JSON(http.StatusTooManyRequests, "locked by other request")
		c.Abort()
		return
	}

	defer lock.Unlock()
	c.Next()

}
