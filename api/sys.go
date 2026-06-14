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

package api

import (
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/lessos/lessgo/types"
)

type SysStatus struct {
	types.TypeMeta  `json:",inline"`
	Meta            types.ObjectMeta `json:"meta,omitempty"`
	InstanceID      string           `json:"instance_id,omitempty"`
	AppVersion      string           `json:"app_version,omitempty"`
	AppRelease      string           `json:"app_release,omitempty"`
	RuntimeVersion  string           `json:"runtime_version,omitempty"`
	Uptime          string           `json:"uptime,omitempty"`
	CoroutineNumber int              `json:"coroutine_number,omitempty"`
	Info            SysStatusInfo    `json:"info"`
	MemStats        SysMemStats      `json:"memstats"`
}

type SysMemStats struct {
	//
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	//
	NextGC       uint64 `json:"next_gc"`
	LastGC       uint64 `json:"last_gc"`
	PauseTotalNs uint64 `json:"pause_total_ns"`
	NumGC        uint32 `json:"num_gc"`
}

type SysStatusInfo struct {
	CpuNum    int       `json:"cpu_num"`
	Uptime    int64     `json:"uptime"`
	Loads     [3]uint64 `json:"loads"`
	MemTotal  uint64    `json:"mem_total"`
	MemFree   uint64    `json:"mem_free"`
	MemShared uint64    `json:"mem_shared"`
	MemBuffer uint64    `json:"mem_buffer"`
	MemUsed   uint64    `json:"mem_used"`
	SwapTotal uint64    `json:"swap_total"`
	SwapFree  uint64    `json:"swap_free"`
	Procs     uint16    `json:"procs"`
}

type SysIamStatus struct {
	types.TypeMeta     `json:",inline"`
	BaseURL            string              `json:"base_url"`
	InstanceSelf       *iamapi.AppInstance `json:"instance_self"`
	InstanceRegistered *iamapi.AppInstance `json:"instance_registered"`
}

type SysConfig struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Comment string `json:"comment,omitempty"`
	Type    string `json:"type,omitempty"`
}

type SysConfigList struct {
	types.TypeMeta `json:",inline"`
	Items          []SysConfig `json:"items,omitempty"`
}

func (ls *SysConfigList) Insert(entry SysConfig) {

	for i, prev := range ls.Items {

		if prev.Key == entry.Key {

			if prev.Value != entry.Value {
				ls.Items[i].Value = entry.Value
			}

			if entry.Comment != "" && prev.Comment != entry.Comment {
				ls.Items[i].Comment = entry.Comment
			}

			return
		}
	}

	ls.Items = append(ls.Items, entry)
}

func (ls *SysConfigList) Fetch(key string) *SysConfig {

	for _, prev := range ls.Items {

		if prev.Key == key {
			return &prev
		}
	}

	return nil
}

func (ls *SysConfigList) FetchString(key string) string {

	if v := ls.Fetch(key); v != nil {
		return v.Value
	}

	return ""
}

func (ls *SysConfigList) Remove(key string) {

	for i, prev := range ls.Items {

		if prev.Key == key {
			ls.Items = append(ls.Items[:i], ls.Items[i:]...)
			break
		}
	}
}
