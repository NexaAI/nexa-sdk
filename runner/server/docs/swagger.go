package docs

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed ui/*
var StaticFiles embed.FS

//go:embed swagger.yaml
var StaticYAML []byte

type embedFileSystem struct {
	http.FileSystem
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	if path != "/" {
		path = strings.TrimSuffix(path, "/")
	}
	_, err := e.Open(path)
	return err == nil
}

func getSwaggerSubFS() fs.FS {
	sub, _ := fs.Sub(StaticFiles, "ui")
	return sub
}

var FS = embedFileSystem{http.FS(getSwaggerSubFS())}

func SwaggerYAMLHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/x-yaml")
		c.Data(http.StatusOK, "application/x-yaml", StaticYAML)
	}
}
