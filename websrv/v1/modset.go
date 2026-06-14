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

	"github.com/hooto/httpsrv"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/hooto/iam/v2/pkg/iamserver"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/modset"
	"github.com/hooto/hpress/store"
)

type ModSet struct {
	*httpsrv.Controller
	us iamserver.UserSession
}

func (c *ModSet) Init() int {

	c.us = iamserver.AppVerifier.Session(c.Request.Request)

	if _, err := c.us.RequireAuth(); err != nil {
		c.Response.Out.WriteHeader(401)
		c.RenderJson(types.NewTypeErrorMeta(iamapi.ErrCodeUnauthorized, "Unauthorized"))
		return 1
	}

	return 0
}

func (c ModSet) SpecListAction() {

	c.AutoRender = false

	rsp := api.SpecList{}

	defer c.RenderJson(&rsp)

	if !c.us.Allow("", "editor.list") {
		rsp.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return
	}

	q := store.Data.NewQueryer().From("hp_modules").Limit(100)
	rs, err := store.Data.Query(q)
	if err != nil {
		rsp.Error = types.NewErrorMeta(api.ErrCodeInternalError, "Can not pull database instance")
		return
	}

	for _, v := range rs {

		var entry api.Spec

		if err := v.Field("body").JsonDecode(&entry); err == nil {
			entry.SrvName, _ = api.SrvNameFilter(v.Field("srvname").String())
			rsp.Items = append(rsp.Items, entry)
		}
	}

	rsp.Kind = "SpecList"
}

func (c ModSet) SpecEntryAction() {

	rsp := api.Spec{}

	defer c.RenderJson(&rsp)

	if !c.us.Allow("", "editor.read") {
		rsp.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return
	}

	if c.Params.Value("name") == "" {
		rsp.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Object Not Found")
		return
	}

	name, err := modset.ModNameFilter(c.Params.Value("name"))
	if err != nil {
		rsp.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return
	}

	q := store.Data.NewQueryer().From("hp_modules").Limit(1)
	q.Where().And("name", name)
	rs, err := store.Data.Query(q)
	if err != nil {
		rsp.Error = types.NewErrorMeta(api.ErrCodeInternalError, "Can not pull database instance")
		return
	}

	if len(rs) < 1 {
		rsp.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Object Not Found")
		return
	}

	if err := rs[0].Field("body").JsonDecode(&rsp); err != nil {
		rsp.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return
	}

	rsp.SrvName, _ = api.SrvNameFilter(rs[0].Field("srvname").String())

	rsp.Kind = "Spec"
}

func (c ModSet) SpecInfoSetAction() {

	var set api.Spec

	defer c.RenderJson(&set)

	if !c.us.Allow("", "editor.write") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return
	}

	err := c.Request.JsonDecode(&set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return
	}

	set.Meta.Name, err = modset.ModNameFilter(set.Meta.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return
	}

	set.SrvName, err = api.SrvNameFilter(set.SrvName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return
	}

	if _, err = modset.SpecFetch(set.Meta.Name); err != nil {

		if err = modset.SpecInfoNew(set); err != nil {
			set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
			return
		}

	} else {

		if err = modset.SpecInfoSet(set); err != nil {
			set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
			return
		}
	}

	seted, err := modset.SpecFetch(set.Meta.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return
	}

	modset.SpecSchemaSync(seted)

	set.Kind = "Spec"
}

func (c ModSet) SpecTermSetAction() {

	var set api.TermModel

	defer c.RenderJson(&set)

	if !c.us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return
	}

	err := c.Request.JsonDecode(&set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return
	}

	set.Meta.Name, err = modset.ModelNameFilter(set.Meta.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return
	}

	set.Type = strings.ToLower(set.Type)
	if set.Type != "tag" && set.Type != "taxonomy" {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Invalid Type")
		return
	}

	_, err = modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "ModName Not Found")
		return
	}

	err = modset.SpecTermSet(set.ModName, set)

	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return
	}

	seted, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return
	}

	modset.SpecSchemaSync(seted)

	set.Kind = "TermModel"
}

func (c ModSet) SpecNodeSetAction() {

	var set api.NodeModel

	defer c.RenderJson(&set)

	if !c.us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return
	}

	err := c.Request.JsonDecode(&set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return
	}

	set.Meta.Name, err = modset.ModelNameFilter(set.Meta.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return
	}

	_, err = modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "ModName Not Found")
		return
	}

	err = modset.SpecNodeSet(set.ModName, &set)

	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return
	}
	seted, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return
	}

	modset.SpecSchemaSync(seted)

	set.Kind = "NodeModel"
}

func (c ModSet) SpecActionSetAction() {

	var set api.Action

	defer c.RenderJson(&set)

	if !c.us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return
	}

	err := c.Request.JsonDecode(&set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return
	}

	set.Name, err = modset.ModelNameFilter(set.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return
	}

	_, err = modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "ModName Not Found")
		return
	}

	err = modset.SpecActionSet(set.ModName, set)

	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return
	}

	seted, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return
	}

	modset.SpecSchemaSync(seted)

	set.Kind = "Action"
}

func (c ModSet) SpecActionDelAction() {

	var set api.Action

	defer c.RenderJson(&set)

	if !c.us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return
	}

	err := c.Request.JsonDecode(&set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return
	}

	set.Name, err = modset.ModelNameFilter(set.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return
	}

	_, err = modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "ModName Not Found")
		return
	}

	err = modset.SpecActionDel(set.ModName, set)

	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return
	}

	seted, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return
	}

	modset.SpecSchemaSync(seted)

	set.Kind = "Action"
}

func (c ModSet) SpecRouteSetAction() {

	var set api.Route

	defer c.RenderJson(&set)

	if !c.us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return
	}

	err := c.Request.JsonDecode(&set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return
	}

	set.Path, err = modset.RoutePathFilter(set.Path)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return
	}

	_, err = modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "ModName Not Found")
		return
	}

	err = modset.SpecRouteSet(set.ModName, set)

	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return
	}

	seted, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return
	}

	modset.SpecSchemaSync(seted)

	set.Kind = "SpecRoute"
}

func (c ModSet) SpecRouteDelAction() {

	var set api.Route

	defer c.RenderJson(&set)

	if !c.us.Allow("", "sys.admin") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return
	}

	err := c.Request.JsonDecode(&set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return
	}

	set.Path, err = modset.RoutePathFilter(set.Path)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return
	}

	set.ModName, err = modset.ModNameFilter(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeBadArgument, err.Error())
		return
	}

	err = modset.SpecRouteDel(set.ModName, set)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return
	}

	seted, err := modset.SpecFetch(set.ModName)
	if err != nil {
		set.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return
	}

	modset.SpecSchemaSync(seted)

	set.Kind = "SpecRoute"
}

func (c ModSet) SpecLangListAction() {
	ls := api.LangList{
		Items: api.LangArray,
	}
	ls.Kind = "SpecLangList"
	c.RenderJson(ls)
}
