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

// LLM (DeepSeek-compatible OpenAI Chat Completions) backend for HTML -> markdown.
//
// Unlike the rule-based "classic" converter, this backend sends the (site-cleaned)
// HTML to a chat model that extracts the article body and emits clean, structured
// markdown, dropping navigation/chrome the rule-based converter leaves in. The
// model output is then fed through the same shared image-download + rewrite +
// finalize pipeline as the classic path, so image handling is identical.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

// llmDefaultTimeoutSec is the default overall per-request timeout (seconds) when
// [llm].timeout is unset. Reasoning models (e.g. deepseek-v4-flash,
// deepseek-reasoner) can spend minutes "thinking" on a large page before
// emitting the answer, so this is deliberately generous; override via
// [llm].timeout or --llm-timeout.
const llmDefaultTimeoutSec = 600

// llmLargeInputWarn is a rough byte heuristic above which a page likely nudges
// the model's context window. We warn (stderr) but still try; the
// finish_reason=length guard below turns a true overflow into a clear error.
const llmLargeInputWarn = 200_000

// llmSystemPrompt instructs the model to extract the article body faithfully,
// to preserve image URLs byte-for-byte, and — in addition to a faithful
// conversion — to fix obvious typos and broken formatting it encounters in the
// source, reporting every such correction so a human can review them.
//
// The URL rules are critical: the shared rewriteImages step downloads each image
// by the captured URL, so any model rewriting (relativizing, query stripping,
// backslash escaping, angle-bracket wrapping) would break that capture. The JSON
// envelope carries the markdown, a self-reported correction list, and (when the
// source has a curated keyword list) the extracted keywords; parseLLMResponse
// decodes it and fails open to plain markdown if the model ignores the contract.
const llmSystemPrompt = `You convert HTML article pages to GitHub-Flavored Markdown, report the corrections you made, and extract any curated keyword list.

OUTPUT CONTRACT (critical): respond with ONLY a single JSON object and nothing else — no prose, no Markdown code fence, no explanation. The object has exactly this shape:

{"markdown": "<the full Markdown document as a JSON string>", "changes": ["<correction 1>", "<correction 2>"], "keywords": ["<keyword 1>", "<keyword 2>"]}

- "markdown": the converted document encoded as a proper JSON string (literal newlines as \n, double quotes as \", backslashes as \\). A downstream step downloads each image by the URL captured from this field, so the decoded value must be valid Markdown with every URL intact.
- "changes": one short human-readable line per correction you applied to the body (see CORRECTIONS). Use [] when you made none.
- "keywords": the source's curated keyword/index list copied verbatim, when one is present (see KEYWORDS); otherwise [].

FIDELITY IS MANDATORY. You are extracting, not authoring:
- Copy the article body text VERBATIM. The ONLY permitted deviation from verbatim is the CORRECTIONS section below. Do not summarize, paraphrase, rephrase, condense, expand, translate, or fabricate any wording.
- Do not add headings, titles, or metadata that are not present in the source.
- Do not include the page title tag, the site logo, breadcrumbs, or any chrome.

CORRECTIONS — while extracting, fix obvious defects in the body and report each fix as one entry in "changes":
- Typos / common misspellings: correct unambiguous errors only (e.g. "teh"->"the", "recieve"->"receive"). Never "correct" valid domain-specific terms, proper names, identifiers, code, or British/American spelling variants. If a fix could change meaning or you are unsure, leave the text unchanged and do not list it.
- Broken formatting that would otherwise convert incorrectly: malformed or unclosed tags, broken list/table/code structure, a missing or wrong code-fence language (apply the inference rule below), doubled or stray characters, broken HTML entity encoding. Do not reflow or restyle content that is already correct.
- Each entry is one concise line, e.g. "typo: 'recieve' -> 'receive' (paragraph 3)" or "repaired broken <ul> nesting in section 2". If you fix nothing, return [].

KEYWORDS — copy a curated metadata term-list into "keywords", but ONLY when the source itself carries one. Recognize it by STRUCTURE, not by an exact label: a compact, explicitly delimited set of short index/tag terms the author assembled as article metadata and set apart from the running prose. It is typically a single line or short block, often near the title or at the end of the article, frequently introduced by a colon (e.g. "Tags: go, cache, concurrency") or rendered as a short row of pill/tag chips, and separated by commas, slashes, semicolons, vertical bars, whitespace, or line breaks. The label varies by language and site — examples (non-exhaustive): 关键词, 关键字, 核心词, 主题词, 标签, 标签词, 索引词, 关键技术, Keywords, Key words, Tags, Tag, Labels, Subjects, Topics, Index terms, Categories. Match the concept, not the wording.
- Count it as a keyword list only when it is clearly a curated set of short index terms — NOT a sentence, NOT body prose, NOT a heading or section title, and NOT navigation.
- Copy each term VERBATIM (same language, source order, original form); apply the same typo discipline and count any keyword fix under "changes".
- Do NOT mine, summarize, paraphrase, split, or invent terms from body prose, headings, captions, or sentences, and do not promote headings/section titles into keywords. When in doubt whether a deliberate keyword list exists, return [].
- Leave the keyword list in the body markdown exactly as-is — it is captured here IN ADDITION, not removed.
- If no curated keyword list is present, return [].

DROP non-article scaffolding: site navigation, headers, footers, sidebars, tables of contents, search boxes, share/related/comment/cookie/consent banners, advertisements, "read next", and pagination.

PRESERVE structure faithfully:
- Heading levels exactly as in source (h1 to #, h2 to ##, and so on); do not promote or demote.
- List nesting and bullet-versus-ordered style exactly as in source.
- Emphasis (bold, italic, strikethrough), inline code, blockquotes, and horizontal rules.
- Tables become pipe tables; keep alignment.
- Code blocks become fenced code blocks. If the source carries a specific language tag, keep it unchanged. If the tag is empty or a generic placeholder (such as text, plain, plaintext, or code), infer the real language from the content and use the matching id as the fence info string: sql for SQL statements; shell (or bash) for shell commands and scripts; and similarly json, yaml, xml, html, css, go, python, javascript, typescript, java, cpp, rust, and so on. Only leave the tag empty when the content has no recognizable language. (A corrected or added fence language counts as one "changes" entry.)

IMAGES - CRITICAL: a downstream step downloads each image by its URL, so the URL must be exact.
- Render every img element as Markdown image syntax: ![alt text](url).
- The url MUST be the value of the img src attribute, copied BYTE FOR BYTE: absolute or relative, with its original query string, fragment, and percent-encoding completely intact.
- DO NOT rewrite, normalize, complete, relativize, or shorten any URL. DO NOT strip the query string. DO NOT percent-decode or re-encode. DO NOT invent or substitute a host for relative URLs; pass them through exactly as in src.
- DO NOT backslash-escape any character inside the URL (no \_ or \*); underscores, asterisks, tildes, parentheses, etc. inside URLs MUST appear literally in the decoded markdown.
- DO NOT wrap the URL in angle brackets; use the plain ![alt](url) form.
- If the img has a non-empty alt, use it verbatim as the alt text; otherwise use a brief literal description. Keep alt text free of ] and ) characters.
- Preserve the original inline ordering of images relative to the surrounding paragraphs.

LINKS: render as [text](href), preserving the href exactly (the same byte-for-byte rule as images; no angle brackets, no backslash escaping).

DROP entirely: script, style, link, iframe, svg, noscript, template, and form elements, and any tracking or markup pixels.

If the input is not HTML or contains no article body, return {"markdown": "", "changes": [], "keywords": []}.`

