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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/ulikunitz/xz"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/hpapi"
	"github.com/hooto/hpress/internal/modset"
	"github.com/hooto/hpress/internal/web"
)

var (
	specUploadSizeMax int64 = 8 * 1024 * 1024
)

func ModSetSpecUploadCommit(c fiber.Ctx) error {

	var set hpapi.SpecUploadCommit

	defer func() { _ = web.JSON(c, &set) }()

	err := web.Bind(c, &set)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Bad Argument "+err.Error())
		return nil
	}

	if set.Size > specUploadSizeMax {
		set.Error = types.NewErrorMeta("400",
			fmt.Sprintf("the max size of Package can not more than %d", specUploadSizeMax))
		return nil
	}

	if len(set.Name) < 10 {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Invalid Name")
		return nil
	}
	ext := filepath.Ext(set.Name)
	if ext != ".txz" && ext != ".tgz" {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Invalid file name extension")
		return nil
	}

	us := web.AuthSession(c)
	if us == nil || !us.Allow("", "editor.write") {
		set.Error = types.NewErrorMeta(iamapi.ErrCodeAccessDenied, "Access Denied")
		return nil
	}

	body64 := strings.SplitAfter(set.Data, ";base64,")
	if len(body64) != 2 {
		return nil
	}
	filedata, err := base64.StdEncoding.DecodeString(body64[1])
	if err != nil {
		set.Error = types.NewErrorMeta("400", "Package Not Found")
		return nil
	}

	if int64(len(filedata)) != set.Size {
		set.Error = types.NewErrorMeta("400", "Invalid Package Size")
		return nil
	}

	var cpr io.Reader

	switch ext {
	case ".txz":
		if cpr, err = xz.NewReader(bytes.NewReader(filedata)); err != nil {
			set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
			return nil
		}

	case ".tgz":
		if cpr, err = gzip.NewReader(bytes.NewReader(filedata)); err != nil {
			set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
			return nil
		}

	default:
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Invalid EXT")
		return nil
	}

	tr := tar.NewReader(cpr)
	if tr == nil {
		set.Error = types.NewErrorMeta("400", "Invalid Encoded Data")
		return nil
	}

	var (
		pkgName = set.Name[:len(set.Name)-4]
		tmpdir  = config.Prefix + "/var/tmp/" + pkgName
		files   = map[string]int64{}
	)

	for {

		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
			return nil
		}
		// fmt.Printf("Contents of %s\n", hdr.Name)

		if hdr.Name[len(hdr.Name)-1] == '/' {
			os.MkdirAll(tmpdir+"/"+hdr.Name, 0755)
			continue
		}

		// if strings.Index(hdr.Name, "/") > 0 {
		// 	os.MkdirAll(tmpdir+"/"+filepath.Dir(hdr.Name), 0755)
		// }

		fpo, err := os.OpenFile(tmpdir+"/"+hdr.Name, os.O_RDWR|os.O_CREATE, os.FileMode(hdr.Mode))
		if err != nil {
			set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
			return nil
		}
		fpo.Seek(0, 0)
		fpo.Truncate(0)

		if _, err := io.Copy(fpo, tr); err != nil {
			set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
			return nil
		}

		fpo.Close()

		files[hdr.Name] = hdr.Mode
	}

	var spec hpapi.Spec
	if err := json.DecodeFile(tmpdir+"/spec.json", &spec); err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	if !hpapi.NewSpecVersion(spec.Meta.Version).Valid() {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Invalid Version Format")
		return nil
	}

	//
	spec.Meta.Name, err = modset.ModNameFilter(spec.Meta.Name)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	spec.SrvName, err = hpapi.SrvNameFilter(spec.SrvName)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}

	if prev, err := modset.SpecFetch(spec.Meta.Name); err == nil {
		if hpapi.NewSpecVersion(prev.Meta.Version).Compare(hpapi.NewSpecVersion(spec.Meta.Version)) == 1 {
			set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Invalid Version")
			return nil
		}
	}

	specDir := config.Prefix + "/modules/" + spec.Meta.Name

	for path, fmode := range files {
		if err := specFileSync(tmpdir+"/"+path, specDir+"/"+path, os.FileMode(fmode)); err != nil {
			set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
			return nil
		}
	}

	modset.SpecSchemaSync(spec)

	set.Kind = "Spec"

	return nil
}

func specFileSync(src, dst string, mod os.FileMode) error {

	fpSrc, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fpSrc.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	fpDst, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, mod)
	if err != nil {
		return err
	}
	defer fpDst.Close()

	fpDst.Seek(0, 0)
	fpDst.Truncate(0)

	if _, err := io.Copy(fpDst, fpSrc); err != nil {
		return err
	}

	return nil
}
