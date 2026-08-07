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

package hpapi

import (
	"strings"

	"github.com/lessos/lessgo/types"
)

var (
	LangArray = []*LangEntry{
		{"en-US", "English"},
		{"zh-CN", "简体中文"},
		{"zh-TW", "繁體中文"},
	}
)

func init() {
	for i, v := range LangArray {
		LangArray[i].Id = strings.ToLower(v.Id)
	}
}

type LangEntry struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type LangList struct {
	types.TypeMeta `json:",inline"`
	Items          []*LangEntry `json:"items"`
}

func LangsStringFilter(str string) string {
	if fls := LangsStringFilterArray(str); len(fls) > 1 {
		return strings.Join(fls, ",")
	}
	return ""
}

func LangHit(ls []*LangEntry, lang string) string {
	if len(ls) > 0 {
		for _, v := range ls {
			if v.Id == lang {
				return v.Id
			}
		}
		return ls[0].Id
	}
	return lang
}

func LangsStringFilterArray(str string) types.ArrayString {

	var (
		fls = types.ArrayString{}
		ls  = strings.Split(strings.ToLower(str), ",")
	)
	for _, v := range ls {
		for _, v2 := range LangArray {
			if v == v2.Id {
				fls.Set(v2.Id)
				break
			}
		}
	}
	return fls
}
