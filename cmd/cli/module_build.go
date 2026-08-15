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

// module-build: pack a module directory into an innerstack v2 .ipk package.
//
// Native IPK1 writer — the dual of the server-side parser in
// internal/api/modset-spec-upload.go (parseIpkPackage). Container layout:
//
//	"IPK1" | uint32le header-len | JSON header | xz-compressed tar data block
//
// The header repeats ipk.toml [metadata] plus a release block whose checksum
// is sha256 over the compressed data block. The tar mirrors the layout emitted
// by `innerstack pkg-build`: a "." root dir, an ".ipk/" meta dir holding
// metadata.json, then the [build].include entries in listed order (on-disk
// permission bits are preserved).

package main

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	stdjson "encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hooto/htoml4g/htoml"
	"github.com/spf13/cobra"
	"github.com/sysinner/innerstack/v2/pkg/inapi"
	"github.com/ulikunitz/xz"
)

const (
	ipkMagic    = "IPK1"
	ipkOS       = "all" // pure content package: no platform binding
	ipkArch     = "src"
	ipkCompress = "xz"
	// ipkDataMax bounds the uncompressed tar (decompression-bomb guard on the
	// reader side; here it keeps the in-memory buffer honest). Mirrors
	// ipkDecompressCap in internal/api/modset-spec-upload.go.
	ipkDataMax = 64 << 20
)

// pkgConfig is the on-disk ipk.toml: [metadata] feeds the package header,
// [build].include selects what lands in the data block.
type pkgConfig struct {
	Metadata inapi.PackageMetadata `toml:"metadata"`
	Build    pkgBuildSection       `toml:"build"`
}

type pkgBuildSection struct {
	Include []string `toml:"include"`
}

// pkgFile is one entry bound for the tar data block.
type pkgFile struct {
	name string // slash-separated archive path
	path string // filesystem source
	dir  bool
	mode fs.FileMode
	size int64
}

func newModuleBuildCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "module-build <dir>",
		Short: "Pack a module directory into an innerstack v2 .ipk package",
		Long: "Pack the module at <dir> per its ipk.toml into an IPK1 container " +
			"(<name>_<version>_all_src.ipk), byte-compatible with `innerstack " +
			"pkg-build` output and accepted by mod-set/spec-upload-commit.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, name, err := pkgBuild(args[0])
			if err != nil {
				return err
			}
			dst := out
			if dst == "" {
				dst = filepath.Join(args[0], name)
			}
			if err := os.WriteFile(dst, data, 0644); err != nil {
				return err
			}
			fmt.Printf("built %s (%d bytes)\n", dst, len(data))
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "output file (default <dir>/<name>_<version>_all_src.ipk)")
	return cmd
}

// pkgBuild packs the module at dir into an IPK1 container, returning the
// package bytes and its file name.
func pkgBuild(dir string) ([]byte, string, error) {

	cfgPath := filepath.Join(dir, "ipk.toml")
	if _, err := os.Stat(cfgPath); err != nil {
		return nil, "", fmt.Errorf("no ipk.toml in %s (run `hpress module-init` first)", dir)
	}
	var cfg pkgConfig
	if err := htoml.DecodeFromFile(cfgPath, &cfg); err != nil {
		return nil, "", fmt.Errorf("decode %s: %w", cfgPath, err)
	}
	if cfg.Metadata.Name == "" || cfg.Metadata.Version == "" {
		return nil, "", fmt.Errorf("ipk.toml needs [metadata].name and [metadata].version")
	}
	if len(cfg.Build.Include) == 0 {
		return nil, "", fmt.Errorf("ipk.toml needs a non-empty [build].include")
	}

	files, err := pkgCollect(dir, cfg.Build.Include)
	if err != nil {
		return nil, "", err
	}
	hasSpec := false
	for _, f := range files {
		if f.name == "spec.json" {
			hasSpec = true
		}
	}
	if !hasSpec {
		return nil, "", fmt.Errorf("package must include spec.json (check [build].include in %s)", cfgPath)
	}

	data, err := pkgDataBlock(files, &cfg.Metadata)
	if err != nil {
		return nil, "", err
	}

	checksum := sha256.Sum256(data)
	header, err := stdjson.Marshal(&inapi.Package{
		Metadata: &cfg.Metadata,
		Release: &inapi.PackageRelease{
			Version:  cfg.Metadata.Version,
			Os:       ipkOS,
			Arch:     ipkArch,
			Built:    time.Now().Unix(),
			Size:     int64(len(data)),
			Checksum: "sha256:" + hex.EncodeToString(checksum[:]),
			Compress: ipkCompress,
		},
	})
	if err != nil {
		return nil, "", err
	}

	// module packages are small (the upload cap is 8 MiB), so assembling the
	// container in memory is fine
	var buf bytes.Buffer
	buf.WriteString(ipkMagic)
	var ln [4]byte
	binary.LittleEndian.PutUint32(ln[:], uint32(len(header)))
	buf.Write(ln[:])
	buf.Write(header)
	buf.Write(data)

	return buf.Bytes(), fmt.Sprintf("%s_%s_%s_%s.ipk", cfg.Metadata.Name, cfg.Metadata.Version, ipkOS, ipkArch), nil
}

