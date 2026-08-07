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
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/datax/sphinxsearch"
	"github.com/hooto/hpress/store"
)

type NodeSphinxSearchEngine struct {
	mu                sync.Mutex
	prefix            string
	binIndexer        string
	binIndexTool      string
	binServer         string
	dataPath          string
	cfgConfigPath     string
	cfgVendorPath     string
	cfgVendorTemplate string
	cfgVendorStopword string
	running           bool
	cfgs              SphinxSearchConfig
	merge_max         int
	actives           []*sphinxSearchBucketActive
}

type SphinxSearchConfig struct {
	Prefix  string                           `json:"prefix"`
	Buckets []*SphinxSearchConfigBucketEntry `json:"buckets"`
	Daemon  struct {
		CpuCoreNum  int `json:"cpu_core_num"`
		MaxChildren int `json:"max_children"`
	} `json:"daemon"`
}

type SphinxSearchConfigBucketEntry struct {
	Name             string         `json:"name"`
	StatsFullIndexed int64          `json:"stats_full_indexed"`
	Model            *api.NodeModel `json:"model"`
	statsActive      bool
}

type sphinxSearchBucketActive struct {
	mu            sync.Mutex
	bukname       string
	deltas        []api.Node
	puts          []api.Node
	deltaIndexNum int
	model         *api.NodeModel
}

const (
	sphXmlPipeHeader = "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<sphinx:docset>\n"
	sphXmlPipeSchema = "<sphinx:schema>%s</sphinx:schema>\n"
	sphXmlPipeFooter = "</sphinx:docset>"
)

func NewNodeSphinxSearchEngine(prefix string) (NodeSearchEngine, error) {

	engine := &NodeSphinxSearchEngine{
		binIndexer:        filepath.Clean(prefix + "/bin/sph-indexer"),
		binIndexTool:      filepath.Clean(prefix + "/bin/sph-indextool"),
		binServer:         filepath.Clean(prefix + "/bin/sph-searchd"),
		dataPath:          filepath.Clean(prefix + "/var/sphinxsearch"),
		cfgConfigPath:     filepath.Clean(prefix + "/etc/sphinxsearch.json"),
		cfgVendorPath:     filepath.Clean(prefix + "/etc/sphinxsearch.conf"),
		cfgVendorTemplate: filepath.Clean(prefix + "/misc/sphinxsearch/sphinxsearch.conf.tpl"),
		cfgVendorStopword: filepath.Clean(prefix + "/misc/sphinxsearch/stopword.conf"),
		merge_max:         1000,
	}

	for _, path := range []string{
		engine.binIndexer,
		engine.binIndexTool,
		engine.binServer,
		engine.cfgVendorTemplate,
		engine.cfgVendorStopword,
	} {
		fp, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		fp.Close()
	}

	if err := os.MkdirAll(engine.dataPath, 0755); err != nil {
		return nil, err
	}

	json.DecodeFile(engine.cfgConfigPath, &engine.cfgs)

	engine.cfgs.Prefix = filepath.Clean(prefix)
	engine.cfgs.Daemon.CpuCoreNum = runtime.NumCPU()
	engine.cfgs.Daemon.MaxChildren = runtime.NumCPU() * 2

	engine.configRefresh()

	go engine.run()

	return engine, nil
}

func (it *NodeSphinxSearchEngine) ModelSet(bukname string, model *api.NodeModel) error {

	//
	var (
		active = it.active(bukname)
		buk    = it.bucket(bukname, false)
	)

	it.mu.Lock()
	defer it.mu.Unlock()

	if active.model == nil {
		active.model = model
	}

	if buk != nil && buk.Model == nil {
		buk.Model = model
		json.EncodeToFile(it.cfgs, it.cfgConfigPath, "  ")
	}

	return nil
}

