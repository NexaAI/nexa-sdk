// Copyright (c) 2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"

	geniex_sdk "github.com/qualcomm/GenieX/bindings/go"
	"github.com/qualcomm/GenieX/cli/internal/config"
	"github.com/qualcomm/GenieX/cli/server/service"
	"github.com/qualcomm/GenieX/cli/server/types"
	"github.com/qualcomm/GenieX/cli/server/utils"
)

type CompletionNewParams openai.CompletionNewParams

type CompletionRequest struct {
	CompletionNewParams
	Stream bool `json:"stream"`

	NCtx       int32  `json:"nctx"`
	Ngl        int32  `json:"ngl"`
	Compute    string `json:"compute"`
	VitCompute string `json:"vit_compute"`
	PowerMode  string `json:"power_mode"`

	TopK              int32   `json:"top_k"`
	MinP              float32 `json:"min_p"`
	RepetitionPenalty float32 `json:"repetition_penalty"`

	PromptEmbeds    string `json:"prompt_embeds"`
	PromptEmbedsDim int32  `json:"prompt_embeds_dim"`
}

func defaultCompletionRequest() CompletionRequest {
	cfg := config.Get()
	return CompletionRequest{
		CompletionNewParams: CompletionNewParams{
			// OpenAI's documented default for the legacy completions endpoint.
			MaxTokens: param.NewOpt[int64](16),
		},
		NCtx:              cfg.NCtx,
		Ngl:               cfg.Ngl,
		Compute:           cfg.Compute,
		VitCompute:        cfg.VitCompute,
		PowerMode:         cfg.PowerMode,
		RepetitionPenalty: 1.0,
	}
}

func completionPrompt(u openai.CompletionNewParamsPromptUnion) (string, error) {
	switch {
	case !param.IsOmitted(u.OfString):
		if u.OfString.Value == "" {
			return "", errors.New("prompt must be non-empty")
		}
		return u.OfString.Value, nil
	case u.OfArrayOfStrings != nil:
		if len(u.OfArrayOfStrings) != 1 || u.OfArrayOfStrings[0] == "" {
			return "", errors.New("only a single non-empty prompt is supported")
		}
		return u.OfArrayOfStrings[0], nil
	case u.OfArrayOfTokens != nil || u.OfArrayOfTokenArrays != nil:
		return "", errors.New("token prompts are not supported; send a string prompt")
	default:
		return "", errors.New("prompt is required")
	}
}

func completionPromptSet(u openai.CompletionNewParamsPromptUnion) bool {
	return !param.IsOmitted(u.OfString) || u.OfArrayOfStrings != nil || u.OfArrayOfTokens != nil || u.OfArrayOfTokenArrays != nil
}

// completionPromptEmbeds decodes prompt_embeds: base64 of little-endian float32,
// row-major [n_rows, prompt_embeds_dim]. It returns nil when the field is unset.
func completionPromptEmbeds(req CompletionRequest) ([]float32, error) {
	if req.PromptEmbeds == "" {
		return nil, nil
	}
	if completionPromptSet(req.Prompt) {
		return nil, errors.New("prompt and prompt_embeds are mutually exclusive")
	}
	if req.Echo.Valid() && req.Echo.Value {
		return nil, errors.New("echo is not supported with prompt_embeds")
	}
	if req.PromptEmbedsDim <= 0 {
		return nil, errors.New("prompt_embeds_dim must be > 0 when prompt_embeds is set")
	}
	raw, err := base64.StdEncoding.DecodeString(req.PromptEmbeds)
	if err != nil {
		return nil, errors.New("prompt_embeds is not valid base64")
	}
	if len(raw) == 0 || len(raw)%4 != 0 {
		return nil, errors.New("prompt_embeds must be a non-empty array of float32")
	}
	embd := make([]float32, len(raw)/4)
	for i := range embd {
		embd[i] = math.Float32frombits(binary.LittleEndian.Uint32(raw[i*4:]))
	}
	if len(embd)%int(req.PromptEmbedsDim) != 0 {
		return nil, errors.New("prompt_embeds length must be a multiple of prompt_embeds_dim")
	}
	return embd, nil
}

var promptEmbedsSeq atomic.Uint64

func completionStop(u openai.CompletionNewParamsStopUnion) []string {
	if !param.IsOmitted(u.OfString) {
		if u.OfString.Value == "" {
			return nil
		}
		return []string{u.OfString.Value}
	}
	return u.OfStringArray
}

