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
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/datax"
	"github.com/hooto/hpress/internal/hpapi"
	"github.com/hooto/hpress/internal/store"
	"github.com/hooto/hpress/websrv/web"
)

var (
	spaceReg                       = regexp.MustCompile(" +")
	term_list_limit          int64 = 15
	term_list_limit_taxonomy int64 = 200
)

func TermList(c fiber.Ctx) error {

	var ls hpapi.TermList

	defer func() { _ = web.JSON(c, &ls) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "editor.list") {
		ls.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	model, err := config.SpecTermModel(web.Param(c, "modname"), web.Param(c, "modelid"))
	if err != nil {
		ls.Error = &types.ErrorMeta{
			Code:    "404",
			Message: "Spec or Model Not Found",
		}
		return nil
	}

	page, limit := web.ParamInt(c, "page"), term_list_limit

	dq := datax.NewQuery(web.Param(c, "modname"), web.Param(c, "modelid"))
	if model.Type == hpapi.TermTaxonomy {
		limit = term_list_limit_taxonomy
		page = 1
	}

	if page < 1 {
		page = 1
	}

	dq.Limit(limit)
	if page > 1 {
		dq.Offset(int64((page - 1) * limit))
	}

	//
	ls = dq.TermList()

	dqc := datax.NewQuery(web.Param(c, "modname"), web.Param(c, "modelid"))

	if web.Param(c, "qry_text") != "" {
		dqc.Filter("title.like", "%"+web.Param(c, "qry_text")+"%")
	}

	count, err := dqc.TermCount()
	if err != nil {
		ls.Error = &types.ErrorMeta{hpapi.ErrCodeInternalError, err.Error()}
		return nil
	}

	ls.Kind = "TermList"
	ls.Meta.TotalResults = uint64(count)
	ls.Meta.StartIndex = uint64((page - 1) * limit)
	ls.Meta.ItemsPerList = uint64(limit)

	return nil
}

func TermEntry(c fiber.Ctx) error {

	rsp := hpapi.Term{
		TypeMeta: types.TypeMeta{
			APIVersion: hpapi.Version,
		},
	}

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "editor.read") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	dq := datax.NewQuery(web.Param(c, "modname"), web.Param(c, "modelid"))
	dq.Limit(100)

	dq.Filter("id", web.Param(c, "id"))

	rsp = dq.TermEntry()

	return nil
}

func TermSet(c fiber.Ctx) error {

	rsp := hpapi.Term{}

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "editor.write") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	model, err := config.SpecTermModel(web.Param(c, "modname"), web.Param(c, "modelid"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{
			Code:    "404",
			Message: "Spec or Model Not Found",
		}
		return nil
	}

	if err := web.Bind(c, &rsp); err != nil {
		rsp.Error = &types.ErrorMeta{
			Code:    "400",
			Message: "Bad Request " + err.Error(),
		}
		return nil
	}

	var (
		set      = map[string]interface{}{}
		username = ""
		table    = hpapi.TermTableName(web.Param(c, "modname"), web.Param(c, "modelid"))
	)

	if s, err := us.Profile(); err == nil {
		username = s.Username
	}

	q := store.Data.NewQueryer().From(table).Limit(1)

	switch model.Type {

	case hpapi.TermTag:

		uniTitle := spaceReg.ReplaceAllString(strings.TrimSpace(strings.ToLower(rsp.Title)), " ")

		h := md5.New()
		io.WriteString(h, uniTitle)
		rsp.UID = fmt.Sprintf("%x", h.Sum(nil))[:16]
		rsp.ID = 0

		q.Where().And("uid", rsp.UID)

		rs, err := store.Data.Query(q)
		if err != nil {
			rsp.Error = &types.ErrorMeta{
				Code:    "500",
				Message: "Can not pull database instance",
			}
			return nil
		}

		if len(rs) == 1 {

			rsp.ID = rs[0].Field("id").Uint32()

			if rs[0].Field("title").String() != rsp.Title {
				set["title"] = rsp.Title
			}

			if rs[0].Field("status").Int16() != rsp.Status {
				set["status"] = rsp.Status
			}

		} else {

			set["uid"] = rsp.UID
			set["title"] = rsp.Title
			set["status"] = rsp.Status
			set["created"] = uint32(time.Now().Unix())
			set["userid"] = username
		}

	case hpapi.TermTaxonomy:

		if rsp.ID > 0 {

			q.Where().And("id", rsp.ID)

			rs, err := store.Data.Query(q)
			if err != nil {
				rsp.Error = &types.ErrorMeta{
					Code:    "500",
					Message: "Can not pull database instance",
				}
				return nil
			}

			if len(rs) < 1 {
				rsp.Error = &types.ErrorMeta{
					Code:    "404",
					Message: "Term Not Found",
				}
				return nil
			}

			if rs[0].Field("title").String() != rsp.Title {
				set["title"] = rsp.Title
			}

			if rs[0].Field("status").Int16() != rsp.Status {
				set["status"] = rsp.Status
			}

			if rs[0].Field("pid").Uint32() != rsp.PID {
				set["pid"] = rsp.PID
			}

			if rs[0].Field("weight").Int32() != rsp.Weight {
				set["weight"] = rsp.Weight
			}

			if rs[0].Field("userid").String() == "" {
				set["userid"] = username
			}

		} else {

			set["pid"] = rsp.PID
			set["title"] = rsp.Title
			set["status"] = rsp.Status
			set["weight"] = rsp.Weight
			set["created"] = uint32(time.Now().Unix())
			set["userid"] = username
		}

		datax.TermTaxonomyCacheClean(web.Param(c, "modname"), web.Param(c, "modelid"))

	default:
		rsp.Error = &types.ErrorMeta{
			Code:    "500",
			Message: "Server Error",
		}
		return nil
	}

	if len(set) > 0 {

		set["updated"] = uint32(time.Now().Unix())

		if rsp.ID > 0 {

			ft := store.Data.NewFilter()
			ft.And("id", rsp.ID)
			_, err = store.Data.Update(table, set, ft)

		} else {

			// fmt.Println("SET", table, "___", set)
			rs, err := store.Data.Insert(table, set)
			if err == nil {
				if incrid, err := rs.LastInsertId(); err == nil && incrid > 0 {
					rsp.ID = uint32(incrid)
				} else {
					err = errors.New("Can Not Get LastInsertId")
				}
			}
		}

		if err != nil {
			rsp.Error = &types.ErrorMeta{
				Code:    "500",
				Message: err.Error(),
			}
			return nil
		}
	}

	rsp.Model = model

	rsp.Kind = "Term"

	return nil
}
