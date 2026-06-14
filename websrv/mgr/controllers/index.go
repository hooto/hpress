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
	"net/http"
	"time"

	"github.com/hooto/httpsrv"
	"github.com/hooto/iam/v2/pkg/iamserver"
	"github.com/sysinner/incore/v2/pkg/inauth"

	"github.com/hooto/hpress/config"
	"github.com/hooto/hpress/status"
)

type Index struct {
	*httpsrv.Controller
}

func (c Index) IndexAction() {

	status.Locker.RLock()
	defer status.Locker.RUnlock()

	if c.Params.Value("_iam_out") != "" {
		c.Redirect(c.UrlBase(""))
		return
	}

	if err := iamserver.AppVerifier.Ping(); err != nil {
		c.RenderError(500, "iam ping fail : "+err.Error())
		return
	}

	session := iamserver.AppVerifier.Session(c.Request.Request)
	if err := session.CheckServer(); err != nil {
		c.RenderError(500, "iam session check fail : "+err.Error())
		return
	}

	if redirectURL, err := session.RequireAuth(); err != nil {
		if redirectURL != "" {
			currentURL := c.Request.RawAbsUrl()
			http.SetCookie(c.Response.Out, &http.Cookie{
				Name:     inauth.AppHttpHeaderKey + "-current-url",
				Value:    currentURL,
				Path:     "/",
				HttpOnly: true,
				Expires:  time.Now().Add(1 * time.Hour),
			})
			c.Redirect(redirectURL)
			return
		}

		c.RenderError(401, "iam auth fail : "+err.Error())
		return
	}

	c.Response.Out.Header().Set("Cache-Control", "no-cache")

	if v := config.SysConfigList.FetchString("http_h_ac_allow_origin"); v != "" {
		c.Response.Out.Header().Set("Access-Control-Allow-Origin", v)
	}

	c.Data["sys_version_sign"] = config.SysVersionSign

	c.Render("index.tpl")
}
