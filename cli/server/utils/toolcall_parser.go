// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package utils

import (
	"strings"

	"github.com/bytedance/sonic"

	"github.com/qualcomm/GenieX/cli/internal/thinkfsm"
)

// ToolCallEvent contains display text or an already parsed function call.
type ToolCallEvent struct {
	Kind thinkfsm.EventKind
	Text string
	Call toolCallFn
}

// ToolCallParser consumes classified events, not protocol tokens.
type ToolCallParser struct {
	scanner *ToolCallScanner
	header  strings.Builder
	args    strings.Builder
	inBody  bool
}

func NewToolCallParser(types map[string]map[string]string) *ToolCallParser {
	return &ToolCallParser{scanner: NewToolCallScanner(types)}
}

func parsedEvents(kind thinkfsm.EventKind, text string, calls []toolCallFn) []ToolCallEvent {
	var events []ToolCallEvent
	if text != "" {
		events = append(events, ToolCallEvent{Kind: kind, Text: text})
	}
	for _, call := range calls {
		events = append(events, ToolCallEvent{Kind: thinkfsm.ToolCall, Call: call})
	}
	return events
}

func (p *ToolCallParser) flushText() []ToolCallEvent {
	text, calls := p.scanner.Tail()
	return parsedEvents(thinkfsm.Content, text, calls)
}

func (p *ToolCallParser) resetTool() {
	p.header.Reset()
	p.args.Reset()
	p.inBody = false
}

func (p *ToolCallParser) finishTool() []ToolCallEvent {
	var calls []toolCallFn
	header, args := p.header.String(), strings.TrimSpace(p.args.String())
	if p.inBody && strings.HasPrefix(header, gptOssRecipient) && strings.HasPrefix(args, "{") && sonic.ValidString(args) {
		name := header[len(gptOssRecipient):]
		name = name[:gptOssNameEnd(name)]
		if name != "" {
			calls = append(calls, toolCallFn{Name: name, Arguments: args})
		}
	}
	p.resetTool()
	return parsedEvents(thinkfsm.Content, "", calls)
}

func (p *ToolCallParser) Feed(ev thinkfsm.Event) []ToolCallEvent {
	switch ev.Kind {
	case thinkfsm.Content:
		events := p.finishTool()
		text, calls := p.scanner.Push(ev.Text)
		return append(events, parsedEvents(thinkfsm.Content, text, calls)...)
	case thinkfsm.Thinking:
		events := append(p.flushText(), p.finishTool()...)
		return append(events, parsedEvents(thinkfsm.Thinking, ev.Text, nil)...)
	case thinkfsm.EnterThinking, thinkfsm.LeaveThinking:
		return append(p.flushText(), p.finishTool()...)
	case thinkfsm.ToolCall:
		if p.inBody {
			p.args.WriteString(ev.Text)
			return nil
		}
		events := p.flushText()
		p.header.WriteString(ev.Text)
		return events
	case thinkfsm.ToolCallStart:
		p.inBody = true
	case thinkfsm.ToolCallEnd:
		return p.finishTool()
	}
	return nil
}

func (p *ToolCallParser) Finish() []ToolCallEvent {
	return append(p.flushText(), p.finishTool()...)
}
