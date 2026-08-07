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

package controllers

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/iam/v2/pkg/iamserver"
	"github.com/sysinner/innerstack/v2/pkg/inauth"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/status"
	"github.com/hooto/hpress/websrv/web"
)

// Index is the management backend entry handler. It enforces IAM auth + the
// PermsManager gate, then renders the admin SPA shell. Replaces httpsrv
// controllers.Index.IndexAction.
func Index(c fiber.Ctx) error {

	status.Locker.RLock()
	defer status.Locker.RUnlock()

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
			currentURL := web.RawAbsUrl(c)
			c.Cookie(&fiber.Cookie{
				Name:     inauth.AppHttpHeaderKey + "-current-url",
				Value:    currentURL,
				Path:     "/",
				HTTPOnly: true,
				Expires:  time.Now().Add(1 * time.Hour),
			})
			return web.Redirect(c, redirectURL)
		}

		return web.RenderError(c, fiber.StatusUnauthorized, "iam auth fail : "+err.Error())
	}

	// Management is gated by backend permissions (PermsManager), which
	// IAM grants only to the app creator. Fail-closed: Allow returns
	// false when the session is unauthenticated or IAM is unreachable.
	if !session.Allow("", config.PermsManager...) {
		return web.RenderError(c, fiber.StatusForbidden, "management access denied")
	}

	c.Set("Cache-Control", "no-cache")

	if v := config.SysConfigList.FetchString("http_h_ac_allow_origin"); v != "" {
		c.Set("Access-Control-Allow-Origin", v)
	}

	data := map[string]interface{}{
		"sys_version_sign": config.SysVersionSign,
		"LANG":             web.ResolveLang(c),
		"URL_MOD_PATH":     "/hp/mgr",
	}

	return web.Render(c, "mgr", "index.tpl", data)
}
