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

package v1

import (
	"strings"

	"github.com/hooto/httpsrv"
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/hooto/iam/v2/pkg/iamserver"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/config"
)

type TermModel struct {
	*httpsrv.Controller
	us iamserver.UserSession
}

func (c *TermModel) Init() int {

	c.us = iamserver.AppVerifier.Session(c.Request.Request)

	if _, err := c.us.RequireAuth(); err != nil {
		c.Response.Out.WriteHeader(401)
		c.RenderJson(types.NewTypeErrorMeta(iamapi.ErrCodeUnauthorized, "Unauthorized"))
		return 1
	}

	return 0
}

func (c TermModel) EntryAction() {

	rsp := api.TermModel{
		TypeMeta: types.TypeMeta{
			APIVersion: api.Version,
		},
	}

	defer c.RenderJson(&rsp)

	if !c.us.Allow("", "editor.read") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return
	}

	modname, modelid := c.Params.Value("modname"), c.Params.Value("modelid")
	if c.Params.Value("id") != "" {
		if s := strings.Split(c.Params.Value("id"), ","); len(s) == 2 {
			modname, modelid = s[0], s[1]
		}
	}

	model, err := config.SpecTermModel(modname, modelid)
	if err != nil {
		rsp.Error = &types.ErrorMeta{
			Code:    api.ErrCodeBadArgument,
			Message: "Model Not Found",
		}
		return
	}

	rsp = *model
	rsp.Kind = "TermModel"
}
