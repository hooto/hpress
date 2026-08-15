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
	stdjson "encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hooto/htoml4g/htoml"

	"github.com/hooto/hpress/internal/hpapi"
)

func TestModuleNameDerive(t *testing.T) {
	tests := []struct {
		dir      string
		override string
		want     string
		wantErr  string
	}{
		{dir: "modules/demo/hello", want: "demo/hello"},
		{dir: "modules/ruilog/notebook", want: "ruilog/notebook"},
		{dir: "hello", want: "hello"},
		{dir: "../hello", want: "hello"},
		{dir: "modules/demo/UPPER", want: "demo/upper"},
		{dir: "some/deep/path/ns/name", want: "ns/name"},
		{dir: "modules/demo/hello", override: "other/name", want: "other/name"},
		{dir: "whatever", override: "demo/hello", want: "demo/hello"},
		{dir: "ok1", override: "A/B", want: "a/b"}, // override is lowercased
		{dir: "", wantErr: "derive"},
		{dir: "..", wantErr: "derive"},
		{dir: "ab", wantErr: "invalid module name"},                  // under 3 chars
		{dir: "ok1", override: "ab", wantErr: "invalid module name"}, // override under 3 chars
		{dir: "ok1", override: "has space", wantErr: "invalid module name"},
	}
	for _, tt := range tests {
		t.Run(tt.dir+"|"+tt.override, func(t *testing.T) {
			got, err := moduleNameDerive(tt.dir, tt.override)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("moduleNameDerive(%q, %q) error = %v, want containing %q",
						tt.dir, tt.override, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("moduleNameDerive(%q, %q): %v", tt.dir, tt.override, err)
			}
			if got != tt.want {
				t.Errorf("moduleNameDerive(%q, %q) = %q, want %q", tt.dir, tt.override, got, tt.want)
			}
		})
	}
}

// TestRunModuleInit scaffolds a module and validates every generated file:
// spec.json decodes into hpapi.Spec with a server-valid version, ipk.toml is
// pkgBuild-compatible, and the views keep their hpress template syntax intact.
func TestRunModuleInit(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "demo", "hello")
	if err := runModuleInit(dir, ""); err != nil {
		t.Fatalf("runModuleInit: %v", err)
	}

	for _, path := range []string{
		"spec.json",
		"ipk.toml",
		"views/entry.tpl",
		"views/list.tpl",
		"views/term/categories.tpl",
	} {
		if _, err := os.Stat(filepath.Join(dir, path)); err != nil {
			t.Errorf("missing scaffolded file %s: %v", path, err)
		}
	}

	// spec.json: decodable, correctly named, server-valid version
	specBs, err := os.ReadFile(filepath.Join(dir, "spec.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec hpapi.Spec
	if err := stdjson.Unmarshal(specBs, &spec); err != nil {
		t.Fatalf("decode spec.json: %v", err)
	}
	if spec.Meta.Name != "demo/hello" {
		t.Errorf("spec meta.name = %q, want demo/hello", spec.Meta.Name)
	}
	if spec.SrvName != "hello" {
		t.Errorf("spec srvname = %q, want hello", spec.SrvName)
	}
	if !hpapi.NewSpecVersion(spec.Meta.Version).Valid() {
		t.Errorf("spec version %q is not server-valid", spec.Meta.Version)
	}
	if spec.NodeModelGet("entry") == nil {
		t.Error("spec has no entry node model")
	}
	if len(spec.Router.Routes) != 2 {
		t.Errorf("spec router routes = %d, want 2", len(spec.Router.Routes))
	}

	// ipk.toml: decodes into the pkgBuild config with a slashed->dashed name
	var cfg pkgConfig
	if err := htoml.DecodeFromFile(filepath.Join(dir, "ipk.toml"), &cfg); err != nil {
		t.Fatalf("decode ipk.toml: %v", err)
	}
	if cfg.Metadata.Name != "demo-hello" {
		t.Errorf("ipk metadata name = %q, want demo-hello", cfg.Metadata.Name)
	}
	if cfg.Metadata.Version != spec.Meta.Version {
		t.Errorf("ipk version %q != spec version %q", cfg.Metadata.Version, spec.Meta.Version)
	}
	if strings.Join(cfg.Build.Include, ",") != "spec.json,views" {
		t.Errorf("ipk include = %v", cfg.Build.Include)
	}

	// views: hpress server-side template syntax must survive verbatim
	listBs, err := os.ReadFile(filepath.Join(dir, "views", "list.tpl"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(listBs), `{{pagelet . "core/general" "v3/html-header.tpl"}}`) {
		t.Error("list.tpl lost its pagelet directive")
	}

	// the scaffold is directly packageable
	if _, name, err := pkgBuild(dir); err != nil {
		t.Errorf("pkgBuild on scaffold: %v", err)
	} else if name != "demo-hello_0.1.0_all_src.ipk" {
		t.Errorf("pkgBuild name = %q", name)
	}

	// second run refuses to clobber
	if err := runModuleInit(dir, ""); err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Errorf("re-run error = %v, want refusing-to-overwrite", err)
	}
}
