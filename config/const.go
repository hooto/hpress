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
	"github.com/hooto/iam/v2/pkg/iamapi"
)

var (
	StorageServiceEndpoint = "/hp/s2/deft"
	coreModules            = []string{
		"core/general",
		"core/comment",
		"core/portal",
		"core/doc",
		"core/gdoc",
		"core/blog",
	}

	Perms = []*iamapi.AppPermission{
		{
			Permission: "frontend.list",
			Summary:    "Frontend - List",
			Roles:      []string{iamapi.Role_Guest, iamapi.Role_User},
		},
		{
			Permission: "frontend.read",
			Summary:    "Frontend - Read",
			Roles:      []string{iamapi.Role_Guest, iamapi.Role_User},
		},
		{
			Permission: "editor.list",
			Summary:    "Editor - List",
			Roles:      []string{},
		},
		{
			Permission: "editor.write",
			Summary:    "Editor - Write",
			Roles:      []string{},
		},
		{
			Permission: "editor.read",
			Summary:    "Editor - Read",
			Roles:      []string{},
		},
		{
			Permission: "sys.admin",
			Summary:    "System Admin",
			Roles:      []string{},
		},
	}
)
