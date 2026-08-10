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

// HTML → markdown extraction with image download/re-encode.

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"github.com/PuerkitoBio/goquery"
	"github.com/cespare/xxhash"
	"golang.org/x/image/webp"
)

var imgUrlRe = regexp.MustCompile(`!\[(.*?)\]\((.*?)\)`)

// nowDate8 returns today's date as "2006/01/02" in UTC+8 (Asia/Shanghai),
// which is the convention used for both the local image directory layout and
// the hpress storage path / markdown placeholder.
func nowDate8() string {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return time.Now().In(loc).Format("2006/01/02")
}

// runExtract reads an HTML file, converts it to markdown (classic or llm
// backend per mode), downloads each referenced image, re-encodes it as JPEG q80,
// and writes "<htmlPath>.md". Images are saved under
// <outDir>/<YYYY/MM/DD>/<hash>.jpg (UTC+8 date) and referenced in the markdown
// via the {{hp_storage_service_endpoint}} placeholder so they resolve correctly
// once uploaded to hpress storage.
//
// mode selects the conversion backend: "classic" (rule-based) or "llm". An
// empty mode resolves to "classic". The two backends share the same site
// cleanup, image download/rewrite, and finalize steps; only the HTML -> markdown
// text step differs.
func runExtract(htmlPath, outDir, mode string, llm ClientLLM) error {
	sw := newStopwatch()

	bs, err := os.ReadFile(htmlPath)
	if err != nil {
		return err
	}
	sw.mark("read html")

	htm := cleanHTML(bs)
	sw.mark("clean html")

	md, changes, keywords, reportOK, err := convertToMarkdown(mode, llm, htm, bs, sw)
	if err != nil {
		return err
	}

	md = rewriteImages(md, outDir)
	sw.mark("download images")

	md = finalizeMarkdown(md)
	sw.mark("finalize markdown")

	outPath := htmlPath
	if strings.HasSuffix(outPath, ".html") || strings.HasSuffix(outPath, ".htm") {
		outPath = strings.TrimSuffix(outPath, filepath.Ext(outPath)) + ".md"
	} else {
		outPath = outPath + ".md"
	}
	if err := os.WriteFile(outPath, []byte(md), 0o644); err != nil {
		return err
	}
	sw.mark("write file")

	fmt.Println("wrote", outPath)
	// In llm mode the model self-reports the textual/formatting corrections it
	// applied to the body; print them so a human can review before publishing.
	if mode == "llm" {
		printCorrections(os.Stderr, changes, reportOK)
		// A curated keyword list (when the source carries one) is written to the
		// per-article state file next to the markdown, with node_id empty —
		// publish fills node_id in later and carries the keywords through.
		if len(keywords) > 0 {
			if err := saveKeywordsState(outPath, keywords); err != nil {
				fmt.Fprintf(os.Stderr, "warn: save keywords state: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "llm: extracted %d keyword(s) -> %s\n", len(keywords), articleStatePath(outPath))
			}
		} else {
			fmt.Fprintln(os.Stderr, "llm: no curated keyword list found")
		}
	}
	fmt.Fprintf(os.Stderr, "total: %s\n", sw.total())
	return nil
}

// saveKeywordsState records the extracted curated keyword list into the
// per-article state file next to the markdown (<file>.toml). It merges with any
// existing state (e.g. a node_id from a prior publish) so re-extracting a page
// refreshes the keywords without losing the published node id. With no keywords
// it is a no-op that leaves any existing state untouched. Pre-publish, node_id
// stays "" and keywords is the only populated field.
func saveKeywordsState(mdPath string, keywords []string) error {
	if len(keywords) == 0 {
		return nil
	}
	state, err := LoadArticleState(mdPath)
	if err != nil {
		return err
	}
	if state == nil {
		state = &ArticleState{}
	}
	state.Keywords = keywords
	return SaveArticleState(mdPath, state)
}

// cleanHTML parses the raw HTML, applies the site-specific (zhihu) cleanup
// (dropping the table-of-contents, rewriting math spans into $$...$$, stripping
// zhihu search links), and re-serializes. Shared by both backends; for non-zhihu
// HTML it is effectively a re-serialize plus a "<!--THE END-->" strip.
func cleanHTML(bs []byte) string {

	// html.Parse is lenient, so a nil doc is rare; guard it anyway so a broken
	// input fails gracefully instead of nil-dereferencing on doc.Find below.
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(bs))
	if err != nil || doc == nil {
		return strings.ReplaceAll(string(bs), "<!--THE END-->", "")
	}

	doc.Find("div.Catalog").Each(func(i int, s *goquery.Selection) { s.Remove() })
	doc.Find("span.ztext-math").Each(func(i int, s *goquery.Selection) {
		if content, exists := s.Attr("data-tex"); exists {
			content = strings.ReplaceAll(content, "_", "\\_")
			s.ReplaceWithHtml("<code>$$" + content + "$$</code>")
		}
	})
	doc.Find("a.RichContent-EntityWord").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok && strings.Contains(href, "zhihu.com/search?") {
			s.ReplaceWithHtml(strings.TrimSpace(s.Text()))
		}
	})

	htm, _ := doc.Html()
	return strings.ReplaceAll(htm, "<!--THE END-->", "")
}

