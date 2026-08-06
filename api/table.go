// Copyright 2015 Eryx <evorui at gmail dot com>, All rights reserved.
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

import "strings"

// Per-module data-table prefixes. Node data lives in hpn_* tables, term
// (tag/taxonomy) data in hpt_*. System tables use the hp_ prefix instead.
const (
	NodeTablePrefix = "hpn_"
	TermTablePrefix = "hpt_"
)

// ModuleTableKey normalizes a module name (e.g. "core/blog") into the
// identifier segment embedded in its data-table names (e.g. "core_blog").
// The "/" namespace separator becomes "_".
func ModuleTableKey(modName string) string {
	return strings.ReplaceAll(modName, "/", "_")
}

// NodeTableName returns the per-module node table name for the given module
// and model, e.g. NodeTableName("core/blog", "post") -> "hpn_core_blog_post".
func NodeTableName(modName, model string) string {
	return NodeTablePrefix + ModuleTableKey(modName) + "_" + model
}

// TermTableName returns the per-module term table name for the given module
// and model, e.g. TermTableName("core/blog", "tag") -> "hpt_core_blog_tag".
func TermTableName(modName, model string) string {
	return TermTablePrefix + ModuleTableKey(modName) + "_" + model
}
