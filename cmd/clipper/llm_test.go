// Copyright 2015 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

// TestParseLLMResponse covers the JSON-envelope decoding and its fail-open
// semantics: a well-formed {"markdown","changes"} object (plain or fenced)
// round-trips exactly, while a non-JSON response is returned verbatim as
// markdown with no change list so the conversion is never lost.
func TestParseLLMResponse(t *testing.T) {
	bodyWithQuote := "# Title\n\nSome \"quoted\" text and a\nnew line."

	tests := []struct {
		name         string
		build        func() string // raw model output
		wantOK       bool
		wantMd       string
		wantChanges  []string
		wantKeywords []string
	}{
		{
			name: "valid envelope with changes and keywords",
			build: func() string {
				return mustJSON(t, llmResponse{Markdown: bodyWithQuote, Changes: []string{"typo: 'teh' -> 'the'", "repaired broken <ul>"}, Keywords: []string{"html", "markdown"}})
			},
			wantOK:       true,
			wantMd:       bodyWithQuote,
			wantChanges:  []string{"typo: 'teh' -> 'the'", "repaired broken <ul>"},
			wantKeywords: []string{"html", "markdown"},
		},
		{
			name:        "valid envelope empty changes",
			build:       func() string { return mustJSON(t, llmResponse{Markdown: "plain body", Changes: []string{}}) },
			wantOK:      true,
			wantMd:      "plain body",
			wantChanges: nil,
		},
		{
			name: "json wrapped in code fence",
			build: func() string {
				return "```json\n" + mustJSON(t, llmResponse{Markdown: "fenced body", Changes: []string{"one"}, Keywords: []string{"fenced-kw"}}) + "\n```"
			},
			wantOK:       true,
			wantMd:       "fenced body",
			wantChanges:  []string{"one"},
			wantKeywords: []string{"fenced-kw"},
		},
		{
			name: "keyword list trimmed and de-duplicated",
			build: func() string {
				return mustJSON(t, llmResponse{Markdown: "b", Changes: []string{}, Keywords: []string{" distributed systems ", "", "Go", "go", "cache"}})
			},
			wantOK:       true,
			wantMd:       "b",
			wantChanges:  nil,
			wantKeywords: []string{"distributed systems", "Go", "go", "cache"},
		},
		{
			name:        "blank change entries dropped",
			build:       func() string { return mustJSON(t, llmResponse{Markdown: "b", Changes: []string{"real", "  ", ""}}) },
			wantOK:      true,
			wantMd:      "b",
			wantChanges: []string{"real"},
		},
		{
			name:        "plain markdown fail-open",
			build:       func() string { return "# Just Markdown\n\nno json here" },
			wantOK:      false,
			wantMd:      "# Just Markdown\n\nno json here",
			wantChanges: nil,
		},
		{
			name:        "plain markdown starting with code fence stays intact on fail-open",
			build:       func() string { return "```go\npackage main\n```\n\nafter" },
			wantOK:      false,
			wantMd:      "```go\npackage main\n```\n\nafter",
			wantChanges: nil,
		},
		{
			name:        "empty body contract",
			build:       func() string { return mustJSON(t, llmResponse{Markdown: "", Changes: []string{}}) },
			wantOK:      true,
			wantMd:      "",
			wantChanges: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, changes, keywords, ok := parseLLMResponse(tt.build())
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if md != tt.wantMd {
				t.Errorf("markdown mismatch:\n got: %q\nwant: %q", md, tt.wantMd)
			}
			if !equalStrings(changes, tt.wantChanges) {
				t.Errorf("changes = %v, want %v", changes, tt.wantChanges)
			}
			if !equalStrings(keywords, tt.wantKeywords) {
				t.Errorf("keywords = %v, want %v", keywords, tt.wantKeywords)
			}
		})
	}
}

// TestPrintCorrections covers the human-review formatting: a header with the
// count and a numbered list when there are fixes, a single "none" line when the
// report parsed but found nothing, and a distinct "unavailable" line when the
// model did not return the JSON envelope.
func TestPrintCorrections(t *testing.T) {
	tests := []struct {
		name     string
		changes  []string
		reportOK bool
		want     string
	}{
		{
			name:     "unavailable report",
			changes:  nil,
			reportOK: false,
			want:     "llm: correction report unavailable (model did not return the JSON envelope) — review the markdown manually\n",
		},
		{
			name:     "parsed but none",
			changes:  []string{},
			reportOK: true,
			want:     "llm: no corrections applied (body copied verbatim)\n",
		},
		{
			name:     "list",
			changes:  []string{"typo: 'teh' -> 'the'", "repaired broken <ul>"},
			reportOK: true,
			want:     "llm: corrections applied (2) — review before publishing:\n  1. typo: 'teh' -> 'the'\n  2. repaired broken <ul>\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printCorrections(&buf, tt.changes, tt.reportOK)
			if buf.String() != tt.want {
				t.Errorf("output mismatch:\n got: %q\nwant: %q", buf.String(), tt.want)
			}
		})
	}
}

// mustJSON marshals v to compact JSON, failing the test on error so the input
// fixture is always well-formed.
func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}

// equalStrings compares two slices order-sensitively, treating nil and an empty
// slice as equal so a json-decoded []string{} and a nil want match.
func equalStrings(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
