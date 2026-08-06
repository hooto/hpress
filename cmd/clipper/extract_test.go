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
	"image"
	"image/png"
	"path/filepath"
	"strings"
	"testing"
)

// TestSanitizeHTMLForLLM verifies the prefilter strips token-wasting noise
// (scripts/styles/non-content elements, decorative attributes, comments) while
// preserving article content and the markdown-relevant attributes (src, href,
// alt, ...). It is a deny-list pass, so unknown tags and their text survive.
func TestSanitizeHTMLForLLM(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		keep   []string // must remain after sanitizing
		drop   []string // must be removed
		shrink bool     // output must be strictly smaller than input
	}{
		{
			name: "scripts styles iframe svg removed",
			input: `<html><body>
				<script>alert(1); var x = "secret";</script>
				<style>.a { color: red; }</style>
				<iframe src="ad.html"></iframe>
				<svg><path d="M0 0"/></svg>
				<h1>Keep Me</h1>
				<p>body text</p>
				</body></html>`,
			keep:   []string{"Keep Me", "body text"},
			drop:   []string{"alert(1)", "secret", "color: red", "ad.html", "<svg", "<iframe", "<script", "<style"},
			shrink: true,
		},
		{
			name: "decorative attrs dropped, markdown attrs kept",
			input: `<html><body>
				<img src="https://e.com/a/b.png" class="hero lazy" style="width:100%" data-id="7" alt="pic">
				<a href="https://e.com/post" onclick="track()" data-x="1" title="t">link</a>
				<table><tr><td colspan="2" class="c" data-v="1">cell</td></tr></table>
				</body></html>`,
			keep: []string{
				`src="https://e.com/a/b.png"`, `alt="pic"`,
				`href="https://e.com/post"`, `title="t"`,
				`colspan="2"`, "cell",
			},
			drop:   []string{`class="hero lazy"`, `style="width:100%"`, "data-id", "onclick", "data-x", `class="c"`, "data-v"},
			shrink: true,
		},
		{
			name: "html comments removed",
			input: `<html><body>
				<!-- tracking comment secret123 -->
				<h2>Head</h2>
				<!--[if IE]>conditional<![endif]-->
				<p>text</p>
				</body></html>`,
			keep:   []string{"Head", "text"},
			drop:   []string{"secret123", "tracking comment", "[if IE]"},
			shrink: true,
		},
		{
			name: "code content preserved",
			input: `<html><body>
				<pre><code class="language-sql">SELECT 1 FROM t;</code></pre>
				<p>after</p>
				</body></html>`,
			keep:   []string{"SELECT 1 FROM t;", "after"},
			drop:   []string{`class="language-sql"`},
			shrink: true,
		},
		{
			name: "form and inputs removed",
			input: `<html><body>
				<form action="/search"><input name="q" value="x"><button>Go</button></form>
				<article><p>content</p></article>
				</body></html>`,
			keep:   []string{"content"},
			drop:   []string{"<form", "<input", "<button", "/search"},
			shrink: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := sanitizeHTMLForLLM(tt.input)
			if tt.shrink && len(out) >= len(tt.input) {
				t.Errorf("expected output to shrink: in=%d out=%d", len(tt.input), len(out))
			}
			low := strings.ToLower(out)
			for _, w := range tt.keep {
				if !strings.Contains(low, strings.ToLower(w)) {
					t.Errorf("expected output to keep %q\noutput: %s", w, out)
				}
			}
			for _, w := range tt.drop {
				if strings.Contains(low, strings.ToLower(w)) {
					t.Errorf("expected output to drop %q\noutput: %s", w, out)
				}
			}
		})
	}
}

// TestPrefilterEnabled confirms the default-on semantics: only "off" disables.
func TestPrefilterEnabled(t *testing.T) {
	for s, want := range map[string]bool{
		"":     true,
		"on":   true,
		"ON":   true, // not "off" -> on
		"off":  false,
		"junk": true, // unknown -> default on
	} {
		if got := prefilterEnabled(s); got != want {
			t.Errorf("prefilterEnabled(%q)=%v want %v", s, got, want)
		}
	}
}

