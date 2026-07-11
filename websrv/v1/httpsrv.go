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

	"github.com/hooto/hpress/websrv/web"
)

// Register mounts the v1 REST API routes on a fiber router. The caller mounts
// the router at "/hp/v1". Each route mirrors the path the httpsrv reflection
// dispatcher generated (camelToKebab of Controller/Action), and accepts any HTTP
// method (httpsrv dispatch was method-agnostic). Routes requiring a session use
// the web.Auth() gate; TextMarkdownRender is public.
func Register(router fiber.Router) {

	// Node
	router.All("/node/list", web.Auth(), NodeList)
	router.All("/node/entry", web.Auth(), NodeEntry)
	router.All("/node/set", web.Auth(), NodeSet)
	router.All("/node/del", web.Auth(), NodeDel)

	// NodeModel
	router.All("/node-model/entry", web.Auth(), NodeModelEntry)

	// Term
	router.All("/term/list", web.Auth(), TermList)
	router.All("/term/entry", web.Auth(), TermEntry)
	router.All("/term/set", web.Auth(), TermSet)

	// TermModel
	router.All("/term-model/entry", web.Auth(), TermModelEntry)

	// Text (public markdown render, no auth)
	router.All("/text/markdown-render", TextMarkdownRender)

	// Sys
	router.All("/sys/config-list", web.Auth(), SysConfigList)
	router.All("/sys/config-set", web.Auth(), SysConfigSet)
	router.All("/sys/status", web.Auth(), SysStatus)
	router.All("/sys/iam-status", web.Auth(), SysIamStatus)
	router.All("/sys/iam-sync", web.Auth(), SysIamSync)

	// S2Obj
	router.All("/s2-obj/list", web.Auth(), S2ObjList)
	router.All("/s2-obj/put", web.Auth(), S2ObjPut)
	router.All("/s2-obj/del", web.Auth(), S2ObjDel)
	router.All("/s2-obj/rename", web.Auth(), S2ObjRename)

	// ModSet
	router.All("/mod-set/spec-list", web.Auth(), ModSetSpecList)
	router.All("/mod-set/spec-entry", web.Auth(), ModSetSpecEntry)
	router.All("/mod-set/spec-info-set", web.Auth(), ModSetSpecInfoSet)
	router.All("/mod-set/spec-term-set", web.Auth(), ModSetSpecTermSet)
	router.All("/mod-set/spec-node-set", web.Auth(), ModSetSpecNodeSet)
	router.All("/mod-set/spec-action-set", web.Auth(), ModSetSpecActionSet)
	router.All("/mod-set/spec-action-del", web.Auth(), ModSetSpecActionDel)
	router.All("/mod-set/spec-route-set", web.Auth(), ModSetSpecRouteSet)
	router.All("/mod-set/spec-route-del", web.Auth(), ModSetSpecRouteDel)
	router.All("/mod-set/spec-lang-list", web.Auth(), ModSetSpecLangList)
	router.All("/mod-set/fs-tpl-list", web.Auth(), ModSetFsTplList)
	router.All("/mod-set/spec-upload-commit", web.Auth(), ModSetSpecUploadCommit)

	// ModSetFs
	router.All("/mod-set-fs/list", web.Auth(), ModSetFsList)
	router.All("/mod-set-fs/get", web.Auth(), ModSetFsGet)
	router.All("/mod-set-fs/put", web.Auth(), ModSetFsPut)
	router.All("/mod-set-fs/del", web.Auth(), ModSetFsDel)
	router.All("/mod-set-fs/rename", web.Auth(), ModSetFsRename)
}
