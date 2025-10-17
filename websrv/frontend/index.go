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

package frontend

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hooto/httpsrv"
	"github.com/hooto/iam/iamapi"
	"github.com/hooto/iam/iamclient"
	"github.com/lessos/lessgo/crypto/idhash"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/config"
	"github.com/hooto/hpress/datax"
	"github.com/hooto/hpress/internal/utils"
	"github.com/hooto/hpress/store"
)

type Index struct {
	*httpsrv.Controller
	urlActionPath string
	hookPosts     []func()
	us            iamapi.UserSession
}

func (c *Index) Init() int {
	c.us, _ = iamclient.SessionInstance(c.Session)
	return 0
}

func (c Index) filter(rt []string, spec *api.Spec) (string, string, bool) {

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
				c.Params.SetValue(k, v)
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

func (c Index) IndexAction() {

	c.AutoRender = false
	start := time.Now().UnixNano()

	if v := config.SysConfigList.FetchString("http_h_ac_allow_origin"); v != "" {
		c.Response.Out.Header().Set("Access-Control-Allow-Origin", v)
	}

	var (
		reqpath = c.Request.UrlPath()
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
	// fmt.Println(uris, srvname, c.Params.Value("referid"), c.Params.Value("id"))

	mod, ok := config.Modules[srvname]
	if !ok {
		srvname = srvnameDefault
		mod, ok = config.Modules[srvname]
		if !ok {
			return
		}
	}

	c.urlActionPath = strings.Join(uris[1:], "/")

	dataAction, template, mat := c.filter(uris[1:], mod)
	if !mat {
		if uris[1] == "" {
			template = "index.tpl"
		} else {
			template = "404.tpl"
		}
	}

	lang := "en"
	if lang, ok := c.Data["LANG"]; ok {
		lang = strings.ToLower(lang.(string))
	}
	c.Data["LANG"] = api.LangHit(config.Languages, lang)

	if len(config.Languages) > 1 {
		c.Data["frontend_langs"] = config.Languages
	}

	// if session, err := c.Session.Instance(); err == nil {
	// 	c.Data["session"] = session
	// }

	c.Data["baseuri"] = "/" + srvname
	c.Data["http_request_path"] = reqpath
	c.Data["srvname"] = srvname
	c.Data["modname"] = mod.Meta.Name
	c.Data["sys_version_sign"] = config.SysVersionSign
	if c.us.IsLogin() {
		c.Data["s_user"] = c.us.UserName
	}

	drs := dataRenderOK

	if dataAction != "" {

		for _, action := range mod.Actions {

			if action.Name != dataAction {
				continue
			}

			for _, datax := range action.Datax {
				drs = c.dataRender(srvname, action.Name, datax)
				c.Data["__datax_table__"] = datax.Query.Table
			}

			break
		}
	}

	switch drs {
	case dataRenderOK:

		// render_start := time.Now()
		c.Render(mod.Meta.Name, template)

		// fmt.Println("render in-time", mod.Meta.Name, template, time.Since(render_start))

		c.RenderString(fmt.Sprintf("<!-- version %s, rt-time/db+render %d ms -->",
			config.Version, (time.Now().UnixNano()-start)/1e6))

		// fmt.Println("hookPosts", len(c.hookPosts))
		for _, fn := range c.hookPosts {
			fn()
		}

	case dataRenderNotFound:
		c.RenderError(404, "Page Not Found")
	}
}

func (c *Index) dataRender(srvname, action_name string, ad api.ActionData) int {

	mod, ok := config.Modules[srvname]
	if !ok {
		return dataRenderNotFound
	}

	qry := datax.NewQuery(mod.Meta.Name, ad.Query.Table)
	if ad.Query.Limit > 0 {
		qry.Limit(ad.Query.Limit)
	}

	if ad.Query.Order != "" {
		qry.Order(ad.Query.Order)
	}

	qry.Filter("status", 1)

	qry.Pager = ad.Pager

	switch ad.Type {

	case "node.list":

		for _, modNode := range mod.NodeModels {

			if ad.Query.Table != modNode.Meta.Name {
				continue
			}

			for _, term := range modNode.Terms {

				if termVal := c.Params.Value("term_" + term.Meta.Name); termVal != "" {

					switch term.Type {

					case api.TermTaxonomy:

						if idxs := datax.TermTaxonomyCacheIndexes(mod.Meta.Name, term.Meta.Name, termVal); len(idxs) > 1 {
							args := []interface{}{}
							for _, idx := range idxs {
								args = append(args, idx)
							}
							qry.Filter("term_"+term.Meta.Name+".in", args...)
						} else {
							qry.Filter("term_"+term.Meta.Name, termVal)
						}

						c.Data["term_"+term.Meta.Name] = termVal

					case api.TermTag:
						// TOPO
						qry.Filter("term_"+term.Meta.Name+".like", "%"+termVal+"%")
						c.Data["term_"+term.Meta.Name] = termVal
					}
				}
			}

			break
		}

		page := c.Params.IntValue("page")
		if page > 1 {
			qry.Offset(ad.Query.Limit * (page - 1))
		}

		if c.Params.Value("qry_text") != "" {
			qry.Filter("field_title.like", "%"+c.Params.Value("qry_text")+"%")
			c.Data["qry_text"] = c.Params.Value("qry_text")
		}

		var ls api.NodeList
		qryhash := qry.Hash()

		if ad.CacheTTL > 0 && (!c.us.IsLogin() || c.us.UserName != config.Config.AppInstance.Meta.User) {
			if rs := store.DataLocal.NewReader([]byte(qryhash)).Exec(); rs.OK() {
				rs.JsonDecode(&ls)
			}
		}

		if len(ls.Items) == 0 {

			if c.Params.Value("qry_text") != "" {
				ls = qry.NodeListSearch(c.Params.Value("qry_text"))
				if ls.Error != nil {
					ls = qry.NodeList([]string{}, []string{})
				}
			} else {
				ls = qry.NodeList([]string{}, []string{})
			}
			// fmt.Println("index node.list")
			if ad.CacheTTL > 0 && len(ls.Items) > 0 {
				c.hookPosts = append(
					c.hookPosts,
					func() {
						store.DataLocal.NewWriter([]byte(qryhash), nil).SetJsonValue(ls).SetTTL(ad.CacheTTL).Exec()
					},
				)
			}
		}

		c.Data[ad.Name] = ls

		if qry.Pager {
			pager := utils.NewPager(uint64(page),
				uint64(ls.Meta.TotalResults),
				uint64(ls.Meta.ItemsPerList),
				10)
			c.Data[ad.Name+"_pager"] = pager
		}

	case "node.entry":

		nodeId := c.Params.Value(ad.Name + "_id")
		if nodeId == "" {
			nodeId = c.Params.Value("id")
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
			if mv, ok := c.Data[action_name+"_nsr_"+nodeModel.Extensions.NodeRefer]; ok {
				nodeRefer = mv.(string)
			}
		}

		var (
			nodeExt = ""
		)

		if mod.Meta.Name == "core/gdoc" {
			if ad.Query.Table == "page" {
				if mat := gdocPathRX.FindAllStringSubmatch(c.urlActionPath, 1); len(mat) == 1 {
					nodeId = strings.ToLower(mat[0][2])
				}
			} else if ad.Query.Table == "doc" && api.NodeIdReg.MatchString(nodeId) {
				nodeExt = "html"
			}
		}
		if i := strings.LastIndex(nodeId, "."); i > 0 {
			nodeExt = nodeId[i+1:]
			nodeId = nodeId[:i]
		}

		if nodeExt == "html" {
			qry.Filter("id", nodeId)
		} else if staticImages.Has(nodeExt) {
			if mod.Meta.Name == "core/gdoc" && ad.Query.Table == "page" {

				if docId := datax.GdocNodeId(c.Params.Value("doc_entry_id")); docId != "" {

					localPath := datax.GdocLocalPath(docId)
					if localPath == "" {
						localPath = fmt.Sprintf("%s/var/vcs/%s", config.Prefix, docId)
					}
					if mat := gdocPathRX.FindAllStringSubmatch(c.urlActionPath, 1); len(mat) == 1 {
						localPath += "/" + mat[0][2]
					}
					localPath = filepath.Clean(localPath)
					s2Server(c.Controller, c.urlActionPath, localPath)
				}
			}
			return dataRenderSkip

		} else if nodeModel.Extensions.Permalink != "" {
			if nodeModel.Extensions.NodeRefer != "" && nodeRefer == "" {
				return dataRenderNotFound
			}
			qry.Filter("ext_permalink_idx", idhash.HashToHexString([]byte(nodeRefer+nodeId), 12))
		} else {
			return dataRenderNotFound
		}

		var entry api.Node
		qryhash := qry.Hash()
		if ad.CacheTTL > 0 && (!c.us.IsLogin() || c.us.UserName != config.Config.AppInstance.Meta.User) {
			if rs := store.DataLocal.NewReader([]byte(qryhash)).Exec(); rs.OK() {
				rs.JsonDecode(&entry)
			}
		}

		if entry.ID == "" {
			entry = qry.NodeEntry()
			if ad.CacheTTL > 0 && entry.Title != "" {
				c.hookPosts = append(
					c.hookPosts,
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

			if ips := strings.Split(c.Request.RemoteAddr, ":"); len(ips) > 1 {

				table := fmt.Sprintf("hpn_%s_%s", idhash.HashToHexString([]byte(mod.Meta.Name), 12), ad.Query.Table)
				store.DataLocal.NewWriter([]byte("access_counter/"+table+"/"+ips[0]+"/"+entry.ID), []byte("1")).Exec()
			}
		}

		if nodeModel.Extensions.NodeSubRefer != "" {
			// fmt.Println("setting", action_name, ad.Query.Table, nodeModel.Extensions.NodeSubRefer, "_id", entry.ID)
			c.Data[action_name+"_nsr_"+ad.Query.Table] = entry.ID
		}

		if entry.Title != "" {
			c.Data["__html_head_title__"] = datax.StringSub(datax.TextHtml2Str(entry.Title), 0, 50)
		}

		c.Data[ad.Name] = entry

	case "term.list":

		var ls api.TermList
		qryhash := qry.Hash()
		if ad.CacheTTL > 0 {
			if rs := store.DataLocal.NewReader([]byte(qryhash)).Exec(); rs.OK() {
				rs.JsonDecode(&ls)
			}
		}

		if len(ls.Items) == 0 {
			ls = qry.TermList()
			if ad.CacheTTL > 0 && len(ls.Items) > 0 {
				store.DataLocal.NewWriter([]byte(qryhash), nil).SetJsonValue(ls).SetTTL(ad.CacheTTL).Exec()
			}
		}

		c.Data[ad.Name] = ls

		if qry.Pager {
			c.Data[ad.Name+"_pager"] = utils.NewPager(0,
				uint64(ls.Meta.TotalResults),
				uint64(ls.Meta.ItemsPerList),
				10)
		}

	case "term.entry":

		var entry api.Term
		qryhash := qry.Hash()

		if ad.CacheTTL > 0 {
			if rs := store.DataLocal.NewReader([]byte(qryhash)).Exec(); rs.OK() {
				rs.JsonDecode(&entry)
			}
		}

		if entry.Title == "" {
			entry = qry.TermEntry()
			if ad.CacheTTL > 0 && entry.Title != "" {
				store.DataLocal.NewWriter([]byte(qryhash), nil).SetJsonValue(entry).SetTTL(ad.CacheTTL).Exec()
			}
		}

		c.Data[ad.Name] = entry
	}

	return dataRenderOK
}