const llmUserPromptPrefix = "Convert the following HTML to Markdown, applying the CORRECTIONS rules. Respond with ONLY the JSON object described above.\n\n"

// llmConvert calls a DeepSeek-compatible chat-completion API and returns the
// markdown body plus the list of corrections the model reports having applied.
// It streams the response: reasoning models (deepseek-v4-flash,
// deepseek-reasoner, ...) emit a long reasoning_content phase before the answer,
// and a non-streaming call would sit idle through it and blow past the timeout.
// Streaming keeps the connection alive and lets us report progress. Only the
// content deltas (the JSON envelope) are accumulated; reasoning deltas are
// counted for the progress line and discarded.
//
// It requires ClientLLM.APIKey and BaseURL; Model defaults to "deepseek-v4-flash".
// Missing config returns an error pointing at the auth subcommand.
func llmConvert(llm ClientLLM, htm string) (md string, changes []string, keywords []string, reportOK bool, err error) {
	if llm.APIKey == "" || llm.BaseURL == "" {
		return "", nil, nil, false, fmt.Errorf("llm mode requires llm.api_key and llm.base_url; run `%s auth --mode llm --llm-base-url <url> --llm-api-key <key>` first",
			os.Args[0])
	}
	model := llm.Model
	if model == "" {
		model = "deepseek-v4-flash"
	}
	timeoutSec := llm.Timeout
	if timeoutSec <= 0 {
		timeoutSec = llmDefaultTimeoutSec
	}

	if len(htm) > llmLargeInputWarn {
		fmt.Fprintf(os.Stderr, "warn: html is %d bytes (~%.0fK tokens); a large page can take minutes for the model to process\n",
			len(htm), float64(len(htm))/4)
	}

	// go-openai posts to {BaseURL}/chat/completions; trim a trailing slash so a
	// user-supplied "https://api.deepseek.com/" does not yield "//chat/completions".
	cfg := openai.DefaultConfig(llm.APIKey)
	cfg.BaseURL = strings.TrimRight(llm.BaseURL, "/")
	client := openai.NewClientWithConfig(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	req := openai.ChatCompletionRequest{
		Model: model,
		// Deterministic extraction, not creative rewriting; faithfulness matters.
		Temperature: 0,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: llmSystemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: llmUserPromptPrefix + htm},
		},
	}
	// Default to disabling the reasoning ("thinking") phase via the
	// OpenAI-style reasoning_effort="none". On reasoning models (deepseek-v4-flash,
	// deepseek-reasoner) the reasoning phase is slow and unnecessary for HTML
	// extraction; reasoning_effort="none" is accepted by both reasoning and
	// non-reasoning DeepSeek models. Enable via [llm].enable_thinking for hard
	// layouts that benefit from the model thinking through the structure.
	if !llm.EnableThinking {
		req.ReasoningEffort = "none"
	}
	fmt.Fprintf(os.Stderr, "llm: model=%s, thinking=%v, timeout=%ds\n", model, llm.EnableThinking, timeoutSec)

	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return "", nil, nil, false, fmt.Errorf("llm create stream: %w", err)
	}
	defer stream.Close()

	var content strings.Builder
	var reasoningChars int
	var finishReason openai.FinishReason
	startedReasoning := false
	startedContent := false
	chunks := 0

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", nil, nil, false, fmt.Errorf("llm stream: %w", err)
		}
		if len(resp.Choices) == 0 {
			continue
		}
		chunks++
		delta := resp.Choices[0].Delta
		if delta.ReasoningContent != "" {
			if !startedReasoning {
				startedReasoning = true
				fmt.Fprintln(os.Stderr, "llm: model is reasoning (can take minutes for large pages)...")
			}
			reasoningChars += len(delta.ReasoningContent)
		}
		if delta.Content != "" {
			if !startedContent {
				startedContent = true
				fmt.Fprintf(os.Stderr, "llm: receiving response (reasoned %d chars)...\n", reasoningChars)
			}
			content.WriteString(delta.Content)
		}
		if fr := resp.Choices[0].FinishReason; fr != "" && fr != openai.FinishReasonNull {
			finishReason = fr
		}
	}

	// Decode the JSON envelope {markdown, changes}. Fail-open: if the model
	// ignored the contract and returned plain markdown, parseLLMResponse returns
	// the raw text as markdown with reportOK=false so the conversion is not lost
	// — only the correction list is.
	md, changes, keywords, reportOK = parseLLMResponse(content.String())
	fmt.Fprintf(os.Stderr, "llm: done (finish=%s, reasoning=%d chars, output=%d chars, %d chunks, json=%v, changes=%d, keywords=%d)\n",
		finishReason, reasoningChars, len(md), chunks, reportOK, len(changes), len(keywords))

	if finishReason == openai.FinishReasonLength {
		return "", nil, nil, false, fmt.Errorf("llm output truncated (finish_reason=length; output capped at the model limit). use a shorter page, a non-reasoning model (e.g. deepseek-chat), or --mode classic")
	}
	if strings.TrimSpace(md) == "" {
		return "", nil, nil, false, fmt.Errorf("llm returned empty content (finish=%s, reasoning=%d chars); try a different model or --mode classic",
			finishReason, reasoningChars)
	}
	return md, changes, keywords, reportOK, nil
}

