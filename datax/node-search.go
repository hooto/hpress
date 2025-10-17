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
	"strings"
	"sync"
	"time"

	"github.com/hooto/hlog4g/hlog"
	"github.com/lessos/lessgo/crypto/idhash"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/config"
	"github.com/hooto/hpress/store"
)

type NodeSearchEngine interface {
	Query(bucket string, q string, qs *QuerySet) api.NodeList
	Put(bucket string, node api.Node) error
	ModelSet(bucket string, model *api.NodeModel) error
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
	termTaxonomy map[string]api.Term
}

func data_search_node_label(name string) string {
	return fmt.Sprintf("hpnode_%s", name)
}

func data_search_sync() error {

	dataSearchOn := false

	q := store.Data.NewQueryer().From("hp_modules").Limit(100)
	rs, err := store.Data.Query(q)
	if err != nil {
		return nil
	}

	dataModOn := map[string]bool{}
	for _, v := range rs {
		var entry api.Spec
		if err := v.Field("body").JsonDecode(&entry); err == nil {
			if entry.Status == 1 {
				dataModOn[entry.Meta.Name] = true
			}
		}
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
		// if engine, err := NewNodeSphinxSearchEngine(config.Prefix); err != nil {
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

		modid := idhash.HashToHexString([]byte(mod.Meta.Name), 12)

		modCache, _ := searchCaches[mod.Meta.Name]
		if modCache == nil {
			modCache = &searchModuleCache{
				termBufs:     map[string][]string{},
				termTaxonomy: map[string]api.Term{},
			}
			searchCaches[mod.Meta.Name] = modCache
		}

		// Fetch Terms
		for _, term := range mod.TermModels {

			switch term.Type {

			case api.TermTaxonomy:

				table := fmt.Sprintf("hpt_%s_%s", modid, term.Meta.Name)
				qs := store.Data.NewQueryer().From(table).Limit(2000)

				if rs, err := store.Data.Query(qs); err == nil && len(rs) > 0 {
					for _, v := range rs {
						modCache.termTaxonomy[term.Meta.Name+"."+v.Field("id").String()] = api.Term{
							ID:    v.Field("id").Uint32(),
							Title: v.Field("title").String(),
						}
						// fmt.Println(table, v.Field("id").Uint32(), v.Field("title").String())
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
				tblname    = fmt.Sprintf("hpn_%s_%s", modid, model.Meta.Name)
				cfgs       types.KvPairs
				offset     = int64(0)
				q          = store.Data.NewQueryer().From(tblname).Limit(limit)
				kvKey      = api.NsSysNodeSearch(tblname)
			)

			nodeSearcher.ModelSet(tblname, model)

			if rs := store.DataLocal.NewReader(kvKey).Exec(); rs.OK() {
				rs.JsonDecode(&cfgs)
				if pv := cfgs.Get("index_updated"); pv.String() != "" {
					q.Where().And("updated.ge", pv.String())
				}
			}

			for {

				rs, err := store.Data.Query(q)
				if err != nil {
					break
				}

				for _, v := range rs {

					id := v.Field("id").String()

					u64 := hex16ToUint64(id)
					if u64 == 0 {
						break
					}

					item := api.Node{
						ID:      v.Field("id").String(),
						PID:     v.Field("pid").String(),
						Status:  v.Field("status").Int16(),
						UserID:  v.Field("userid").String(),
						Created: v.Field("created").Uint32(),
						Updated: v.Field("updated").Uint32(),
					}

					if model.Extensions.AccessCounter {
						item.ExtAccessCounter = v.Field("ext_access_counter").Uint32()
					}

					if model.Extensions.CommentEnable {
						if model.Extensions.CommentPerEntry && v.Field("ext_comment_perentry").Bool() == false {
							item.ExtCommentEnable = false
							item.ExtCommentPerEntry = false
						} else {
							item.ExtCommentEnable = true
							item.ExtCommentPerEntry = true
						}
					}

					if model.Extensions.Permalink != "" && v.Field("ext_permalink_name").String() != "" {
						item.ExtPermalinkName = v.Field("ext_permalink_name").String()
						item.SelfLink = fmt.Sprintf("%s", item.ExtPermalinkName)
					} else {
						item.SelfLink = fmt.Sprintf("%s.html", item.ID)
					}

					for _, field := range model.Fields {

						nodeField := api.NodeField{
							Name:  field.Name,
							Value: v.Field("field_" + field.Name).String(),
						}

						if field.Type == "text" &&
							len(v.Field("field_"+field.Name+"_attrs").String()) > 10 {

							var attrs types.KvPairs
							if err := v.Field("field_" + field.Name + "_attrs").JsonDecode(&attrs); err == nil {
								nodeField.Attrs = attrs
							}
						}

						if l := field.Attrs.Get("langs"); len(l) > 3 {

							if len(v.Field("field_"+field.Name+"_langs").String()) > 5 {
								var node_langs api.NodeFieldLangs
								if err := v.Field("field_" + field.Name + "_langs").JsonDecode(&node_langs); err == nil {
									nodeField.Langs = &node_langs
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
						case api.TermTaxonomy:

							if ttv, ok := modCache.termTaxonomy[term.Meta.Name+"."+v.Field("term_"+term.Meta.Name).String()]; ok {
								termItem := api.NodeTerm{
									Name:  term.Meta.Name,
									Value: v.Field("term_" + term.Meta.Name).String(),
									Type:  term.Type,
								}

								termItem.Items = append(termItem.Items, ttv)

								item.Terms = append(item.Terms, termItem)
							}

						case api.TermTag:

							tags := strings.Split(v.Field("term_"+term.Meta.Name).String(), ",")

							if len(tags) > 0 {
								termItem := api.NodeTerm{
									Name:  term.Meta.Name,
									Value: v.Field("term_" + term.Meta.Name).String(),
									Type:  term.Type,
								}

								for _, vtag := range tags {
									termItem.Items = append(termItem.Items, api.Term{
										Title: vtag,
									})

								}

								item.Terms = append(item.Terms, termItem)
							}
						}
					}

					// fmt.Println(id, v.Field("title").String())

					if err := nodeSearcher.Put(tblname, item); err != nil {
						return err
					}

					indexNum += 1
				}

				if len(rs) < int(limit) {
					break
				}

				offset += limit
				q.Offset(offset)
			}

			if indexNum > 0 {
				cfgs.Set("index_updated", indexUpdated)
				if rs := store.DataLocal.NewWriter(kvKey, nil).SetJsonValue(cfgs).Exec(); !rs.OK() {
					hlog.Printf("warn", "search index error")
				}
				hlog.Printf("info", "search data sync %d at %v",
					indexNum, time.Since(indexStart))
			}
		}
	}

	return nil
}

func (q *QuerySet) NodeListSearch(qry string) api.NodeList {

	var rsp api.NodeList

	if !searchInited || nodeSearcher == nil {
		rsp.Error = types.NewErrorMeta(api.ErrCodeBadArgument, "Server Not Ready")
		return rsp
	}

	table := fmt.Sprintf("hpn_%s_%s",
		idhash.HashToHexString([]byte(q.ModName), 12), q.Table)

	return nodeSearcher.Query(table, qry, q)
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
