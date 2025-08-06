package nexa_sdk

import (
	"log/slog"
	"testing"
)

var (
	vlm         *VLM
	vlmMessages = []VlmChatMessage{
		{
			Role: "system",
			Contents: []VlmContent{
				{
					Type: "text",
					Text: "You are a helpful assistant that can see images.",
				},
			},
		},
		{
			Role: "user",
			Contents: []VlmContent{
				{
					Type: "text",
					Text: "What do you see in this image?",
				},
				{
					Type: "image",
					Text: "modelfiles/assets/test_image.png",
				},
			},
		},
	}
)

func initVLM() {
	slog.Debug("initVLM called")

	var err error

	input := VlmCreateInput{
		ModelPath:  "modelfiles/llama_cpp/SmolVLM-256M-Instruct-Q8_0.gguf",
		MmprojPath: "modelfiles/llama_cpp/mmproj-SmolVLM-256M-Instruct-Q8_0.gguf",
		Config: ModelConfig{
			NCtx:    512,
			NSeqMax: 64,
		},

		PluginID: "llama_cpp",
	}

	vlm, err = NewVLM(input)
	if err != nil {
		panic("Error creating VLM: " + err.Error())
	}
}

func deinitVLM() {
	if vlm != nil {
		vlm.Destroy()
	}
}

func TestVLMReset(t *testing.T) {
	// Skip if VLM is not available
	if vlm == nil {
		t.Skip("VLM not initialized, skipping test")
	}

	err := vlm.Reset()
	if err != nil {
		t.Errorf("Reset failed: %v", err)
		return
	}

	t.Logf("Reset completed successfully")
}

func TestVLMApplyChatTemplate(t *testing.T) {
	// Skip if VLM is not available
	if vlm == nil {
		t.Skip("VLM not initialized, skipping test")
	}

	input := VlmApplyChatTemplateInput{
		Messages: vlmMessages,
	}

	output, err := vlm.ApplyChatTemplate(input)
	if err != nil {
		t.Errorf("ApplyChatTemplate failed: %v", err)
		return
	}

	t.Logf("ApplyChatTemplate: %s", output.FormattedText)
	vlm.Reset()
}

func TestVLMGenerate(t *testing.T) {
	// Skip if VLM is not available
	if vlm == nil {
		t.Skip("VLM not initialized, skipping test")
	}

	tpl, err := vlm.ApplyChatTemplate(VlmApplyChatTemplateInput{
		Messages: vlmMessages,
	})
	if err != nil {
		t.Fatalf("Failed to apply chat template: %v", err)
	}
	cfg := &GenerationConfig{
		MaxTokens: 100,
	}

	for _, content := range vlmMessages[1].Contents {
		if content.Type == "image" {
			cfg.ImagePaths = append(cfg.ImagePaths, content.Text)
		}
	}

	stream, err := vlm.Generate(VlmGenerateInput{
		PromptUTF8: tpl.FormattedText,
		Config:     cfg,
		OnToken: func(token string) bool {
			// t.Logf("<< %s", token)
			return true
		},
	})
	if err != nil {
		t.Fatalf("Failed to generate text: %v", err)
	}

	t.Logf("GenerateStream: %s", stream.FullText)
	vlm.Reset()
}

func TestVLMGenerateMulti(t *testing.T) {
	// Skip if VLM is not available
	if vlm == nil {
		t.Skip("VLM not initialized, skipping test")
	}

	// Define test rounds
	testRounds := []struct {
		name    string
		message VlmChatMessage
	}{
		{
			"Initial conversation",
			VlmChatMessage{
				Role: "user",
				Contents: []VlmContent{
					{
						Type: "text",
						Text: "What do you see in this image?",
					},
					{
						Type: "image",
						Text: "modelfiles/assets/test_image.png",
					},
				},
			},
		},
		{
			"Repeat the number 42 three times",
			VlmChatMessage{
				Role: "user",
				Contents: []VlmContent{
					{
						Type: "text",
						Text: "Please repeat the number 42 three times.",
					},
				},
			},
		},
		{
			"What colors do you see in this image? Please describe the color scheme.",
			VlmChatMessage{
				Role: "user",
				Contents: []VlmContent{
					{
						Type: "text",
						Text: "What colors do you see in this image? Please describe the color scheme.",
					},
					{
						Type: "image",
						Text: "modelfiles/assets/test_image.png",
					},
				},
			},
		},
	}

	// Set prompt
	history := []VlmChatMessage{
		{
			Role: "system",
			Contents: []VlmContent{
				{
					Type: "text",
					Text: "You are a helpful assistant that can see images.",
				},
			},
		},
	}

	for i, round := range testRounds {
		t.Logf("=== Round %d: %s ===", i+1, round.name)

		// Add user message to history
		history = append(history, round.message)

		// Apply chat template to get formatted text
		formated, err := vlm.ApplyChatTemplate(VlmApplyChatTemplateInput{
			Messages: history,
		})
		if err != nil {
			t.Fatalf("Failed to apply chat template for round %d: %v", i+1, err)
		}

		cfg := &GenerationConfig{MaxTokens: 100}

		// Add image paths to generation config if this round has image
		for _, content := range round.message.Contents {
			if content.Type == "image" {
				cfg.ImagePaths = append(cfg.ImagePaths, content.Text)
			}
		}

		stream, err := vlm.Generate(VlmGenerateInput{
			PromptUTF8: formated.FormattedText,
			Config:     cfg,
			OnToken: func(token string) bool {
				return true
			},
		})
		if err != nil {
			t.Fatalf("Failed to generate text for round %d: %v", i+1, err)
		}

		t.Logf("Round %d response: %s", i+1, stream.FullText)

		// Add assistant response to conversation history
		history = append(history, VlmChatMessage{
			Role: VlmRoleAssistant,
			Contents: []VlmContent{
				{
					Type: VlmContentTypeText,
					Text: stream.FullText,
				},
			},
		})
	}

	t.Logf("Multi-round conversation completed successfully")
	vlm.Reset()
}
