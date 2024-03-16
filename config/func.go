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

package config

import (
	"path/filepath"

	"github.com/hooto/httpsrv"
)

func init() {
	httpsrv.DefaultService.Config.RegisterTemplateFunc("SysConfig", SysConfigList.FetchString)
	httpsrv.DefaultService.Config.RegisterTemplateFunc("ThemeConfig", ThemeConfigFetchString)
	httpsrv.DefaultService.Config.RegisterTemplateFunc("HttpSrvBasePath", HttpSrvBasePath)
}

func HttpSrvBasePath(uri string) string {

	if httpsrv.DefaultService.Config.UrlBasePath == "" {
		return filepath.Clean("/" + uri)
	}

	return filepath.Clean("/" + httpsrv.DefaultService.Config.UrlBasePath + "/" + uri)
}
