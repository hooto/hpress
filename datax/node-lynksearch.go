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
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/hooto/hlog4g/hlog"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/store"

	"github.com/lynkdb/lynkapi/go/codec"
	"github.com/lynkdb/lynkapi/go/lynkapi"
	"github.com/lynkdb/lynksearch/pkg/lynksearch"
)

type NodeLynkSearchEngine struct {
	mu            sync.Mutex
	dataPath      string
	cfgConfigPath string
	running       bool
	cfgs          LynkSearchConfig
	actives       map[string]*lynkSearchBucketActive
	arrActives    []*lynkSearchBucketActive
}

type LynkSearchConfig struct {
	Prefix  string                         `json:"prefix"`
	Buckets []*LynkSearchConfigBucketEntry `json:"buckets"`
	Daemon  struct {
		CpuCoreNum  int `json:"cpu_core_num"`
		MaxChildren int `json:"max_children"`
	} `json:"daemon"`
}

type LynkSearchConfigBucketEntry struct {
	Name             string         `json:"name"`
	StatsFullIndexed int64          `json:"-"` // `json:"stats_full_indexed"`
	Model            *api.NodeModel `json:"model"`
}

type lynkSearchBucketActive struct {
	mu sync.Mutex

	config *LynkSearchConfigBucketEntry

	lynkSearch lynksearch.Instance

	updateNum int64
}

func NewNodeLynkSearchEngine(prefix string) (NodeSearchEngine, error) {

	engine := &NodeLynkSearchEngine{
		dataPath:      filepath.Clean(prefix + "/var/lynksearch"),
		cfgConfigPath: filepath.Clean(prefix + "/etc/lynksearch.json"),
		actives:       map[string]*lynkSearchBucketActive{},
	}

	json.DecodeFile(engine.cfgConfigPath, &engine.cfgs)

	engine.cfgs.Prefix = filepath.Clean(prefix)

	for _, buk := range engine.cfgs.Buckets {
		if _, ok := engine.actives[buk.Name]; ok {
			continue
		}
		active := &lynkSearchBucketActive{
			config:    buk,
			updateNum: -1,
		}
		engine.actives[buk.Name] = active
		engine.arrActives = append(engine.arrActives, active)
	}

	go engine.run()

	return engine, nil
}

func (it *NodeLynkSearchEngine) configRefresh() error {
	return json.EncodeToFile(it.cfgs, it.cfgConfigPath, "  ")
}

func (it *NodeLynkSearchEngine) ModelSet(bukName string, model *api.NodeModel) error {

	if model == nil {
		return fmt.Errorf("NodeModel required")
	}

	it.mu.Lock()
	defer it.mu.Unlock()

	//
	active, ok := it.actives[bukName]
	if !ok {

		config := &LynkSearchConfigBucketEntry{
			Name:  bukName,
			Model: model,
		}

		active = &lynkSearchBucketActive{
			config:    config,
			updateNum: -1,
		}

		it.actives[bukName] = active
		it.arrActives = append(it.arrActives, active)

		it.cfgs.Buckets = append(it.cfgs.Buckets, config)

		if err := it.configRefresh(); err != nil {
			return err
		}
	}

	return nil
}

func (it *NodeLynkSearchEngine) active(bukName string) (*lynkSearchBucketActive, error) {

	it.mu.Lock()
	defer it.mu.Unlock()

	active, ok := it.actives[bukName]
	if !ok {
		return nil, fmt.Errorf("bucket (%s) not setup", bukName)
	}

	return active, it.trySetupBucket(active)
}

func (it *NodeLynkSearchEngine) trySetupBucket(active *lynkSearchBucketActive) error {

	if active.lynkSearch != nil {
		return nil
	}

	spec := &lynkapi.TableSpec{}

	spec.SetField("content", lynkapi.FieldSpec_String)
	spec.SetIndex("content", lynkapi.TableSpec_Index_FullTextSearch)

	if ins, err := lynksearch.NewInstance(lynksearch.InstanceConfig{
		Dir: it.dataPath + "/" + active.config.Name,
	}, spec); err != nil {
		return fmt.Errorf("bucket (%s) indexer setup fail : %s", active.config.Name, err.Error())
	} else {
		active.lynkSearch = ins
	}

	hlog.Printf("info", "search bucket (%s) setup ok", active.config.Name)
	return nil
}

func (it *NodeLynkSearchEngine) run() {

	for {
		time.Sleep(1e9)
		it.runAction()
	}
}

