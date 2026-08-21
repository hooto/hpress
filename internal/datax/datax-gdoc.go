// Copyright 2019 Eryx <evorui аt gmail dοt com>, All rights reserved.
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
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"log/slog"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lessos/lessgo/crypto/idhash"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/hpapi"
	"github.com/hooto/hpress/internal/store"
)

var (
	gdocMu         sync.Mutex
	gdocPending    = false
	gdocUpdated    = uint32(0)
	gdocRepoUrlReg = regexp.MustCompile("^https?://([0-9a-zA-Z.\\-_/]{1,100})\\.git$")
	gdocNameReg    = regexp.MustCompile("^[0-9a-zA-Z.\\-_/]{1,100}$")
	gdocLangs      = types.ArrayString([]string{
		"en",
		"zh", "zh-cn", "zh-hk", "zh-tw",
	}) // TODO
	gdocAttrsJS = `[{"key":"format", "value":"md"}]`
	gdocVerDef  = "0000"
)

var (
	gdocNodeReferToPermalinkName = map[string]string{}
	gdocPermalinkNameToNodeRefer = map[string]string{}
)

func gdocNodePermalinkNameSet(nodeId, name string) {
	if name == "" || nodeId == name {
		return
	}
	if pn, ok := gdocNodeReferToPermalinkName[nodeId]; !ok || name != pn {
		gdocNodeReferToPermalinkName[nodeId] = name
		gdocPermalinkNameToNodeRefer[name] = nodeId
	}
}

func gdocNodePermalinkName(nodeId string) string {
	if pn, ok := gdocNodeReferToPermalinkName[nodeId]; ok {
		return pn
	}
	return nodeId
}

func GdocNodeId(name string) string {
	if id, ok := gdocPermalinkNameToNodeRefer[name]; ok {
		return id
	}
	return name
}

func gdocTextFilter(txt string) string {
	txt = strings.TrimSpace(txt)
	if len(txt) > 8 && txt[:3] == "---" {
		if n := strings.Index(txt[3:], "---"); n > 0 {
			txt = strings.TrimSpace(txt[n+6:])
		}
	}
	return txt
}

func gdocTextSummaryFilter(txt string) string {
	if len(txt) > 10 && txt[:2] == "# " {
		if n := strings.Index(txt, "\n"); n > 0 && n < 20 {
			if line1 := strings.ToLower(txt[2:n]); line1 == "summary" {
				txt = strings.TrimSpace(txt[n+1:])
			}
		}
	}
	return txt
}

func gdocNameLangHit(name string) (string, string, bool) {
	if n := strings.LastIndex(name, "."); n > 0 && (n+2) < len(name) {
		if gdocLangs.Has(name[n+1:]) {
			return name[:n], name[n+1:], true
		}
	}
	return name, "", false
}

func gdocLangExts(sets map[string]map[string]string, name string) string {
	if lx, ok := sets[name]; ok {

		var langs hpapi.NodeFieldLangs
		for l, v := range lx {
			langs.Items.Set(l, v)
		}
		if len(langs.Items) > 0 {
			langJs, _ := json.Encode(langs, "")
			return string(langJs)
		}
	}
	return ""
}

