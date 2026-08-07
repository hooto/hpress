// Copyright 2015 Eryx <evorui at gmail dot com>, All rights reserved.
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

package mgr2

import (
	"io/fs"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/iam/v2/pkg/iamserver"
	"github.com/sysinner/innerstack/v2/pkg/inauth"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/websrv/web"
)

// index serves the embedded Svelte SPA at /hp/mgr2.
//
// Hashed/static assets (everything under dist/ except the shell) are served
// openly with immutable caching — this mirrors the legacy admin, whose JS/CSS
// are served from /hp/~/hpm/* without the management gate. Serving assets
// openly also avoids returning an IAM login redirect for a script request.
//
// The SPA shell (index.html, and any unknown sub-path used as a deep link) is
// gated by the SAME IAM auth + config.PermsManager check as /hp/mgr
// (websrv/mgr/controllers/index.go:41-71), inlined verbatim below so the legacy
// package is not modified.
func index(c fiber.Ctx) error {

	rel := c.Params("*")
	if rel == "" || rel == "/" {
		rel = "index.html"
	}
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" {
		rel = "index.html"
	}

	// 1. Hashed/static asset: serve openly, cache forever (filename is hashed).
	if rel != "index.html" {
		if b, err := fs.ReadFile(distFS, "dist/"+rel); err == nil {
			c.Type(web.Ext(rel))
			c.Set("Cache-Control", "public, max-age=31536000, immutable")
			return c.Send(b)
		}
		// not a real file → fall through to SPA shell (deep-link fallback)
	}

	// 2. SPA shell — enforce the management access gate (mirrors
	//    websrv/mgr/controllers/index.go lines 41-71).
	if web.Param(c, "_iam_out") != "" {
		return web.Redirect(c, web.UrlBase(c, ""))
	}

	if err := iamserver.AppVerifier.Ping(); err != nil {
		return web.RenderError(c, fiber.StatusInternalServerError, "iam ping fail : "+err.Error())
	}

	session := iamserver.AppVerifier.Session(c.Cookies(inauth.AppHttpHeaderKey))
	if err := session.CheckServer(); err != nil {
		return web.RenderError(c, fiber.StatusInternalServerError, "iam session check fail : "+err.Error())
	}

	if redirectURL, err := session.RequireAuth(); err != nil {
		if redirectURL != "" {
			c.Cookie(&fiber.Cookie{
				Name:     inauth.AppHttpHeaderKey + "-current-url",
				Value:    web.RawAbsUrl(c),
				Path:     "/",
				HTTPOnly: true,
				Expires:  time.Now().Add(1 * time.Hour),
			})
			return web.Redirect(c, redirectURL)
		}
		// Session missing/expired (e.g. "auth-denied : iat expired"). The SPA
		// has not loaded yet, so render a centered modal page whose Sign In
		// button re-authenticates via IAM and returns here after login.
		msg := "Your login session has expired or you are not signed in. Please sign in again."
		if strings.Contains(err.Error(), "iat expired") {
			msg = "Your login session has expired (IAM access token is no longer valid). Please sign in again."
		}
		return web.RenderAuthRequired(c, web.UrlBase(c, "hp/user-auth/sign-in"), web.RawAbsUrl(c), msg)
	}

	// Allow returns false when unauthenticated OR when the session lacks any
	// PermsManager permission (sys.admin / editor.write/list/read). Fail-closed.
	if !session.Allow("", config.PermsManager...) {
		return web.RenderError(c, fiber.StatusForbidden, "management access denied")
	}

	// 3. Authenticated manager — serve the SPA shell (never cache the shell).
	b, err := fs.ReadFile(distFS, "dist/index.html")
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	c.Set("Cache-Control", "no-cache")
	c.Type("html")
	return c.Send(b)
}
