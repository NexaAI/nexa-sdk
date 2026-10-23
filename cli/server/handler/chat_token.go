// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"strings"

	"github.com/openai/openai-go/v3"

	"github.com/qualcomm/GenieX/cli/internal/thinkfsm"
	"github.com/qualcomm/GenieX/cli/server/utils"
)

// "" / "none" keep thinking inline in content (default); others separate it.
func reasoningSeparated(format string) bool {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "none":
		return false
	default:
		return true
	}
}

// tokenClass classifies one token without choosing a display or scanner policy.
type tokenClass func(string) thinkfsm.Event

// plainClass passes tokens through verbatim, keeping the default byte-identical.
func plainClass(token string) thinkfsm.Event {
	return thinkfsm.Event{Kind: thinkfsm.Content, Text: token}
}

func reasoningClass() tokenClass {
	return thinkfsm.New().Feed
}

func sink(class tokenClass, content, reasoning *strings.Builder) func(string) bool {
	return func(token string) bool {
		ev := class(token)
		switch ev.Kind {
		case thinkfsm.Content:
			content.WriteString(ev.Text)
		case thinkfsm.Thinking:
			reasoning.WriteString(ev.Text)
		}
		return true
	}
}

func sinkWithToolCalls(class tokenClass, content, reasoning *strings.Builder, tools string) (func(string) bool, func() []openai.ChatCompletionMessageFunctionToolCallFunction) {
	parser := utils.NewToolCallParser(utils.ToolParameterTypesFromTools(tools))
	var calls []openai.ChatCompletionMessageFunctionToolCallFunction
	write := func(events []utils.ToolCallEvent) {
		for _, ev := range events {
			switch ev.Kind {
			case thinkfsm.Content:
				content.WriteString(ev.Text)
			case thinkfsm.Thinking:
				reasoning.WriteString(ev.Text)
			case thinkfsm.ToolCall:
				calls = append(calls, ev.Call)
			}
		}
	}
	onToken := func(token string) bool {
		write(parser.Feed(class(token)))
		return true
	}
	finish := func() []openai.ChatCompletionMessageFunctionToolCallFunction {
		write(parser.Finish())
		return calls
	}
	return onToken, finish
}
