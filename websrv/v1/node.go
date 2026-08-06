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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/lessos/lessgo/crypto/idhash"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/lessos/lessgo/utilx"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/config"
	"github.com/hooto/hpress/datax"
	"github.com/hooto/hpress/store"
	"github.com/hooto/hpress/websrv/web"
)

var (
	node_id_length         = 12
	node_pid_default       = "00"
	node_list_limit  int64 = 15
	node_set_lock    sync.Mutex
)

func NodeList(c fiber.Ctx) error {

	ls := api.NodeList{}

	defer func() { _ = web.JSON(c, &ls) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "editor.list") {
		ls.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	model, err := config.SpecNodeModel(web.Param(c, "modname"), web.Param(c, "modelid"))
	if err != nil {
		ls.Error = types.NewErrorMeta("400", "Invalid modname or modelid")
		return nil
	}

	dq := datax.NewQuery(web.Param(c, "modname"), web.Param(c, "modelid"))
	dq.Limit(node_list_limit)
	dq.Filter("status.gt", 0)

	page := web.ParamInt(c, "page")
	if page < 1 {
		page = 1
	}

	if page > 1 {
		dq.Offset(int64((page - 1) * node_list_limit))
	}

	dqc := datax.NewQuery(web.Param(c, "modname"), web.Param(c, "modelid"))
	dqc.Filter("status.gt", 0)

	node_refer := web.Param(c, "ext_node_refer")
	if model.Extensions.NodeRefer != "" &&
		api.NodeExtNodeReferReg.MatchString(web.Param(c, "ext_node_refer")) {
		dq.Filter("ext_node_refer", node_refer)
		dqc.Filter("ext_node_refer", node_refer)
	}

	if web.Param(c, "qry_text") != "" {
		dq.Filter("field_title.like", "%"+web.Param(c, "qry_text")+"%")
		dqc.Filter("field_title.like", "%"+web.Param(c, "qry_text")+"%")
	}

	var (
		fields = strings.Split(web.Param(c, "fields"), ",")
		terms  = strings.Split(web.Param(c, "terms"), ",")
	)

	count, err := dqc.NodeCount()
	if err != nil {
		ls.Error = &types.ErrorMeta{api.ErrCodeInternalError, err.Error()}
		return nil
	}

	ls = dq.NodeList(fields, terms)

	ls.Meta.TotalResults = uint64(count)
	ls.Meta.StartIndex = uint64((page - 1) * node_list_limit)
	ls.Meta.ItemsPerList = uint64(node_list_limit)

	return nil
}

func NodeEntry(c fiber.Ctx) error {

	rsp := api.Node{}

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "editor.read") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	dq := datax.NewQuery(web.Param(c, "modname"), web.Param(c, "modelid"))
	dq.Limit(100)
	dq.Filter("status.gt", 0)

	dq.Filter("id", web.Param(c, "id"))

	rsp = dq.NodeEntry()

	return nil
}

