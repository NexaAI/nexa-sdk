package nexa_sdk

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

// llm is the global LLM instance used across all tests
var llm *LLM

// initLLM creates a new LLM instance for testing with a predefined model
// Uses the nexaml/Qwen3-0.6B model from the local cache
func initLLM() {
	var err error
	llm, err = NewLLM(
		path.Join(os.Getenv("HOME"), ".cache", "models", "Qwen3-0.6B-Q8_0.gguf"),
		nil, 8192, nil)
	if err != nil {
		panic("Error creating LLM: " + err.Error())
	}
}

// deinitLLM cleans up the LLM instance after testing
func deinitLLM() {
	llm.Destroy()
}

// TestEncode tests the tokenization functionality
// Verifies that text can be converted to token IDs
func TestEncode(t *testing.T) {
	ids, e := llm.Encode("hello world")
	if e != nil {
		t.Error(e)
	}
	t.Log(ids)
}

// TestDecode tests the detokenization functionality
// Verifies that token IDs can be converted back to text
func TestDecode(t *testing.T) {
	res, e := llm.Decode([]int32{14990, 1879})
	if e != nil {
		t.Error(e)
	}
	t.Log(res)
}

// TestSaveKVCache tests saving the key-value cache to disk
// This can improve performance for subsequent generations
func TestSaveKVCache(t *testing.T) {
	e := llm.SaveKVCache(path.Join(nexaPath, "kvcache"))
	if e != nil {
		t.Error(e)
	}
}

// SKIP_TestLoadKVCache tests loading a previously saved key-value cache
// Currently skipped (TODO) - likely due to implementation issues
func TestLoadKVCache(t *testing.T) {
	e := llm.LoadKVCache(path.Join(nexaPath, "kvcache"))
	if e != nil {
		t.Error(e)
	}
}

// TestApplyChatTemplate tests the chat template formatting functionality
// Verifies that chat messages can be properly formatted for the model
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

func TestApplyJinjaTemplate(t *testing.T) {
	tests := []struct {
		name    string
		param   ChatTemplateParam
		wantErr bool
	}{
		{
			name: "No tools, multiple messages",
			param: ChatTemplateParam{
				Messages: []ChatMessage{
					{Role: LLMRoleUser, Content: "hello"},
					{Role: LLMRoleAssistant, Content: "yes, you are a so cute cat"},
					{Role: LLMRoleUser, Content: "can you give me a new cute name"},
				},
			},
		},
		{
			name: "No tools, single message",
			param: ChatTemplateParam{
				Messages: []ChatMessage{
					{Role: LLMRoleUser, Content: "hello"},
				},
			},
		},
		{
			name:    "No tools, empty message list",
			param:   ChatTemplateParam{},
			wantErr: true,
		},
		{
			name: "With tools, multiple messages",
			param: ChatTemplateParam{
				Messages: []ChatMessage{
					{Role: LLMRoleUser, Content: "hello"},
					{Role: LLMRoleUser, Content: "what is the weather"},
				},
				Tools: []ChatTool{
					{
						Type: "function",
						Function: ChatToolFunction{
							Name:        "get_weather",
							Description: "Get current weather info",
							Parameters: map[string]any{
								"location": map[string]string{"type": "string"},
							},
						},
					},
				},
			},
		},
		{
			name: "With tools, single message",
			param: ChatTemplateParam{
				Messages: []ChatMessage{
					{Role: LLMRoleUser, Content: "get temperature"},
				},
				Tools: []ChatTool{
					{
						Type: "function",
						Function: ChatToolFunction{
							Name:        "get_temp",
							Description: "Get current temperature",
							Parameters:  map[string]any{},
						},
					},
				},
			},
		},
		{
			name: "With tools, empty messages",
			param: ChatTemplateParam{
				Messages: []ChatMessage{},
				Tools: []ChatTool{
					{
						Type: "function",
						Function: ChatToolFunction{
							Name:        "noop",
							Description: "do nothing",
							Parameters:  map[string]any{},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := llm.ApplyJinjaTemplate(tt.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyJinjaTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				t.Logf("Generated template:\n%s\n", msg)
			}
		})
	}
}

// TestGenerate tests basic text generation functionality
// Verifies that the model can complete a given prompt
func TestGenerate(t *testing.T) {
	res, e := llm.Generate("i am lihua, ")
	if e != nil {
		t.Error(e)
	}
	t.Log(res)
}

// TestGetChatTemplate tests retrieval of the model's chat template
// Verifies that the default chat template can be obtained
func TestGetChatTemplate(t *testing.T) {
	msg, e := llm.GetChatTemplate(nil)
	if e != nil {
		t.Error(e)
	}
	t.Log(msg)
}

// TestChat tests end-to-end chat functionality
// Combines chat template application with text generation
func TestChat(t *testing.T) {
	// Format the user message using chat template
	msg, e := llm.ApplyChatTemplate([]ChatMessage{
		{LLMRoleUser, "i am lihua, i am a cat"},
	})
	if e != nil {
		t.Error(e)
	}

	// Generate response using the formatted prompt
	res, e := llm.Generate(msg)
	if e != nil {
		t.Error(e)
	}
	t.Log(res)
}

func TestChatMultiTurn(t *testing.T) {
	type turn struct {
		user      string
		expectSub string
	}

	turns := []turn{
		{
			user:      "a=11, b=22, what is a+b?",
			expectSub: "33",
		},
		{
			user:      "what is b-a?",
			expectSub: "11",
		},
		{
			user:      "what is b/a?",
			expectSub: "2",
		},
	}

	var history []ChatMessage
	for i, turn := range turns {
		history = append(history, ChatMessage{Role: LLMRoleUser, Content: turn.user})
		msg, err := llm.ApplyChatTemplate(history)
		if err != nil {
			t.Fatalf("ApplyChatTemplate failed at turn %d: %v", i+1, err)
		}
		res, err := llm.Generate(msg)
		if err != nil {
			t.Fatalf("Generate failed at turn %d: %v", i+1, err)
		}
		t.Logf("Turn %d reply: %s", i+1, res)
		if turn.expectSub != "" && !strings.Contains(res, turn.expectSub) {
			t.Errorf("Turn %d: expected reply to contain %q, got %q", i+1, turn.expectSub, res)
		}
		// Append assistant reply to history for next turn
		history = append(history, ChatMessage{Role: LLMRoleAssistant, Content: res})
	}
}

// TestGenerateStream tests streaming text generation functionality
// Measures generation speed and verifies that tokens are streamed properly
func TestGenerateStream(t *testing.T) {
	dataCh, errCh := llm.GenerateStream(context.Background(), "i am lihua, ")

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
