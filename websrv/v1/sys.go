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

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/hooto/iam/v2/pkg/iamserver"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/hpapi"
	"github.com/hooto/hpress/internal/status"
	"github.com/hooto/hpress/internal/store"
	"github.com/hooto/hpress/websrv/web"
)

var (
	uptime time.Time
)

func init() {
	uptime = time.Now()
}

func SysConfigList(c fiber.Ctx) error {

	us := web.AuthSession(c)

	if us == nil || !us.Allow("", "sys.admin") {
		return web.JSON(c, types.TypeMeta{
			Error: &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"},
		})
	}

	return web.JSON(c, config.SysConfigList)
}

func SysConfigSet(c fiber.Ctx) error {

	var ls hpapi.SysConfigList

	defer func() { _ = web.JSON(c, &ls) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		ls.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	err := web.Bind(c, &ls)
	if err != nil {
		ls.Error = &types.ErrorMeta{hpapi.ErrCodeBadArgument, "Bad Argument " + err.Error()}
		return nil
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
				Code:    hpapi.ErrCodeInternalError,
				Message: "Can not pull database instance",
			}
			return nil
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
				Code:    hpapi.ErrCodeInternalError,
				Message: err.Error(),
			}
			return nil
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
			config.Languages = []*hpapi.LangEntry{}
			if langs := hpapi.LangsStringFilterArray(entry.Value); len(langs) > 0 {
				for _, lv := range langs {
					for _, lv2 := range hpapi.LangArray {
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

	return nil
}

func SysStatus(c fiber.Ctx) error {

	set := hpapi.SysStatus{}

	defer func() { _ = web.JSON(c, &set) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		set.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
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

	return nil
}

func SysIamStatus(c fiber.Ctx) error {

	var sets hpapi.SysIamStatus

	defer func() { _ = web.JSON(c, &sets) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		sets.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	inst_url := "://" + c.Host()
	if c.Secure() {
		inst_url = "https" + inst_url
	} else {
		inst_url = "http" + inst_url
	}

	if len(web.UrlBasePath) > 0 {
		inst_url += "/" + web.UrlBasePath
	}

	cfg := iamserver.AppVerifier.Config()

	sets = hpapi.SysIamStatus{
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
		sets.AppID = cfg.AppId
		sets.SecretKey = maskSecret(cfg.SecretKey)
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

	return nil
}

func SysIamSync(c fiber.Ctx) error {

	var rsp struct {
		types.TypeMeta `json:",inline"`
	}

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	status.Refresh()

	if status.IamServiceStatus == status.IamServiceOK {
		rsp.Kind = "AppInstanceRegister"
	} else {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeInternalError, "IAM sync failed"}
	}

	return nil
}

func memStatsFetch() runtime.MemStats {

	var ms runtime.MemStats

	runtime.ReadMemStats(&ms)

	return ms
}

func sysinfoFetch() hpapi.SysStatusInfo {

	// var si syscall.Sysinfo_t
	// syscall.Sysinfo(&si)

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	return hpapi.SysStatusInfo{
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

// maskSecret returns a redacted view of a secret for display, keeping the
// first 4 and last 4 characters and replacing the middle with asterisks.
// Short or empty values are fully masked so no usable material leaks.
func maskSecret(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "******"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
