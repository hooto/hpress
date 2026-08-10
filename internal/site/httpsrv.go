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

package site

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/hchart/pkg/webui"
	"github.com/hooto/iam/v2/pkg/iamserver/authfiber"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/web"
)

// Register mounts the public frontend (the spec-driven catch-all page renderer)
// on a fiber router. The caller mounts the router at the app root group ("/")
// and MUST register it LAST, so the explicit /hp/* routes take priority over
// the /* catch-all. This replaces httpsrv NewModule (mounted at "/").
func Register(router fiber.Router) {
	router.All("/error/browser", ErrorBrowser)
	router.All("/*", IndexPage)
}

// RegisterHtp mounts the /hp application routes: the S2 image service, the
// webui, hchart and per-module static assets, and the IAM user-auth routes.
// The caller mounts the router at "/hp". This replaces httpsrv NewHtpModule.
func RegisterHtp(router fiber.Router) {

	// S2 image serve/resize (controller S2 -> /s2, method-agnostic).
	router.All("/s2", S2Index)
	router.All("/s2/*", S2Index)

	// Embedded hchart assets (registered before /~/* so the more-specific
	// /~/hchart path wins).
	router.Get("/~/hchart/*", hchartStatic)

	// On-disk webui assets.
	router.Get("/~/*", web.DiskStatic(config.Prefix+"/webui/"))

	// Per-module static assets: /hp/-/static/<mod>/...
	// (merged from the former websrv/module package).
	router.All("/-/static/*", moduleStatic)

	// IAM user-auth routes: /hp/user-auth/{session,sign-in,callback,sign-out}.
	authfiber.RegisterAuthRoutes(router.Group("/user-auth"))
}

// hchartStatic serves the embedded hchart assets (webui.NewFs, an
// http.FileSystem) under /hp/~/hchart/*. It reads file bytes directly — no
// net/http server or request-cycle coupling.
func hchartStatic(c fiber.Ctx) error {
	fsys := webui.NewFs()

	rel := c.Params("*")
	if rel == "" || rel == "/" {
		rel = "/index.html"
	}
	if !strings.HasPrefix(rel, "/") {
		rel = "/" + rel
	}

	f, err := fsys.Open(rel)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil || st.IsDir() {
		return c.SendStatus(fiber.StatusNotFound)
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	c.Type(web.Ext(rel))
	return c.Send(b)
}

// moduleStatic serves per-module static assets at /hp/-/static/<mod>/...
// It dispatches by file extension and serves the file from the module's on-disk
// static directory with Range support. (Merged from the former websrv/module
// package, where it was StaticIndex.)
func moduleStatic(c fiber.Ctx) error {

	objectPath := strings.TrimPrefix(web.UrlPath(c), "/hp/-/static/")

	n := strings.Index(objectPath, "/")
	if n < 1 {
		return nil
	}
	srvname := objectPath[:n]

	ext := strings.ToLower(filepath.Ext(objectPath))
	switch ext {
	case ".js", ".css", ".jpg", ".png", ".svg", ".gif", ".ico":
	default:
		return nil
	}

	mod, ok := config.Modules[srvname]
	if !ok {
		return nil
	}

	absPath := config.Prefix + "/modules/" + mod.Meta.Name + "/static/" + objectPath[n+1:]

	if _, err := os.Stat(absPath); err != nil {
		return web.RenderError(c, fiber.StatusNotFound, "Object Not Found")
	}

	c.Set("Cache-Control", "max-age=86400")
	return c.SendFile(absPath)
}
