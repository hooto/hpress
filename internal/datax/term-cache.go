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
	termCMap   = map[string]*termCates{}
	termCMapMu sync.RWMutex
)

type termCates struct {
	ls  hpapi.TermList
	dps map[uint32][]uint32
}

func termTaxonomyCacheRefresh(modname, table string) {

	if _, ok := termCMap[modname+table]; ok {
		return
	}

	model, err := config.SpecTermModel(modname, table)
	if err != nil {
		return
	}

	txTable := hpapi.TermTableName(modname, table)

	qs := store.Data.NewQueryer().From(txTable).Limit(200).Order("weight desc")
	qs.Where().And("status", 1)

	rs, err := store.Data.Query(qs)
	if err != nil || len(rs) < 1 {
		return
	}

	termCMapMu.Lock()
	defer termCMapMu.Unlock()

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

	tcm := &termCates{
		ls:  ls,
		dps: map[uint32][]uint32{},
	}

	for _, termEntry := range tcm.ls.Items {
		tcm.dps[termEntry.ID] = termCateSubtree(&tcm.ls, []uint32{}, termEntry.ID)
	}

	termCMap[modname+table] = tcm
}

func TermTaxonomyCacheIndexes(modname, table, termIDStr string) []uint32 {

	tid, _ := strconv.ParseUint(termIDStr, 10, 32)

	if _, ok := termCMap[modname+table]; !ok {
		termTaxonomyCacheRefresh(modname, table)
	}

	termCMapMu.RLock()
	defer termCMapMu.RUnlock()

	if t, ok := termCMap[modname+table]; ok {
		if tis, ok := t.dps[uint32(tid)]; ok {
			return tis
		}
	}

	return []uint32{}
}

func TermTaxonomyCacheEntry(modname, table string, termid uint32) *hpapi.Term {

	termCMapMu.RLock()
	defer termCMapMu.RUnlock()

	if t, ok := termCMap[modname+table]; ok {

		for _, entry := range t.ls.Items {

			if entry.ID == termid {
				return &entry
			}
		}
	}

	return nil
}

func TermTaxonomyCacheClean(modname, table string) {

	termCMapMu.Lock()
	defer termCMapMu.Unlock()

	if _, ok := termCMap[modname+table]; ok {
		delete(termCMap, modname+table)
	}
}
