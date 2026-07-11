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
	"github.com/gofiber/fiber/v3"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/config"
	"github.com/hooto/hpress/websrv/web"
)

func NodeModelEntry(c fiber.Ctx) error {

	rsp := api.NodeModel{}

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if !us.Allow("", "editor.read") {
		return nil
	}

	modname, modelid := web.Param(c, "modname"), web.Param(c, "modelid")

	nmodel, err := config.SpecNodeModel(modname, modelid)
	if err != nil {
		rsp.Error = &types.ErrorMeta{
			Code:    api.ErrCodeBadArgument,
			Message: "Model Not Found",
		}
		return nil
	}

	rsp = *nmodel
	rsp.Kind = "NodeModel"

	return nil
}