func completionUnsupported(p CompletionNewParams) error {
	switch {
	case p.Suffix.Valid() && p.Suffix.Value != "":
		return errors.New("suffix is not supported; send a prompt containing the model's FIM template instead")
	case p.N.Valid() && p.N.Value > 1:
		return errors.New("n > 1 is not supported")
	case p.BestOf.Valid() && p.BestOf.Value > 1:
		return errors.New("best_of > 1 is not supported")
	case p.Logprobs.Valid():
		return errors.New("logprobs is not supported")
	case len(p.LogitBias) > 0:
		return errors.New("logit_bias is not supported")
	default:
		return nil
	}
}

// Completions serves POST /v1/completions: the prompt reaches the model verbatim
// — no chat template — so FIM-style clients get a raw continuation back.
func Completions(c *gin.Context) {
	req := defaultCompletionRequest()
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	logReq := req
	logReq.PromptEmbeds = ""
	slog.Info("Completions", "param", logReq, "prompt_embeds_bytes", len(req.PromptEmbeds))
	embd, err := completionPromptEmbeds(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	var prompt string
	if embd == nil {
		prompt, err = completionPrompt(req.Prompt)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
	}
	if err := completionUnsupported(req.CompletionNewParams); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	paths, err := geniex_sdk.ModelGetPaths(string(req.Model))
	if err != nil {
		slog.Error("Failed to resolve model paths", "model", req.Model, "error", err)
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if paths.ModelType != geniex_sdk.ModelTypeLLM {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "completions is only supported for LLM models"})
		return
	}

	modelParam, err := service.ResolveModelParam(paths.RuntimeID, paths.ModelName, req.NCtx, req.Ngl, req.Compute, req.VitCompute, req.PowerMode, service.Chipset(), types.SpecParam{})
	if err != nil {
		slog.Error("Failed to resolve model params", "model", req.Model, "error", err)
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	// Automatically adjust NCtx if MaxTokens is larger (llama_cpp only — QAIRT
	// does not use NCtx and the 0-default must not be overwritten for non-llama_cpp plugins).
	if paths.RuntimeID == geniex_sdk.RuntimeLlamaCpp && modelParam.NCtx < int32(req.MaxTokens.Value) {
		slog.Debug("Adjust NCtx to MaxTokens", "from", modelParam.NCtx, "to", req.MaxTokens.Value)
		modelParam.NCtx = int32(req.MaxTokens.Value)
	}

	session := utils.HashText(prompt)
	if embd != nil {
		session = utils.SessionKey{promptEmbedsSeq.Add(1)}
	}
	acquired, err := service.KeepAliveGet[geniex_sdk.LLM](
		string(req.Model),
		modelParam,
		session,
	)
	if writeKeepAliveError(c, err) {
		return
	}
	p := acquired.Model

	if embd != nil {
		info, err := p.GetModelInfo()
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": geniex_sdk.SDKErrorCode(err)})
			return
		}
		if info.EmbdDim == 0 {
			c.JSON(http.StatusBadRequest, map[string]any{"error": "prompt_embeds is not supported by this model"})
			return
		}
		if req.PromptEmbedsDim != info.EmbdDim {
			c.JSON(http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("prompt_embeds_dim (%d) does not match the model's embedding width (%d)", req.PromptEmbedsDim, info.EmbdDim)})
			return
		}
	}

	genConfig := &geniex_sdk.GenerationConfig{
		MaxTokens: int32(req.MaxTokens.Value),
		Stop:      completionStop(req.Stop),
		SamplerConfig: &geniex_sdk.SamplerConfig{
			Temperature:       float32(req.Temperature.Value),
			TopP:              float32(req.TopP.Value),
			TopK:              req.TopK,
			MinP:              req.MinP,
			RepetitionPenalty: req.RepetitionPenalty,
			PresencePenalty:   float32(req.PresencePenalty.Value),
			FrequencyPenalty:  float32(req.FrequencyPenalty.Value),
			Seed:              int32(req.Seed.Value),
		},
	}
	echo := ""
	if req.Echo.Valid() && req.Echo.Value {
		echo = prompt
	}

	// ---- generate: stream SSE chunks or write one blocking response ----
	if req.Stream {
		// streaming
		var stopGen atomic.Bool
		dataCh := make(chan string)

		var (
			profile geniex_sdk.ProfileData
			genErr  error
			wg      sync.WaitGroup
		)

		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := p.Generate(geniex_sdk.LlmGenerateInput{
				PromptUTF8:   prompt,
				InputEmbd:    embd,
				InputEmbdDim: req.PromptEmbedsDim,
				OnToken: func(token string) bool {
					if stopGen.Load() {
						return false
					}
					dataCh <- token
					return true
				},
				Config: genConfig,
			})
			genErr = err
			if out != nil {
				profile = out.ProfileData
			}
			close(dataCh)
		}()

		wait := func() error { wg.Wait(); return genErr }
		streamCompletion(c, dataCh, wait, req.StreamOptions.IncludeUsage.Value, echo, &profile)

		stopGen.Store(true)
		for range dataCh {
		}
	} else {
		// blocking
		out, err := p.Generate(geniex_sdk.LlmGenerateInput{
			PromptUTF8:   prompt,
			InputEmbd:    embd,
			InputEmbdDim: req.PromptEmbedsDim,
			Config:       genConfig,
		})
		// A prompt that never fit is a 400; a window exhausted mid-generation is a
		// normal truncated completion (finish_reason=length), handled below.
		if errors.Is(err, geniex_sdk.ErrLlmGenerationPromptTooLong) {
			writeCompletionPromptTooLong(c, out.ProfileData)
			return
		}
		if err != nil && !errors.Is(err, geniex_sdk.ErrLlmTokenizationContextLength) {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": geniex_sdk.SDKErrorCode(err)})
			return
		}
		logProfile(out.ProfileData)
		writeCompletionResponse(c, echo+out.FullText, out.ProfileData)
	}
}

