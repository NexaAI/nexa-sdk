package nexa_sdk

import (
	"path"
	"testing"
)

var embedder Embedder

func initEmbeder() {
	embedder = NewEmbedder(
		path.Join(nexaPath, "models", "UXdlbi9Rd2VuMy0wLjZCLUdHVUY=", "modelfile"),
		nil, nil,
	)

}

func deinitEmbeder() {
	embedder.Destroy()
}

func TestEmbed(t *testing.T) {
	data := []string{
		"hello world",
		"data test test test data",
	}

	res, e := embedder.Embed(data)
	if e != nil {
		t.Error(e)
	}

	t.Log(res)
}