func NodeSet(c fiber.Ctx) error {

	rsp := api.Node{}
	defer func() { _ = web.JSON(c, &rsp) }()

	if err := web.Bind(c, &rsp); err != nil {
		rsp.Error = &types.ErrorMeta{
			Code:    "400",
			Message: "Bad Request: " + err.Error(),
		}
		return nil
	}

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "editor.write") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	model, err := config.SpecNodeModel(web.Param(c, "modname"), web.Param(c, "modelid"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{
			Code:    "404",
			Message: "Spec or Model Not Found",
		}
		return nil
	}

	node_set_lock.Lock()
	defer node_set_lock.Unlock()

	var (
		set        = map[string]interface{}{}
		table      = api.NodeTableName(web.Param(c, "modname"), web.Param(c, "modelid"))
		node_refer = ""
	)

	//
	if model.Extensions.Permalink != "" && rsp.ExtPermalinkName != "" {
		rsp.ExtPermalinkName, err = api.PermalinkNameFilter(rsp.ExtPermalinkName)
		if err != nil || rsp.ExtPermalinkName == "" {
			rsp.Error = types.NewErrorMeta("400", "Invalid Permalink Name")
			return nil
		}
	}

	if model.Extensions.NodeRefer != "" {
		if !api.NodeExtNodeReferReg.MatchString(rsp.ExtNodeRefer) {
			rsp.Error = types.NewErrorMeta("400", "Invalid Node Refer ID")
			return nil
		}
		node_refer = rsp.ExtNodeRefer
	}

	if ft := rsp.Field("title"); ft == nil {
		rsp.Error = types.NewErrorMeta("400", "Title Not Found")
		return nil
	} else if ft.Value = strings.TrimSpace(ft.Value); ft.Value == "" {
		rsp.Error = types.NewErrorMeta("400", "Title can not be empty")
		return nil
	}

	if len(rsp.ID) > 0 {

		q := store.Data.NewQueryer().From(table).Limit(1)
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
				Message: "Node Not Found",
			}
			return nil
		}

		/*
			if rs[0].Field("title").String() != rsp.Title {
				set["title"] = rsp.Title
			}
		*/

		if rs[0].Field("status").Int16() != rsp.Status {
			set["status"] = rsp.Status
		}

		if model.Extensions.Permalink != "" {
			set["ext_permalink_name"] = rs[0].Field("ext_permalink_name").String()
		}

		if model.Extensions.NodeRefer != "" {
			set["ext_node_refer"] = rs[0].Field("ext_node_refer").String()
		}

		//
		for _, valField := range rsp.Fields {

			for _, modField := range model.Fields {

				if modField.Name != valField.Name {
					continue
				}

				if rs[0].Field("field_"+modField.Name).String() != valField.Value {
					set["field_"+modField.Name] = valField.Value
				}

				// upgrade
				if modField.Name == "title" {
					set["title"] = valField.Value
				}

				if modField.Type == "text" {
					attrs := types.KvPairs{}

					for _, attr := range valField.Attrs {
						if modField.Type == "text" && attr.Key == "format" &&
							utilx.ArrayContain(attr.Value, []string{"md", "text", "html", "shtml"}) {
							attrs.Set(attr.Key, attr.Value)
						}
					}

					if len(attrs) > 0 {
						attrs_js, _ := json.Encode(attrs, "  ")
						if string(attrs_js) != rs[0].Field("field_"+modField.Name+"_attrs").String() {
							set["field_"+modField.Name+"_attrs"] = string(attrs_js)
						}
					}
				}

				if modField.Type == "text" || modField.Type == "string" {
					// langs
					if attr := modField.Attrs.Get("langs"); attr != nil && valField.Langs != nil {

						var langs api.NodeFieldLangs
						if len(rs[0].Field("field_"+modField.Name+"_langs").String()) > 5 {
							rs[0].Field("field_" + modField.Name + "_langs").JsonDecode(&langs)
						}

						attr_langs := api.LangsStringFilterArray(attr.String())
						for li := 1; li < len(attr_langs); li++ {
							if lang_entry := valField.Langs.Items.Get(attr_langs[li]); lang_entry != nil {
								langs.Items.Set(attr_langs[li], lang_entry.String())
							}
						}

						if len(langs.Items) > 0 {
							langs_js, _ := json.Encode(langs, "")
							set["field_"+modField.Name+"_langs"] = string(langs_js)
						}
					}
				}

				break
			}
		}

		//
		for _, modTerm := range model.Terms {

			for _, term := range rsp.Terms {

				if modTerm.Meta.Name != term.Name {
					continue
				}

				switch modTerm.Type {

				case api.TermTag:

					tags, _ := datax.TermSync(web.Param(c, "modname"), modTerm.Meta.Name, term.Value)

					if rs[0].Field("term_"+term.Name).String() != term.Value {
						set["term_"+modTerm.Meta.Name] = tags.Content()
						set["term_"+modTerm.Meta.Name+"_idx"] = tags.Index()
					}

				case api.TermTaxonomy:

					set["term_"+modTerm.Meta.Name] = term.Value
				}
			}
		}

	} else {

		set["id"] = idhash.RandHexString(node_id_length)
		// set["title"] = rsp.Title
		set["status"] = rsp.Status
		set["created"] = uint32(time.Now().Unix())

		// TODO
		if p, _ := us.Profile(); p != nil {
			set["userid"] = p.Username
		} else {
			set["userid"] = ""
		}

		set["pid"] = node_pid_default
		if model.Extensions.AccessCounter {
			set["ext_access_counter"] = "0"
		}

		//
		for _, modField := range model.Fields {

			for _, valField := range rsp.Fields {

				if modField.Name != valField.Name {
					continue
				}

				set["field_"+valField.Name] = valField.Value

				// upgrade
				if modField.Name == "title" {
					set["title"] = valField.Value
				}

				if modField.Type == "text" {

					attrs := types.KvPairs{}

					for _, attr := range valField.Attrs {
						if modField.Type == "text" && attr.Key == "format" &&
							utilx.ArrayContain(attr.Value, []string{"md", "text", "html", "shtml"}) {
							attrs.Set(attr.Key, attr.Value)
						}
					}

					if len(attrs) > 0 {
						attrs_js, _ := json.Encode(attrs, "  ")
						set["field_"+modField.Name+"_attrs"] = string(attrs_js)
					}
				}

				if modField.Type == "text" || modField.Type == "string" {
					// langs
					if attr := modField.Attrs.Get("langs"); attr != nil && valField.Langs != nil {

						var langs api.NodeFieldLangs

						attr_langs := api.LangsStringFilterArray(attr.String())
						for li := 1; li < len(attr_langs); li++ {
							if lang_entry := valField.Langs.Items.Get(attr_langs[li]); lang_entry != nil {
								langs.Items.Set(attr_langs[li], lang_entry.String())
							}
						}

						if len(langs.Items) > 0 {
							langs_js, _ := json.Encode(langs, "")
							set["field_"+modField.Name+"_langs"] = string(langs_js)
						}
					}
				}

				break
			}

			if _, ok := set["field_"+modField.Name]; !ok {

				switch modField.Type {

				case "bool":
					set["field_"+modField.Name] = false

				case "string":
					set["field_"+modField.Name] = ""

				case "text":
					set["field_"+modField.Name] = ""
					set["field_"+modField.Name+"_attrs"] = "[]"

				case "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64":
					set["field_"+modField.Name] = "0"

				case "float", "decimal":
					set["field_"+modField.Name] = "0"

				default:
					set["field_"+modField.Name] = ""
				}
			}
		}

		//
		for _, modTerm := range model.Terms {

			for _, term := range rsp.Terms {

				if modTerm.Meta.Name != term.Name {
					continue
				}

				switch modTerm.Type {

				case api.TermTag:

					tags, _ := datax.TermSync(web.Param(c, "modname"), modTerm.Meta.Name, term.Value)
					set["term_"+modTerm.Meta.Name] = tags.Content()
					set["term_"+modTerm.Meta.Name+"_idx"] = tags.Index()

				case api.TermTaxonomy:

					set["term_"+modTerm.Meta.Name] = term.Value
				}

				break
			}

			if _, ok := set["term_"+modTerm.Meta.Name]; !ok {

				switch modTerm.Type {

				case api.TermTag:
					set["term_"+modTerm.Meta.Name+"_idx"] = ""
					set["term_"+modTerm.Meta.Name] = ""

				case api.TermTaxonomy:
					set["term_"+modTerm.Meta.Name] = ""
				}
			}
		}
	}

	if model.Extensions.Permalink != "" {

		if prev, ok := set["ext_permalink_name"]; !ok || prev.(string) != rsp.ExtPermalinkName {

			if rsp.ExtPermalinkName == "" {
				if len(rsp.ID) > 0 {
					set["ext_permalink_idx"] = rsp.ID
				} else {
					set["ext_permalink_idx"], _ = set["id"]
				}
				set["ext_permalink_name"] = ""
			} else {

				permaname := rsp.ExtPermalinkName

				for i := 0; i < 10; i++ {

					if i > 0 {
						permaname = fmt.Sprintf("%s-%d", rsp.ExtPermalinkName, i)
					}

					permaidx := idhash.HashToHexString([]byte(node_refer+permaname), 12)

					q := store.Data.NewQueryer().From(table).Limit(1)
					q.Where().And("ext_permalink_idx", permaidx)
					q.Where().And("status", 1)

					if len(rsp.ID) > 0 {
						q.Where().And("id.ne", rsp.ID)
					}

					if rs, err := store.Data.Query(q); err == nil && len(rs) < 1 {

						set["ext_permalink_name"] = permaname
						set["ext_permalink_idx"] = permaidx
						break
					}
				}

				if _, ok := set["ext_permalink_idx"]; !ok {

					rsp.Error = &types.ErrorMeta{
						Code:    "400",
						Message: "Permalink Name Conflict",
					}
					return nil
				}

			}
		}
	}

	if model.Extensions.CommentPerEntry {
		if model.Extensions.CommentEnable && !rsp.ExtCommentPerEntry {
			set["ext_comment_perentry"] = 0
		} else {
			set["ext_comment_perentry"] = 1
		}
	}

	if model.Extensions.NodeRefer != "" {

		if prev, ok := set["ext_node_refer"]; !ok || prev != rsp.ExtNodeRefer {
			ref_q := store.Data.NewQueryer().From(api.NodeTableName(web.Param(c, "modname"), model.Extensions.NodeRefer)).Limit(1)
			ref_q.Where().And("id", rsp.ExtNodeRefer)
			if rs, err := store.Data.Query(ref_q); err != nil {
				rsp.Error = types.NewErrorMeta("500", "Server Error")
				return nil
			} else if len(rs) < 1 {
				rsp.Error = types.NewErrorMeta("400", "Node Refer ID Not Found")
				return nil
			}
			set["ext_node_refer"] = rsp.ExtNodeRefer
		}
	}

	if len(set) > 0 {

		set["updated"] = uint32(time.Now().Unix())

		if len(rsp.ID) > 0 {

			ft := store.Data.NewFilter()
			ft.And("id", rsp.ID)
			_, err = store.Data.Update(table, set, ft)

		} else {
			rsp.ID = set["id"].(string)
			_, err = store.Data.Insert(table, set)
		}

		// clean frontend cache
		qry := datax.NewQuery(web.Param(c, "modname"), model.Meta.Name)
		qry.Filter("status", 1)
		qry.Filter("id", rsp.ID)

		store.DataLocal.NewDeleter([]byte(qry.Hash())).Exec()

		if err != nil {
			rsp.Error = &types.ErrorMeta{
				Code:    "500",
				Message: err.Error(),
			}
			return nil
		}
	}

	rsp.Kind = "Node"

	return nil
}

