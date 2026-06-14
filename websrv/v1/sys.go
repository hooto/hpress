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
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hooto/httpsrv"
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/hooto/iam/v2/pkg/iamserver"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/config"
	"github.com/hooto/hpress/status"
	"github.com/hooto/hpress/store"
)

var (
	uptime time.Time
)

func init() {
	uptime = time.Now()
}

type Sys struct {
	*httpsrv.Controller
	us iamserver.UserSession
}

func (c *Sys) Init() int {

	c.us = iamserver.AppVerifier.Session(c.Request.Request)

	if _, err := c.us.RequireAuth(); err != nil {
		c.Response.Out.WriteHeader(401)
		c.RenderJson(types.NewTypeErrorMeta(iamapi.ErrCodeUnauthorized, "Unauthorized"))
		return 1
	}

	return 0
}

func (c Sys) ConfigListAction() {

	if !c.us.Allow("", "sys.admin") {
		c.RenderJson(types.TypeMeta{
			Error: &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"},
		})
		return
	}

	c.RenderJson(config.SysConfigList)
}

func (c Sys) ConfigSetAction() {

	var ls api.SysConfigList

	defer c.RenderJson(&ls)

	if !c.us.Allow("", "sys.admin") {
		ls.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return
	}

	err := c.Request.JsonDecode(&ls)
	if err != nil {
		ls.Error = &types.ErrorMeta{api.ErrCodeBadArgument, "Bad Argument " + err.Error()}
		return
	}

	for _, entry := range ls.Items {

		if prev := config.SysConfigList.Fetch(entry.Key); prev == nil {
			continue
		}

		q := store.Data.NewQueryer().From("hp_sys_config").Limit(1)
		q.Where().And("key", entry.Key)

		rs, err := store.Data.Query(q)
		if err != nil {
			ls.Error = &types.ErrorMeta{
				Code:    api.ErrCodeInternalError,
				Message: "Can not pull database instance",
			}
			return
		}

		set := map[string]interface{}{
			"value": entry.Value,
		}

		sync := false

		if len(rs) > 0 {

			if rs[0].Field("value").String() != entry.Value {

				ft := store.Data.NewFilter()
				ft.And("key", entry.Key)
				_, err = store.Data.Update("hp_sys_config", set, ft)
				sync = true
			}

		} else {

			set["key"] = entry.Key

			_, err = store.Data.Insert("hp_sys_config", set)
			sync = true
		}

		if err != nil {
			ls.Error = &types.ErrorMeta{
				Code:    api.ErrCodeInternalError,
				Message: err.Error(),
			}
			return
		}

		if entry.Key == "router_basepath_default" {
			entry.Value = filepath.Clean("/" + strings.TrimSpace(entry.Value))
			if entry.Value == "" || entry.Value == "." || entry.Value == "/" {
				entry.Value = "/"
				config.RouterBasepathDefaults = []string{}
			} else {
				config.RouterBasepathDefaults = strings.Split(strings.Trim(entry.Value, "/"), "/")
			}
			config.RouterBasepathDefault = entry.Value
		}

		if sync && entry.Key == "frontend_languages" {
			config.Languages = []*api.LangEntry{}
			if langs := api.LangsStringFilterArray(entry.Value); len(langs) > 0 {
				for _, lv := range langs {
					for _, lv2 := range api.LangArray {
						if lv == lv2.Id {
							config.Languages = append(config.Languages, lv2)
						}
					}
				}
			}
		}

		config.SysConfigList.Insert(entry)
	}

	ls.Kind = "SysConfigList"
}

func (c Sys) StatusAction() {

	set := api.SysStatus{}

	defer c.RenderJson(&set)

	if !c.us.Allow("", "sys.admin") {
		set.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return
	}

	set.InstanceID = config.Config.InstanceID
	set.AppVersion = config.Version
	set.AppRelease = config.Release
	set.RuntimeVersion = runtime.Version()
	set.Uptime = status.Uptime
	set.CoroutineNumber = runtime.NumGoroutine()

	ms := memStatsFetch()

	set.MemStats.Alloc = ms.Alloc
	set.MemStats.TotalAlloc = ms.TotalAlloc
	set.MemStats.Sys = ms.Sys

	set.MemStats.NextGC = ms.NextGC
	set.MemStats.LastGC = ms.LastGC
	set.MemStats.PauseTotalNs = ms.TotalAlloc
	set.MemStats.NumGC = ms.NumGC

	set.Info = sysinfoFetch()

	set.Kind = "SysStatus"
}

func (c Sys) IamStatusAction() {

	var sets api.SysIamStatus

	if !c.us.Allow("", "sys.admin") {
		sets.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return
	}

	inst_url := "://" + c.Request.Host
	if c.Request.TLS != nil {
		inst_url = "https" + inst_url
	} else {
		inst_url = "http" + inst_url
	}

	if len(httpsrv.DefaultService.Config.UrlBasePath) > 0 {
		inst_url += "/" + httpsrv.DefaultService.Config.UrlBasePath
	}

	cfg := iamserver.AppVerifier.Config()

	sets = api.SysIamStatus{
		InstanceSelf: &iamapi.AppInstance{
			ID:          config.Config.InstanceID,
			Name:        config.AppName,
			Version:     config.Version,
			Permissions: config.Perms,
			Url:         inst_url,
		},
	}

	if cfg != nil {
		sets.BaseURL = cfg.BaseURL
	}

	if status.IamServiceStatus == status.IamServiceOK {
		sets.InstanceRegistered = &iamapi.AppInstance{
			ID:          config.Config.InstanceID,
			Name:        config.AppName,
			Version:     config.Version,
			Permissions: config.Perms,
			Url:         inst_url,
		}
	}

	sets.Kind = "SysIamStatus"

	c.RenderJson(sets)
}

func (c Sys) IamSyncAction() {

	var rsp struct {
		types.TypeMeta `json:",inline"`
	}

	defer c.RenderJson(&rsp)

	if !c.us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return
	}

	status.Refresh()

	if status.IamServiceStatus == status.IamServiceOK {
		rsp.Kind = "AppInstanceRegister"
	} else {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeInternalError, "IAM sync failed"}
	}
}

func memStatsFetch() runtime.MemStats {

	var ms runtime.MemStats

	runtime.ReadMemStats(&ms)

	return ms
}

func sysinfoFetch() api.SysStatusInfo {

	// var si syscall.Sysinfo_t
	// syscall.Sysinfo(&si)

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	return api.SysStatusInfo{
		CpuNum:    runtime.NumCPU(),
		Uptime:    uptime.Unix(),       // si.Uptime,
		Loads:     [3]uint64{0, 0, 0},  // si.Loads,
		MemTotal:  ms.Alloc,            // si.Totalram,
		MemFree:   ms.Frees,            // si.Freeram,
		MemShared: 0,                   // si.Sharedram,
		MemBuffer: 0,                   // si.Bufferram,
		MemUsed:   ms.Alloc - ms.Frees, // si.Totalram - si.Freeram,
		SwapTotal: 0,                   // si.Totalswap,
		SwapFree:  0,                   // si.Freeswap,
		Procs:     0,                   // si.Procs,
		// TimeNow: time.Now().Format(time.RFC3339),
	}
}
