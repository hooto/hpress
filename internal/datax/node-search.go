// Copyright 2018 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
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
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/hpapi"
	"github.com/hooto/hpress/internal/store"
)

type NodeSearchEngine interface {
	Query(bucket string, q string, qs *QuerySet) hpapi.NodeList
	Put(bucket string, node hpapi.Node) error
	ModelSet(bucket string, model *hpapi.NodeModel) error
}

var (
	searchInited   = false
	searchLocker   sync.Mutex
	searchIndexNum = 0
	searchCaches   = map[string]*searchModuleCache{}
	nodeSearcher   NodeSearchEngine
)

type searchModuleCache struct {
	termBufs     map[string][]string
	termTaxonomy map[string]hpapi.Term
}

func dataSearchSync() error {

	dataSearchOn := false

	q := store.Data.NewQueryer().From("hp_modules").Limit(100)
	rs := store.Data.Query(q)
	if rs.Err() != nil {
		return nil
	}

	dataModOn := map[string]bool{}
	for rs.Valid() {
		var entry hpapi.Spec
		if err := rs.Field("body").JsonDecode(&entry); err == nil {
			if entry.Status == 1 {
				dataModOn[entry.Meta.Name] = true
			}
		}

		rs.Next()
	}

	for _, mod := range config.Modules {

		if mod.Meta.Name == "core/comment" {
			continue
		}

		if _, ok := dataModOn[mod.Meta.Name]; !ok {
			continue
		}

		for _, model := range mod.NodeModels {
			if model.Extensions.TextSearch {
				dataSearchOn = true
				break
			}
		}
		if dataSearchOn {
			break
		}
	}

	if !dataSearchOn {
		return nil
	}

	searchLocker.Lock()
	if !searchInited {
		if engine, err := NewNodeLynkSearchEngine(config.Prefix); err != nil {
			return err
		} else {
			nodeSearcher = engine
		}
		searchInited = true
	}
	searchLocker.Unlock()

	var (
		limit        int64 = 100
		indexUpdated       = uint32(time.Now().Unix())
	)

	if nodeSearcher == nil {
		return errors.New("server error")
	}

	for _, mod := range config.Modules {

		if mod.Meta.Name == "core/comment" {
			continue
		}

		extTextSearch := false
		for _, model := range mod.NodeModels {
			if model.Extensions.TextSearch {
				extTextSearch = true
				break
			}
		}
		if !extTextSearch {
			continue
		}

		modCache, _ := searchCaches[mod.Meta.Name]
		if modCache == nil {
			modCache = &searchModuleCache{
				termBufs:     map[string][]string{},
				termTaxonomy: map[string]hpapi.Term{},
			}
			searchCaches[mod.Meta.Name] = modCache
		}

		// Fetch Terms
		for _, term := range mod.TermModels {

			switch term.Type {

			case hpapi.TermTaxonomy:

				table := hpapi.TermTableName(mod.Meta.Name, term.Meta.Name)
				qs := store.Data.NewQueryer().From(table).Limit(2000)

				if rs2 := store.Data.Query(qs); rs2.OK() {
					for rs2.Valid() {
						modCache.termTaxonomy[term.Meta.Name+"."+rs2.Field("id").String()] = hpapi.Term{
							ID:    rs2.Field("id").Uint32(),
							Title: rs2.Field("title").String(),
						}
						// fmt.Println(table, rs2.Field("id").Uint32(), rs2.Field("title").String())

						rs2.Next()
					}
				}
			}
		}

		for _, model := range mod.NodeModels {

			if !model.Extensions.TextSearch {
				continue
			}

			var (
				indexStart = time.Now()
				indexNum   = 0
				tblname    = hpapi.NodeTableName(mod.Meta.Name, model.Meta.Name)
				cfgs       types.KvPairs
				offset     = int64(0)
				q          = store.Data.NewQueryer().From(tblname).Limit(limit)
				kvKey      = hpapi.NsSysNodeSearch(tblname)
			)

			nodeSearcher.ModelSet(tblname, model)

			if rs := store.DataLocal.NewReader(kvKey).Exec(); rs.OK() {
				rs.JsonDecode(&cfgs)
				if pv := cfgs.Get("index_updated"); pv.String() != "" {
					q.Where().And("updated.ge", pv.String())
				}
			}

			for {

				rs := store.Data.Query(q)
				if rs.Err() != nil {
					break
				}

				for rs.Valid() {

					id := rs.Field("id").String()

					u64 := hex16ToUint64(id)
					if u64 == 0 {
						break
					}

					item := hpapi.Node{
						ID:      rs.Field("id").String(),
						PID:     rs.Field("pid").String(),
						Status:  rs.Field("status").Int16(),
						UserID:  rs.Field("userid").String(),
						Created: rs.Field("created").Uint32(),
						Updated: rs.Field("updated").Uint32(),
					}

					if model.Extensions.AccessCounter {
						item.ExtAccessCounter = rs.Field("ext_access_counter").Uint32()
					}

					if model.Extensions.CommentEnable {
						if model.Extensions.CommentPerEntry && rs.Field("ext_comment_perentry").Bool() == false {
							item.ExtCommentEnable = false
							item.ExtCommentPerEntry = false
						} else {
							item.ExtCommentEnable = true
							item.ExtCommentPerEntry = true
						}
					}

					if model.Extensions.Permalink != "" && rs.Field("ext_permalink_name").String() != "" {
						item.ExtPermalinkName = rs.Field("ext_permalink_name").String()
						item.SelfLink = fmt.Sprintf("%s", item.ExtPermalinkName)
					} else {
						item.SelfLink = fmt.Sprintf("%s.html", item.ID)
					}

					for _, field := range model.Fields {

						nodeField := hpapi.NodeField{
							Name:  field.Name,
							Value: rs.Field("field_" + field.Name).String(),
						}

						if field.Type == "text" &&
							len(rs.Field("field_"+field.Name+"_attrs").String()) > 10 {

							var attrs types.KvPairs
							if err := rs.Field("field_" + field.Name + "_attrs").JsonDecode(&attrs); err == nil {
								nodeField.Attrs = attrs
							}
						}

						if l := field.Attrs.Get("langs"); len(l) > 3 {

							if len(rs.Field("field_"+field.Name+"_langs").String()) > 5 {
								var nodeLangs hpapi.NodeFieldLangs
								if err := rs.Field("field_" + field.Name + "_langs").JsonDecode(&nodeLangs); err == nil {
									nodeField.Langs = &nodeLangs
								}
							}
						}

						if field.Name == "title" {
							item.Title = nodeField.Value
						}

						item.Fields = append(item.Fields, &nodeField)
					}

					for _, term := range model.Terms {

						switch term.Type {
						case hpapi.TermTaxonomy:

							if ttv, ok := modCache.termTaxonomy[term.Meta.Name+"."+rs.Field("term_"+term.Meta.Name).String()]; ok {
								termItem := hpapi.NodeTerm{
									Name:  term.Meta.Name,
									Value: rs.Field("term_" + term.Meta.Name).String(),
									Type:  term.Type,
								}

								termItem.Items = append(termItem.Items, ttv)

								item.Terms = append(item.Terms, termItem)
							}

						case hpapi.TermTag:

							tags := strings.Split(rs.Field("term_"+term.Meta.Name).String(), ",")

							if len(tags) > 0 {
								termItem := hpapi.NodeTerm{
									Name:  term.Meta.Name,
									Value: rs.Field("term_" + term.Meta.Name).String(),
									Type:  term.Type,
								}

								for _, vtag := range tags {
									termItem.Items = append(termItem.Items, hpapi.Term{
										Title: vtag,
									})

								}

								item.Terms = append(item.Terms, termItem)
							}
						}
					}

					// fmt.Println(id, rs.Field("title").String())

					if err := nodeSearcher.Put(tblname, item); err != nil {
						return err
					}

					indexNum += 1

					rs.Next()
				}

				if n, _ := rs.RowsAffected(); n < limit {
					break
				}

				offset += limit
				q.Offset(offset)
			}

			if indexNum > 0 {
				cfgs.Set("index_updated", indexUpdated)
				if rs := store.DataLocal.NewWriter(kvKey, nil).SetJsonValue(cfgs).Exec(); !rs.OK() {
					slog.Warn("search index error")
				}
				slog.Info(fmt.Sprintf("search data sync %d at %v",
					indexNum, time.Since(indexStart)))
			}
		}
	}

	return nil
}

func (q *QuerySet) NodeListSearch(query string) hpapi.NodeList {

	var rsp hpapi.NodeList

	if !searchInited || nodeSearcher == nil {
		rsp.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Server Not Ready")
		return rsp
	}

	table := hpapi.NodeTableName(q.ModName, q.Table)

	return nodeSearcher.Query(table, query, q)
}

func hex16ToUint64(str string) uint64 {
	if n := len(str); n > 0 {
		if n < 16 {
			str = strings.Repeat("0", 16-n) + str
		}
		if bs, err := hex.DecodeString(str); err == nil && len(bs) >= 8 {
			return binary.BigEndian.Uint64(bs)
		}
	}
	return 0
}