func NodeDel(c fiber.Ctx) error {

	rsp := api.Node{}
	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "editor.write") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	if _, err := config.SpecNodeModel(web.Param(c, "modname"), web.Param(c, "modelid")); err != nil {
		rsp.Error = &types.ErrorMeta{
			Code:    "404",
			Message: "Spec or Model Not Found",
		}
		return nil
	}

	//
	set := map[string]interface{}{
		"updated": uint32(time.Now().Unix()),
		"status":  0,
	}

	//
	table := api.NodeTableName(web.Param(c, "modname"), web.Param(c, "modelid"))

	//
	ids := strings.Split(web.Param(c, "id"), ",")

	for _, id := range ids {

		q := store.Data.NewQueryer().From(table).Limit(1)
		q.Where().And("id", id)

		if rs, err := store.Data.Query(q); err != nil {
			rsp.Error = &types.ErrorMeta{
				Code:    "500",
				Message: "Can not pull database instance",
			}
			return nil
		} else if len(rs) < 1 {
			rsp.Error = &types.ErrorMeta{
				Code:    "404",
				Message: "Node Not Found",
			}
			return nil
		}

		ft := store.Data.NewFilter()
		ft.And("id", id)

		if _, err := store.Data.Update(table, set, ft); err != nil {
			rsp.Error = &types.ErrorMeta{
				Code:    "500",
				Message: fmt.Sprintf("id:%s err:%s", id, err.Error()),
			}
			return nil
		}
	}

	rsp.Kind = "Node"

	return nil
}
