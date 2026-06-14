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

package v1

import (
	"encoding/base64"
	"errors"
	"html/template"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/hooto/httpsrv"
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/hooto/iam/v2/pkg/iamserver"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/lessos/lessgo/utils"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/config"
	"github.com/hooto/hpress/modset"
)

type ModSetFs struct {
	*httpsrv.Controller
	us iamserver.UserSession
}

func (c *ModSetFs) Init() int {

	//
	c.us = iamserver.AppVerifier.Session(c.Request.Request)

	if _, err := c.us.RequireAuth(); err != nil {
		c.Response.Out.WriteHeader(401)
		c.RenderJson(types.NewTypeErrorMeta(iamapi.ErrCodeUnauthorized, "Unauthorized"))
		return 1
	}

	return 0
}

func (c ModSetFs) RenameAction() {

	var (
		rsp api.FsFile
		req api.FsFile
	)

	defer c.RenderJson(&rsp)

	if !c.us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return
	}

	if err := c.Request.JsonDecode(&req); err != nil {
		rsp.Error = &types.ErrorMeta{"400", "Bad Request"}
		return
	}

	modname, err := modset.ModNameFilter(c.Params.Value("modname"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{"403", "Forbidden"}
		return
	}

	path := filepath.Clean(req.Path)
	path = filepath.Clean(config.Config.ModuleDir + "/" + modname + "/" + path)

	pathset := filepath.Clean(req.PathSet)
	pathset = filepath.Clean(config.Config.ModuleDir + "/" + modname + "/" + pathset)

	dir := filepath.Dir(pathset)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fsMakeDir(dir, config.User.Uid, config.User.Gid, 0750)
	}

	if err := os.Rename(path, pathset); err != nil {
		rsp.Error = &types.ErrorMeta{"500", err.Error()}
		return
	}

	rsp.Kind = "FsFile"
}

func (c ModSetFs) DelAction() {

	var (
		rsp api.FsFile
		req api.FsFile
	)

	defer c.RenderJson(&rsp)

	if !c.us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return
	}

	if err := c.Request.JsonDecode(&req); err != nil {
		rsp.Error = &types.ErrorMeta{"400", "Bad Request"}
		return
	}

	//
	modname, err := modset.ModNameFilter(c.Params.Value("modname"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{"403", "Forbidden"}
		return
	}

	path := filepath.Clean(req.Path)

	path = filepath.Clean(config.Config.ModuleDir + "/" + modname + "/" + path)

	if err := os.Remove(path); err != nil {
		rsp.Error = &types.ErrorMeta{"500", err.Error()}
		return
	}

	if path[len(path)-4:] == ".tpl" {
		config.SpecRefresh(modname)
	}

	rsp.Kind = "FsFile"
}

func (c ModSetFs) PutAction() {

	var (
		rsp api.FsFile
		req api.FsFile
		err error
	)

	defer c.RenderJson(&rsp)

	if !c.us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return
	}

	if err := c.Request.JsonDecode(&req); err != nil {
		rsp.Error = &types.ErrorMeta{"400", "Bad Request"}
		return
	}

	modname, err := modset.ModNameFilter(c.Params.Value("modname"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{"403", "Forbidden"}
		return
	}

	path := filepath.Clean(req.Path)

	path = filepath.Clean(config.Config.ModuleDir + "/" + modname + "/" + path)

	var body []byte
	if req.Encode == "base64" {

		dataurl := strings.SplitAfter(req.Body, ";base64,")
		if len(dataurl) != 2 {
			rsp.Error = &types.ErrorMeta{"400", "Bad Request"}
			return
		}

		body, err = base64.StdEncoding.DecodeString(dataurl[1])
		if err != nil {
			rsp.Error = &types.ErrorMeta{"400", err.Error()}
			return
		}

	} else if req.Encode == "text" || req.Encode == "jm" {
		body = []byte(req.Body)
	} else {
		rsp.Error = &types.ErrorMeta{"400", "Bad Request"}
		return
	}

	projfp := filepath.Clean(path)

	if req.Encode == "jm" {

		var jsPrev, jsAppend map[string]interface{}

		if err := json.Decode([]byte(body), &jsAppend); err != nil {
			rsp.Error = &types.ErrorMeta{"400", err.Error()}
			return
		}

		file, _, err := fsFileGetRead(projfp)
		if err != nil {
			rsp.Error = &types.ErrorMeta{"500", err.Error()}
			return
		}

		if err = json.Decode([]byte(file.Body), &jsPrev); err != nil {
			rsp.Error = &types.ErrorMeta{"400", err.Error()}
			return
		}

		jsMerged := utils.JsonMerge(jsPrev, jsAppend)
		// fmt.Println(jsPrev, "\n\n", jsAppend, "\n\n", jsMerged)

		body, _ = json.Encode(jsMerged, "  ")
	}

	if err := fsFilePutWrite(projfp, body); err != nil {
		rsp.Error = &types.ErrorMeta{"500", err.Error()}
		return
	}

	if path[len(path)-4:] == ".tpl" {

		loaderTemplate := template.New("templateName").Funcs(httpsrv.TemplateFuncs)

		if _, err := loaderTemplate.ParseFiles(path); err != nil {

			rsp.Error = &types.ErrorMeta{"400", err.Error()}
			return
		}

		config.SpecRefresh(modname)
	}

	rsp.Kind = "FsFile"
}

