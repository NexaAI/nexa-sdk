package nexa_sdk

import (
	"context"
	"fmt"
	"path"
	"testing"
	"time"
)

var (
	// vlm is the global vlm instance used across all tests
	vlm *VLM
)

// initvlm creates a new vlm instance for testing with a predefined model
// Uses the Qwen3-0.6B-GGUF model from the local cache
func initVLM() {
	mmproj := path.Join(nexaPath, "models", "bmV4YW1sL25leGFtbC1tb2RlbHM=", "mmproj-model-f16.gguf")
	vlm, _ = NewVLM(
		path.Join(nexaPath, "models", "bmV4YW1sL25leGFtbC1tb2RlbHM=", "gemma-3-4b-it-Q8_0.gguf"),
		&mmproj, 8192, nil)
}

// deinitvlm cleans up the vlm instance after testing
func deinitVLM() {
	vlm.Destroy()
}

// TestEncode tests the tokenization functionality
// Verifies that text can be converted to token IDs
func TestVLMEncode(t *testing.T) {
	ids, e := vlm.Encode("hello world")
	if e != nil {
		t.Error(e)
	}
	t.Log(ids)
}

// TestDecode tests the detokenization functionality
// Verifies that token IDs can be converted back to text
func TestVLMDecode(t *testing.T) {
	res, e := vlm.Decode([]int32{2, 23391, 1902})
	if e != nil {
		t.Error(e)
	}
	t.Log(res)
}

// TestApplyChatTemplate tests the chat template formatting functionality
// Verifies that chat messages can be properly formatted for the model
func TestVLMApplyChatTemplate(t *testing.T) {
	msg, e := vlm.ApplyChatTemplate([]ChatMessage{
		{LLMRoleUser, "hello"},
		{LLMRoleAssistant, "yes, you are a so cute cat"},
		{LLMRoleUser, "can you give me a new cute name"},
	}, nil, nil)

	if e != nil {
		t.Error(e)
	}
	t.Log(msg)
}

// TestGenerate tests basic text generation functionality
// Verifies that the model can complete a given prompt
//func TestVLMGenerate(t *testing.T) {
//	pic := "~/Pictures/ScreenShot/20200201_182517.png"
//	res, e := vlm.Generate("what does the picture say", &pic)
//	if e != nil {
//		t.Error(e)
//	}
//	t.Log(res)
//}

// TestGetChatTemplate tests retrieval of the model's chat template
// Verifies that the default chat template can be obtained
func TestVLMGetChatTemplate(t *testing.T) {
	msg, e := vlm.GetChatTemplate(nil)
	if e != nil {
		t.Error(e)
	}
	t.Log(msg)
}

// TestChat tests end-to-end chat functionality
// Combines chat template application with text generation
func TestVLMChat(t *testing.T) {
	// Format the user message using chat template
	pic := "~/Pictures/ScreenShot/20200201_182517.png"
	msg, e := vlm.ApplyChatTemplate([]ChatMessage{
		{LLMRoleUser, "what does the picture say"},
	}, []string{pic}, nil)
	if e != nil {
		t.Error(e)
	}

	// Generate response using the formatted prompt
	res, e := vlm.Generate(msg, []string{pic}, nil)
	if e != nil {
		t.Error(e)
	}
	t.Log(res)
}

// TestGenerateStream tests streaming text generation functionality
// Measures generation speed and verifies that tokens are streamed properly
func TestVLMGenerateStream(t *testing.T) {
	pic := "/home/remilia/Pictures/ScreenShot/20200201_182517.png"
	dataCh, errCh := vlm.GenerateStream(context.Background(), "what does the picture say", []string{pic}, nil)

	start := time.Now()
	count := 0

	// Receive and print each token as it's generated
	for r := range dataCh {
		fmt.Print(r)
		count++
	}
	fmt.Print("\n")

	// Check for any errors during generation
	e, ok := <-errCh
	if ok {
		t.Error(e)
		return
	}

	// Calculate and report generation speed
	duration := time.Since(start).Seconds()
	t.Logf("\033[34mGenerate %d token in %f s, speed is %f token/s\033[0m\n",
		count,
		duration,
		float64(count)/duration)
}
