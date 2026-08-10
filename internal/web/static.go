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
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// DiskStatic returns a fiber.Handler that serves files from a root directory.
// Mount it on a wildcard route, e.g. router.Get("/~/*", web.DiskStatic(dir));
// the wildcard value (c.Params("*")) is the file path relative to rootDir.
// It is the fiber replacement for httpsrv Module.RegisterFileServer on a disk
// directory. Path traversal outside rootDir is rejected with 404.
func DiskStatic(rootDir string) fiber.Handler {
	rootClean := filepath.Clean(rootDir)
	return func(c fiber.Ctx) error {
		rel := c.Params("*")
		if rel == "" || rel == "/" {
			rel = "/index.html"
		}
		if !strings.HasPrefix(rel, "/") {
			rel = "/" + rel
		}

		// filepath.Clean collapses any ".." that would escape, and Join anchors
		// the result under rootDir; the HasPrefix guard is belt-and-suspenders.
		abs := filepath.Join(rootClean, filepath.Clean(rel))
		if !strings.HasPrefix(abs, rootClean) {
			return c.SendStatus(fiber.StatusNotFound)
		}

		if st, err := os.Stat(abs); err != nil || st.IsDir() {
			return c.SendStatus(fiber.StatusNotFound)
		}

		return c.SendFile(abs)
	}
}

// Ext returns the file extension (without dot) of name, for fiber's Ctx.Type().
func Ext(name string) string {
	return strings.TrimPrefix(path.Ext(name), ".")
}
