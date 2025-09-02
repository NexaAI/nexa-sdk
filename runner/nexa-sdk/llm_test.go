package nexa_sdk

import (
	"log/slog"
	"os"
	"testing"
)

var llm *LLM
var messages []LlmChatMessage = []LlmChatMessage{
	{
		Role:    "system",
		Content: "You are a helpful repeater. You will repeat the user's message back to them. Repeat the message exactly as it is.",
	},
	{
		Role:    "user",
		Content: "Happy coding! 🚀",
	},
}

func initLLM() {
	slog.Debug("initLLM called")

	var err error

	input := LlmCreateInput{
		ModelPath:     "./modelfiles/llama_cpp/Qwen3-0.6B-Q8_0.gguf",
		TokenizerPath: "",
		Config: ModelConfig{
			NCtx:    512,
			NSeqMax: 64,
		},
		PluginID: "llama_cpp",
	}

	llm, err = NewLLM(input)
	if err != nil {
		panic("Error creating LLM: " + err.Error())
	}
}

func deinitLLM() {
	llm.Destroy()
}

func TestApplyChatTemplate(t *testing.T) {
	output, err := llm.ApplyChatTemplate(LlmApplyChatTemplateInput{Messages: messages})
	if err != nil {
		t.Errorf("ApplyChatTemplate failed: %v", err)
		return
	}

	t.Logf("ApplyChatTemplate: %s", output.FormattedText)
}

func TestLLMGenerateStream(t *testing.T) {
	tpl, err := llm.ApplyChatTemplate(LlmApplyChatTemplateInput{Messages: messages})
	if err != nil {
		t.Fatalf("Failed to generate text: %v", err)
	}

	input := LlmGenerateInput{
		PromptUTF8: tpl.FormattedText,
		OnToken: func(token string) bool {
			t.Logf("<< %s", token)
			return true
		},
	}

	stream, err := llm.Generate(input)
	if err != nil {
		t.Fatalf("Failed to generate text: %v", err)
	}

	t.Logf("GenerateStream: %s", stream.FullText)
}

func TestLLMMultiTurn(t *testing.T) {
	testRounds := []LlmChatMessage{
		{Role: "user", Content: "let a = 1; let b = 2;"},
		{Role: "user", Content: "what is the value of a + b?"},
	}
	history := []LlmChatMessage{}

	for i, round := range testRounds {
		history = append(history, round)
		tpl, err := llm.ApplyChatTemplate(LlmApplyChatTemplateInput{Messages: history,EnableThink: true})
		if err != nil {
			t.Fatalf("Failed to generate text: %v", err)
		}
		stream, err := llm.Generate(LlmGenerateInput{
			PromptUTF8: tpl.FormattedText,
			OnToken: func(token string) bool {
				t.Logf("<< %s", token)
				return true
			},
		})
		if err != nil {
			t.Fatalf("Failed to generate text: %v", err)
		}
		t.Logf("Turn %d reply: %s", i+1, stream.FullText)
		history = append(history, LlmChatMessage{Role: LLMRoleAssistant, Content: stream.FullText})
	}
}

func TestLLMSaveKVCache(t *testing.T) {
	_, err := llm.SaveKVCache(LlmSaveKVCacheInput{Path: "./test_kv_cache.bin"})
	if err != nil {
		t.Errorf("SaveKVCache failed: %v", err)
		return
	}

	t.Logf("SaveKVCache completed successfully")
}

func TestLLMLoadKVCache(t *testing.T) {
	_, err := llm.LoadKVCache(LlmLoadKVCacheInput{Path: "./test_kv_cache.bin"})
	if err != nil {
		t.Errorf("LoadKVCache failed: %v", err)
		return
	}

	_ = os.Remove("./test_kv_cache.bin")

	t.Logf("LoadKVCache completed successfully")
}
