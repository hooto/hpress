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
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/modset"
	"github.com/hooto/hpress/store"
	"github.com/hooto/hpress/websrv/web"
)

func ModSetSpecList(c fiber.Ctx) error {

	rsp := api.SpecList{}

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "editor.list") {
		rsp.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	q := store.Data.NewQueryer().From("hp_modules").Limit(100)
	rs, err := store.Data.Query(q)
	if err != nil {
		rsp.Error = types.NewErrorMeta(api.ErrCodeInternalError, "Can not pull database instance")
		return nil
	}

	for _, v := range rs {

		var entry api.Spec

		if err := v.Field("body").JsonDecode(&entry); err == nil {
			entry.SrvName, _ = api.SrvNameFilter(v.Field("srvname").String())
			rsp.Items = append(rsp.Items, entry)
		}
	}

	rsp.Kind = "SpecList"

	return nil
}

func ModSetSpecEntry(c fiber.Ctx) error {

	rsp := api.Spec{}

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "editor.read") {
		rsp.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	if web.Param(c, "name") == "" {
		rsp.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Object Not Found")
		return nil
	}

	name, err := modset.ModNameFilter(web.Param(c, "name"))
	if err != nil {
		rsp.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return nil
	}

	q := store.Data.NewQueryer().From("hp_modules").Limit(1)
	q.Where().And("name", name)
	rs, err := store.Data.Query(q)
	if err != nil {
		rsp.Error = types.NewErrorMeta(api.ErrCodeInternalError, "Can not pull database instance")
		return nil
	}

	if len(rs) < 1 {
		rsp.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Object Not Found")
		return nil
	}

	if err := rs[0].Field("body").JsonDecode(&rsp); err != nil {
		rsp.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return nil
	}

	rsp.SrvName, _ = api.SrvNameFilter(rs[0].Field("srvname").String())

	rsp.Kind = "Spec"

	return nil
}

func ModSetSpecInfoSet(c fiber.Ctx) error {

	var set api.Spec

	defer func() { _ = web.JSON(c, &set) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "editor.write") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	err := web.Bind(c, &set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return nil
	}

	set.Meta.Name, err = modset.ModNameFilter(set.Meta.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.SrvName, err = api.SrvNameFilter(set.SrvName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return nil
	}

	if _, err = modset.SpecFetch(set.Meta.Name); err != nil {

		if err = modset.SpecInfoNew(set); err != nil {
			set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
			return nil
		}

	} else {

		if err = modset.SpecInfoSet(set); err != nil {
			set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
			return nil
		}
	}

	seted, err := modset.SpecFetch(set.Meta.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return nil
	}

	modset.SpecSchemaSync(seted)

	set.Kind = "Spec"

	return nil
}

func ModSetSpecTermSet(c fiber.Ctx) error {

	var set api.TermModel

	defer func() { _ = web.JSON(c, &set) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	err := web.Bind(c, &set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return nil
	}

	set.Meta.Name, err = modset.ModelNameFilter(set.Meta.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.Type = strings.ToLower(set.Type)
	if set.Type != "tag" && set.Type != "taxonomy" {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Invalid Type")
		return nil
	}

	_, err = modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "ModName Not Found")
		return nil
	}

	err = modset.SpecTermSet(set.ModName, set)

	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return nil
	}

	seted, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return nil
	}

	modset.SpecSchemaSync(seted)

	set.Kind = "TermModel"

	return nil
}

func ModSetSpecNodeSet(c fiber.Ctx) error {

	var set api.NodeModel

	defer func() { _ = web.JSON(c, &set) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	err := web.Bind(c, &set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return nil
	}

	set.Meta.Name, err = modset.ModelNameFilter(set.Meta.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return nil
	}

	_, err = modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "ModName Not Found")
		return nil
	}

	err = modset.SpecNodeSet(set.ModName, &set)

	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return nil
	}
	seted, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return nil
	}

	modset.SpecSchemaSync(seted)

	set.Kind = "NodeModel"

	return nil
}

func ModSetSpecActionSet(c fiber.Ctx) error {

	var set api.Action

	defer func() { _ = web.JSON(c, &set) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	err := web.Bind(c, &set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return nil
	}

	set.Name, err = modset.ModelNameFilter(set.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return nil
	}

	_, err = modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "ModName Not Found")
		return nil
	}

	err = modset.SpecActionSet(set.ModName, set)

	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return nil
	}

	seted, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return nil
	}

	modset.SpecSchemaSync(seted)

	set.Kind = "Action"

	return nil
}

func ModSetSpecActionDel(c fiber.Ctx) error {

	var set api.Action

	defer func() { _ = web.JSON(c, &set) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	err := web.Bind(c, &set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return nil
	}

	set.Name, err = modset.ModelNameFilter(set.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return nil
	}

	_, err = modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "ModName Not Found")
		return nil
	}

	err = modset.SpecActionDel(set.ModName, set)

	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return nil
	}

	seted, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return nil
	}

	modset.SpecSchemaSync(seted)

	set.Kind = "Action"

	return nil
}

func ModSetSpecRouteSet(c fiber.Ctx) error {

	var set api.Route

	defer func() { _ = web.JSON(c, &set) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	err := web.Bind(c, &set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return nil
	}

	set.Path, err = modset.RoutePathFilter(set.Path)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return nil
	}

	_, err = modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "ModName Not Found")
		return nil
	}

	err = modset.SpecRouteSet(set.ModName, set)

	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return nil
	}

	seted, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return nil
	}

	modset.SpecSchemaSync(seted)

	set.Kind = "SpecRoute"

	return nil
}

func ModSetSpecRouteDel(c fiber.Ctx) error {

	var set api.Route

	defer func() { _ = web.JSON(c, &set) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	err := web.Bind(c, &set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return nil
	}

	set.Path, err = modset.RoutePathFilter(set.Path)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return nil
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return nil
	}

	err = modset.SpecRouteDel(set.ModName, set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return nil
	}

	seted, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return nil
	}

	modset.SpecSchemaSync(seted)

	set.Kind = "SpecRoute"

	return nil
}

func ModSetSpecLangList(c fiber.Ctx) error {
	ls := api.LangList{
		Items: api.LangArray,
	}
	ls.Kind = "SpecLangList"
	return web.JSON(c, ls)
}
