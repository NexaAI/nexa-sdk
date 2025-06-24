//go:build rerank

package nexa_sdk

import (
	"path"
	"testing"
)

var reranker Reranker

func initReranker() {
	reranker = NewReranker(
		path.Join(nexaPath, "models", "UXdlbi9Rd2VuMy0wLjZCLUdHVUY=", "modelfile"),
		nil, nil,
	)

}

func deinitReranker() {
	reranker.Destroy()
}

func TestRerank(t *testing.T) {
	data := []string{
		"hello world",
		"data test test test data",
	}

	res, e := reranker.Rerank("hello", data)
	if e != nil {
		t.Error(e)
	}

	t.Log(res)
}
