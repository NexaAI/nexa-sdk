package nexa_sdk

import (
	"testing"
)

var reranker *Reranker

func initReranker() {
	tokenizer := "/home/remilia/Workspace/github/nexasdk-bridge/modelfiles/jina_rerank_tokenizer.json"
	reranker = NewReranker(
		"/home/remilia/Workspace/github/nexasdk-bridge/modelfiles/jina-reranker-v2-base-multilingual.F16.gguf",
		&tokenizer,
		nil,
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
