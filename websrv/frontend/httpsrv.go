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

package frontend

import (
	"io"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/hchart/pkg/webui"
	"github.com/hooto/iam/v2/pkg/iamserver/authfiber"

	"github.com/hooto/hpress/config"
	"github.com/hooto/hpress/websrv/web"
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
// webui and hchart static assets, and the IAM user-auth routes. The caller
// mounts the router at "/hp". This replaces httpsrv NewHtpModule.
func RegisterHtp(router fiber.Router) {

	// S2 image serve/resize (controller S2 -> /s2, method-agnostic).
	router.All("/s2", S2Index)
	router.All("/s2/*", S2Index)

	// Embedded hchart assets (registered before /~/* so the more-specific
	// /~/hchart path wins).
	router.Get("/~/hchart/*", hchartStatic)

	// On-disk webui assets.
	router.Get("/~/*", web.DiskStatic(config.Prefix+"/webui/"))

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
