package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
)

var lock sync.Mutex

func GIL(c *gin.Context) {
	lock.Lock()
	defer lock.Unlock()
	c.Next()
}
