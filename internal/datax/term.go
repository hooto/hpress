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

package datax

import (
	"crypto/md5"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/hpapi"
	"github.com/hooto/hpress/internal/store"
)

var (
	spaceReg = regexp.MustCompile(" +")
)

func (q *QuerySet) TermCount() (int64, error) {

	table := hpapi.TermTableName(q.ModName, q.Table)

	fr := store.Data.NewFilter()
	fr.And("status", 1)

	return store.Data.Count(table, fr)
}

func (q *QuerySet) TermList() hpapi.TermList {

	rsp := hpapi.TermList{}

	model, err := config.SpecTermModel(q.ModName, q.Table)
	if err != nil {
		rsp.Error = &types.ErrorMeta{
			Code:    hpapi.ErrCodeBadArgument,
			Message: "Term Not Found",
		}
		return rsp
	}

	if model.Type == hpapi.TermTaxonomy {
		if tc, ok := term_cmap[q.ModName+q.Table]; ok {
			return tc.ls
		}
	}

	// q.limit = 100
	table := hpapi.TermTableName(q.ModName, q.Table)

	qs := store.Data.NewQueryer().
		Select(q.cols).
		From(table).
		Offset(q.offset)

	if model.Type == hpapi.TermTag {
		qs.Order("updated desc")
	} else if model.Type == hpapi.TermTaxonomy {
		q.limit = 200
		qs.Order("weight desc")
	}

	qs.Limit(q.limit)

	qs.SetFilter(q.filter)

	rs, err := store.Data.Query(qs)
	if err != nil {
		rsp.Error = &types.ErrorMeta{
			Code:    hpapi.ErrCodeInternalError,
			Message: "Can not pull database instance",
		}
		return rsp
	}

	if len(rs) > 0 {

		for _, v := range rs {

			item := hpapi.Term{
				ID:      v.Field("id").Uint32(),
				PID:     v.Field("pid").Uint32(),
				Status:  v.Field("status").Int16(),
				UserID:  v.Field("userid").String(),
				Title:   v.Field("title").String(),
				Created: v.Field("created").Uint32(),
				Updated: v.Field("updated").Uint32(),
			}

			switch model.Type {
			case hpapi.TermTag:
				item.UID = v.Field("uid").String()
			case hpapi.TermTaxonomy:
				item.PID = v.Field("pid").Uint32()
				item.Weight = v.Field("weight").Int32()
			}

			rsp.Items = append(rsp.Items, item)
		}
	}

	rsp.Model = model

	if q.Pager {
		num, _ := store.Data.Count(table, q.filter)
		rsp.Meta.TotalResults = uint64(num)
		rsp.Meta.StartIndex = uint64(q.offset)
		rsp.Meta.ItemsPerList = uint64(q.limit)
	}

	rsp.Kind = "TermList"

	if model.Type == hpapi.TermTaxonomy {

		tcm := &term_cates{
			ls:  rsp,
			dps: map[uint32][]uint32{},
		}

		for _, term_entry := range tcm.ls.Items {
			tcm.dps[term_entry.ID] = _term_cate_subtree(&tcm.ls, []uint32{}, term_entry.ID)
		}

		term_cmap_mu.Lock()
		term_cmap[q.ModName+q.Table] = tcm
		term_cmap_mu.Unlock()
	}

	// qryhash := q.Hash()

	// if model.CacheTTL > 0 && entry.Title != "" {
	// 	store.CacheSetJson(qryhash, rsp, model.CacheTTL)
	// }

	return rsp
}

func _term_cate_subtree(termls *hpapi.TermList, prs []uint32, pid uint32) []uint32 {

	if _term_in_array(prs, pid) {
		return prs
	}

	prs = append(prs, pid)

	for _, entry := range termls.Items {

		if entry.PID == pid {
			prs = _term_cate_subtree(termls, prs, entry.ID)
		}
	}

	return prs
}

func _term_in_array(arr []uint32, a uint32) bool {

	for _, ar := range arr {
		if ar == a {
			return true
		}
	}

	return false
}

