// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"

	geniex_sdk "github.com/qualcomm/GenieX/bindings/go"
	"github.com/qualcomm/GenieX/cli/internal/config"
	"github.com/qualcomm/GenieX/cli/internal/decision"
	"github.com/qualcomm/GenieX/cli/server/service"
	"github.com/qualcomm/GenieX/cli/server/types"
	"github.com/qualcomm/GenieX/cli/server/utils"
)

type decisionsRequest struct {
	Model      string                                  `json:"model"`
	Input      json.RawMessage                         `json:"input"`
	Questions  []openai.DecisionNewParamsQuestionUnion `json:"questions"`
	NCtx       int32                                   `json:"nctx"`
	Ngl        int32                                   `json:"ngl"`
	Compute    string                                  `json:"compute"`
	VitCompute string                                  `json:"vit_compute"`
	PowerMode  string                                  `json:"power_mode"`
}

type decisionChoice = decision.Choice

type decisionLevel = decision.Level

type decisionQuestion struct {
	Kind         decision.Kind
	Name         string
	Instructions string
	Choices      []decisionChoice
	Levels       []decisionLevel
}

type decisionProbability = decision.Probability

type decisionAnswer struct {
	Type          string                `json:"type"`
	Name          *string               `json:"name"`
	Probability   *float64              `json:"probability,omitempty"`
	Choice        any                   `json:"choice,omitempty"`
	Probabilities []decisionProbability `json:"probabilities,omitempty"`
	Confidence    *float64              `json:"confidence,omitempty"`
	Score         *float64              `json:"score,omitempty"`
}

type decisionsResponse struct {
	Answers []decisionAnswer     `json:"answers"`
	Model   string               `json:"model"`
	Usage   openai.DecisionUsage `json:"usage"`
}

type compiledDecisionInput struct {
	Text      string
	ImageURLs []string
	AudioURLs []string
}

