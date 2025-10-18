// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

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
