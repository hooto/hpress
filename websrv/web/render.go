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
	"html"
	"net/url"
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

// RenderAuthRequired renders a centered, self-contained modal page prompting the
// user to sign in. SPA shell routes (e.g. /hp/mgr2) call this when the IAM
// session is missing or expired (e.g. "auth-denied : iat expired"): at that
// point the SPA itself has not loaded, so it cannot show its in-app modal —
// this server-rendered page stands in for it. The Sign In button targets the
// local user-auth/sign-in handler with current_url set to returnURL, so a
// successful login returns the user to the page they were trying to reach.
func RenderAuthRequired(c fiber.Ctx, signInPath, returnURL, msg string) error {
	href := signInPath
	if returnURL != "" {
		sep := "?"
		if strings.Contains(signInPath, "?") {
			sep = "&"
		}
		href = signInPath + sep + "current_url=" + url.QueryEscape(returnURL)
	}
	if msg == "" {
		msg = "Your login session has expired or you are not signed in. Please sign in again."
	}
	c.Status(fiber.StatusUnauthorized)
	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Set("Cache-Control", "no-cache")
	_, err := c.Write([]byte(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Session Expired</title>
<style>
  html,body{height:100%;margin:0;font-family:system-ui,Segoe UI,Roboto,Helvetica,Arial,sans-serif;background:#f4f5f7}
  .overlay{position:fixed;inset:0;display:flex;align-items:center;justify-content:center}
  .box{background:#fff;border-radius:8px;width:440px;max-width:92vw;box-shadow:0 12px 40px rgba(0,0,0,.18);overflow:hidden}
  .body{padding:24px 24px 8px}
  .alert{border:1px solid #ffe08a;background:#fff8e1;color:#8a6d3b;border-radius:4px;padding:12px 14px;font-size:14px;line-height:1.55}
  .foot{padding:16px 24px 24px;text-align:right}
  .btn{display:inline-block;background:#0d6efd;color:#fff;text-decoration:none;border-radius:4px;padding:8px 20px;font-size:14px;font-weight:600}
  .btn:hover{background:#0b5ed7}
</style>
</head>
<body>
  <div class="overlay"><div class="box">
    <div class="body"><div class="alert">` + html.EscapeString(msg) + `</div></div>
    <div class="foot"><a class="btn" href="` + html.EscapeString(href) + `">Sign In</a></div>
  </div></div>
</body>
</html>`))
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
