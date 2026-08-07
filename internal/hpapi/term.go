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
	"github.com/lessos/lessgo/types"
)

type Term struct {
	types.TypeMeta `json:",inline"`
	Model          *TermModel `json:"model,omitempty"`
	ID             uint32     `json:"id,omitempty"`
	PID            uint32     `json:"pid,omitempty"`
	UID            string     `json:"uid,omitempty"`
	Status         int16      `json:"status,omitempty"`
	UserID         string     `json:"userid,omitempty"`
	Title          string     `json:"title"`
	Weight         int32      `json:"weight,omitempty"`
	Created        uint32     `json:"created,omitempty"`
	Updated        uint32     `json:"updated,omitempty"`
}

type TermList struct {
	types.TypeMeta `json:",inline"`
	Meta           types.ListMeta `json:"meta,omitempty"`
	Model          *TermModel     `json:"model,omitempty"`
	Items          []Term         `json:"items"`
}