func gdocRefresh() {

	var (
		tn        = uint32(time.Now().Unix())
		timeRange = uint32(10) // DEBUG
	)

	if gdocUpdated+600 > tn {
		return
	}

	gdocMu.Lock()
	if gdocPending {
		gdocMu.Unlock()
		return
	}
	gdocPending = true
	gdocMu.Unlock()
	defer func() {
		gdocPending = false
	}()

	gdocUpdated = tn

	spec := config.SpecGet("core/gdoc")
	if spec == nil {
		return
	}

	var (
		offset = int64(0)
		limit  = int64(10)
		query  = NewQuery("core/gdoc", "doc")
	)
	query.Limit(limit)

	for {
		query.Offset(offset)

		ls := query.NodeList(nil, nil)
		for _, v := range ls.Items {

			if !gdocRepoUrlReg.MatchString(v.Field("repo_url").Value) {
				continue
			}

			if v.Updated+timeRange > tn {
				continue
			}

			// dir, _ := filepath.Abs(fmt.Sprintf("%s/var/vcs/test", config.Prefix))
			dir, _ := filepath.Abs(fmt.Sprintf("%s/var/vcs/%s/%s",
				config.Prefix, v.ID, v.Field("repo_dir").Value))

			repo := &VcsRepoItem{
				Url:      v.Field("repo_url").Value,
				Branch:   v.Field("repo_branch").Value,
				Dir:      fmt.Sprintf("%s/var/vcs/%s", config.Prefix, v.ID),
				AuthUser: v.Field("repo_auth_user").Value,
				AuthPass: v.Field("repo_auth_key").Value,
			}
			if repo.Branch == "" {
				repo.Branch = "master"
			}

			gdocNodePermalinkNameSet(v.ID, v.ExtPermalinkName)

			ver, err := vcsAction(repo)
			if err != nil {
				slog.Warn(fmt.Sprintf("vcs %s, err %s", v.ID, err.Error()))
				continue
			}

			if ver != gdocVerDef && ver == v.Field("repo_version").Value {
				slog.Debug(fmt.Sprintf("vcs %s, version %s, skip", v.ID, ver))
				continue
			}

			if err := gdocRefreshItem(v.ID, v.UserID, ver, dir); err != nil {
				slog.Warn(fmt.Sprintf("vcs %s, version %s, err %s", v.ID, ver, err.Error()))
			}
		}

		if len(ls.Items) < int(limit) {
			break
		}

		offset += limit
	}

	expGdocRefreshPath()
}

var (
	gdocLpMu       sync.RWMutex
	gdocLocalPaths = map[string]string{}
)

func GdocLocalPath(docId string) string {
	gdocLpMu.RLock()
	defer gdocLpMu.RUnlock()
	p, ok := gdocLocalPaths[docId]
	if ok {
		return p
	}
	return ""
}

func expGdocRefreshPath() {

	gdocLpMu.Lock()
	defer gdocLpMu.Unlock()

	ver := "00000000"

	for _, dir := range config.Config.ExpGdocPaths {

		dir = filepath.Clean(dir)

		docId := idhash.HashToHexString([]byte(dir), 12)

		if err := gdocRefreshItem(docId, "sysadmin", ver, dir); err != nil {
			slog.Warn(fmt.Sprintf("vcs %s, version %s, err %s", docId, ver, err.Error()))
		} else {
			gdocLocalPaths[docId] = dir
		}
	}
}