func (it *NodeSphinxSearchEngine) bucket(bukname string, create bool) *SphinxSearchConfigBucketEntry {

	it.mu.Lock()
	defer it.mu.Unlock()

	for _, v := range it.cfgs.Buckets {
		if v.Name == bukname {
			return v
		}
	}

	if create {

		buk := &SphinxSearchConfigBucketEntry{
			Name: bukname,
		}

		it.cfgs.Buckets = append(it.cfgs.Buckets, buk)

		return buk
	}

	return nil
}

func (it *NodeSphinxSearchEngine) active(bukname string) *sphinxSearchBucketActive {

	it.mu.Lock()
	defer it.mu.Unlock()

	for _, v := range it.actives {
		if v.bukname == bukname {
			return v
		}
	}

	active := &sphinxSearchBucketActive{
		bukname: bukname,
	}

	for _, v := range it.cfgs.Buckets {
		if v.Name == bukname {
			active.model = v.Model
			break
		}
	}

	it.actives = append(it.actives, active)

	return active
}

func (it *NodeSphinxSearchEngine) configRefresh() error {

	if len(it.cfgs.Buckets) < 1 {
		return nil
	}

	fp, err := os.Open(it.cfgVendorTemplate)
	if err != nil {
		return err
	}
	defer fp.Close()

	txt, err := ioutil.ReadAll(fp)
	if err != nil {
		return err
	}

	fp2, err := os.OpenFile(it.cfgVendorPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	fp2.Seek(0, 0)
	fp2.Truncate(0)
	defer fp2.Close()

	tpl, err := template.New("tmp").Parse(string(txt))
	if err != nil {
		return err
	}

	var bs bytes.Buffer
	if err := tpl.Execute(&bs, map[string]interface{}{
		"config": it.cfgs,
	}); err != nil {
		return err
	}

	if _, err := fp2.Write(bs.Bytes()); err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("config/refresh buckets: %d", len(it.cfgs.Buckets)))

	return nil
}

func (it *NodeSphinxSearchEngine) setupServer() error {

	if len(it.cfgs.Buckets) < 1 {
		return nil
	}

	var (
		out, err = exec.Command("pidof", it.binServer).Output()
		pids     = strings.Split(strings.TrimSpace(string(out)), " ")
		running  = false
	)
	if err == nil && len(pids) > 0 {
		running = true
	}

	for _, buk := range it.cfgs.Buckets {

		if buk.statsActive {
			continue
		}

		for _, v := range []string{"full", "delta"} {

			if fp, err := os.Open(it.dataPath + "/" + buk.Name + "/" + v + ".sph"); err == nil {
				fp.Close()
				continue
			}

			idxname := buk.Name + "_" + v

			if fp, err := os.Open(it.dataPath + "/" + buk.Name); err != nil {
				os.MkdirAll(it.dataPath+"/"+buk.Name, 0755)
			} else {
				fp.Close()
			}

			if _, err := it.dataRefresh(buk.Name, v, nil); err != nil {
				slog.Error(fmt.Sprintf("setup/DataSource %s/%s, ER %s", buk.Name, v, err.Error()))
			}

			args := []string{
				"-c",
				it.cfgVendorPath,
				idxname,
			}
			if running {
				args = append(args, "--rotate")
			}

			if _, err := exec.Command(it.binIndexer, args...).Output(); err != nil {
				slog.Warn(fmt.Sprintf("setup/DataIndex %s, ER %s", idxname, err.Error()))
			}
		}

		slog.Info(fmt.Sprintf("server/setup %s", buk.Name))
		buk.statsActive = true
	}

	if !running {
		slog.Info(fmt.Sprintf("setup/ServerStart"))
		if _, err = exec.Command(it.binServer, "-c", it.cfgVendorPath).Output(); err != nil {
			return err
		}
	}

	return nil
}

func (it *NodeSphinxSearchEngine) run() {

	for {
		time.Sleep(1e9)
		it.runAction()
	}
}

