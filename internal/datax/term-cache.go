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
	"strconv"
	"sync"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/hpapi"
	"github.com/hooto/hpress/internal/store"
)

var (
	term_cmap    = map[string]*term_cates{}
	term_cmap_mu sync.RWMutex
)

type term_cates struct {
	ls  hpapi.TermList
	dps map[uint32][]uint32
}

func _termTaxonomyCacheRefresh(modname, table string) {

	if _, ok := term_cmap[modname+table]; ok {
		return
	}

	model, err := config.SpecTermModel(modname, table)
	if err != nil {
		return
	}

	tx_table := hpapi.TermTableName(modname, table)

	qs := store.Data.NewQueryer().From(tx_table).Limit(200).Order("weight desc")
	qs.Where().And("status", 1)

	rs, err := store.Data.Query(qs)
	if err != nil || len(rs) < 1 {
		return
	}

	term_cmap_mu.Lock()
	defer term_cmap_mu.Unlock()

	ls := hpapi.TermList{}

	for _, v := range rs {

		ls.Items = append(ls.Items, hpapi.Term{
			ID:      v.Field("id").Uint32(),
			PID:     v.Field("pid").Uint32(),
			Status:  v.Field("status").Int16(),
			UserID:  v.Field("userid").String(),
			Title:   v.Field("title").String(),
			Weight:  v.Field("weight").Int32(),
			Created: v.Field("created").Uint32(),
			Updated: v.Field("updated").Uint32(),
		})
	}

	ls.Model = model
	ls.Meta.TotalResults = uint64(len(ls.Items))
	ls.Meta.StartIndex = 0
	ls.Meta.ItemsPerList = 200

	tcm := &term_cates{
		ls:  ls,
		dps: map[uint32][]uint32{},
	}

	for _, term_entry := range tcm.ls.Items {
		tcm.dps[term_entry.ID] = _term_cate_subtree(&tcm.ls, []uint32{}, term_entry.ID)
	}

	term_cmap[modname+table] = tcm
}

func TermTaxonomyCacheIndexes(modname, table, termid_s string) []uint32 {

	tid, _ := strconv.ParseUint(termid_s, 10, 32)

	if _, ok := term_cmap[modname+table]; !ok {
		_termTaxonomyCacheRefresh(modname, table)
	}

	term_cmap_mu.RLock()
	defer term_cmap_mu.RUnlock()

	if t, ok := term_cmap[modname+table]; ok {
		if tis, ok := t.dps[uint32(tid)]; ok {
			return tis
		}
	}

	return []uint32{}
}

func TermTaxonomyCacheEntry(modname, table string, termid uint32) *hpapi.Term {

	term_cmap_mu.RLock()
	defer term_cmap_mu.RUnlock()

	if t, ok := term_cmap[modname+table]; ok {

		for _, entry := range t.ls.Items {

			if entry.ID == termid {
				return &entry
			}
		}
	}

	return nil
}

func TermTaxonomyCacheClean(modname, table string) {

	term_cmap_mu.Lock()
	defer term_cmap_mu.Unlock()

	if _, ok := term_cmap[modname+table]; ok {
		delete(term_cmap, modname+table)
	}
}