// llmResponse is the JSON envelope the model is asked to emit: the converted
// markdown plus a human-readable list of the corrections it applied.
type llmResponse struct {
	Markdown string   `json:"markdown"`
	Changes  []string `json:"changes"`
	Keywords []string `json:"keywords"`
}

// parseLLMResponse decodes the model's JSON envelope {markdown, changes} and
// trims/drops blank change entries. It is tolerant of two contract violations:
// the model wrapping the JSON in a Markdown code fence (```json ... ```), and
// the model ignoring JSON entirely and returning plain markdown.
//
// Fail-open: if neither plain JSON nor fenced JSON parses, the raw (trimmed)
// text is returned as markdown with a nil change list and ok=false, so a
// well-formed-but-non-JSON conversion is still usable — only the correction
// report is lost.
func parseLLMResponse(raw string) (md string, changes []string, keywords []string, ok bool) {
	trimmed := strings.TrimSpace(raw)

	var resp llmResponse
	if err := json.Unmarshal([]byte(trimmed), &resp); err == nil {
		return resp.Markdown, normalizeChanges(resp.Changes), normalizeKeywords(resp.Keywords), true
	}

	// Some models wrap the JSON in a ```json fence despite the instruction not to.
	if fenced := stripMarkdownFence(trimmed); fenced != trimmed {
		if err := json.Unmarshal([]byte(fenced), &resp); err == nil {
			return resp.Markdown, normalizeChanges(resp.Changes), normalizeKeywords(resp.Keywords), true
		}
	}

	// Fail-open: treat the raw output as plain markdown (no correction list).
	return trimmed, nil, nil, false
}

// normalizeChanges trims surrounding whitespace from each reported correction
// and drops empties, so the printed review list has no blank/noise rows.
func normalizeChanges(in []string) []string {
	out := make([]string, 0, len(in))
	for _, c := range in {
		if t := strings.TrimSpace(c); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// normalizeKeywords trims each keyword, drops empties, and removes duplicates
// (case-sensitive, preserving first-occurrence order) so a curated keyword list
// has no blank or repeated rows.
func normalizeKeywords(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]bool, len(in))
	for _, k := range in {
		k = strings.TrimSpace(k)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, k)
	}
	return out
}

// stripMarkdownFence defensively removes a single outer code fence if the model
// wrapped its whole output despite the instruction not to. It only strips a
// fence that spans the entire output, so legitimate leading/trailing code blocks
// that are part of the article body are left intact.
func stripMarkdownFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	// drop the opening fence line (``` or ```lang)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[i+1:]
	} else {
		return s // fence with no newline is just a bare "```"; nothing to unwrap
	}
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
