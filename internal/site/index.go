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

package site

import (
	"bytes"
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/iam/v2/pkg/iamserver"
	"github.com/lessos/lessgo/crypto/idhash"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/datax"
	"github.com/hooto/hpress/internal/hpapi"
	"github.com/hooto/hpress/internal/store"
	"github.com/hooto/hpress/internal/utils"
	"github.com/hooto/hpress/internal/web"
)

// indexContext holds the per-request state previously captured by the legacy
// controller layer. It is request-scoped: one instance is created at the top of
// IndexPage and threaded through filter and dataRender.
type indexContext struct {
	c      fiber.Ctx
	data   map[string]any
	ovr    map[string]string // route-param overrides set by filter
	us     iamserver.UserSession
	hooks  []func() // post-render callbacks (cache writers)
	urlAct string
}

func newIndexContext(c fiber.Ctx) *indexContext {
	return &indexContext{
		c:    c,
		data: map[string]any{},
		ovr:  map[string]string{},
		us:   web.AuthSession(c),
	}
}

func (ix *indexContext) param(key string) string {
	if v, ok := ix.ovr[key]; ok && v != "" {
		return v
	}
	return web.Param(ix.c, key)
}

func (ix *indexContext) paramInt(key string) int64 {
	v := ix.param(key)
	if v == "" {
		return 0
	}
	n, _ := strconv.ParseInt(v, 10, 64)
	return n
}

func (ix *indexContext) filter(rt []string, spec *hpapi.Spec) (string, string, bool) {

	for _, route := range spec.Router.Routes {

		matlen, params := 0, map[string]string{}

		for i, node := range route.Tree {

			if len(node) < 1 || i >= len(rt) {
				break
			}

			if node[0] == ':' {

				params[node[1:]] = rt[i]

				matlen++

			} else if node == rt[i] {

				matlen++
			}
		}

		if matlen == len(route.Tree) {

			for k, v := range params {
				ix.ovr[k] = v
			}

			return route.DataAction, route.Template, true
		}
	}

	for _, route := range spec.Router.Routes {
		if route.Default {
			return route.DataAction, route.Template, true
		}
	}

	return "", "", false
}

var (
	srvnameDefault     = "core-genereal"
	urisDefault        = []string{"core-general"}
	dataRenderOK       = 0
	dataRenderNotFound = 1
	dataRenderSkip     = 2
	staticImages       = types.ArrayString([]string{
		"png", "jpg", "jpeg", "gif", "webp", "svg", "ico",
	})

	gdocPathRX = regexp.MustCompile(`^view\/([a-zA-Z-_0-9]+)\/(.*)$`)
)

func IndexPage(c fiber.Ctx) error {

	start := time.Now().UnixNano()

	if v := config.SysConfigList.FetchString("http_h_ac_allow_origin"); v != "" {
		c.Set("Access-Control-Allow-Origin", v)
	}

	ix := newIndexContext(c)

	var (
		reqpath = web.UrlPath(c)
		uris    = []string{}
	)
	if reqpath == "" || reqpath == "." {
		reqpath = "/"
	}
	if len(reqpath) > 0 && reqpath != "/" {
		uris = strings.Split(strings.Trim(reqpath, "/"), "/")
	}

	if len(uris) < 1 {
		if config.RouterBasepathDefault != "/" {
			reqpath = config.RouterBasepathDefault
			uris = config.RouterBasepathDefaults
		} else {
			uris = urisDefault
		}
	}
	srvname := uris[0]

	if len(uris) < 2 {
		uris = append(uris, "")
	}
	// fmt.Println(uris, srvname, ix.param("referid"), ix.param("id"))

	mod, ok := config.Modules[srvname]
	if !ok {
		srvname = srvnameDefault
		mod, ok = config.Modules[srvname]
		if !ok {
			return nil
		}
	}

	ix.urlAct = strings.Join(uris[1:], "/")

	dataAction, template, mat := ix.filter(uris[1:], mod)
	if !mat {
		if uris[1] == "" {
			template = "index.tpl"
		} else {
			template = "404.tpl"
		}
	}

	ix.data["LANG"] = hpapi.LangHit(config.Languages, web.ResolveLang(c))

	if len(config.Languages) > 1 {
		ix.data["frontend_langs"] = config.Languages
	}

	// if session, err := c.Session.Instance(); err == nil {
	// 	ix.data["session"] = session
	// }

	ix.data["baseuri"] = "/" + srvname
	ix.data["http_request_path"] = reqpath
	ix.data["srvname"] = srvname
	ix.data["modname"] = mod.Meta.Name
	ix.data["sys_version_sign"] = config.SysVersionSign
	if s, err := ix.us.Profile(); err == nil {
		ix.data["s_user"] = s.Username
	}

	drs := dataRenderOK

	if dataAction != "" {

		for _, action := range mod.Actions {

			if action.Name != dataAction {
				continue
			}

			for _, datax := range action.Datax {
				drs = ix.dataRender(srvname, action.Name, datax)
				ix.data["__datax_table__"] = datax.Query.Table
			}

			break
		}
	}

	switch drs {
	case dataRenderOK:

		// render_start := time.Now()
		var buf bytes.Buffer
		if err := web.Templates.Render(&buf, mod.Meta.Name, template, ix.data); err != nil {
			return err
		}

		// fmt.Println("render in-time", mod.Meta.Name, template, time.Since(render_start))

		fmt.Fprintf(&buf, "<!-- version %s, rt-time/db+render %d ms -->",
			config.Version, (time.Now().UnixNano()-start)/1e6)

		c.Set("Content-Type", "text/html; charset=utf-8")
		if err := c.Send(buf.Bytes()); err != nil {
			return err
		}

		// fmt.Println("hookPosts", len(ix.hooks))
		for _, fn := range ix.hooks {
			fn()
		}

		return nil

	case dataRenderNotFound:
		return web.RenderError(c, fiber.StatusNotFound, "Page Not Found")
	}

	return nil
}