// ---- streaming SSE chunks ----

type completionStreamChoice struct {
	Index        int64   `json:"index"`
	Text         string  `json:"text"`
	FinishReason *string `json:"finish_reason"` // null until the final chunk
}

type completionStreamChunk struct {
	Object  string                   `json:"object"`
	Choices []completionStreamChoice `json:"choices"`
	Usage   *openai.CompletionUsage  `json:"usage,omitempty"`
	Timings *timings                 `json:"timings,omitempty"`
}

const completionObject = "text_completion"

func completionTextChunk(text string) completionStreamChunk {
	return completionStreamChunk{
		Object:  completionObject,
		Choices: []completionStreamChoice{{Text: text}},
	}
}

func completionFinishChunk(reason string) completionStreamChunk {
	return completionStreamChunk{
		Object:  completionObject,
		Choices: []completionStreamChoice{{FinishReason: &reason}},
	}
}

func completionUsageChunk(u openai.CompletionUsage, t timings) completionStreamChunk {
	return completionStreamChunk{Object: completionObject, Choices: []completionStreamChoice{}, Usage: &u, Timings: &t}
}

// profile is read only after wait() returns, when generation has filled it.
func streamCompletion(c *gin.Context, dataCh <-chan string, wait func() error, includeUsage bool, echo string, profile *geniex_sdk.ProfileData) {
	first := echo != ""
	c.Stream(func(w io.Writer) bool {
		if first {
			first = false
			c.SSEvent("", completionTextChunk(echo))
			return true
		}
		r, ok := <-dataCh
		if ok {
			c.SSEvent("", completionTextChunk(r))
			return true
		}
		// A context window exhausted mid-stream is a normal truncated completion:
		// fall through to the finish chunk. Other errors become an error event.
		if err := wait(); err != nil && !errors.Is(err, geniex_sdk.ErrLlmTokenizationContextLength) {
			c.SSEvent("", map[string]any{"error": err.Error(), "code": geniex_sdk.SDKErrorCode(err)})
			return false
		}
		logProfile(*profile)
		c.SSEvent("", completionFinishChunk(mapFinishReason(profile.StopReason)))
		if includeUsage {
			c.SSEvent("", completionUsageChunk(profile2Usage(*profile), profile2Timings(*profile)))
		}
		c.SSEvent("", "[DONE]")
		return false
	})
}

// ---- blocking response ----

type completionChoice struct {
	Index        int64  `json:"index"`
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason"`
}

type completionResponse struct {
	Object  string                 `json:"object"`
	Choices []completionChoice     `json:"choices"`
	Usage   openai.CompletionUsage `json:"usage"`
	Timings timings                `json:"timings"`
}

func writeCompletionResponse(c *gin.Context, text string, profile geniex_sdk.ProfileData) {
	c.JSON(http.StatusOK, completionResponse{
		Object: completionObject,
		Choices: []completionChoice{{
			Text:         text,
			FinishReason: mapFinishReason(profile.StopReason),
		}},
		Usage:   profile2Usage(profile),
		Timings: profile2Timings(profile),
	})
}

// OpenAI's 400 body for a prompt that is longer than the context window. No
// partial output exists (nothing was generated), so this returns only the error.
func writeCompletionPromptTooLong(c *gin.Context, profile geniex_sdk.ProfileData) {
	c.JSON(http.StatusBadRequest, map[string]any{
		"error": map[string]any{
			"message": "prompt is longer than the model's context window",
			"type":    "invalid_request_error",
			"code":    "context_length_exceeded",
		},
		"usage": profile2Usage(profile),
	})
}
