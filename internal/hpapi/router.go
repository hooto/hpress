// Copyright 2014 lessOS.com. All rights reserved.
//
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.
package hpapi

import (
	"github.com/lessos/lessgo/types"
)

type Router struct {
	Routes []Route `json:"routes,omitempty"`
	// DefaultPagelet string  `json:"defaultPagelet,omitempty"` // e.g. index.tpl
}

type Route struct {
	types.TypeMeta `json:",inline"`
	Path           string            `json:"path"` // e.g. /app/:id
	DataAction     string            `json:"dataAction,omitempty"`
	Template       string            `json:"template,omitempty"` // e.g. index.tpl
	Params         map[string]string `json:"params,omitempty"`
	Tree           []string          `json:"-"`
	ModName        string            `json:"modname,omitempty"`
	Default        bool              `json:"default,omitempty"`
}
