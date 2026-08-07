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
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lessos/lessgo/types"
)

var (
	SrvNameReg = regexp.MustCompile("^[0-9a-z\\-_]{1,50}$")
)

type Spec struct {
	types.TypeMeta `json:",inline"`
	Meta           types.InnerObjectMeta `json:"meta,omitempty"`
	SrvName        string                `json:"srvname"`
	Status         int16                 `json:"status,omitempty"`
	Title          string                `json:"title"`
	Comment        string                `json:"comment,omitempty"`
	NodeModels     []*NodeModel          `json:"nodeModels,omitempty"`
	TermModels     []TermModel           `json:"termModels,omitempty"`
	Actions        []Action              `json:"actions,omitempty"`
	Views          []View                `json:"views,omitempty"`
	Router         Router                `json:"router,omitempty"`
	ThemeConfig    string                `json:"theme_config,omitempty"`
}

func (it *Spec) NodeModelGet(name string) *NodeModel {
	for _, v := range it.NodeModels {
		if v.Meta.Name == name {
			return v
		}
	}
	return nil
}

type SpecUploadCommit struct {
	types.TypeMeta `json:",inline"`
	Size           int64  `json:"size"`
	Name           string `json:"name"`
	Data           string `json:"data"`
}

func SrvNameFilter(name string) (string, error) {

	name = strings.Replace(strings.Trim(filepath.Clean(strings.ToLower(name)), "/"), "/", "-", -1)

	if mat := SrvNameReg.MatchString(name); !mat {
		return "", fmt.Errorf("Invalid Service Name (%s)", name)
	}

	return name, nil
}

type View struct {
	Path string `json:"path"`
}

type ViewList struct {
	types.TypeMeta `json:",inline"`
	Items          []View `json:"items,omitempty"`
}

type SpecList struct {
	types.TypeMeta `json:",inline"`
	Items          []Spec `json:"items,omitempty"`
}

type FieldModel struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Length      string        `json:"length,omitempty"`
	Extra       []string      `json:"extra,omitempty"`
	Attrs       types.KvPairs `json:"attrs,omitempty"`
	IndexType   int           `json:"indexType,omitempty"`
	Title       string        `json:"title"`
	EditDisable bool          `json:"edit_disable,omitempty"`
	Comment     string        `json:"comment,omitempty"`
}

type NodeModel struct {
	types.TypeMeta `json:",inline"`
	Meta           types.ObjectMeta `json:"meta,omitempty"`
	ModName        string           `json:"modname,omitempty"`
	SrvName        string           `json:"-"`
	Status         int16            `json:"status,omitempty"`
	Title          string           `json:"title,omitempty"`
	Comment        string           `json:"comment,omitempty"`
	Fields         []FieldModel     `json:"fields,omitempty"`
	Terms          []TermModel      `json:"terms,omitempty"`
	Extensions     SpecExtensions   `json:"extensions,omitempty"`
}

func (item *NodeModel) Field(name string) *FieldModel {
	for _, v := range item.Fields {
		if name == v.Name {
			return &v
		}
	}
	return nil
}

var (
	PermalinkNameReg = regexp.MustCompile("^[0-9a-z_-]{1,100}$")
)

func PermalinkNameFilter(name string) (string, error) {

	name = strings.Replace(filepath.Clean(strings.ToLower(strings.TrimSpace(name))), " ", "-", -1)

	if mat := PermalinkNameReg.MatchString(name); !mat {
		return "", fmt.Errorf("Invalid Permalink Name (%s)", name)
	}

	return name, nil
}

type SpecExtensions struct {
	AccessCounter   bool   `json:"access_counter,omitempty"`
	CommentEnable   bool   `json:"comment_enable,omitempty"`
	CommentPerEntry bool   `json:"comment_perentry,omitempty"`
	Permalink       string `json:"permalink,omitempty"`
	NodeRefer       string `json:"node_refer,omitempty"`
	NodeSubRefer    string `json:"node_sub_refer,omitempty"`
	TextSearch      bool   `json:"text_search,omitempty"`
}

type NodeModelList struct {
	types.TypeMeta `json:",inline"`
	Items          []NodeModel `json:"items,omitempty"`
}

const (
	TermTag      = "tag"
	TermTaxonomy = "taxonomy"
)

type TermModel struct {
	types.TypeMeta `json:",inline"`
	Meta           types.ObjectMeta `json:"meta,omitempty"`
	ModName        string           `json:"modname,omitempty"`
	Status         int16            `json:"status,omitempty"`
	Type           string           `json:"type,omitempty"`
	Title          string           `json:"title,omitempty"`
	Comment        string           `json:"comment,omitempty"`
}

type TermModelList struct {
	types.TypeMeta `json:",inline"`
	Items          []TermModel `json:"items,omitempty"`
}

type Query struct {
	Spec   string   `json:"spec,omitempty"`
	Table  string   `json:"table,omitempty"`
	Fields string   `json:"fields,omitempty"`
	Order  string   `json:"order,omitempty"`
	Limit  int64    `json:"limit,omitempty"`
	Offset int64    `json:"offset,omitempty"`
	Filter []string `json:"filter,omitempty"`
}

type Action struct {
	types.TypeMeta `json:",inline"`
	Name           string       `json:"name"`
	ModName        string       `json:"modname,omitempty"`
	Datax          []ActionData `json:"datax,omitempty"`
}

type ActionData struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Pager    bool   `json:"pager,omitempty"`
	Query    Query  `json:"query,omitempty"`
	CacheTTL int64  `json:"cache_ttl,omitempty"` // cache time to live in milliseconds
}
