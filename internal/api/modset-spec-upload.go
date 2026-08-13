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
	"encoding/binary"
	stdjson "encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/sysinner/innerstack/v2/pkg/inapi"
	"github.com/ulikunitz/xz"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/hpapi"
	"github.com/hooto/hpress/internal/modset"
	"github.com/hooto/hpress/internal/web"
)

const (
	// ipkMagic is the 4-byte magic prefix of an innerstack v2 .ipk container.
	ipkMagic = "IPK1"
	// ipkMaxHeaderLen bounds the declared JSON header length (DoS guard);
	// real headers are a few KB, 1 MiB is a generous upper bound.
	ipkMaxHeaderLen = 1 << 20
	// ipkDecompressCap bounds the uncompressed payload read from a package
	// (decompression-bomb guard; the upload size cap only limits compressed size).
	ipkDecompressCap = 64 << 20
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
	if ext != ".ipk" {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, "Invalid file name extension, expected .ipk")
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

	_, cpr, err := parseIpkPackage(filedata)
	if err != nil {
		set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
		return nil
	}
	// Guard against decompression bombs: the upload size cap only bounds the
	// compressed bytes, so cap the uncompressed stream explicitly.
	cpr = io.LimitReader(cpr, ipkDecompressCap)

	tr := tar.NewReader(cpr)

	var (
		pkgName = strings.TrimSuffix(set.Name, filepath.Ext(set.Name))
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

		// Guard against path traversal (zip-slip): reject absolute or
		// parent-escaping names and confine every extraction under tmpdir.
		clean := filepath.Clean(hdr.Name)
		if filepath.IsAbs(clean) || clean == ".." ||
			strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
			set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument,
				"unsafe path in archive: "+hdr.Name)
			return nil
		}
		full := filepath.Join(tmpdir, clean)
		if !strings.HasPrefix(full, tmpdir+string(os.PathSeparator)) && full != tmpdir {
			set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument,
				"path escapes extract dir: "+hdr.Name)
			return nil
		}

		if hdr.FileInfo().IsDir() {
			if err := os.MkdirAll(full, 0755); err != nil {
				set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
				return nil
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
			return nil
		}

		fpo, err := os.OpenFile(full, os.O_RDWR|os.O_CREATE, os.FileMode(hdr.Mode))
		if err != nil {
			set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
			return nil
		}
		fpo.Seek(0, 0)
		fpo.Truncate(0)

		if _, err := io.Copy(fpo, tr); err != nil {
			fpo.Close()
			set.Error = types.NewErrorMeta(hpapi.ErrCodeBadArgument, err.Error())
			return nil
		}

		fpo.Close()

		files[clean] = hdr.Mode
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

// parseIpkPackage parses an innerstack v2 IPK1 container and returns the
// package header plus a reader over the (still-compressed) data block. The
// data block is a tar archive compressed per the header's release.compress.
// Mirrors the canonical reader in innerstack internal/cli/pkg_info.go.
func parseIpkPackage(filedata []byte) (*inapi.Package, io.Reader, error) {
	if len(filedata) < 8 {
		return nil, nil, fmt.Errorf("ipk: package too small")
	}
	if string(filedata[:4]) != ipkMagic {
		return nil, nil, fmt.Errorf("ipk: bad magic (expected %q)", ipkMagic)
	}

	headerLen := binary.LittleEndian.Uint32(filedata[4:8])
	if headerLen > ipkMaxHeaderLen {
		return nil, nil, fmt.Errorf("ipk: header too large (%d)", headerLen)
	}
	// Overflow-safe bounds check (a plain uint32 add can wrap past the cap).
	if uint64(8)+uint64(headerLen) > uint64(len(filedata)) {
		return nil, nil, fmt.Errorf("ipk: header length out of bounds")
	}
	headerEnd := 8 + int(headerLen)

	var pkg inapi.Package
	if err := stdjson.Unmarshal(filedata[8:headerEnd], &pkg); err != nil {
		return nil, nil, fmt.Errorf("ipk: bad header json: %w", err)
	}
	if pkg.Release == nil {
		return nil, nil, fmt.Errorf("ipk: header missing release")
	}

	dataBlock := filedata[headerEnd:]
	if len(dataBlock) == 0 {
		return nil, nil, fmt.Errorf("ipk: empty data block")
	}

	var cpr io.Reader
	switch pkg.Release.GetCompress() {
	case "xz":
		r, err := xz.NewReader(bytes.NewReader(dataBlock))
		if err != nil {
			return nil, nil, fmt.Errorf("ipk: xz init: %w", err)
		}
		cpr = r
	case "gzip":
		r, err := gzip.NewReader(bytes.NewReader(dataBlock))
		if err != nil {
			return nil, nil, fmt.Errorf("ipk: gzip init: %w", err)
		}
		cpr = r
	case "":
		cpr = bytes.NewReader(dataBlock) // uncompressed tar
	default:
		return nil, nil, fmt.Errorf("ipk: unsupported compression %q", pkg.Release.GetCompress())
	}

	return &pkg, cpr, nil
}
