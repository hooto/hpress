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

package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/hooto/hlang4g/hlang"
	"github.com/hooto/hlog4g/hlog"

	"github.com/lynkdb/lynkui/go/lynkui"
	"github.com/lynkdb/lynkui/go/uiserver"

	"github.com/hooto/hpress/config"
	"github.com/hooto/hpress/datax"
	"github.com/hooto/hpress/status"
	"github.com/hooto/hpress/websrv/web"

	cdef "github.com/hooto/hpress/websrv/frontend"
	cmgr "github.com/hooto/hpress/websrv/mgr"
	cmod "github.com/hooto/hpress/websrv/module"
	capi "github.com/hooto/hpress/websrv/v1"

	ext_captcha "github.com/hooto/hcaptcha/captcha4g/webfiber"
	ext_comment "github.com/hooto/hpress/modules/core/comment/websrv"
)

var (
	version = ""
	release = ""
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

	if version != "" {
		config.Version = version
	}
	if release != "" {
		config.Release = release
	}

	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("Version: %s, Release: %s\n", config.Version, config.Release)
		return
	}

	//
	retry := time.Second * 3
	for i := 0; ; i++ {

		err := config.Setup()
		if err == nil {
			break
		}

		if i >= 100 {
			fmt.Println(err)
			os.Exit(1)
		}

		hlog.Printf("error", "config.Setup error: %v", err)
		time.Sleep(retry)

		if retry < time.Minute {
			retry += time.Second
		}
	}

	// i18n
	hlang.StdLangFeed.LoadMessages(config.Prefix+"/i18n/en.json", true)
	hlang.StdLangFeed.LoadMessages(config.Prefix+"/i18n/zh-CN.json", true)
	hlang.StdLangFeed.Init()

	// status
	status.Init()
	datax.Worker()

	// fiber app + global panic recovery
	app := fiber.New()
	app.Use(recover.New())

	// UrlBasePath is the global prefix prepended to every route (mirrors the
	// prior httpsrv Config.UrlBasePath, applied once via the root group).
	web.UrlBasePath = config.Config.UrlBasePath
	root := app.Group(web.UrlBasePath)

	// external modules (note: "/hp/+" uses a literal "+", escaped for fiber)
	ext_comment.Register(root.Group("/hp/\\+/comment"))
	ext_captcha.Register(root.Group("/hp/\\+/hcaptcha"))

	// module static assets
	cmod.Register(root.Group("/hp/-"))

	// REST API + admin + /hp app routes
	capi.Register(root.Group("/hp/v1"))
	cmgr.Register(root.Group("/hp/mgr"))
	cdef.RegisterHtp(root.Group("/hp"))

	// lynkui admin UI (mounts /hp/lynkui on the app via its fiber entry).
	// MUST be registered before the public frontend "/*" catch-all below: fiber
	// v3 is order-sensitive with greedy wildcards, so a route mounted after a
	// root "/*" catch-all is shadowed by it (every /hp/lynkui/... request would
	// fall through to IndexPage and return an empty 200 instead of the asset).
	if _, err := uiserver.NewServiceFiber(app, &lynkui.ServiceConfig{
		AppProjectPath: config.Prefix + "/webui",
		UrlEntryPath:   "/hp/lynkui",
		RunMode:        "prod",
	}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// public frontend catch-all — registered LAST so the explicit /hp/* routes
	// (and lynkui's /hp/lynkui/* above) take priority over the "/*" match.
	cdef.Register(root.Group("/"))

	if err := app.Listen(fmt.Sprintf(":%d", config.Config.HttpPort)); err != nil {
		fmt.Println("app.Listen error", err)
		os.Exit(1)
	}
}