func (it *NodeSphinxSearchEngine) runAction() {

	tn := int64(time.Now().Unix())

	for _, active := range it.actives {
		if buk := it.bucket(active.bukname, false); buk == nil {
			it.bucket(active.bukname, true)
			it.configRefresh()
		}
	}

	if err := it.setupServer(); err != nil {
		slog.Error(fmt.Sprintf("server %s", err.Error()))
		return
	}

	for _, active := range it.actives {

		buk := it.bucket(active.bukname, false)
		if buk == nil {
			continue
		}

		it.deltaRefresh(active)

		if (tn - buk.StatsFullIndexed) > 86400 {
			if err := it.indexFull(active); err == nil {
				buk.StatsFullIndexed = tn
				json.EncodeToFile(it.cfgs, it.cfgConfigPath, "  ")
			} else {
				slog.Error(fmt.Sprintf("index/full ER %s", err.Error()))
			}
		}

		if len(active.deltas) > active.deltaIndexNum {
			if n, err := it.indexDelta(active); err == nil {
				active.deltaIndexNum = n
			} else {
				slog.Error(fmt.Sprintf("index/delta ER %s", err.Error()))
			}
		}

		if len(active.deltas) > it.merge_max {
			slog.Info(fmt.Sprintf("merge %d", len(active.deltas)))
			if err := it.indexMerge(active.bukname); err == nil {
				active.deltas = []api.Node{}
				active.deltaIndexNum = 0
			} else {
				slog.Error(fmt.Sprintf("index/merge ER %s", err.Error()))
			}
		}
	}
}

func (it *NodeSphinxSearchEngine) deltaRefresh(active *sphinxSearchBucketActive) {

	active.mu.Lock()
	defer active.mu.Unlock()

	for _, v := range active.puts {

		found := false

		for i2, v2 := range active.deltas {
			if v.ID == v2.ID {
				active.deltas[i2] = v
				found = true
				break
			}
		}

		if !found {
			active.deltas = append(active.deltas, v)
		}
	}

	active.puts = []api.Node{}
}

func (it *NodeSphinxSearchEngine) indexFull(active *sphinxSearchBucketActive) error {

	active.mu.Lock()
	if len(active.deltas) > 0 {
		active.deltas = []api.Node{}
	}
	active.mu.Unlock()

	n, err := it.dataRefresh(active.bukname, "full", func(fpbuf *bufio.Writer) int {

		var (
			offset = api.NsTextSearchCacheNodeEntry(active.bukname, "")
			cutset = api.NsTextSearchCacheNodeEntry(active.bukname, "")
			limit  = 1000
			num    = 0
		)

		for {

			ls := store.DataLocal.NewRanger(offset, cutset).
				SetLimit(int64(limit)).Exec()

			for _, v := range ls.Items {
				var nv api.Node
				if err := v.JsonDecode(&nv); err == nil {
					if _, err := fpbuf.WriteString(sphDocumentXml(&nv, active)); err != nil {
						return 0
					}
				}
				num += 1
				offset = v.Key
			}

			if !ls.NextResultSet {
				break
			}
		}

		return num
	})

	if err != nil {
		return err
	}

	if n > 0 {
		if err = it.index(active, "full"); err != nil {
			slog.Error(fmt.Sprintf("index/full ER %s", err.Error()))
		} else {
			slog.Info(fmt.Sprintf("index/full %s, num %d", active.bukname, n))
		}
	}

	return err
}

func (it *NodeSphinxSearchEngine) indexDelta(active *sphinxSearchBucketActive) (int, error) {

	n, err := it.dataRefresh(active.bukname, "delta", func(fpbuf *bufio.Writer) int {

		for _, nv := range active.deltas {
			if _, err := fpbuf.WriteString(sphDocumentXml(&nv, active)); err != nil {
				return 0
			}
		}

		return len(active.deltas)
	})

	if err != nil {
		return n, err
	}

	if n > 0 {
		if err = it.index(active, "delta"); err != nil {
			slog.Error(fmt.Sprintf("index/delta ER %s", err.Error()))
		} else {
			slog.Info(fmt.Sprintf("index/delta %s, num %d", active.bukname, n))
		}
	}

	return n, err
}

