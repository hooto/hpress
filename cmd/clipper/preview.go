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

// Local markdown → HTML preview server (renders the .md with its local images).

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/russross/blackfriday"
)

const previewPlaceholder = "{{hp_storage_service_endpoint}}"

// runPreview renders mdPath to HTML and serves it on localhost:port, mapping
// the {{hp_storage_service_endpoint}} placeholder to the local image directory
// so the extracted images appear inline. Blocks until interrupted.
func runPreview(mdPath, outDir string, port int, openBrowser bool) error {

	md, err := os.ReadFile(mdPath)
	if err != nil {
		return err
	}

	// render with the same engine the hpress frontend uses (blackfriday common)
	htmlBody := string(blackfriday.MarkdownCommon(md))
	// point image refs at the local serve route (/img/<date>/<file>)
	htmlBody = strings.ReplaceAll(htmlBody, previewPlaceholder, "/img")

	page := []byte(`<!DOCTYPE html>
<html><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>clipper preview</title>
<style>
  body{max-width:820px;margin:24px auto;padding:0 16px;font:16px/1.6 -apple-system,Segoe UI,Roboto,sans-serif;color:#222}
  img{max-width:100%;height:auto}
  pre{background:#f5f5f5;padding:12px;overflow:auto;border-radius:4px}
  code{background:#f5f5f5;padding:1px 4px;border-radius:3px}
  pre code{background:none;padding:0}
  table{border-collapse:collapse}
  th,td{border:1px solid #ddd;padding:4px 8px}
</style>
</head><body>` + htmlBody + `
</body></html>`)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(page)
	})
	mux.HandleFunc("/img/", func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(r.URL.Path, "/img/")
		rel = strings.TrimSuffix(rel, "/")
		// strip any resize query (ipn=/ipl=) — serve the raw local file
		http.ServeFile(w, r, filepath.Join(outDir, rel))
	})

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	url := "http://" + addr + "/"
	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		fmt.Printf("preview serving %s  (Ctrl-C to stop)\n", url)
		if openBrowser {
			openURL(url)
		}
	}()

	return srv.ListenAndServe()
}

// openURL attempts to open url in the default browser.
func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
