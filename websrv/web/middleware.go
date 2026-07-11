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

package web

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/hooto/iam/v2/pkg/iamserver"
	"github.com/lessos/lessgo/types"
	"github.com/sysinner/innerstack/v2/pkg/inauth"
)

// ctxKey is an unexported type used for fiber Locals keys owned by this package.
type ctxKey int

const (
	sessionKey ctxKey = iota
	langKey
)

// cookieKeyLocale is the locale cookie name (httpsrv Config.CookieKeyLocale).
const cookieKeyLocale = "lang"

// Auth is the authentication gate middleware. It resolves the IAM user session
// from the access-token cookie (exactly as iamserver.verifier.Session does for
// an *http.Request), calls RequireAuth, and on failure returns 401 + the same
// JSON error body the per-controller Init() gates produced. On success the
// session is stored for retrieval via AuthSession.
func Auth() fiber.Handler {
	return func(c fiber.Ctx) error {
		us := iamserver.AppVerifier.Session(c.Cookies(inauth.AppHttpHeaderKey))
		if _, err := us.RequireAuth(); err != nil {
			c.Status(401)
			return JSON(c, types.NewTypeErrorMeta(iamapi.ErrCodeUnauthorized, "Unauthorized"))
		}
		c.Locals(sessionKey, us)
		return c.Next()
	}
}

// AuthSession returns the IAM user session stashed by Auth. If Auth did not run
// (or stored nothing) it resolves a fresh session from the cookie so handlers
// always have a usable session, mirroring the old c.us field.
func AuthSession(c fiber.Ctx) iamserver.UserSession {
	if v := c.Locals(sessionKey); v != nil {
		if us, ok := v.(iamserver.UserSession); ok {
			return us
		}
	}
	return iamserver.AppVerifier.Session(c.Cookies(inauth.AppHttpHeaderKey))
}

// Locale is middleware that resolves the request language and stores it for
// handlers/templates (replaces httpsrv I18nFilter setting Data["LANG"]).
func Locale() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals(langKey, ResolveLang(c))
		return c.Next()
	}
}

// ResolveLang returns the request language: locale cookie → first
// Accept-Language tag → "en" (replaces httpsrv I18nFilter resolution).
func ResolveLang(c fiber.Ctx) string {
	if v := c.Cookies(cookieKeyLocale); v != "" {
		return strings.ToLower(strings.TrimSpace(v))
	}
	if al := c.Get("Accept-Language"); al != "" {
		if tag := strings.TrimSpace(strings.Split(al, ",")[0]); tag != "" {
			if i := strings.IndexAny(tag, ";-"); i > 0 {
				tag = tag[:i]
			}
			if tag = strings.ToLower(tag); tag != "" {
				return tag
			}
		}
	}
	return "en"
}

// Lang returns the language stored by Locale (or resolved on demand).
func Lang(c fiber.Ctx) string {
	if v := c.Locals(langKey); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return ResolveLang(c)
}