func (ix *indexContext) dataRender(srvname, actionName string, ad hpapi.ActionData) int {

	mod, ok := config.Modules[srvname]
	if !ok {
		return dataRenderNotFound
	}

	query := datax.NewQuery(mod.Meta.Name, ad.Query.Table)
	if ad.Query.Limit > 0 {
		query.Limit(ad.Query.Limit)
	}

	if ad.Query.Order != "" {
		query.Order(ad.Query.Order)
	}

	query.Filter("status", 1)

	query.Pager = ad.Pager

	switch ad.Type {

	case "node.list":

		for _, modNode := range mod.NodeModels {

			if ad.Query.Table != modNode.Meta.Name {
				continue
			}

			for _, term := range modNode.Terms {

				if termVal := ix.param("term_" + term.Meta.Name); termVal != "" {

					switch term.Type {

					case hpapi.TermTaxonomy:

						if idxs := datax.TermTaxonomyCacheIndexes(mod.Meta.Name, term.Meta.Name, termVal); len(idxs) > 1 {
							args := []interface{}{}
							for _, idx := range idxs {
								args = append(args, idx)
							}
							query.Filter("term_"+term.Meta.Name+".in", args...)
						} else {
							query.Filter("term_"+term.Meta.Name, termVal)
						}

						ix.data["term_"+term.Meta.Name] = termVal

					case hpapi.TermTag:
						// TOPO
						query.Filter("term_"+term.Meta.Name+".like", "%"+termVal+"%")
						ix.data["term_"+term.Meta.Name] = termVal
					}
				}
			}

			break
		}

		page := ix.paramInt("page")
		if page > 1 {
			query.Offset(ad.Query.Limit * (page - 1))
		}

		if ix.param("qry_text") != "" {
			query.Filter("field_title.like", "%"+ix.param("qry_text")+"%")
			ix.data["qry_text"] = ix.param("qry_text")
		}

		var ls hpapi.NodeList
		qryhash := query.Hash()

		if ad.CacheTTL > 0 && !ix.us.Allow("", "editor.write") {
			if rs := store.DataLocal.NewReader([]byte(qryhash)).Exec(); rs.OK() {
				rs.JsonDecode(&ls)
			}
		}

		if len(ls.Items) == 0 {

			if ix.param("qry_text") != "" {
				ls = query.NodeListSearch(ix.param("qry_text"))
				if ls.Error != nil {
					ls = query.NodeList([]string{}, []string{})
				}
			} else {
				ls = query.NodeList([]string{}, []string{})
			}
			// fmt.Println("index node.list")
			if ad.CacheTTL > 0 && len(ls.Items) > 0 {
				ix.hooks = append(
					ix.hooks,
					func() {
						store.DataLocal.NewWriter([]byte(qryhash), nil).SetJsonValue(ls).SetTTL(ad.CacheTTL).Exec()
					},
				)
			}
		}

		ix.data[ad.Name] = ls

		if query.Pager {
			pager := utils.NewPager(uint64(page),
				uint64(ls.Meta.TotalResults),
				uint64(ls.Meta.ItemsPerList),
				10)
			ix.data[ad.Name+"_pager"] = pager
		}

	case "node.entry":

		nodeId := ix.param(ad.Name + "_id")
		if nodeId == "" {
			nodeId = ix.param("id")
			if nodeId == "" {
				return dataRenderNotFound
			}
		}

		nodeModel, err := config.SpecNodeModel(mod.Meta.Name, ad.Query.Table)
		if err != nil {
			return dataRenderNotFound
		}

		nodeRefer := ""
		if nodeModel.Extensions.NodeRefer != "" {
			if mv, ok := ix.data[actionName+"_nsr_"+nodeModel.Extensions.NodeRefer]; ok {
				nodeRefer = mv.(string)
			}
		}

		var (
			nodeExt = ""
		)

		if mod.Meta.Name == "core/gdoc" {
			if ad.Query.Table == "page" {
				if mat := gdocPathRX.FindAllStringSubmatch(ix.urlAct, 1); len(mat) == 1 {
					nodeId = strings.ToLower(mat[0][2])
				}
			} else if ad.Query.Table == "doc" && hpapi.NodeIdReg.MatchString(nodeId) {
				nodeExt = "html"
			}
		}
		if i := strings.LastIndex(nodeId, "."); i > 0 {
			nodeExt = nodeId[i+1:]
			nodeId = nodeId[:i]
		}

		if nodeExt == "html" {
			query.Filter("id", nodeId)
		} else if staticImages.Has(nodeExt) {
			if mod.Meta.Name == "core/gdoc" && ad.Query.Table == "page" {

				if docId := datax.GdocNodeId(ix.param("doc_entry_id")); docId != "" {

					localPath := datax.GdocLocalPath(docId)
					if localPath == "" {
						localPath = fmt.Sprintf("%s/var/vcs/%s", config.Prefix, docId)
					}
					if mat := gdocPathRX.FindAllStringSubmatch(ix.urlAct, 1); len(mat) == 1 {
						localPath += "/" + mat[0][2]
					}
					localPath = filepath.Clean(localPath)
					if err := s2Server(ix.c, ix.urlAct, localPath); err != nil {
						slog.Warn(fmt.Sprintf("s2Server: %v", err))
					}
				}
			}
			return dataRenderSkip

		} else if nodeModel.Extensions.Permalink != "" {
			if nodeModel.Extensions.NodeRefer != "" && nodeRefer == "" {
				return dataRenderNotFound
			}
			query.Filter("ext_permalink_idx", idhash.HashToHexString([]byte(nodeRefer+nodeId), 12))
		} else {
			return dataRenderNotFound
		}

		var entry hpapi.Node
		qryhash := query.Hash()
		if ad.CacheTTL > 0 && !ix.us.Allow("", "editor.write") {
			if rs := store.DataLocal.NewReader([]byte(qryhash)).Exec(); rs.OK() {
				rs.JsonDecode(&entry)
			}
		}

		if entry.ID == "" {
			entry = query.NodeEntry()
			if ad.CacheTTL > 0 && entry.Title != "" {
				ix.hooks = append(
					ix.hooks,
					func() {
						store.DataLocal.NewWriter([]byte(qryhash), nil).SetJsonValue(entry).SetTTL(ad.CacheTTL).Exec()
					},
				)
			}
		}

		if entry.ID == "" {
			return dataRenderNotFound
		}

		if nodeModel.Extensions.AccessCounter {

			if ips := strings.Split(ix.c.IP(), ":"); len(ips) > 1 {

				table := hpapi.NodeTableName(mod.Meta.Name, ad.Query.Table)
				store.DataLocal.NewWriter([]byte("access_counter/"+table+"/"+ips[0]+"/"+entry.ID), []byte("1")).Exec()
			}
		}

		if nodeModel.Extensions.NodeSubRefer != "" {
			// fmt.Println("setting", actionName, ad.Query.Table, nodeModel.Extensions.NodeSubRefer, "_id", entry.ID)
			ix.data[actionName+"_nsr_"+ad.Query.Table] = entry.ID
		}

		if entry.Title != "" {
			ix.data["__html_head_title__"] = datax.StringSub(datax.TextHtml2Str(entry.Title), 0, 50)
		}

		ix.data[ad.Name] = entry

	case "term.list":

		var ls hpapi.TermList
		qryhash := query.Hash()
		if ad.CacheTTL > 0 {
			if rs := store.DataLocal.NewReader([]byte(qryhash)).Exec(); rs.OK() {
				rs.JsonDecode(&ls)
			}
		}

		if len(ls.Items) == 0 {
			ls = query.TermList()
			if ad.CacheTTL > 0 && len(ls.Items) > 0 {
				store.DataLocal.NewWriter([]byte(qryhash), nil).SetJsonValue(ls).SetTTL(ad.CacheTTL).Exec()
			}
		}

		ix.data[ad.Name] = ls

		if query.Pager {
			ix.data[ad.Name+"_pager"] = utils.NewPager(0,
				uint64(ls.Meta.TotalResults),
				uint64(ls.Meta.ItemsPerList),
				10)
		}

	case "term.entry":

		var entry hpapi.Term
		qryhash := query.Hash()

		if ad.CacheTTL > 0 {
			if rs := store.DataLocal.NewReader([]byte(qryhash)).Exec(); rs.OK() {
				rs.JsonDecode(&entry)
			}
		}

		if entry.Title == "" {
			entry = query.TermEntry()
			if ad.CacheTTL > 0 && entry.Title != "" {
				store.DataLocal.NewWriter([]byte(qryhash), nil).SetJsonValue(entry).SetTTL(ad.CacheTTL).Exec()
			}
		}

		ix.data[ad.Name] = entry
	}

	return dataRenderOK
}
