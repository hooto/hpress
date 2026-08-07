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

package module

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/websrv/web"
)

// StaticIndex serves per-module static assets at /hp/-/static/<mod>/...
// (replacing the httpsrv Static.IndexAction). It dispatches by file extension
// and serves the file from the module's static directory with Range support.
func StaticIndex(c fiber.Ctx) error {

	object_path := strings.TrimPrefix(web.UrlPath(c), "/hp/-/static/")

	n := strings.Index(object_path, "/")
	if n < 1 {
		return nil
	}
	srvname := object_path[:n]

	ext := strings.ToLower(filepath.Ext(object_path))
	switch ext {
	case ".js", ".css", ".jpg", ".png", ".svg", ".git", ".ico":
	default:
		return nil
	}

	mod, ok := config.Modules[srvname]
	if !ok {
		return nil
	}

	abs_path := config.Prefix + "/modules/" + mod.Meta.Name + "/static/" + object_path[n+1:]

	if _, err := os.Stat(abs_path); err != nil {
		return web.RenderError(c, fiber.StatusNotFound, "Object Not Found")
	}

	c.Set("Cache-Control", "max-age=86400")
	return c.SendFile(abs_path)
}
