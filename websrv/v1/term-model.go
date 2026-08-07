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

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/hpapi"
	"github.com/hooto/hpress/websrv/web"
)

func TermModelEntry(c fiber.Ctx) error {

	rsp := hpapi.TermModel{
		TypeMeta: types.TypeMeta{
			APIVersion: hpapi.Version,
		},
	}

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "editor.read") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	modname, modelid := web.Param(c, "modname"), web.Param(c, "modelid")
	if web.Param(c, "id") != "" {
		if s := strings.Split(web.Param(c, "id"), ","); len(s) == 2 {
			modname, modelid = s[0], s[1]
		}
	}

	model, err := config.SpecTermModel(modname, modelid)
	if err != nil {
		rsp.Error = &types.ErrorMeta{
			Code:    hpapi.ErrCodeBadArgument,
			Message: "Model Not Found",
		}
		return nil
	}

	rsp = *model
	rsp.Kind = "TermModel"

	return nil
}