func fsFilePutWrite(path string, body []byte) error {

	defer func() {
		if r := recover(); r != nil {
			//
		}
	}()

	dir := filepath.Dir(path)

	if st, err := os.Stat(dir); os.IsNotExist(err) {

		fsMakeDir(dir, config.User.Uid, config.User.Gid, 0750)

	} else if !st.IsDir() {
		return errors.New("Can not create directory, File exists")
	}

	fp, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer fp.Close()

	fp.Seek(0, 0)
	fp.Truncate(int64(len(body))) // TODO
	if _, err = fp.Write(body); err != nil {
		return err
	}

	iUid, _ := strconv.Atoi(config.User.Uid)
	iGid, _ := strconv.Atoi(config.User.Gid)

	os.Chmod(path, 0644)
	os.Chown(path, iUid, iGid)

	return nil
}

func fsMakeDir(path, uuid, ugid string, mode os.FileMode) error {

	if _, err := os.Stat(path); err == nil {
		return nil
	}

	iUid, _ := strconv.Atoi(uuid)
	iGid, _ := strconv.Atoi(ugid)

	paths := strings.Split(strings.Trim(path, "/"), "/")

	path = ""

	for _, v := range paths {

		path += "/" + v

		if _, err := os.Stat(path); err == nil {
			continue
		}

		if err := os.Mkdir(path, mode); err != nil {
			return err
		}

		os.Chmod(path, mode)
		os.Chown(path, iUid, iGid)
	}

	return nil
}

func (c ModSetFs) ListAction() {

	var rsp api.FsFileList

	defer c.RenderJson(&rsp)

	if !c.us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return
	}

	modname, err := modset.ModNameFilter(c.Params.Value("modname"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{"403", "Forbidden"}
		return
	}

	path := filepath.Clean(c.Params.Value("path"))

	projfp := filepath.Clean(config.Config.ModuleDir + "/" + modname + "/" + path)

	rsp.Path = path
	rsp.Items = fsDirList(projfp, "", false)

	rsp.Kind = "FsFileList"
}

func fsDirList(path, ppath string, subdir bool) []api.FsFile {

	var ret []api.FsFile

	globpath := path
	if !strings.Contains(globpath, "*") {
		globpath += "/*"
	}

	rs, err := filepath.Glob(globpath)

	if err != nil {
		return ret
	}

	if len(ppath) > 0 {
		ppath += "/"
	}

	for _, v := range rs {

		var file api.FsFile
		// file.Path = v

		st, err := os.Stat(v)
		if os.IsNotExist(err) {
			continue
		}

		file.Name = ppath + st.Name()
		file.Size = st.Size()
		file.IsDir = st.IsDir()
		file.ModTime = st.ModTime().Format("2006-01-02T15:04:05Z07:00")

		if !st.IsDir() {
			file.Mime = fsFileMime(v)
		} else if subdir {
			subret := fsDirList(path+"/"+st.Name(), ppath+st.Name(), subdir)
			for _, v := range subret {
				ret = append(ret, v)
			}
		}

		ret = append(ret, file)
	}

	return ret
}

func fsFileMime(v string) string {

	// TODO
	//  ... add more extension types
	ctype := mime.TypeByExtension(filepath.Ext(v))

	if ctype == "" {
		fp, err := os.Open(v)
		if err == nil {

			defer fp.Close()

			if ctn, err := ioutil.ReadAll(fp); err == nil {
				ctype = http.DetectContentType(ctn)
			}
		}
	}

	ctypes := strings.Split(ctype, ";")
	if len(ctypes) > 0 {
		ctype = ctypes[0]
	}

	return ctype
}

func (c ModSetFs) GetAction() {

	var rsp api.FsFile
	var err error

	defer c.RenderJson(&rsp)

	if !c.us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return
	}

	modname, err := modset.ModNameFilter(c.Params.Value("modname"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{"403", "Forbidden"}
		return
	}

	path := filepath.Clean(c.Params.Value("path"))

	path = filepath.Clean(config.Config.ModuleDir + "/" + modname + "/" + path)

	rsp, _, err = fsFileGetRead(path)
	if err != nil {
		rsp.Error = &types.ErrorMeta{"400", err.Error()}
		return
	}

	rsp.Kind = "FsFile"
}

func fsFileGetRead(path string) (api.FsFile, int, error) {

	var file api.FsFile
	file.Path = path

	reg, _ := regexp.Compile("/+")
	path = "/" + strings.Trim(reg.ReplaceAllString(path, "/"), "/")

	st, err := os.Stat(path)
	if err != nil || os.IsNotExist(err) {
		return file, 404, errors.New("File Not Found")
	}
	file.Size = st.Size()

	if st.Size() > (2 * 1024 * 1024) {
		return file, 413, errors.New("File size is too large") // Request Entity Too Large
	}

	fp, err := os.OpenFile(path, os.O_RDWR, 0754)
	if err != nil {
		return file, 500, errors.New("File Can Not Open")
	}
	defer fp.Close()

	ctn, err := ioutil.ReadAll(fp)
	if err != nil {
		return file, 500, errors.New("File Can Not Readable")
	}
	file.Body = string(ctn)

	// TODO
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		ctype = http.DetectContentType(ctn)
	}
	ctypes := strings.Split(ctype, ";")
	if len(ctypes) > 0 {
		ctype = ctypes[0]
	}
	file.Mime = ctype

	return file, 200, nil
}