// convertToMarkdown dispatches the conversion backend by mode and returns the
// markdown plus, for the llm backend, the corrections the model reports having
// applied and whether that correction report was available (reportOK). For the
// classic backend changes is nil and reportOK is false (printing is llm-only).
// The classic backend receives the raw HTML bytes so it can preserve today's
// behavior of falling back to the raw HTML on converter error. sw records
// per-step timing.
func convertToMarkdown(mode string, llm ClientLLM, htm string, rawBs []byte, sw *stopwatch) (string, []string, []string, bool, error) {
	switch mode {
	case "llm":
		// Default-on prefilter: strip token-wasting noise (scripts/styles,
		// non-content elements, decorative attributes) before the LLM call.
		if prefilterEnabled(llm.Prefilter) {
			before := len(htm)
			htm = sanitizeHTMLForLLM(htm)
			sw.mark(fmt.Sprintf("llm: pre-filter %d -> %d bytes (-%d%%)",
				before, len(htm), percentDrop(before, len(htm))))
		} else {
			sw.mark("llm: pre-filter (off)")
		}
		md, changes, keywords, reportOK, err := llmConvert(llm, htm)
		sw.mark("llm: convert")
		return md, changes, keywords, reportOK, err
	default: // "" or "classic"
		md := classicConvert(htm, rawBs)
		sw.mark("classic: convert")
		return md, nil, nil, false, nil
	}
}

// printCorrections writes the llm-reported correction list to w for human
// review before publishing. When reportOK is false the model did not return the
// JSON envelope, so the correction list is unavailable and the operator is told
// to review the markdown manually (we cannot claim the body was verbatim).
// reportOK true with an empty list prints a one-line "none" notice.
func printCorrections(w io.Writer, changes []string, reportOK bool) {
	if !reportOK {
		fmt.Fprintln(w, "llm: correction report unavailable (model did not return the JSON envelope) — review the markdown manually")
		return
	}
	if len(changes) == 0 {
		fmt.Fprintln(w, "llm: no corrections applied (body copied verbatim)")
		return
	}
	fmt.Fprintf(w, "llm: corrections applied (%d) — review before publishing:\n", len(changes))
	for i, c := range changes {
		fmt.Fprintf(w, "  %d. %s\n", i+1, c)
	}
}

// prefilterEnabled reports whether the LLM HTML prefilter is on. Default ("") is
// on; only an explicit "off" disables it.
func prefilterEnabled(s string) bool { return s != "off" }

func percentDrop(before, after int) int {
	if before <= 0 {
		return 0
	}
	return (before - after) * 100 / before
}

// stopwatch is a simple per-step stderr timing printer for the extract pipeline.
// Each mark() reports the elapsed time since the previous mark; total() reports
// the elapsed time since the stopwatch was created. Output goes to stderr so it
// does not pollute the generated markdown.
type stopwatch struct {
	start time.Time
	last  time.Time
}

func newStopwatch() *stopwatch {
	now := time.Now()
	return &stopwatch{start: now, last: now}
}

// mark prints the time elapsed since the previous mark under the given label.
func (s *stopwatch) mark(label string) {
	now := time.Now()
	fmt.Fprintf(os.Stderr, "  %-42s %s\n", label, now.Sub(s.last).Round(time.Millisecond))
	s.last = now
}

// total returns the elapsed time since the stopwatch was created.
func (s *stopwatch) total() time.Duration {
	return time.Since(s.start).Round(time.Millisecond)
}

var htmlCommentRe = regexp.MustCompile(`(?s)<!--.*?-->`)