type decisionInputMessage struct {
	Role    string          `json:"role"`
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

type decisionInputPart struct {
	Type       string `json:"type"`
	Text       string `json:"text"`
	ImageURL   string `json:"image_url"`
	InputAudio *struct {
		Data   string `json:"data"`
		Format string `json:"format"`
	} `json:"input_audio"`
}

// Flatten supported user messages to one state string and separate media arrays.
func compileDecisionInput(input json.RawMessage) (compiledDecisionInput, error) {
	var text string
	if json.Unmarshal(input, &text) == nil {
		if strings.TrimSpace(text) == "" {
			return compiledDecisionInput{}, errors.New("input must not be empty")
		}
		return compiledDecisionInput{Text: text}, nil
	}
	var inputMessages []decisionInputMessage
	if json.Unmarshal(input, &inputMessages) != nil || len(inputMessages) == 0 {
		return compiledDecisionInput{}, errors.New("input must be a nonempty string or an array of user messages")
	}

	var result compiledDecisionInput
	var messages []string
	for i, message := range inputMessages {
		if message.Role != "" && message.Role != "user" {
			return compiledDecisionInput{}, fmt.Errorf("input message %d: only user messages are supported", i)
		}
		if message.Type != "" && message.Type != "message" {
			return compiledDecisionInput{}, fmt.Errorf("input message %d: type must be message", i)
		}
		var contentText string
		if json.Unmarshal(message.Content, &contentText) == nil {
			messages = append(messages, contentText)
			continue
		}
		var contentParts []decisionInputPart
		if json.Unmarshal(message.Content, &contentParts) != nil || len(contentParts) == 0 {
			return compiledDecisionInput{}, fmt.Errorf("input message %d: content must be text or media parts", i)
		}
		var parts []string
		for j, part := range contentParts {
			switch part.Type {
			case "input_text":
				parts = append(parts, part.Text)
			case "input_image":
				if strings.TrimSpace(part.ImageURL) == "" {
					return compiledDecisionInput{}, fmt.Errorf("input message %d part %d: image_url must not be empty", i, j)
				}
				result.ImageURLs = append(result.ImageURLs, part.ImageURL)
			case "input_audio":
				if part.InputAudio == nil || strings.TrimSpace(part.InputAudio.Data) == "" {
					return compiledDecisionInput{}, fmt.Errorf("input message %d part %d: input_audio.data must not be empty", i, j)
				}
				// Reuse the URI loader by wrapping chat-style base64 audio as a data URI.
				data := part.InputAudio.Data
				if !strings.HasPrefix(data, "data:") && part.InputAudio.Format != "" {
					data = "data:audio/" + part.InputAudio.Format + ";base64," + data
				}
				result.AudioURLs = append(result.AudioURLs, data)
			default:
				return compiledDecisionInput{}, fmt.Errorf("input message %d part %d: only input_text, input_image, and input_audio parts are supported", i, j)
			}
		}
		if len(parts) > 0 {
			messages = append(messages, strings.Join(parts, "\n"))
		}
	}
	if len(result.ImageURLs) > 8 || len(result.AudioURLs) > 8 {
		return compiledDecisionInput{}, errors.New("input supports at most 8 images and 8 audio files")
	}
	result.Text = strings.Join(messages, "\n")
	if strings.TrimSpace(result.Text) == "" && len(result.ImageURLs) == 0 && len(result.AudioURLs) == 0 {
		return compiledDecisionInput{}, errors.New("input must contain text, an image, or audio")
	}
	return result, nil
}

func decisionQuestionName(question openai.DecisionNewParamsQuestionUnion) (string, bool, error) {
	var name param.Opt[string]
	switch {
	case question.OfPredicate != nil:
		name = question.OfPredicate.Name
	case question.OfChoice != nil:
		name = question.OfChoice.Name
	case question.OfScore != nil:
		name = question.OfScore.Name
	default:
		return "", false, errors.New("question type must be predicate, choice, or score")
	}
	if param.IsNull(name) {
		return "", false, errors.New("question name must be a string when provided")
	}
	if !name.Valid() {
		return "", false, nil
	}
	if strings.TrimSpace(name.Value) == "" {
		return "", false, errors.New("question name must not be empty")
	}
	return name.Value, true, nil
}

// Validate the OpenAI question union before converting it to SDK-neutral fields.
func compileDecisionQuestions(questions []openai.DecisionNewParamsQuestionUnion) ([]decisionQuestion, error) {
	if len(questions) == 0 || len(questions) > 64 {
		return nil, errors.New("questions must contain 1–64 questions")
	}

	compiled := make([]decisionQuestion, 0, len(questions))
	usedNames := make(map[string]struct{}, len(questions))
	for i, question := range questions {
		name, hasName, err := decisionQuestionName(question)
		if err != nil {
			return nil, fmt.Errorf("question %d: %w", i, err)
		}
		if hasName {
			if _, exists := usedNames[name]; exists {
				return nil, fmt.Errorf("question %d: duplicate question name %q", i, name)
			}
			usedNames[name] = struct{}{}
		}

		var field decisionQuestion
		switch {
		case question.OfPredicate != nil:
			field.Kind = decision.KindPredicate
			field.Instructions = question.OfPredicate.Instructions
		case question.OfChoice != nil:
			field.Kind = decision.KindChoice
			field.Instructions = question.OfChoice.Instructions
		case question.OfScore != nil:
			field.Kind = decision.KindScore
			field.Instructions = question.OfScore.Instructions
		default:
			return nil, fmt.Errorf("question %d: type must be predicate, choice, or score", i)
		}
		if strings.TrimSpace(field.Instructions) == "" {
			return nil, fmt.Errorf("question %d: instructions must be a nonempty string", i)
		}
		field.Name = name

		switch {
		case question.OfPredicate != nil:
			field.Choices = []decisionChoice{{Value: true}, {Value: false}}
		case question.OfChoice != nil:
			choices := question.OfChoice.Choices
			if len(choices) < 2 || len(choices) > 255 {
				return nil, fmt.Errorf("question %q: choices must contain 2–255 options", field.Name)
			}
			// Boolean false and the string "false" are distinct choice values.
			seenValues := make(map[any]struct{}, len(choices))
			for j, choice := range choices {
				value := choice.Value
				switch {
				case value.OfString.Valid():
					if strings.TrimSpace(value.OfString.Value) == "" {
						return nil, fmt.Errorf("question %q choice %d: value must not be empty", field.Name, j)
					}
					field.Choices = append(field.Choices, decisionChoice{Value: value.OfString.Value})
				case value.OfBool.Valid():
					field.Choices = append(field.Choices, decisionChoice{Value: value.OfBool.Value})
				default:
					return nil, fmt.Errorf("question %q choice %d: value must be a string or boolean", field.Name, j)
				}
				if _, exists := seenValues[field.Choices[j].Value]; exists {
					return nil, fmt.Errorf("question %q choice %d: duplicate value", field.Name, j)
				}
				seenValues[field.Choices[j].Value] = struct{}{}
				if param.IsNull(choice.Description) {
					return nil, fmt.Errorf("question %q choice %d: description must be a string when provided", field.Name, j)
				}
				if choice.Description.Valid() {
					field.Choices[j].Description = choice.Description.Value
				}
			}
		case question.OfScore != nil:
			levels := question.OfScore.Levels
			if len(levels) < 2 || len(levels) > 10 {
				return nil, fmt.Errorf("question %q: levels must contain 2–10 items", field.Name)
			}
			for j, level := range levels {
				if strings.TrimSpace(level.Label) == "" {
					return nil, fmt.Errorf("question %q level %d: label must not be empty", field.Name, j)
				}
				if param.IsNull(level.Description) {
					return nil, fmt.Errorf("question %q level %d: description must be a string when provided", field.Name, j)
				}
				description := level.Label
				if level.Description.Valid() {
					description = level.Description.Value
				}
				field.Levels = append(field.Levels, decisionLevel{Label: level.Label, Description: description})
			}
		}
		compiled = append(compiled, field)
	}

	return compiled, nil
}

func (question decisionQuestion) modelQuestion() decision.Question {
	return decision.Question{
		ID: question.Name, Kind: question.Kind, Instructions: question.Instructions,
		Choices: question.Choices, Levels: question.Levels,
	}
}

func decisionAnswerFor(question decisionQuestion, logits []float32, temperature float32) (decisionAnswer, error) {
	result, err := decision.Answer(question.modelQuestion(), logits, float64(temperature))
	if err != nil {
		return decisionAnswer{}, err
	}
	return decisionAnswerFromResult(question, result), nil
}

func decisionAnswerFromResult(question decisionQuestion, result decision.Result) decisionAnswer {
	answer := decisionAnswer{Type: string(question.Kind)}
	if question.Name != "" {
		answer.Name = &question.Name
	}
	answer.Probability = result.Probability
	answer.Choice = result.Choice
	answer.Probabilities = result.Probabilities
	answer.Confidence = result.Confidence
	answer.Score = result.Score
	return answer
}

type decisionsModel interface {
	Decide(geniex_sdk.DecideInput) (geniex_sdk.DecideOutput, error)
}

func evaluateDecision(model decisionsModel, input geniex_sdk.DecideInput, question decisionQuestion) (decisionAnswer, int, error) {
	scored, err := model.Decide(input)
	if err != nil {
		return decisionAnswer{}, 0, err
	}
	// Avoid reapplying temperature to SDK-calibrated LEV probabilities.
	if len(scored.Probabilities) > 0 {
		result, err := decision.AnswerProbabilities(question.modelQuestion(), scored.Probabilities)
		if err != nil {
			return decisionAnswer{}, 0, err
		}
		return decisionAnswerFromResult(question, result), scored.InputTokens, nil
	}
	answer, err := decisionAnswerFor(question, scored.Logits, scored.Temperature)
	if err != nil {
		return decisionAnswer{}, 0, err
	}
	return answer, scored.InputTokens, nil
}

func decisionErrorStatus(code int32) int {
	switch code {
	case int32(geniex_sdk.ErrCommonInvalidInput), int32(geniex_sdk.ErrCommonParamNotSupported),
		int32(geniex_sdk.ErrLlmTokenizationContextLength), int32(geniex_sdk.ErrLlmGenerationPromptTooLong),
		int32(geniex_sdk.ErrVlmImageLoad), int32(geniex_sdk.ErrVlmAudioLoad):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func defaultDecisionsRequest() decisionsRequest {
	cfg := config.Get()
	return decisionsRequest{
		Model: cfg.Model, NCtx: cfg.NCtx, Ngl: cfg.Ngl, Compute: cfg.Compute,
		VitCompute: cfg.VitCompute, PowerMode: cfg.PowerMode,
	}
}

func Decisions(c *gin.Context) {
	req := defaultDecisionsRequest()
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	if err := c.ShouldBindJSON(&req); err != nil {
		var tooLarge *http.MaxBytesError
		status := http.StatusUnprocessableEntity
		if errors.As(err, &tooLarge) {
			status = http.StatusRequestEntityTooLarge
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Model) == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "model is required"})
		return
	}
	input, err := compileDecisionInput(req.Input)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	questions, err := compileDecisionQuestions(req.Questions)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	modelQuestions := make([]decision.Question, len(questions))
	for i, question := range questions {
		modelQuestions[i] = question.modelQuestion()
	}
	// Keep the full schema for templates such as Nimble while scoring each question separately.
	modelOptions, err := decision.Questions(modelQuestions)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	paths, err := geniex_sdk.ModelGetPaths(req.Model)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if paths.RuntimeID != geniex_sdk.RuntimeLlamaCpp ||
		(paths.ModelType != geniex_sdk.ModelTypeLLM && paths.ModelType != geniex_sdk.ModelTypeVLM) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "decisions requires a llama_cpp LLM or VLM model"})
		return
	}
	if (len(input.ImageURLs) > 0 || len(input.AudioURLs) > 0) && paths.ModelType != geniex_sdk.ModelTypeVLM {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input_image and input_audio require a VLM model"})
		return
	}
	modelParam, err := service.ResolveModelParam(paths.RuntimeID, paths.ModelName, req.NCtx, req.Ngl, req.Compute, req.VitCompute, req.PowerMode, service.Chipset(), types.SpecParam{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var model decisionsModel
	var images, audios []string
	if paths.ModelType == geniex_sdk.ModelTypeVLM {
		// Media files must outlive all per-question SDK forwards in this request.
		for _, url := range input.ImageURLs {
			image, err := utils.SaveURIToTempFile(url)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			images = append(images, image)
			defer os.Remove(image)
		}
		for _, url := range input.AudioURLs {
			audio, err := utils.SaveURIToTempFile(url)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			audios = append(audios, audio)
			defer os.Remove(audio)
		}
		cacheKey := input.Text + strings.Join(input.ImageURLs, "\x00") + strings.Join(input.AudioURLs, "\x00")
		acquired, err := service.KeepAliveGet[geniex_sdk.VLM](req.Model, modelParam, utils.HashText(cacheKey))
		if writeKeepAliveError(c, err) {
			return
		}
		if len(images) > 0 || len(audios) > 0 {
			caps, err := acquired.Model.Capabilities()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if len(images) > 0 && !caps.SupportsVision {
				c.JSON(http.StatusBadRequest, gin.H{"error": "input_image requires a VLM with vision support"})
				return
			}
			if len(audios) > 0 && !caps.SupportsAudio {
				c.JSON(http.StatusBadRequest, gin.H{"error": "input_audio requires a VLM with audio support"})
				return
			}
		}
		model = acquired.Model
	} else {
		acquired, err := service.KeepAliveGet[geniex_sdk.LLM](req.Model, modelParam, utils.HashTokens(nil))
		if writeKeepAliveError(c, err) {
			return
		}
		model = acquired.Model
	}

	// Sum tokens from each independent logits pass for the response usage.
	answers := make([]decisionAnswer, 0, len(questions))
	var inputTokens int64
	for index, question := range questions {
		answer, tokens, err := evaluateDecision(model, geniex_sdk.DecideInput{
			State: input.Text, Questions: modelOptions, QuestionIndex: index,
			ImagePaths: images, AudioPaths: audios,
		}, question)
		if err != nil {
			code := geniex_sdk.SDKErrorCode(err)
			status := decisionErrorStatus(code)
			body := gin.H{"error": err.Error()}
			if code != -1 || paths.ModelType == geniex_sdk.ModelTypeVLM {
				body["code"] = code
			}
			c.JSON(status, body)
			return
		}
		inputTokens += int64(tokens)
		answers = append(answers, answer)
	}

	c.JSON(http.StatusOK, decisionsResponse{
		Answers: answers,
		Model:   req.Model,
		Usage: openai.DecisionUsage{
			InputTokens:         inputTokens,
			OutputTokens:        0,
			InputTokensDetails:  openai.DecisionUsageInputTokensDetails{},
			OutputTokensDetails: openai.DecisionUsageOutputTokensDetails{},
			TotalTokens:         inputTokens,
		},
	})
}
