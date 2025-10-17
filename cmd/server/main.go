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

	"github.com/hooto/hlang4g/hlang"
	"github.com/hooto/hlog4g/hlog"
	"github.com/hooto/httpsrv"
	"github.com/hooto/iam/iamclient"

	"github.com/lynkdb/lynkui/go/lynkui"
	"github.com/lynkdb/lynkui/go/uiserver"

	"github.com/hooto/hpress/config"
	"github.com/hooto/hpress/datax"
	"github.com/hooto/hpress/status"

	cdef "github.com/hooto/hpress/websrv/frontend"
	cmgr "github.com/hooto/hpress/websrv/mgr"
	cmod "github.com/hooto/hpress/websrv/module"
	capi "github.com/hooto/hpress/websrv/v1"

	ext_captcha "github.com/hooto/hcaptcha/captcha4g"
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

		// fmt.Println("Error on config.Setup", err)
		hlog.Printf("error", "config.Setup error: %v", err)
		time.Sleep(retry)

		if retry < time.Minute {
			retry += time.Second
		}
	}

	// ext_captcha.DataConnector = store.DataLocal
	// if err := ext_captcha.Config(config.CaptchaConfig); err != nil {
	// 	hlog.Printf("error", "ext_captcha.Config error: %v", err)
	// 	fmt.Println("ext_captcha.Config error", err)
	// 	os.Exit(1)
	// }

	if _, err := uiserver.NewService(httpsrv.DefaultService, &lynkui.ServiceConfig{
		AppProjectPath: config.Prefix + "/webui",
		UrlEntryPath:   "/hp/lynkui",
		RunMode:        "prod",
	}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	iamclient.ServiceUrl = config.Config.IamServiceUrl
	iamclient.ServiceUrlFrontend = config.Config.IamServiceUrlFrontend

	iamclient.InstanceID = config.Config.InstanceID
	iamclient.InstanceOwner = config.Config.AppInstance.Meta.User

	httpsrv.DefaultService.SetLogger(httpsrv.NewRawLogger())

	httpsrv.DefaultService.Config.UrlBasePath = config.Config.UrlBasePath
	httpsrv.DefaultService.Config.HttpPort = config.Config.HttpPort
	// httpsrv.DefaultService.Config.CompressResponse = true

	// i18n
	hlang.StdLangFeed.LoadMessages(config.Prefix+"/i18n/en.json", true)
	hlang.StdLangFeed.LoadMessages(config.Prefix+"/i18n/zh-CN.json", true)
	hlang.StdLangFeed.Init()

	// status
	status.Init()
	datax.Worker()

	//
	httpsrv.DefaultService.HandleModule("/hp/+/comment", ext_comment.NewModule())
	httpsrv.DefaultService.HandleModule("/hp/+/hcaptcha", ext_captcha.WebServerModule())

	//
	httpsrv.DefaultService.HandleModule("/hp/-", cmod.NewModule())

	//
	httpsrv.DefaultService.HandleModule("/hp/v1", capi.NewModule())
	httpsrv.DefaultService.HandleModule("/hp/mgr", cmgr.NewModule())
	httpsrv.DefaultService.HandleModule("/hp", cdef.NewHtpModule())
	httpsrv.DefaultService.HandleModule("/", cdef.NewModule())

	// if err := httpsrv.DefaultService.Start(); err != nil {
	// 	fmt.Println("srv.Start error", err)
	// 	os.Exit(1)
	// }

	httpsrv.DefaultService.Start()
}