func (it *NodeLynkSearchEngine) runAction() {

	tn := int64(time.Now().Unix())

	for _, active := range it.arrActives {

		if (tn-active.config.StatsFullIndexed) < 86400 &&
			active.updateNum == 0 {
			continue
		}
		active.config.StatsFullIndexed = tn

		if err := it.trySetupBucket(active); err != nil {
			continue
		}

		if err := it.indexFull(active); err == nil {
			active.config.StatsFullIndexed = tn
			json.EncodeToFile(it.cfgs, it.cfgConfigPath, "  ")
		} else {
			hlog.Printf("error", "index/full ER %s", err.Error())
		}
	}
}

func (it *NodeLynkSearchEngine) indexFull(active *lynkSearchBucketActive) error {

	if active.lynkSearch == nil {
		return fmt.Errorf("search not setup (buk %s)", active.config.Name)
	}

	var (
		offset = api.NsTextSearchCacheNodeEntry(active.config.Name, "")
		cutset = api.NsTextSearchCacheNodeEntry(active.config.Name, "")
		num    = 0
	)

	for {

		ls := store.DataLocal.NewRanger(offset, cutset).
			SetLimit(1000).Exec()

		for _, v := range ls.Items {
			var nv api.Node
			if err := v.JsonDecode(&nv); err == nil && nv.Status == 1 {
				lynkSearchAddDocument(&nv, active)
				num += 1
			}
			offset = v.Key
		}

		if !ls.NextResultSet {
			break
		}
	}

	if num > 0 {
		if err := active.lynkSearch.Flush(); err != nil {
			hlog.Printf("error", "index/full buk %s, doc-num %d,  error %s",
				active.config.Name, num, err.Error())
			return err
		} else {
			hlog.Printf("info", "index/full buk %s, doc-num %d",
				active.config.Name, num)
		}
	}

	active.updateNum = 0

	return nil
}

func (it *NodeLynkSearchEngine) Put(bukName string, node api.Node) error {

	key := api.NsTextSearchCacheNodeEntry(bukName, node.ID)

	if node.Status == 1 {
		if rs := store.DataLocal.NewWriter(key, nil).SetJsonValue(node).Exec(); !rs.OK() {
			return errors.New("DataLocal/Put Error")
		}
	} else {
		if rs := store.DataLocal.NewDeleter(key).Exec(); !rs.OK() {
			return errors.New("DataLocal/Del Error")
		}
	}

	active, err := it.active(bukName)
	if err == nil && active.lynkSearch != nil {
		if node.Status == 1 {
			lynkSearchAddDocument(&node, active)
		} else {
			active.lynkSearch.DelDocument(node.ID)
		}
		active.updateNum += 1
	}

	return nil
}

func (it *NodeLynkSearchEngine) Query(bukName string, q string, qs *QuerySet) api.NodeList {

	var (
		ls api.NodeList

		active, err = it.active(bukName)
	)

	if err != nil {
		ls.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return ls
	}

	qr := lynkapi.NewDataQuery().
		AddFuncFilter("search", q).
		SetLimit(int32(1000 - qs.offset)).
		SetOffset(int32(qs.offset))

	rs := active.lynkSearch.Search(qr)
	if !rs.OK() {
		ls.Error = types.NewErrorMeta(api.ErrCodeInternalError, rs.Err().Error())
		return ls
	}
	{
		rsjs, _ := codec.Json.Encode(rs)
		hlog.Printf("info", "query result : %s", string(rsjs))
	}

	rows := rs.Rows

	if len(rows) == 0 {
		ls.Error = types.NewErrorMeta(api.ErrCodeInternalError, "server error")
		return ls
	}

	for _, row := range rows {
		if rs := store.DataLocal.NewReader(
			api.NsTextSearchCacheNodeEntry(bukName, row.Id)).Exec(); rs.OK() {
			var node api.Node
			if err := rs.JsonDecode(&node); err == nil {
				ls.Items = append(ls.Items, node)
			}
		}
	}

	ls.Kind = "NodeList"

	if qs.Pager {
		ls.Meta.TotalResults = uint64(rs.Stats.RowsHit)
		ls.Meta.StartIndex = uint64(qs.offset)
		ls.Meta.ItemsPerList = uint64(qs.limit)
	}

	return ls
}

func lynkSearchAddDocument(node *api.Node, active *lynkSearchBucketActive) {

	if node.Status != 1 {
		active.lynkSearch.DelDocument(node.ID)
		return
	}

	content := node.Title

	if len(node.Terms) > 0 {
		for _, nt := range node.Terms {
			if nt.Type != api.TermTag {
				continue
			}
			for _, ntv := range nt.Items {
				if content != "" {
					content += ","
				}
				content += ntv.Title
			}
		}
	}

	for _, mf := range node.Fields {
		if ft := mf.Attrs.Get("format"); len(ft) > 1 {
			content += "," + mf.Value
		}
	}

	active.lynkSearch.AddDocument(node.ID, map[string]any{
		"content": content,
	})
}