func gdocRefreshItem(docId, userId, ver, dir string) error {

	var (
		tableDoc = hpapi.NodeTableName("core/gdoc", "doc")
		table    = hpapi.NodeTableName("core/gdoc", "page")
	)

	slog.Debug(fmt.Sprintf("vcs %s, version %s", docId, ver))

	args := []string{
		dir,
		"-type",
		"f",
		"-name",
		"*.md",
	}

	out, err := exec.Command("find", args...).Output()
	if err != nil {
		return err
	}

	var (
		files   = strings.Split(strings.TrimSpace(string(out)), "\n")
		summary = ""
		readme  = ""
		langs   = map[string]map[string]string{}
	)

	sort.Slice(files, func(i, j int) bool {
		return (strings.Compare(files[i], files[j]) == 1)
	})

	for _, path := range files {

		subPath := strings.TrimSpace(path[len(dir)+1:])
		if !gdocNameReg.MatchString(subPath) {
			continue
		}
		subPath = strings.ToLower(subPath)
		if len(subPath) < 4 {
			continue
		}
		subPath = subPath[:len(subPath)-3]

		var (
			nodeId = idhash.HashToHexString([]byte(docId+subPath), 12)
		)

		bs, err := ioutil.ReadFile(path)
		if err != nil {
			slog.Warn(fmt.Sprintf("doc %s, path walk err %s", docId, err.Error()))
			continue
		}
		txt := gdocTextFilter(string(bs))

		if subPath == "readme" {
			readme = txt
		} else if subPath == "summary" {
			summary = gdocTextSummaryFilter(txt)
			continue
		}

		if pSubPath, lang, hit := gdocNameLangHit(subPath); hit {
			lx, ok := langs[pSubPath]
			if !ok {
				lx = map[string]string{}
			}
			if pSubPath == "summary" {
				lx[lang] = gdocTextSummaryFilter(txt)
			} else {
				lx[lang] = txt
			}
			langs[pSubPath] = lx
			continue
		}

		q := store.Data.NewQueryer().
			Select("id,field_repo_sumcheck").
			From(table)

		q.Where().And("id", nodeId)

		rs := store.Data.Query(q)
		if rs.Err() != nil {
			continue
		}

		var (
			langTxt  = gdocLangExts(langs, subPath)
			sumCheck = fmt.Sprintf("crc32:%d", crc32.ChecksumIEEE([]byte(txt+langTxt)))
		)

		sets := map[string]interface{}{
			"userid":              userId,
			"pid":                 docId,
			"title":               subPath,
			"field_title":         subPath,
			"field_content":       txt,
			"field_content_attrs": gdocAttrsJS,
			"field_repo_sumcheck": sumCheck,
			"ext_permalink_name":  subPath,
			"ext_permalink_idx":   nodeId,
			"ext_node_refer":      docId,
			"status":              1,
			"updated":             time.Now().Unix(),
		}
		if len(langTxt) > 2 {
			sets["field_content_langs"] = langTxt
		}

		if rs.NotFound() {
			sets["id"] = nodeId
			sets["created"] = sets["updated"]
			err = store.Data.Insert(table, sets).Err()
		} else {
			if rs.Field("field_repo_sumcheck").String() != sumCheck {
				fr := store.Data.NewFilter().And("id", nodeId)
				err = store.Data.Update(table, sets, fr).Err()
			} else {
				slog.Debug(fmt.Sprintf("doc %s, page %s, path %s, skip",
					docId, nodeId, subPath))
			}
		}

		if err != nil {
			slog.Info(fmt.Sprintf("doc %s, page %s, path %s, refreshed err %s",
				docId, nodeId, subPath, err.Error()))
		} else {
			slog.Debug(fmt.Sprintf("doc %s, page %s, path %s, refreshed %d",
				docId, nodeId, subPath, len(bs)))
		}
	}

	sets := map[string]interface{}{
		"field_preface":       readme,
		"field_preface_attrs": gdocAttrsJS,
		"field_content":       summary,
		"field_content_attrs": gdocAttrsJS,
		"field_repo_version":  ver,
		"updated":             time.Now().Unix(),
	}
	if langTxt := gdocLangExts(langs, "readme"); len(langTxt) > 2 {
		sets["field_preface_langs"] = langTxt
	}
	if langTxt := gdocLangExts(langs, "summary"); len(langTxt) > 2 {
		sets["field_content_langs"] = langTxt
	}

	q := store.Data.NewQueryer().Select("id").From(tableDoc)
	q.Where().And("id", docId)

	// store.Data.Delete(tableDoc, store.Data.NewFilter().And("id", docId))

	rs := store.Data.Query(q)

	if rs.NotFound() {

		sets["id"] = docId
		sets["created"] = time.Now().Unix()
		sets["userid"] = userId
		sets["title"] = docId
		sets["field_title"] = docId
		sets["status"] = 1
		sets["ext_permalink_name"] = docId

		err = store.Data.Insert(tableDoc, sets).Err()
	} else {
		err = store.Data.Update(tableDoc, sets, store.Data.NewFilter().And("id", docId)).Err()
	}

	if err != nil {
		return err
	}

	return nil
}
