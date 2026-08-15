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

package main

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	stdjson "encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sysinner/innerstack/v2/pkg/inapi"
	"github.com/ulikunitz/xz"
)

const fixtureIpkToml = `[metadata]
name        = "test-demo"
version     = "0.1.0"
description = "test fixture"

[build]
include = ["spec.json", "views"]
`

const fixtureSpec = `{"meta":{"name":"demo/test","version":"0.1.0"},"srvname":"test","title":"Test"}`

// fixtureModule writes a minimal module tree into a fresh temp dir.
func fixtureModule(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"ipk.toml":        fixtureIpkToml,
		"spec.json":       fixtureSpec,
		"views/entry.tpl": "<html>{{.entry}}</html>",
		"views/list.tpl":  "<html>{{.list}}</html>",
		"views/.DS_Store": "finder noise inside an included dir",
		"notes/extra.md":  "packed when included",
	}
	for path, body := range files {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestModuleBuildRoundTrip builds a fixture module and re-parses the container
// the way the server (parseIpkPackage) and any innerstack consumer would:
// magic + LE32 header length + JSON header + xz(tar) data block, sha256
// checksum over the compressed block, and the mirrored tar layout.
func TestModuleBuildRoundTrip(t *testing.T) {
	dir := fixtureModule(t)

	data, name, err := pkgBuild(dir)
	if err != nil {
		t.Fatalf("pkgBuild: %v", err)
	}
	if want := "test-demo_0.1.0_all_src.ipk"; name != want {
		t.Errorf("package name = %q, want %q", name, want)
	}

	if string(data[:4]) != ipkMagic {
		t.Fatalf("magic = %q, want %q", data[:4], ipkMagic)
	}
	hlen := binary.LittleEndian.Uint32(data[4:8])
	if uint64(8)+uint64(hlen) > uint64(len(data)) {
		t.Fatalf("header length %d out of bounds (total %d)", hlen, len(data))
	}

	var pkg inapi.Package
	if err := stdjson.Unmarshal(data[8:8+hlen], &pkg); err != nil {
		t.Fatalf("decode header json: %v", err)
	}
	block := data[8+hlen:]

	if pkg.Metadata == nil || pkg.Metadata.Name != "test-demo" || pkg.Metadata.Version != "0.1.0" {
		t.Errorf("header metadata = %+v", pkg.Metadata)
	}
	if pkg.Release == nil {
		t.Fatal("header missing release")
	}
	if pkg.Release.Os != "all" || pkg.Release.Arch != "src" || pkg.Release.Compress != "xz" {
		t.Errorf("release os/arch/compress = %q/%q/%q", pkg.Release.Os, pkg.Release.Arch, pkg.Release.Compress)
	}
	if pkg.Release.Size != int64(len(block)) {
		t.Errorf("release size = %d, want %d", pkg.Release.Size, len(block))
	}
	sum := sha256.Sum256(block)
	if want := "sha256:" + hex.EncodeToString(sum[:]); pkg.Release.Checksum != want {
		t.Errorf("release checksum = %q, want %q", pkg.Release.Checksum, want)
	}

	zr, err := xz.NewReader(bytes.NewReader(block))
	if err != nil {
		t.Fatalf("xz init: %v", err)
	}
	tr := tar.NewReader(zr)

	var (
		got      []string
		specBody string
		metaBody string
	)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("tar next: %v", err)
		}
		got = append(got, hdr.Name)
		switch hdr.Name {
		case "spec.json":
			bs, err := io.ReadAll(tr)
			if err != nil {
				t.Fatal(err)
			}
			specBody = string(bs)
		case ".ipk/metadata.json":
			bs, err := io.ReadAll(tr)
			if err != nil {
				t.Fatal(err)
			}
			metaBody = string(bs)
		}
	}

	// innerstack-mirrored layout, include order, .DS_Store skipped
	want := []string{
		".", ".ipk", ".ipk/metadata.json",
		"spec.json",
		"views", "views/entry.tpl", "views/list.tpl",
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("tar entries = %v, want %v", got, want)
	}
	if specBody != fixtureSpec {
		t.Errorf("spec.json body = %q, want %q", specBody, fixtureSpec)
	}

	var meta inapi.Package
	if err := stdjson.Unmarshal([]byte(metaBody), &meta); err != nil {
		t.Fatalf("decode .ipk/metadata.json: %v", err)
	}
	if meta.Metadata == nil || meta.Metadata.Name != "test-demo" {
		t.Errorf("metadata.json = %+v", meta.Metadata)
	}
	if meta.Release == nil || meta.Release.Version != "0.1.0" ||
		meta.Release.Os != "all" || meta.Release.Arch != "src" {
		t.Errorf("metadata.json release = %+v", meta.Release)
	}
}

func TestModuleBuildErrors(t *testing.T) {
	tests := []struct {
		name string
		ipk  string
		want string
	}{
		{"no include", "[metadata]\nname=\"a\"\nversion=\"0.1.0\"\n", "include"},
		{"no metadata", "[build]\ninclude=[\"spec.json\"]\n", "name"},
		{"spec missing", "[metadata]\nname=\"a\"\nversion=\"0.1.0\"\n[build]\ninclude=[\"ipk.toml\"]\n", "spec.json"},
		{"escaping include", "[metadata]\nname=\"a\"\nversion=\"0.1.0\"\n[build]\ninclude=[\"../escape\"]\n", "unsafe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "ipk.toml"), []byte(tt.ipk), 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "spec.json"), []byte(fixtureSpec), 0644); err != nil {
				t.Fatal(err)
			}
			_, _, err := pkgBuild(dir)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("pkgBuild error = %v, want containing %q", err, tt.want)
			}
		})
	}

	t.Run("no ipk.toml", func(t *testing.T) {
		_, _, err := pkgBuild(t.TempDir())
		if err == nil || !strings.Contains(err.Error(), "module-init") {
			t.Errorf("pkgBuild error = %v, want module-init hint", err)
		}
	})
}
