package nexa_sdk

import (
	"fmt"
	"path"
	"testing"
	"time"
)

var (
	llm LLM
)

func initLLM() {
	llm = NewLLM(
		path.Join(nexaPath, "models", "UXdlbi9Rd2VuMy0wLjZCLUdHVUY=", "modelfile"),
		nil, 4096, nil)
}

func deinitLLM() {
	llm.Destroy()
}

func TestEncode(t *testing.T) {
	ids, e := llm.Encode("hello world")
	if e != nil {
		t.Error(e)
	}
	t.Log(ids)
}

func TestDecode(t *testing.T) {
	res, e := llm.Decode([]int32{14990, 1879})
	if e != nil {
		t.Error(e)
	}
	t.Log(res)
}

func TestSaveKVCache(t *testing.T) {
	e := llm.SaveKVCache(path.Join(nexaPath, "kvcache"))
	if e != nil {
		t.Error(e)
	}
}

// TODO
func SKIP_TestLoadKVCache(t *testing.T) {
	e := llm.LoadKVCache(path.Join(nexaPath, "kvcache"))
	if e != nil {
		t.Error(e)
	}
}

func TestApplyChatTemplate(t *testing.T) {
	msg, e := llm.ApplyChatTemplate([]ChatMessage{
		{LLMRoleUser, "hello"},
		{LLMRoleAssistant, "yes, you are a so cute cat"},
		{LLMRoleUser, "can you give me a new cute name"},
	})

	if e != nil {
		t.Error(e)
	}
	t.Log(msg)
}

func TestGenerate(t *testing.T) {
	res, e := llm.Generate("i am lihua, ")
	if e != nil {
		t.Error(e)
	}
	t.Log(res)
}

func TestGetChatTemplate(t *testing.T) {
	msg, e := llm.GetChatTemplate(nil)
	if e != nil {
		t.Error(e)
	}
	t.Log(msg)
}

func TestChat(t *testing.T) {
	msg, e := llm.ApplyChatTemplate([]ChatMessage{
		{LLMRoleUser, "i am lihua, i am a cat"},
	})
	if e != nil {
		t.Error(e)
	}
	res, e := llm.Generate(msg)
	if e != nil {
		t.Error(e)
	}
	t.Log(res)
}

func TestGenerateStream(t *testing.T) {
	dataCh, errCh := llm.GenerateStream("i am lihua, ")

	start := time.Now()
	count := 0

	for r := range dataCh {
		fmt.Print(r)
		count++
	}
	fmt.Print("\n")

	e, ok := <-errCh
	if ok {
		t.Error(e)
		return
	}
	duration := time.Since(start).Seconds()

	t.Logf("\033[34mGenerate %d token in %f s, speed is %f token/s\033[0m\n",
		count,
		duration,
		float64(count)/duration)
}