func (q *QuerySet) TermEntry() hpapi.Term {

	var (
		rsp = hpapi.Term{}
		err error
	)

	rsp.Model, err = config.SpecTermModel(q.ModName, q.Table)
	if err != nil {
		rsp.Error = &types.ErrorMeta{
			Code:    hpapi.ErrCodeBadArgument,
			Message: "Term Not Found",
		}
		return rsp
	}

	table := hpapi.TermTableName(q.ModName, q.Table)

	qs := store.Data.NewQueryer().
		Select(q.cols).
		From(table).
		Order(q.order).
		Limit(1).
		Offset(q.offset)

	qs.SetFilter(q.filter)

	rs, err := store.Data.Query(qs)
	if err != nil {
		rsp.Error = &types.ErrorMeta{
			Code:    hpapi.ErrCodeInternalError,
			Message: "Can not pull database instance",
		}
		return rsp
	}

	if len(rs) < 1 {
		rsp.Error = &types.ErrorMeta{
			Code:    hpapi.ErrCodeBadArgument,
			Message: "Term Not Found",
		}
		return rsp
	}

	switch rsp.Model.Type {
	case hpapi.TermTaxonomy:
		rsp.PID = rs[0].Field("pid").Uint32()
		rsp.Weight = rs[0].Field("weight").Int32()
	case hpapi.TermTag:
		rsp.UID = rs[0].Field("uid").String()
	default:
		rsp.Error = &types.ErrorMeta{
			Code:    hpapi.ErrCodeInternalError,
			Message: "Server Error",
		}
		return rsp
	}

	rsp.ID = rs[0].Field("id").Uint32()
	rsp.PID = rs[0].Field("pid").Uint32()
	rsp.Status = rs[0].Field("status").Int16()
	rsp.UserID = rs[0].Field("userid").String()
	rsp.Title = rs[0].Field("title").String()
	rsp.Created = rs[0].Field("created").Uint32()
	rsp.Updated = rs[0].Field("updated").Uint32()

	rsp.Kind = "Term"

	// qryhash := q.Hash()

	// if rsp.Model.CacheTTL > 0 && entry.Title != "" {
	// 	store.CacheSetJson(qryhash, rsp, rsp.Model.CacheTTL)
	// }

	return rsp
}

type TermList hpapi.TermList

func (t *TermList) Index() string {

	if len(t.Items) < 1 {
		return ""
	}

	idxs := []string{}
	for _, v := range t.Items {
		idxs = append(idxs, fmt.Sprintf("%v", v.ID))
	}

	return strings.Join(idxs, ",")
}

func (t *TermList) Content() string {

	if len(t.Items) < 1 {
		return ""
	}

	ts := []string{}
	for _, v := range t.Items {
		ts = append(ts, v.Title)
	}

	return strings.Join(ts, ",")
}

func NodeTermQuery(modname string, model *hpapi.NodeModel, terms []hpapi.NodeTerm) []hpapi.NodeTerm {

	for _, modTerm := range model.Terms {

		for k, term := range terms {

			if modTerm.Meta.Name != term.Name {
				continue
			}

			switch modTerm.Type {

			case hpapi.TermTag:

				tags := strings.Split(term.Value, ",")

				for _, vtag := range tags {

					terms[k].Items = append(terms[k].Items, hpapi.Term{
						Title: vtag,
					})
				}

			case hpapi.TermTaxonomy:

				table := hpapi.TermTableName(modname, modTerm.Meta.Name)

				q := store.Data.NewQueryer().From(table)
				q.Limit(1)
				q.Where().And("id", term.Value)

				if rs, err := store.Data.Query(q); err == nil && len(rs) > 0 {

					terms[k].Items = append(terms[k].Items, hpapi.Term{
						ID:    rs[0].Field("id").Uint32(),
						Title: rs[0].Field("title").String(),
					})
				}
			}

			// terms[k].Type = modTerm.Type

			break
		}
	}

	return terms
}

func TermSync(modname, modelid, terms string) (TermList, error) {

	ls := TermList{}

	terms = spaceReg.ReplaceAllString(terms, " ")

	tars := strings.Split(terms, ",")

	ids := []interface{}{}

	for _, term := range tars {

		tag := hpapi.Term{
			Title: strings.TrimSpace(term),
		}

		if len(tag.Title) < 1 {
			continue
		}

		h := md5.New()
		io.WriteString(h, strings.ToLower(tag.Title))
		tag.UID = fmt.Sprintf("%x", h.Sum(nil))[:16]

		exist := false
		for _, prev := range ids {
			if prev.(string) == tag.UID {
				exist = true
				break
			}
		}
		if exist {
			continue
		}

		ls.Items = append(ls.Items, tag)

		ids = append(ids, tag.UID)
	}

	table := hpapi.TermTableName(modname, modelid)

	if len(ids) > 0 {

		q := store.Data.NewQueryer().From(table).Limit(int64(len(ids)))
		q.Where().And("uid.in", ids...)

		if rs, err := store.Data.Query(q); err == nil {

			for _, v := range rs {

				for tk, tv := range ls.Items {

					if v.Field("uid").String() == tv.UID {

						ls.Items[tk].ID = v.Field("id").Uint32()
						break
					}
				}
			}
		}
	}

	timenow := uint32(time.Now().Unix())

	for tk, tv := range ls.Items {

		if tv.ID > 0 {
			continue
		}

		if _, err := store.Data.Insert(table, map[string]interface{}{
			"uid":     tv.UID,
			"title":   tv.Title,
			"userid":  "sysadmin",
			"status":  1,
			"created": timenow,
			"updated": timenow,
		}); err == nil {

			// LastInsertId() is unsupported on pgsqlgo, so re-read the row by
			// its deterministic uid to get the server-assigned id.
			q := store.Data.NewQueryer().From(table).Limit(1)
			q.Where().And("uid", tv.UID)
			if rs2, err := store.Data.Query(q); err == nil && len(rs2) > 0 {
				ls.Items[tk].ID = rs2[0].Field("id").Uint32()
			}
		}
	}

	return ls, nil
}
