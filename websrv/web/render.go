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
	"bytes"
	"encoding/json"
	stdpath "path"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// JSON writes a JSON response, replicating httpsrv Controller.RenderJson: it
// sets Access-Control-Allow-Origin: * and Content-Type: application/json.
func JSON(c fiber.Ctx, data any) error {
	c.Set("Access-Control-Allow-Origin", "*")
	return c.JSON(data)
}

// JSONIndent writes an indented JSON response (httpsrv Controller.RenderJsonIndent).
func JSONIndent(c fiber.Ctx, data any, indent string) error {
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Content-Type", "application/json")
	js, err := json.MarshalIndent(data, "", indent)
	if err != nil {
		return err
	}
	return c.Send(js)
}

// Render executes a module template and writes it as text/html, replacing
// httpsrv Controller.Render. module selects the template set; tplPath names the
// template within it; data is the template data model.
func Render(c fiber.Ctx, module, tplPath string, data any) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	out, err := Templates.renderToBuffer(module, tplPath, data)
	if err != nil {
		return err
	}
	return c.Send(out)
}

// RenderHTML parses and executes an inline template string as text/html,
// replacing httpsrv Controller.RenderHTML.
func RenderHTML(c fiber.Ctx, html string, data any) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	if err := Templates.rawRender(&buf, html, data); err != nil {
		c.Status(fiber.StatusBadRequest)
		_, _ = c.Write([]byte("400 Bad Request"))
		return nil
	}
	return c.Send(buf.Bytes())
}

// RenderString writes a raw string to the response without template parsing,
// replacing httpsrv Controller.RenderString.
func RenderString(c fiber.Ctx, v string) error {
	return c.SendString(v)
}

// RenderError writes a status code with a raw text/html body, replacing httpsrv
// Controller.RenderError.
func RenderError(c fiber.Ctx, status int, msg string) error {
	c.Status(status)
	c.Set("Content-Type", "text/html; charset=utf-8")
	_, err := c.Write([]byte(msg))
	return err
}

// Redirect sends a 302 Found, prefixing UrlBasePath onto relative locations
// (those without a leading "/" or scheme), exactly as httpsrv
// Controller.Redirect does.
func Redirect(c fiber.Ctx, location string) error {
	if location == "" {
		return nil
	}
	if location[0] != '/' && !strings.HasPrefix(location, "http") {
		elems := []string{"/"}
		if UrlBasePath != "" {
			elems = append(elems, UrlBasePath)
		}
		elems = append(elems, location)
		location = stdpath.Join(elems...)
	}
	return c.Redirect().Status(fiber.StatusFound).To(location)
}

// UrlBase builds an absolute URL (scheme://host + UrlBasePath + path), replacing
// httpsrv Controller.UrlBase.
func UrlBase(c fiber.Ctx, p string) string {
	scheme := "http"
	if c.Secure() {
		scheme = "https"
	}
	urlBase := scheme + "://" + c.Host()
	elems := []string{"/"}
	if UrlBasePath != "" {
		elems = append(elems, UrlBasePath)
	}
	if p != "" {
		elems = append(elems, p)
	}
	return urlBase + stdpath.Join(elems...)
}

// UrlPath returns the request path, replacing httpsrv Request.UrlPath.
func UrlPath(c fiber.Ctx) string {
	return c.Path()
}

// RawAbsUrl returns the absolute URL of the current request, replacing httpsrv
// Request.RawAbsUrl.
func RawAbsUrl(c fiber.Ctx) string {
	scheme := "http"
	if c.Secure() {
		scheme = "https"
	}
	return scheme + "://" + c.Host() + c.OriginalURL()
}
