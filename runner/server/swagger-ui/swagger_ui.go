package swagger_ui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/basic/docs"
)

//go:embed dist/*
var StaticFiles embed.FS

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

// 获取 dist 子目录的 fs.FS
func getSwaggerSubFS() fs.FS {
	sub, err := fs.Sub(StaticFiles, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}

var FS = embedFileSystem{http.FS(getSwaggerSubFS())}

// SwaggerJSONHandler 处理 swagger.json 请求
func SwaggerJSONHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		docs.SwaggerInfo.BasePath = "/v1"
		c.Header("Content-Type", "application/json")
		swaggerJSON := docs.SwaggerInfo.ReadDoc()
		c.Data(http.StatusOK, "application/json", []byte(swaggerJSON))
	}
}