// pkgCollect expands the [build].include list (in order) into tar entries:
// a bare directory name is walked recursively (lexical order), a file is taken
// as-is. Archive names are slash-separated paths relative to the module dir.
func pkgCollect(dir string, includes []string) ([]pkgFile, error) {

	var files []pkgFile
	var total int64

	for _, inc := range includes {
		clean := filepath.Clean(inc)
		if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
			return nil, fmt.Errorf("unsafe include entry %q", inc)
		}
		full := filepath.Join(dir, clean)

		fi, err := os.Stat(full)
		if err != nil {
			return nil, fmt.Errorf("include %q: %w", inc, err)
		}

		if !fi.IsDir() {
			files = append(files, pkgFile{
				name: filepath.ToSlash(clean),
				path: full,
				mode: fi.Mode().Perm(),
				size: fi.Size(),
			})
			total += fi.Size()
			continue
		}

		err = filepath.WalkDir(full, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			base := filepath.Base(p)
			if base == ".DS_Store" { // macOS finder noise, never wanted in a package
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			rel, err := filepath.Rel(dir, p)
			if err != nil {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			pf := pkgFile{
				name: filepath.ToSlash(rel),
				path: p,
				dir:  d.IsDir(),
				mode: info.Mode().Perm(),
			}
			if !pf.dir {
				pf.size = info.Size()
				total += pf.size
			}
			files = append(files, pf)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("include %q: %w", inc, err)
		}
	}

	if total > ipkDataMax {
		return nil, fmt.Errorf("module content is %d bytes (max %d)", total, ipkDataMax)
	}
	return files, nil
}

// pkgDataBlock assembles the tar archive (mirroring the innerstack pkg-build
// layout) and returns it xz-compressed.
func pkgDataBlock(files []pkgFile, md *inapi.PackageMetadata) ([]byte, error) {

	now := time.Now()

	var raw bytes.Buffer
	tw := tar.NewWriter(&raw)

	writeEntry := func(name string, mode fs.FileMode, dir bool, size int64, body io.Reader) error {
		hdr := &tar.Header{
			Name:    name,
			Mode:    int64(mode.Perm()),
			Size:    size,
			ModTime: now,
		}
		if dir {
			hdr.Typeflag = tar.TypeDir
		} else {
			hdr.Typeflag = tar.TypeReg
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if body != nil {
			if _, err := io.Copy(tw, body); err != nil {
				return err
			}
		}
		return nil
	}

	// container meta, mirroring innerstack pkg-build
	meta, err := stdjson.Marshal(&inapi.Package{
		Metadata: md,
		Release:  &inapi.PackageRelease{Version: md.Version, Os: ipkOS, Arch: ipkArch},
	})
	if err != nil {
		return nil, err
	}
	if err := writeEntry(".", 0755, true, 0, nil); err != nil {
		return nil, err
	}
	if err := writeEntry(".ipk", 0755, true, 0, nil); err != nil {
		return nil, err
	}
	if err := writeEntry(".ipk/metadata.json", 0644, false, int64(len(meta)), bytes.NewReader(meta)); err != nil {
		return nil, err
	}

	// module content
	for _, f := range files {
		if f.dir {
			if err := writeEntry(f.name, f.mode, true, 0, nil); err != nil {
				return nil, err
			}
			continue
		}
		fp, err := os.Open(f.path)
		if err != nil {
			return nil, err
		}
		err = writeEntry(f.name, f.mode, false, f.size, fp)
		fp.Close()
		if err != nil {
			return nil, err
		}
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	var cbuf bytes.Buffer
	xw, err := xz.NewWriter(&cbuf)
	if err != nil {
		return nil, err
	}
	if _, err := xw.Write(raw.Bytes()); err != nil {
		return nil, err
	}
	if err := xw.Close(); err != nil {
		return nil, err
	}
	return cbuf.Bytes(), nil
}
