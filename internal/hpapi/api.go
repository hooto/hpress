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
	"fmt"

	"github.com/lessos/lessgo/encoding/json"
)

const (
	Version = "0.0.0.dev00"
)

const (
	ErrCodeBadArgument   = "BadArgument"
	ErrCodeInternalError = "InternalError"
	ErrCodeNotFound      = "NotFound"
)

func NsSysDataPull() []byte {
	return []byte("hp:sys:config:ext_data_pull")
}

func NsSysNodeSearch(bukname string) []byte {
	return []byte("hp:sys:config:ext_node_search:" + bukname)
}

func NsTextSearchCacheNodeEntry(bukname, id string) []byte {
	return []byte("hp:cache:node:" + bukname + ":" + id)
}

func ObjPrint(name string, obj interface{}) {
	js, _ := json.Encode(obj, "  ")
	fmt.Println(name, string(js))
}
