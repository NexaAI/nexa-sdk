package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
)

var lock sync.Mutex

func GIL(c *gin.Context) {
	// Block and wait for lock instead of immediately failing
	// This prevents 429 errors when requests queue up briefly
	lock.Lock()
	defer lock.Unlock()

	c.Next()
}
