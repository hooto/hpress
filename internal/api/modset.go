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
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/internal/hpapi"
	"github.com/hooto/hpress/internal/modset"
	"github.com/hooto/hpress/internal/store"
	"github.com/hooto/hpress/internal/web"
)

func ModSetSpecList(c fiber.Ctx) error {

	rsp := hpapi.SpecList{}

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "editor.list") {
		rsp.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	q := store.Data.NewQueryer().From("hp_modules").Limit(100)
	rs := store.Data.Query(q)
	if rs.Err() != nil {
		rsp.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, "Can not pull database instance")
		return nil
	}

	for rs.Valid() {

		var entry hpapi.Spec

		if err := rs.Field("body").JsonDecode(&entry); err == nil {
			entry.SrvName, _ = hpapi.SrvNameFilter(rs.Field("srvname").String())
			rsp.Items = append(rsp.Items, entry)
		}

		rs.Next()
	}

	rsp.Kind = "SpecList"

	return nil
}

func ModSetSpecEntry(c fiber.Ctx) error {

	rsp := hpapi.Spec{}

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "editor.read") {
		rsp.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	if web.Param(c, "name") == "" {
		rsp.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Object Not Found")
		return nil
	}

	name, err := modset.ModNameFilter(web.Param(c, "name"))
	if err != nil {
		rsp.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	q := store.Data.NewQueryer().From("hp_modules").Limit(1)
	q.Where().And("name", name)
	rs := store.Data.Query(q)
	if rs.Err() != nil {
		rsp.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, "Can not pull database instance")
		return nil
	}

	if rs.NotFound() {
		rsp.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Object Not Found")
		return nil
	}

	if err := rs.Field("body").JsonDecode(&rsp); err != nil {
		rsp.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, err.Error())
		return nil
	}

	rsp.SrvName, _ = hpapi.SrvNameFilter(rs.Field("srvname").String())

	rsp.Kind = "Spec"

	return nil
}

func ModSetSpecInfoSet(c fiber.Ctx) error {

	var set hpapi.Spec

	defer func() { _ = web.JSON(c, &set) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "editor.write") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	err := web.Bind(c, &set)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return nil
	}

	set.Meta.Name, err = modset.ModNameFilter(set.Meta.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.SrvName, err = hpapi.SrvNameFilter(set.SrvName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	if _, err = modset.SpecFetch(set.Meta.Name); err != nil {
		// Module creation is CLI-only (hpress module-init + .ipk upload);
		// this endpoint updates an installed module's info only.
		set.Error = types.NewErrorMeta(hpapi.ErrCodeNotFound,
			"Spec Not Found, please install a module package first")
		return nil
	}

	if err = modset.SpecInfoSet(set); err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, err.Error())
		return nil
	}

	specUpdated, err := modset.SpecFetch(set.Meta.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, err.Error())
		return nil
	}

	modset.SpecSchemaSync(specUpdated)

	set.Kind = "Spec"

	return nil
}

func ModSetSpecTermSet(c fiber.Ctx) error {

	var set hpapi.TermModel

	defer func() { _ = web.JSON(c, &set) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	err := web.Bind(c, &set)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return nil
	}

	set.Meta.Name, err = modset.ModelNameFilter(set.Meta.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.Type = strings.ToLower(set.Type)
	if set.Type != "tag" && set.Type != "taxonomy" {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Invalid Type")
		return nil
	}

	_, err = modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "ModName Not Found")
		return nil
	}

	err = modset.SpecTermSet(set.ModName, set)

	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, err.Error())
		return nil
	}

	specUpdated, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, err.Error())
		return nil
	}

	modset.SpecSchemaSync(specUpdated)

	set.Kind = "TermModel"

	return nil
}

func ModSetSpecNodeSet(c fiber.Ctx) error {

	var set hpapi.NodeModel

	defer func() { _ = web.JSON(c, &set) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	err := web.Bind(c, &set)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return nil
	}

	set.Meta.Name, err = modset.ModelNameFilter(set.Meta.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	_, err = modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "ModName Not Found")
		return nil
	}

	err = modset.SpecNodeSet(set.ModName, &set)

	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, err.Error())
		return nil
	}
	specUpdated, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, err.Error())
		return nil
	}

	modset.SpecSchemaSync(specUpdated)

	set.Kind = "NodeModel"

	return nil
}

func ModSetSpecActionSet(c fiber.Ctx) error {

	var set hpapi.Action

	defer func() { _ = web.JSON(c, &set) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	err := web.Bind(c, &set)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return nil
	}

	set.Name, err = modset.ModelNameFilter(set.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	_, err = modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "ModName Not Found")
		return nil
	}

	err = modset.SpecActionSet(set.ModName, set)

	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, err.Error())
		return nil
	}

	specUpdated, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, err.Error())
		return nil
	}

	modset.SpecSchemaSync(specUpdated)

	set.Kind = "Action"

	return nil
}

func ModSetSpecActionDel(c fiber.Ctx) error {

	var set hpapi.Action

	defer func() { _ = web.JSON(c, &set) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	err := web.Bind(c, &set)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return nil
	}

	set.Name, err = modset.ModelNameFilter(set.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	_, err = modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "ModName Not Found")
		return nil
	}

	err = modset.SpecActionDel(set.ModName, set)

	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, err.Error())
		return nil
	}

	specUpdated, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, err.Error())
		return nil
	}

	modset.SpecSchemaSync(specUpdated)

	set.Kind = "Action"

	return nil
}

func ModSetSpecRouteSet(c fiber.Ctx) error {

	var set hpapi.Route

	defer func() { _ = web.JSON(c, &set) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	err := web.Bind(c, &set)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return nil
	}

	set.Path, err = modset.RoutePathFilter(set.Path)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	_, err = modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "ModName Not Found")
		return nil
	}

	err = modset.SpecRouteSet(set.ModName, set)

	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, err.Error())
		return nil
	}

	specUpdated, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, err.Error())
		return nil
	}

	modset.SpecSchemaSync(specUpdated)

	set.Kind = "SpecRoute"

	return nil
}

func ModSetSpecRouteDel(c fiber.Ctx) error {

	var set hpapi.Route

	defer func() { _ = web.JSON(c, &set) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	err := web.Bind(c, &set)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return nil
	}

	set.Path, err = modset.RoutePathFilter(set.Path)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	err = modset.SpecRouteDel(set.ModName, set)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, err.Error())
		return nil
	}

	specUpdated, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeInternalError, err.Error())
		return nil
	}

	modset.SpecSchemaSync(specUpdated)

	set.Kind = "SpecRoute"

	return nil
}

func ModSetSpecLangList(c fiber.Ctx) error {
	ls := hpapi.LangList{
		Items: hpapi.LangArray,
	}
	ls.Kind = "SpecLangList"
	return web.JSON(c, ls)
}
