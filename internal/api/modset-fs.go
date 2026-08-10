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

package api

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

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/lessos/lessgo/utils"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/hpapi"
	"github.com/hooto/hpress/internal/modset"
	"github.com/hooto/hpress/internal/web"
)

func ModSetFsRename(c fiber.Ctx) error {

	var (
		rsp hpapi.FsFile
		req hpapi.FsFile
	)

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	if err := web.Bind(c, &req); err != nil {
		rsp.Error = &types.ErrorMeta{"400", "Bad Request"}
		return nil
	}

	modname, err := modset.ModNameFilter(web.Param(c, "modname"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{"403", "Forbidden"}
		return nil
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
		return nil
	}

	rsp.Kind = "FsFile"

	return nil
}

func ModSetFsDel(c fiber.Ctx) error {

	var (
		rsp hpapi.FsFile
		req hpapi.FsFile
	)

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	if err := web.Bind(c, &req); err != nil {
		rsp.Error = &types.ErrorMeta{"400", "Bad Request"}
		return nil
	}

	//
	modname, err := modset.ModNameFilter(web.Param(c, "modname"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{"403", "Forbidden"}
		return nil
	}

	path := filepath.Clean(req.Path)

	path = filepath.Clean(config.Config.ModuleDir + "/" + modname + "/" + path)

	if err := os.Remove(path); err != nil {
		rsp.Error = &types.ErrorMeta{"500", err.Error()}
		return nil
	}

	if path[len(path)-4:] == ".tpl" {
		config.SpecRefresh(modname)
	}

	rsp.Kind = "FsFile"

	return nil
}

func ModSetFsPut(c fiber.Ctx) error {

	var (
		rsp hpapi.FsFile
		req hpapi.FsFile
		err error
	)

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	if err := web.Bind(c, &req); err != nil {
		rsp.Error = &types.ErrorMeta{"400", "Bad Request"}
		return nil
	}

	modname, err := modset.ModNameFilter(web.Param(c, "modname"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{"403", "Forbidden"}
		return nil
	}

	path := filepath.Clean(req.Path)

	path = filepath.Clean(config.Config.ModuleDir + "/" + modname + "/" + path)

	// Directory creation (New Folder). No body needed; just ensure the path
	// exists as a directory (recursive, chowned to the module user).
	if req.IsDir {
		if err := fsMakeDir(path, config.User.Uid, config.User.Gid, 0755); err != nil {
			rsp.Error = &types.ErrorMeta{"500", err.Error()}
			return nil
		}
		rsp.Kind = "FsFile"
		return nil
	}

	var body []byte
	if req.Encode == "base64" {

		dataurl := strings.SplitAfter(req.Body, ";base64,")
		if len(dataurl) != 2 {
			rsp.Error = &types.ErrorMeta{"400", "Bad Request"}
			return nil
		}

		body, err = base64.StdEncoding.DecodeString(dataurl[1])
		if err != nil {
			rsp.Error = &types.ErrorMeta{"400", err.Error()}
			return nil
		}

	} else if req.Encode == "text" || req.Encode == "jm" {
		body = []byte(req.Body)
	} else {
		rsp.Error = &types.ErrorMeta{"400", "Bad Request"}
		return nil
	}

	projfp := filepath.Clean(path)

	if req.Encode == "jm" {

		var jsPrev, jsAppend map[string]interface{}

		if err := json.Decode([]byte(body), &jsAppend); err != nil {
			rsp.Error = &types.ErrorMeta{"400", err.Error()}
			return nil
		}

		file, _, err := fsFileGetRead(projfp)
		if err != nil {
			rsp.Error = &types.ErrorMeta{"500", err.Error()}
			return nil
		}

		if err = json.Decode([]byte(file.Body), &jsPrev); err != nil {
			rsp.Error = &types.ErrorMeta{"400", err.Error()}
			return nil
		}

		jsMerged := utils.JsonMerge(jsPrev, jsAppend)
		// fmt.Println(jsPrev, "\n\n", jsAppend, "\n\n", jsMerged)

		body, _ = json.Encode(jsMerged, "  ")
	}

	if err := fsFilePutWrite(projfp, body); err != nil {
		rsp.Error = &types.ErrorMeta{"500", err.Error()}
		return nil
	}

	if path[len(path)-4:] == ".tpl" {

		loaderTemplate := template.New("templateName").Funcs(web.Templates.FuncMap())

		if _, err := loaderTemplate.ParseFiles(path); err != nil {

			rsp.Error = &types.ErrorMeta{"400", err.Error()}
			return nil
		}

		config.SpecRefresh(modname)
	}

	rsp.Kind = "FsFile"

	return nil
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

func fsMakeDir(path, uid, gid string, mode os.FileMode) error {

	if _, err := os.Stat(path); err == nil {
		return nil
	}

	iUid, _ := strconv.Atoi(uid)
	iGid, _ := strconv.Atoi(gid)

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

func ModSetFsList(c fiber.Ctx) error {

	var rsp hpapi.FsFileList

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	modname, err := modset.ModNameFilter(web.Param(c, "modname"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{"403", "Forbidden"}
		return nil
	}

	path := filepath.Clean(web.Param(c, "path"))

	projfp := filepath.Clean(config.Config.ModuleDir + "/" + modname + "/" + path)

	rsp.Path = path
	rsp.Items = fsDirList(projfp, "", false)

	rsp.Kind = "FsFileList"

	return nil
}

func fsDirList(path, ppath string, subdir bool) []hpapi.FsFile {

	var ret []hpapi.FsFile

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

		var file hpapi.FsFile
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

func ModSetFsGet(c fiber.Ctx) error {

	var rsp hpapi.FsFile
	var err error

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	modname, err := modset.ModNameFilter(web.Param(c, "modname"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{"403", "Forbidden"}
		return nil
	}

	path := filepath.Clean(web.Param(c, "path"))

	path = filepath.Clean(config.Config.ModuleDir + "/" + modname + "/" + path)

	rsp, _, err = fsFileGetRead(path)
	if err != nil {
		rsp.Error = &types.ErrorMeta{"400", err.Error()}
		return nil
	}

	rsp.Kind = "FsFile"

	return nil
}

func fsFileGetRead(path string) (hpapi.FsFile, int, error) {

	var file hpapi.FsFile
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
