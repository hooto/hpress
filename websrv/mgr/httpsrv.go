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

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/websrv/mgr/controllers"
	"github.com/hooto/hpress/websrv/web"
)

// Register mounts the management backend routes on a fiber router. The caller
// mounts the router at "/hp/mgr". It loads the admin view templates and serves
// the Index handler at the module root and /index (reproducing the httpsrv
// Index.IndexAction special-casing for a controller named "Index").
func Register(router fiber.Router) {

	web.Templates.Set("mgr", []string{config.Prefix + "/websrv/mgr/views"})

	router.All("/", controllers.Index)
	router.All("/index", controllers.Index)
}
