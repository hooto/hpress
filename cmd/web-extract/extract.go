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

// runExtract reads an HTML file, converts it to markdown, downloads each
// referenced image, re-encodes it as JPEG q80, and writes "<htmlPath>.md".
// Images are saved under <outDir>/<YYYY/MM/DD>/<hash>.jpg (UTC+8 date) and
// referenced in the markdown via the {{hp_storage_service_endpoint}} placeholder
// so they resolve correctly once uploaded to hpress storage.
func runExtract(htmlPath, outDir string) error {

	bs, err := os.ReadFile(htmlPath)
	if err != nil {
		return err
	}

	doc, _ := goquery.NewDocumentFromReader(bytes.NewBuffer(bs))

	// site-specific cleanup (zhihu)
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
	htm = strings.ReplaceAll(htm, "<!--THE END-->", "")

	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
		),
	)
	if st, err := conv.ConvertString(htm); err == nil {
		st = strings.ReplaceAll(st, "<!--THE END-->\n", "")
		bs = []byte(st)
	}

	dateStr := nowDate8()
	downloaded := map[string]bool{}

	lines := strings.Split(string(bs), "\n")
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

		var imgBytes []byte
		if downloaded[fileName] {
			// already fetched this run; reuse the hashed output if present
		} else {
			imgBytes, err = downloadImage(v2)
			if err != nil {
				fmt.Fprintln(os.Stderr, "skip image (download):", err)
				continue
			}
			downloaded[fileName] = true
		}

		ext := extensionByMime(imgBytes)
		if ext == "" {
			fmt.Fprintln(os.Stderr, "skip image (unsupported type):", fileName)
			continue
		}

		img, err := decodeImage(imgBytes, ext)
		if err != nil || img == nil {
			fmt.Fprintln(os.Stderr, "skip image (decode):", fileName, err)
			continue
		}

		outputFile := fmt.Sprintf("%x.jpg", xxhash.Sum64String(fileName))

		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
			fmt.Fprintln(os.Stderr, "skip image (encode):", fileName, err)
			continue
		}

		relDir := filepath.Join(dateStr)
		fullDir := filepath.Join(outDir, relDir)
		if err := os.MkdirAll(fullDir, 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "skip image (mkdir):", fileName, err)
			continue
		}
		if err := os.WriteFile(filepath.Join(fullDir, outputFile), buf.Bytes(), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "skip image (write):", fileName, err)
			continue
		}

		lines[i] = fmt.Sprintf("![%s]({{hp_storage_service_endpoint}}/%s/%s?ipn=s800x)",
			alt, dateStr, outputFile)
	}

	md := finalizeMarkdown(strings.Join(lines, "\n"))

	outPath := htmlPath
	if strings.HasSuffix(outPath, ".html") || strings.HasSuffix(outPath, ".htm") {
		outPath = strings.TrimSuffix(outPath, filepath.Ext(outPath)) + ".md"
	} else {
		outPath = outPath + ".md"
	}
	if err := os.WriteFile(outPath, []byte(md), 0o644); err != nil {
		return err
	}
	fmt.Println("wrote", outPath)
	return nil
}

// downloadImage fetches a URL into memory.
func downloadImage(u string) ([]byte, error) {
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

// extensionByMime sniffs the first 512 bytes for an image MIME type.
func extensionByMime(b []byte) string {
	n := len(b)
	if n > 512 {
		n = 512
	}
	switch http.DetectContentType(b[:n]) {
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

	// multi-level list indentation
	for _, re := range []*regexp.Regexp{
		regexp.MustCompile(`(\n)( +)(\- )`),
		regexp.MustCompile(`(\n)( +)(\+ )`),
		regexp.MustCompile(`(\n)( +)(\* )`),
	} {
		md = re.ReplaceAllString(md, "$1$2$2$3")
	}
	md = regexp.MustCompile(`(\n)( +)(\n)`).ReplaceAllString(md, "$1$3")

	return md
}
