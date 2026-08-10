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

// Package web is the gofiber/fiber/v3 helper layer for hpress. It provides the
// stateless request/response helpers, the html/template loader, and the
// middleware that previously lived inside github.com/hooto/httpsrv. Handlers
// remain plain fiber handlers (func(fiber.Ctx) error); this package holds no
// controller/dispatch machinery.
package web

import (
	"html/template"
	"strings"
	"time"
)

// UrlBasePath is the global URL prefix applied to every route. It mirrors
// httpsrv Config.UrlBasePath and is set once during startup from
// config.Config.UrlBasePath.
var UrlBasePath string

// Templates is the application-wide html/template loader (per-module template
// sets), replacing httpsrv.DefaultService.TemplateLoader.
var Templates = NewTemplateLoader()

// RegisterFunc registers a template function on the global loader, replacing
// httpsrv Config.RegisterTemplateFunc.
func RegisterFunc(name string, fn any) {
	Templates.RegisterFunc(name, fn)
}

// builtinFuncs are the framework-level template funcs, ported verbatim from
// httpsrv/template-func.go. The "T" func is intentionally omitted here — the
// application registers its own i18n "T" (hlang) via RegisterFunc, which
// overrides any builtin, matching prior behavior.
var builtinFuncs = template.FuncMap{
	"raw": func(text string) template.HTML {
		return template.HTML(text)
	},
	"replace": func(s, old, new string) string {
		return strings.Replace(s, old, new, -1)
	},
	"date": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
	"datetime": func(t time.Time) string {
		return t.Format("2006-01-02 15:04")
	},
	"upper": func(s string) string {
		return strings.ToUpper(s)
	},
	"lower": func(s string) string {
		return strings.ToLower(s)
	},
}
