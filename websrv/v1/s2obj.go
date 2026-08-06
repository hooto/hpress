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

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/lessos/lessgo/utils"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/config"
	"github.com/hooto/hpress/websrv/web"
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

func S2ObjRename(c fiber.Ctx) error {

	var (
		rsp api.FsFile
		req api.FsFile
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

	path, err := path_filter(req.Path)
	if err != nil {
		rsp.Error = &types.ErrorMeta{"400", err.Error()}
		return nil
	}

	pathset, err := path_filter(req.PathSet)
	if err != nil {
		rsp.Error = &types.ErrorMeta{"400", err.Error()}
		return nil
	}

	path = abs_path(path)
	pathset = abs_path(pathset)

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

func S2ObjDel(c fiber.Ctx) error {

	var (
		rsp api.FsFile
	)

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	//
	path, err := path_filter(web.Param(c, "path"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{"400", err.Error()}
		return nil
	}
	path = abs_path(path)

	if err := os.Remove(path); err != nil {
		rsp.Error = &types.ErrorMeta{"500", err.Error()}
		return nil
	}

	rsp.Kind = "FsFile"

	return nil
}

func S2ObjPut(c fiber.Ctx) error {

	var (
		rsp api.FsFile
		req api.FsFile
		err error
	)

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	// Publishing a node (editor.write) inherently includes uploading its
	// images, so s2-obj/put is gated on editor.write rather than sys.admin —
	// this lets an editor's access-key-signed request (clipper publish)
	// upload without needing admin rights.
	if us == nil || !us.Allow("", "editor.write") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	if err := web.Bind(c, &req); err != nil {
		rsp.Error = &types.ErrorMeta{"400", "Bad Request"}
		return nil
	}

	path, err := path_filter(req.Path)
	if err != nil {
		rsp.Error = &types.ErrorMeta{"400", err.Error()}
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

	path = abs_path(path)

	if req.Encode == "jm" {

		var jsPrev, jsAppend map[string]interface{}

		err := json.Decode([]byte(body), &jsAppend)
		if err != nil {
			rsp.Error = &types.ErrorMeta{"400", err.Error()}
			return nil
		}

		file, _, err := fsFileGetRead(path)
		if err != nil {
			rsp.Error = &types.ErrorMeta{"500", err.Error()}
			return nil
		}

		err = json.Decode([]byte(file.Body), &jsPrev)
		if err != nil {
			rsp.Error = &types.ErrorMeta{"400", err.Error()}
			return nil
		}

		jsMerged := utils.JsonMerge(jsPrev, jsAppend)
		// fmt.Println(jsPrev, "\n\n", jsAppend, "\n\n", jsMerged)

		body, _ = json.Encode(jsMerged, "")
	}

	if err := fsFilePutWrite(path, body); err != nil {
		rsp.Error = &types.ErrorMeta{"500", err.Error()}
		return nil
	}

	rsp.Kind = "FsFile"

	return nil
}

func S2ObjList(c fiber.Ctx) error {

	var rsp api.FsFileList

	defer func() { _ = web.JSON(c, &rsp) }()

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "sys.admin") {
		rsp.Error = &types.ErrorMeta{iamapi.ErrCodeAccessDenied, "Access Denied"}
		return nil
	}

	path, err := path_filter(web.Param(c, "path"))
	if err != nil {
		rsp.Error = &types.ErrorMeta{"400", err.Error()}
		return nil
	}

	rsp.Path = path
	rsp.Items = fsDirList(abs_path(path), "", false)

	relpath := strings.Replace(path, s2_bucket_deft, "", -1)

	for i := range rsp.Items {
		rsp.Items[i].SelfLink = config.SysConfigList.FetchString("storage_service_endpoint") +
			relpath + "/" + rsp.Items[i].Name
	}

	rsp.Kind = "FsFileList"

	return nil
}