func (it *NodeSphinxSearchEngine) indexMerge(bukname string) error {

	args := []string{
		"-c",
		it.cfgVendorPath,
		"--merge",
		bukname + "_full",
		bukname + "_delta",
		"--merge-killlists",
		"--rotate",
	}

	out, _ := exec.Command(it.binIndexer, args...).Output()

	if strings.Contains(string(out), "indices NOT rotated") {
		it.indexRepair(bukname + "_full")
		it.indexRepair(bukname + "_delta")
		out, _ = exec.Command(it.binIndexer, args...).Output()
	}

	if strings.Contains(string(out), "successfully") {
		return nil
	}

	return errors.New("server error")
}

func (it *NodeSphinxSearchEngine) dataRefresh(
	bukname, idxtype string, fn func(fpbuf *bufio.Writer) int) (int, error) {

	var (
		path = fmt.Sprintf("%s/%s/%s.xml", it.dataPath, bukname, idxtype)
		num  = 0
	)

	// slog.Info(fmt.Sprintf("dataRefresh %s %s", bukname, idxtype))

	fp, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return 0, err
	}
	fp.Seek(0, 0)
	fp.Truncate(0)
	defer fp.Close()

	fpbuf := bufio.NewWriter(fp)

	fpbuf.WriteString(sphXmlPipeHeader)

	schemas := []string{
		fmt.Sprintf(`<sphinx:attr name="%s" type="string"/>`, "nid"),
		fmt.Sprintf(`<sphinx:attr name="%s" type="int" bits="8" default="0"/>`, "status"),
		fmt.Sprintf(`<sphinx:field name="%s"/>`, "title"),
		fmt.Sprintf(`<sphinx:field name="%s"/>`, "term_tags"),
		fmt.Sprintf(`<sphinx:field name="%s"/>`, "content"),
		fmt.Sprintf(`<sphinx:attr name="%s" type="timestamp"/>`, "created"),
	}

	fpbuf.WriteString(fmt.Sprintf(sphXmlPipeSchema, strings.Join(schemas, "\n")))

	if fn != nil {
		if num = fn(fpbuf); num > 0 {
			slog.Info(fmt.Sprintf("data/refresh %s/%s, num %d", bukname, idxtype, num))
		}
	}

	fpbuf.WriteString(sphXmlPipeFooter)
	fpbuf.Flush()

	return num, nil
}

func (it *NodeSphinxSearchEngine) index(active *sphinxSearchBucketActive, idxtype string) error {

	args := []string{
		"-c",
		it.cfgVendorPath,
		active.bukname + "_" + idxtype,
		"--rotate",
	}

	out, _ := exec.Command(it.binIndexer, args...).Output()

	if strings.Contains(string(out), "indices NOT rotated") {
		it.indexRepair(active.bukname + "_" + idxtype)
		out, _ = exec.Command(it.binIndexer, args...).Output()
	}

	if strings.Contains(string(out), "successfully") {
		return nil
	}

	return errors.New("server error " + string(out))
}

func (it *NodeSphinxSearchEngine) indexRepair(idxname string) error {
	_, err := exec.Command(it.binIndexTool, "-c", it.cfgVendorPath, "--check", idxname).Output()
	if err != nil {
		slog.Error(fmt.Sprintf("indexRepair %s, err %s", idxname, err.Error()))
	} else {
		slog.Info(fmt.Sprintf("indexRepair %s", idxname))
	}
	return err
}

func (it *NodeSphinxSearchEngine) Put(bukname string, node api.Node) error {

	if rs := store.DataLocal.NewWriter(api.NsTextSearchCacheNodeEntry(bukname, node.ID), nil).SetJsonValue(node).Exec(); !rs.OK() {
		return errors.New("DataLocal/Put Error")
	}

	active := it.active(bukname)

	active.mu.Lock()
	defer active.mu.Unlock()

	for i, v := range active.puts {
		if v.ID == node.ID {
			active.puts[i] = node
			return nil
		}
	}

	active.puts = append(active.puts, node)

	return nil
}