// sanitizeHTMLForLLM strips token-wasting, non-content noise from the HTML before
// sending it to the LLM. It removes <script>/<style>/<noscript>/<iframe>/<svg>/
// <template>/<form>/etc. elements, HTML comments, and every attribute not in the
// markdown-relevant allow-list (src, href, alt, title, colspan, ...). Article
// text, headings, lists, tables, code, links, and images (with src) are kept.
//
// It is a deny-list pass: it removes known noise but keeps everything else, so it
// never drops article content. Fail-open: on a parse error it returns the input
// unchanged so the LLM still sees the HTML. Applies only to the llm backend —
// the classic converter is cheap and the raw HTML loses it nothing.
func sanitizeHTMLForLLM(htm string) string {

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htm))
	if err != nil || doc == nil {
		return htm
	}

	// elements that carry no article content but burn tokens
	doc.Find("script,style,noscript,template,iframe,object,embed,svg,canvas," +
		"link,meta,form,fieldset,button,input,select,textarea,option,map,area").Remove()

	// drop every attribute that is irrelevant to markdown conversion (class,
	// style, id, data-*, on*, aria-*, srcset, integrity, crossorigin, ...)
	doc.Find("*").Each(func(_ int, s *goquery.Selection) {
		for _, n := range s.Nodes {
			keep := n.Attr[:0]
			for _, a := range n.Attr {
				if attrRelevant(a.Key) {
					keep = append(keep, a)
				}
			}
			n.Attr = keep
		}
	})

	out, err := doc.Html()
	if err != nil || out == "" {
		return htm
	}
	return htmlCommentRe.ReplaceAllString(out, "")
}

// attrRelevant reports whether an attribute is useful for markdown conversion.
func attrRelevant(k string) bool {
	switch k {
	case "src", "href", "alt", "title", "colspan", "rowspan",
		"width", "height", "start", "type", "value", "datetime",
		"cite", "lang", "dir", "name":
		return true
	}
	return false
}

// classicConvert is the rule-based JohannesKaufmann converter. On conversion
// error it returns the raw HTML verbatim, matching the prior inline behavior.
func classicConvert(htm string, rawBs []byte) string {
	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
		),
	)
	st, err := conv.ConvertString(htm)
	if err != nil {
		return string(rawBs)
	}
	return strings.ReplaceAll(st, "<!--THE END-->\n", "")
}

// rewriteImages scans the markdown for image references, downloads each image,
// re-encodes it as JPEG q80, writes it under <outDir>/<date>/<hash>.jpg, and
// rewrites the reference to the {{hp_storage_service_endpoint}} placeholder.
// Failures are logged to stderr and skipped, matching the original behavior.
// Shared by both backends, so image handling is identical regardless of how the
// markdown text was produced.
func rewriteImages(md, outDir string) string {

	dateStr := nowDate8()
	// outputByName caches the full URL -> the <hash>.jpg already written this run.
	// Two refs that share a URL resolve to one file: process it once, then reuse
	// it for every later ref. Keying on the URL (not the basename) keeps distinct
	// images that happen to share a filename separate.
	outputByName := map[string]string{}

	lines := strings.Split(md, "\n")
	for i, v := range lines {

		mat := imgUrlRe.FindStringSubmatch(v)
		if len(mat) != 3 {
			continue
		}

		alt := mat[1]
		v2 := mat[2]
		up, err := url.Parse(v2)
		if err != nil {
			fmt.Fprintln(os.Stderr, "skip image (bad url):", err)
			continue
		}

		fileName := filepath.Base(up.Path)
		if fileName == "" || fileName == "." || fileName == "/" {
			continue
		}

		// Identity key is the FULL URL, not the basename: many CDNs serve every
		// asset under a generic name (e.g. ".../Desktop-Light.png"), so keying on
		// the basename collapses distinct images into one file. The full URL is the
		// real identity; refs that share a URL still reuse one file (real dedup).
		key := v2
		if outputFile := outputByName[key]; outputFile != "" {
			lines[i] = fmt.Sprintf("![%s]({{hp_storage_service_endpoint}}/%s/%s%s)",
				alt, dateStr, outputFile, imageRefSuffix(outputFile))
			continue
		}

		imgBytes, err := downloadImage(v2)
		if err != nil {
			fmt.Fprintln(os.Stderr, "skip image (download):", err)
			continue
		}

		ext := extensionByMime(imgBytes)
		if ext == "" {
			fmt.Fprintln(os.Stderr, "skip image (unsupported type):", fileName)
			continue
		}

		// Pick the stored bytes + output file name. SVG is a vector format: store
		// the source verbatim with no decode and no re-encode/resize — the Go
		// stdlib has no SVG decoder, and rasterizing a vector would only lose
		// fidelity. Raster formats keep the original q80 JPEG re-encode path.
		var (
			outputFile string
			outBytes   []byte
		)
		if ext == ".svg" {
			outputFile = fmt.Sprintf("%x.svg", xxhash.Sum64String(key))
			outBytes = imgBytes
		} else {
			img, err := decodeImage(imgBytes, ext)
			if err != nil || img == nil {
				fmt.Fprintln(os.Stderr, "skip image (decode):", fileName, err)
				continue
			}
			var buf bytes.Buffer
			if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
				fmt.Fprintln(os.Stderr, "skip image (encode):", fileName, err)
				continue
			}
			outputFile = fmt.Sprintf("%x.jpg", xxhash.Sum64String(key))
			outBytes = buf.Bytes()
		}

		relDir := filepath.Join(dateStr)
		fullDir := filepath.Join(outDir, relDir)
		if err := os.MkdirAll(fullDir, 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "skip image (mkdir):", fileName, err)
			continue
		}
		if err := os.WriteFile(filepath.Join(fullDir, outputFile), outBytes, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "skip image (write):", fileName, err)
			continue
		}
		outputByName[key] = outputFile

		lines[i] = fmt.Sprintf("![%s]({{hp_storage_service_endpoint}}/%s/%s%s)",
			alt, dateStr, outputFile, imageRefSuffix(outputFile))
	}

	return strings.Join(lines, "\n")
}

