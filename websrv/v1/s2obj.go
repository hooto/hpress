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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hooto/httpsrv"
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/hooto/iam/v2/pkg/iamserver"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/lessos/lessgo/utils"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/config"
)

var (
	s2_path_reg    = regexp.MustCompile("^[0-9a-zA-Z_\\-\\.\\/]{1,100}$")
	s2_bucket_deft = "/deft"
)

func path_filter(path string) (string, error) {

	path = filepath.Clean(strings.Replace(strings.TrimSpace(path), " ", "-", -1))
	if !s2_path_reg.MatchString(path) {
		return path, fmt.Errorf("Invalid File Name")
	}

	if !strings.HasPrefix(path, s2_bucket_deft) ||
		(len(path) > len(s2_bucket_deft) && path[len(s2_bucket_deft)] != '/') {
		return "", errors.New("Invalid Bucket Name")
	}

	return path, nil
}

func abs_path(path string) string {
	return filepath.Clean(config.Prefix + "/var/storage/" + path)
}

type S2Obj struct {
	*httpsrv.Controller
	us iamserver.UserSession
}

func (c *S2Obj) Init() int {

	//
	c.us = iamserver.AppVerifier.Session(c.Request.Request)

	if _, err := c.us.RequireAuth(); err != nil {
		c.Response.Out.WriteHeader(401)
		c.RenderJson(types.NewTypeErrorMeta(iamapi.ErrCodeUnauthorized, "Unauthorized"))
		return 1
	}

	return 0
}

func (c S2Obj) RenameAction() {

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

	path, err := path_filter(req.Path)
	if err != nil {
		rsp.Error = &types.ErrorMeta{"400", err.Error()}
		return
	}

	pathset, err := path_filter(req.PathSet)
	if err != nil {
		rsp.Error = &types.ErrorMeta{"400", err.Error()}
		return
	}

	path = abs_path(path)
	pathset = abs_path(pathset)

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

func (c S2Obj) DelAction() {

	var (
		rsp api.FsFile
	)

	defer c.RenderJson(&rsp)

	if !c.us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return
	}

	//
	path, err := path_filter(c.Params.Value("path"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{"400", err.Error()}
		return
	}
	path = abs_path(path)

	if err := os.Remove(path); err != nil {
		rsp.Error = &types.ErrorMeta{"500", err.Error()}
		return
	}

	rsp.Kind = "FsFile"
}

func (c S2Obj) PutAction() {

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

	path, err := path_filter(req.Path)
	if err != nil {
		rsp.Error = &types.ErrorMeta{"400", err.Error()}
		return
	}

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

	path = abs_path(path)

	if req.Encode == "jm" {

		var jsPrev, jsAppend map[string]interface{}

		err := json.Decode([]byte(body), &jsAppend)
		if err != nil {
			rsp.Error = &types.ErrorMeta{"400", err.Error()}
			return
		}

		file, _, err := fsFileGetRead(path)
		if err != nil {
			rsp.Error = &types.ErrorMeta{"500", err.Error()}
			return
		}

		err = json.Decode([]byte(file.Body), &jsPrev)
		if err != nil {
			rsp.Error = &types.ErrorMeta{"400", err.Error()}
			return
		}

		jsMerged := utils.JsonMerge(jsPrev, jsAppend)
		// fmt.Println(jsPrev, "\n\n", jsAppend, "\n\n", jsMerged)

		body, _ = json.Encode(jsMerged, "")
	}

	if err := fsFilePutWrite(path, body); err != nil {
		rsp.Error = &types.ErrorMeta{"500", err.Error()}
		return
	}

	rsp.Kind = "FsFile"
}

func (c S2Obj) ListAction() {

	var rsp api.FsFileList

	defer c.RenderJson(&rsp)

	if !c.us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return
	}

	path, err := path_filter(c.Params.Value("path"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{"400", err.Error()}
		return
	}

	rsp.Path = path
	rsp.Items = fsDirList(abs_path(path), "", false)

	relpath := strings.Replace(path, s2_bucket_deft, "", -1)

	for i := range rsp.Items {
		rsp.Items[i].SelfLink = config.SysConfigList.FetchString("storage_service_endpoint") +
			relpath + "/" + rsp.Items[i].Name
	}

	rsp.Kind = "FsFileList"
}