// TestSaveKeywordsState covers the per-article keyword state write: a fresh write
// leaves node_id empty (pre-publish), a re-extract merges keywords onto an
// existing published state without losing the node id, and an empty keyword list
// is a no-op that leaves any existing state untouched.
func TestSaveKeywordsState(t *testing.T) {
	mdPath := filepath.Join(t.TempDir(), "article.md")

	// 1) fresh: no existing state -> node_id stays "", keywords recorded.
	if err := saveKeywordsState(mdPath, []string{"go", "cache"}); err != nil {
		t.Fatalf("save (fresh): %v", err)
	}
	st, err := LoadArticleState(mdPath)
	if err != nil {
		t.Fatalf("load (fresh): %v", err)
	}
	if st.NodeID != "" {
		t.Errorf("fresh node_id = %q, want empty (pre-publish)", st.NodeID)
	}
	if !equalStrings(st.Keywords, []string{"go", "cache"}) {
		t.Errorf("fresh keywords = %v, want [go cache]", st.Keywords)
	}

	// 2) simulate a publish that set node_id, then re-extract keywords: node_id
	//    must be preserved while keywords refresh.
	st.NodeID = "node-123"
	st.Title = "Published Title"
	if err := SaveArticleState(mdPath, st); err != nil {
		t.Fatalf("seed published state: %v", err)
	}
	if err := saveKeywordsState(mdPath, []string{"go", "concurrency"}); err != nil {
		t.Fatalf("save (merge): %v", err)
	}
	st, err = LoadArticleState(mdPath)
	if err != nil {
		t.Fatalf("load (merge): %v", err)
	}
	if st.NodeID != "node-123" {
		t.Errorf("merge node_id = %q, want node-123 (preserved)", st.NodeID)
	}
	if st.Title != "Published Title" {
		t.Errorf("merge title = %q, want preserved", st.Title)
	}
	if !equalStrings(st.Keywords, []string{"go", "concurrency"}) {
		t.Errorf("merge keywords = %v, want [go concurrency]", st.Keywords)
	}

	// 3) empty keyword list is a no-op: the file is untouched (node_id + old
	//    keywords still present).
	if err := saveKeywordsState(mdPath, nil); err != nil {
		t.Fatalf("save (noop): %v", err)
	}
	st, err = LoadArticleState(mdPath)
	if err != nil {
		t.Fatalf("load (noop): %v", err)
	}
	if st.NodeID != "node-123" {
		t.Errorf("noop node_id = %q, want node-123 (untouched)", st.NodeID)
	}
	if !equalStrings(st.Keywords, []string{"go", "concurrency"}) {
		t.Errorf("noop keywords = %v, want [go concurrency] (untouched)", st.Keywords)
	}
}

// TestRewriteImagesDedupSameBasename locks in the image dedup fix: two image
// refs that share a basename (different paths) must both be rewritten to the SAME
// on-disk file, and the second one must NOT be re-downloaded. The previous code
// left the second ref as a raw external URL (it cached "already downloaded" but
// never reused the bytes, so a nil image was skipped). No real network: the
// downloadImage func is stubbed.
func TestRewriteImagesDedupSameBasename(t *testing.T) {
	pngBytes := testPNGBytes(t)
	calls := map[string]int{}
	orig := downloadImage
	downloadImage = func(u string) ([]byte, error) {
		calls[u]++
		return pngBytes, nil
	}
	t.Cleanup(func() { downloadImage = orig })

	outDir := t.TempDir()
	md := "![a](http://h/x/p1/1.png)\n" +
		"![b](http://h/x/p2/1.png)\n" + // same basename "1.png", different path
		"![c](http://h/x/p3/2.png)\n" // distinct basename "2.png"
	got := rewriteImages(md, outDir)

	lines := strings.Split(got, "\n")

	// every ref is rewritten to the placeholder; no raw URL survives (the bug
	// left the second same-basename ref untouched).
	for i := range 3 {
		if strings.Contains(lines[i], "http://h") {
			t.Errorf("line %d still carries a raw url: %q", i, lines[i])
		}
		if !strings.Contains(lines[i], previewPlaceholder) {
			t.Errorf("line %d not rewritten to placeholder: %q", i, lines[i])
		}
	}

	// the two same-basename refs point at the same file; the third at a different one.
	refA := placeholderFile(t, lines[0])
	refB := placeholderFile(t, lines[1])
	refC := placeholderFile(t, lines[2])
	if refA == "" || refB == "" || refC == "" {
		t.Fatalf("missing placeholder: a=%q b=%q c=%q", refA, refB, refC)
	}
	if refA != refB {
		t.Errorf("same-basename refs should share one file: a=%s b=%s", refA, refB)
	}
	if refC == refA {
		t.Errorf("distinct basename should map to a different file: c=%s a=%s", refC, refA)
	}

	// dedup: first and third URLs fetched once each; the second (same basename)
	// is never fetched because the file is reused.
	if calls["http://h/x/p1/1.png"] != 1 {
		t.Errorf("p1/1.png fetched %d times, want 1", calls["http://h/x/p1/1.png"])
	}
	if calls["http://h/x/p3/2.png"] != 1 {
		t.Errorf("p3/2.png fetched %d times, want 1", calls["http://h/x/p3/2.png"])
	}
	if calls["http://h/x/p2/1.png"] != 0 {
		t.Errorf("p2/1.png fetched %d times, want 0 (reused)", calls["http://h/x/p2/1.png"])
	}
}

// placeholderFile extracts the "<date>/<hash>.jpg" tail from a rewritten image
// line (the part after the storage placeholder, before the resize query).
func placeholderFile(t *testing.T, line string) string {
	t.Helper()
	i := strings.Index(line, previewPlaceholder+"/")
	if i < 0 {
		return ""
	}
	tail := line[i+len(previewPlaceholder)+1:]
	if q := strings.IndexAny(tail, "?)"); q >= 0 {
		tail = tail[:q]
	}
	return tail
}

// testPNGBytes encodes a valid 1x1 PNG so extensionByMime/decodeImage succeed
// without shipping fixture bytes or hitting the network.
func testPNGBytes(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}
