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
	"fmt"
	"image"
	"image/png"
	"os"
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

// TestRewriteImagesKeysOnURLNotBasename locks in the image-identity fix: the
// dedup key and output filename are derived from the FULL URL, not the basename.
// Many CDNs serve every asset under a generic name (e.g. "Desktop-Light.png"), so
// two distinct images can share a basename; they must each be downloaded and
// stored under their own file. Only refs with the SAME url reuse one file. No
// real network: the downloadImage func is stubbed.
func TestRewriteImagesKeysOnURLNotBasename(t *testing.T) {
	pngBytes := testPNGBytes(t)
	calls := map[string]int{}
	orig := downloadImage
	downloadImage = func(u string) ([]byte, error) {
		calls[u]++
		return pngBytes, nil
	}
	t.Cleanup(func() { downloadImage = orig })

	outDir := t.TempDir()
	md := "![a](http://h/x/p1/1.png)\n" + // basename 1.png, url A
		"![b](http://h/x/p2/1.png)\n" + // SAME basename, DIFFERENT url -> distinct file
		"![c](http://h/x/p1/1.png)\n" + // SAME url as line 1 -> reuse (real dedup)
		"![d](http://h/x/p3/2.png)\n" // distinct basename -> distinct file
	got := rewriteImages(md, outDir)

	lines := strings.Split(got, "\n")

	// every ref is rewritten to the placeholder; no raw URL survives.
	for i := range 4 {
		if strings.Contains(lines[i], "http://h") {
			t.Errorf("line %d still carries a raw url: %q", i, lines[i])
		}
		if !strings.Contains(lines[i], previewPlaceholder) {
			t.Errorf("line %d not rewritten to placeholder: %q", i, lines[i])
		}
	}

	refA := placeholderFile(t, lines[0])
	refB := placeholderFile(t, lines[1])
	refC := placeholderFile(t, lines[2])
	refD := placeholderFile(t, lines[3])
	if refA == "" || refB == "" || refC == "" || refD == "" {
		t.Fatalf("missing placeholder: a=%q b=%q c=%q d=%q", refA, refB, refC, refD)
	}

	// same basename, DIFFERENT url -> distinct files (the old code collapsed them).
	if refA == refB {
		t.Errorf("distinct urls sharing a basename must NOT share a file: a=%s b=%s", refA, refB)
	}
	// identical url -> reuse the same file (legitimate dedup).
	if refA != refC {
		t.Errorf("identical url should reuse one file: a=%s c=%s", refA, refC)
	}
	// distinct url -> its own file.
	if refD == refA || refD == refB {
		t.Errorf("distinct url should map to its own file: d=%s a=%s b=%s", refD, refA, refB)
	}

	// each distinct url is fetched exactly once; the repeated url is not re-fetched.
	if calls["http://h/x/p1/1.png"] != 1 {
		t.Errorf("p1/1.png fetched %d times, want 1", calls["http://h/x/p1/1.png"])
	}
	if calls["http://h/x/p2/1.png"] != 1 {
		t.Errorf("p2/1.png fetched %d times, want 1", calls["http://h/x/p2/1.png"])
	}
	if calls["http://h/x/p3/2.png"] != 1 {
		t.Errorf("p3/2.png fetched %d times, want 1", calls["http://h/x/p3/2.png"])
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

// testSVGBytes returns a minimal inline SVG so the extract path is exercised
// without shipping fixture bytes or hitting the network.
func testSVGBytes() []byte {
	return []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10">` +
		`<circle cx="5" cy="5" r="4"/></svg>`)
}

// TestIsSVG verifies SVG detection. Go's http.DetectContentType reports SVG as
// text, so extensionByMime relies on isSVG; this locks in the cases it must
// cover (bare svg root, xml-prolog-prefixed, leading whitespace) and the
// negatives it must not over-match (plain text, raster bytes).
func TestIsSVG(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want bool
	}{
		{"bare svg root", []byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect/></svg>`), true},
		{"xml prolog", append([]byte(`<?xml version="1.0"?>`+"\n"), []byte(`<svg><rect/></svg>`)...), true},
		{"leading whitespace", []byte("\n  <svg><rect/></svg>"), true},
		{"upper case", []byte(`<SVG><RECT/></SVG>`), true},
		{"plain text", []byte("hello world"), false},
		{"png bytes", testPNGBytes(t), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSVG(tt.in); got != tt.want {
				t.Errorf("isSVG(%q) = %v, want %v", tt.in, got, tt.want)
			}
			if tt.want && extensionByMime(tt.in) != ".svg" {
				t.Errorf("extensionByMime(%q) = %q, want .svg", tt.in, extensionByMime(tt.in))
			}
		})
	}
}

// TestRewriteImagesSVG locks in the SVG path: an inline SVG is stored verbatim
// under a .svg name with no resize query (vector images are not resized), while a
// raster image in the same run keeps the q80 JPEG + ipn=s800x behavior. No real
// network: downloadImage is stubbed.
func TestRewriteImagesSVG(t *testing.T) {
	svgBytes := testSVGBytes()
	pngBytes := testPNGBytes(t)
	orig := downloadImage
	downloadImage = func(u string) ([]byte, error) {
		switch u {
		case "http://h/vector.svg":
			return svgBytes, nil
		case "http://h/raster.png":
			return pngBytes, nil
		}
		return nil, fmt.Errorf("unexpected url %s", u)
	}
	t.Cleanup(func() { downloadImage = orig })

	outDir := t.TempDir()
	md := "![vec](http://h/vector.svg)\n" +
		"![ras](http://h/raster.png)\n"
	got := rewriteImages(md, outDir)
	lines := strings.Split(got, "\n")

	vecTail := placeholderFile(t, lines[0]) // <date>/<hash>.svg
	rasTail := placeholderFile(t, lines[1]) // <date>/<hash>.jpg
	if vecTail == "" || rasTail == "" {
		t.Fatalf("missing placeholder tails: vec=%q ras=%q", vecTail, rasTail)
	}

	// SVG: .svg name, verbatim bytes, no resize query.
	if !strings.HasSuffix(vecTail, ".svg") {
		t.Errorf("svg tail = %q, want .svg suffix", vecTail)
	}
	if strings.Contains(lines[0], "ipn=") {
		t.Errorf("svg ref must not carry a resize hint: %q", lines[0])
	}
	if got := readFile(t, filepath.Join(outDir, vecTail)); !bytes.Equal(got, svgBytes) {
		t.Errorf("svg stored bytes differ from source (len got=%d want=%d)", len(got), len(svgBytes))
	}

	// Raster: still .jpg + resize hint (unchanged behavior).
	if !strings.HasSuffix(rasTail, ".jpg") {
		t.Errorf("raster tail = %q, want .jpg suffix", rasTail)
	}
	if !strings.Contains(lines[1], "ipn=s800x") {
		t.Errorf("raster ref must keep the resize hint: %q", lines[1])
	}
}

// TestMdImageRefReMatchesSVG confirms the publish-side regex still finds SVG
// references (which carry no resize query) as well as JPEG references.
func TestMdImageRefReMatchesSVG(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string // captured <date>/<file> tail
	}{
		{"jpg with resize", `![]({{hp_storage_service_endpoint}}/2026/08/10/1a2b3c.jpg?ipn=s800x)`, "2026/08/10/1a2b3c.jpg"},
		{"svg no resize", `![]({{hp_storage_service_endpoint}}/2026/08/10/1a2b3c.svg)`, "2026/08/10/1a2b3c.svg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mdImageRefRe.FindStringSubmatch(tt.line)
			if m == nil {
				t.Fatalf("no match for %q", tt.line)
			}
			if m[1] != tt.want {
				t.Errorf("capture = %q, want %q", m[1], tt.want)
			}
		})
	}
}

// readFile reads a file, failing the test on any error.
func readFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}
