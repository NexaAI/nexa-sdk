package handler

import (
	"errors"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"

	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/server/utils"
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
		for q, f := range m.ModelFile {
			if !f.Downloaded {
				continue
			}
			id := m.Name
			if q != "N/A" {
				id += ":" + q
			}
			res = append(res, openai.Model{
				ID:      id,
				OwnedBy: strings.Split(m.Name, "/")[0],
			})
		}
	}

	c.JSON(http.StatusOK, map[string]any{
		"object": "list",
		"data":   res,
	})
}

func RetrieveModel(c *gin.Context) {
	name := strings.TrimPrefix(c.Param("model"), "/")
	name, quant := utils.NormalizeModelName(name)

	// check model exist
	s := store.Get()
	manifest, err := s.GetManifest(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.JSON(http.StatusNotFound, nil)
		} else {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		}
		return
	}

	// fill quant if not specified
	if quant == "" {
		quants := make([]string, 0, len(manifest.ModelFile))
		for quant, v := range manifest.ModelFile {
			if v.Downloaded {
				quants = append(quants, quant)
				break
			}
		}
		slices.Sort(quants)
		quant = quants[0]
	}

	// check quant exist
	if _, ok := manifest.ModelFile[quant]; !ok {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	// compact with openai format
	var res map[string]any
	ms, _ := sonic.Marshal(manifest)
	_ = sonic.Unmarshal(ms, &res)
	model := openai.Model{}
	model.ID = name
	if quant != "N/A" {
		model.ID += ":" + quant
	}
	model.OwnedBy = strings.Split(manifest.Name, "/")[0]
	ms, _ = sonic.Marshal(model)
	_ = sonic.Unmarshal(ms, &res)

	c.JSON(http.StatusOK, res)
}