func (it *NodeSphinxSearchEngine) Query(bukname string, q string, qs *QuerySet) api.NodeList {

	var ls api.NodeList

	opts := sphinxsearch.DefaultOptions
	opts.MaxQueryTime = 5000
	opts.Socket = it.dataPath + "/searchd.sock"

	client := sphinxsearch.NewClient(opts)
	if err := client.Error(); err != nil {
		ls.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return ls
	}
	defer client.Close()

	client.SetFieldWeights(map[string]int{
		"title":     40,
		"term_tags": 20,
		"content":   1,
	})

	client.SetLimits(int(qs.offset), int(qs.limit), 1000, 0)
	client.SetFilter("status", []uint64{1}, false)
	client.SetMatchMode(sphinxsearch.SPH_MATCH_EXTENDED)

	rss, err := client.Query(q, bukname, "")
	if err != nil {
		ls.Error = types.NewErrorMeta(api.ErrCodeInternalError, err.Error())
		return ls
	}

	id_idx := -1
	for i, v := range rss.AttrNames {
		if v == "nid" {
			id_idx = i
			break
		}
	}

	if id_idx == -1 {
		ls.Error = types.NewErrorMeta(api.ErrCodeInternalError, "server error")
		return ls
	}

	for _, v := range rss.Matches {
		if id_idx >= len(v.AttrValues) {
			continue
		}

		if rs := store.DataLocal.NewReader(
			api.NsTextSearchCacheNodeEntry(bukname, fmt.Sprintf("%s", v.AttrValues[id_idx]))).Exec(); rs.OK() {
			var node api.Node
			if err := rs.JsonDecode(&node); err == nil {
				ls.Items = append(ls.Items, node)
			}
		}
	}

	ls.Kind = "NodeList"

	if qs.Pager {
		ls.Meta.TotalResults = uint64(rss.TotalFound)
		ls.Meta.StartIndex = uint64(qs.offset)
		ls.Meta.ItemsPerList = uint64(qs.limit)
	}

	return ls
}

func sphTextFilter(txt string) string {

	for _, v := range []string{
		`"`,
		"'",
		"<",
		">",
		"\r",
		"\n",
	} {
		txt = strings.Replace(txt, v, " ", -1)
	}

	return txt
}

func sphDocumentXml(node *api.Node, active *sphinxSearchBucketActive) string {

	u64 := sphHex16ToUint64(node.ID)
	if u64 < 1 {
		return ""
	}

	xml := fmt.Sprintf(`<sphinx:document id="%d">`, u64)

	xml += fmt.Sprintf(`<%s>%s</%s>`, "nid", node.ID, "nid")
	xml += fmt.Sprintf(`<%s>%d</%s>`, "status", node.Status, "status")
	xml += fmt.Sprintf(`<%s><![CDATA[%s]]></%s>`, "title", sphTextFilter(node.Title), "title")

	if len(node.Terms) > 0 {
		terms := ""
		for _, nt := range node.Terms {
			if nt.Type != api.TermTag {
				continue
			}
			for _, ntv := range nt.Items {
				if terms != "" {
					terms += ","
				}
				terms += ntv.Title
			}
		}
		if len(terms) > 0 {
			xml += fmt.Sprintf(`<%s><![CDATA[%s]]></%s>`, "term_tags", sphTextFilter(terms), "term_tags")
		}
	}

	content := ""
	for _, mf := range node.Fields {
		if ft := mf.Attrs.Get("format"); len(ft) > 1 {
			content += mf.Value
		}
	}
	if len(content) > 0 {
		xml += fmt.Sprintf(`<%s><![CDATA[%s]]></%s>`, "content", sphTextFilter(content), "content")
	}

	xml += "</sphinx:document>\n"

	return xml
}

func sphHex16ToUint64(str string) uint64 {
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