// downloadImage fetches a URL into memory. It is a package-level variable so
// tests can substitute a stub and avoid real network I/O.
var downloadImage = func(u string) ([]byte, error) {
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, u)
	}
	return io.ReadAll(resp.Body)
}

// extensionByMime sniffs the leading bytes for an image MIME type. SVG is
// detected directly because http.DetectContentType reports it as text.
func extensionByMime(b []byte) string {
	if isSVG(b) {
		return ".svg"
	}
	switch http.DetectContentType(b) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/bmp":
		return ".bmp"
	default:
		return ""
	}
}

// isSVG reports whether b is an inline SVG image. SVG is XML text, so
// http.DetectContentType reports "text/plain"/"text/xml" rather than
// "image/svg+xml"; detect it directly by scanning the leading bytes for an <svg
// root, which may be preceded by an XML prolog or comment.
func isSVG(b []byte) bool {
	head := b
	if len(head) > 1024 {
		head = head[:1024]
	}
	return strings.Contains(strings.ToLower(string(head)), "<svg")
}

// imageRefSuffix returns the query hint appended to a storage image placeholder
// reference. Raster images carry the on-the-fly resize hint (ipn=s800x); SVG is
// vector and skips it — the front end serves SVG verbatim regardless of the
// query, so the hint would only imply a resize that never happens.
func imageRefSuffix(name string) string {
	if strings.HasSuffix(name, ".svg") {
		return ""
	}
	return "?ipn=s800x"
}

func decodeImage(b []byte, ext string) (image.Image, error) {
	switch strings.ToLower(ext) {
	case ".webp":
		return webp.Decode(bytes.NewBuffer(b))
	case ".jpg", ".jpeg":
		return jpeg.Decode(bytes.NewBuffer(b))
	case ".png":
		return png.Decode(bytes.NewBuffer(b))
	default:
		return nil, fmt.Errorf("unsupported extension %s", ext)
	}
}

// finalizeMarkdown applies the small formatting cleanups the original tool did
// (list indentation, zhihu code-fence fixups, blank-line collapse in lists).
func finalizeMarkdown(md string) string {
	lines := strings.Split(md, "\n")
	for i, line := range lines {
		for _, s := range []string{"-", "+", "*"} {
			if strings.HasPrefix(line, s+"\t") {
				lines[i] = strings.Replace(line, s+"\t", s+" ", 1)
			}
		}
	}

	// collapse blank lines between consecutive list items
	fmtLines := make([]string, 0, len(lines))
	for i, line := range lines {
		trs := strings.TrimSpace(line)
		if trs == "" && i > 0 && i+1 < len(lines) {
			ct := false
			for _, s := range []string{"-", "+", "*"} {
				if strings.HasPrefix(lines[i-1], s+" ") &&
					strings.HasPrefix(lines[i+1], s+" ") {
					ct = true
					break
				}
			}
			if ct {
				continue
			}
		}
		fmtLines = append(fmtLines, line)
	}

	md = strings.Join(fmtLines, "\n")
	for _, v := range [][]string{
		{"```python3\n", "```python\n"},
		{"`$$", "$$"},
		{"\\$$`", "$$"},
		{"$$`", "$$"},
	} {
		md = strings.ReplaceAll(md, v[0], v[1])
	}

	// multi-level list indentation: double the indent in front of any list marker.
	md = listIndentRe.ReplaceAllString(md, "$1$2$2$3")
	md = blankAfterIndentRe.ReplaceAllString(md, "$1$3")

	return md
}

// listIndentRe matches a newline + indent spaces + a list marker ("- "/"+"/"* ");
// the replacement doubles the indent. blankAfterIndentRe collapses a lone run of
// indent spaces on an otherwise blank line. Compiled once at package load.
var (
	listIndentRe       = regexp.MustCompile(`(\n)( +)([-+*] )`)
	blankAfterIndentRe = regexp.MustCompile(`(\n)( +)(\n)`)
)
