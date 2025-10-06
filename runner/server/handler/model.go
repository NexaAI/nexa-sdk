package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"

	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
)

func ListModels(c *gin.Context) {
	s := store.Get()

	models, err := s.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	res := make([]openai.Model, 0, len(models))
	for _, m := range models {
		res = append(res, openai.Model{
			ID:      m.Name,
			OwnedBy: strings.Split(m.Name, "/")[0],
		})
	}

	c.JSON(http.StatusOK, res)
}

func RetrieveModel(c *gin.Context) {
	model := strings.TrimPrefix(c.Param("model"), "/")

	s := store.Get()

	if manifest, err := s.GetManifest(model); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.JSON(http.StatusNotFound, nil)
		} else {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		}
	} else {
		c.JSON(http.StatusOK, manifest)
	}
}

func PullModel(c *gin.Context) {
	manifest := types.ModelManifest{}
	if err := c.ShouldBindJSON(&manifest); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if manifest.Name == "" || len(manifest.ModelFile) == 0 {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "name or modelfile is empty"})
		return
	}

	s := store.Get()
	ctx, cancel := context.WithCancel(context.Background())

	infoCh, errCh := s.Pull(ctx, manifest)

	c.Stream(func(w io.Writer) bool {
		info, ok := <-infoCh
		if ok {
			c.SSEvent("", info)
			return true
		} else {
			err, ok := <-errCh
			if ok {
				c.SSEvent("", map[string]any{"error": err.Error()})
				return true
			}
		}

		return false
	})

	cancel()
	for range infoCh {
	}
	for range errCh {
	}
}
