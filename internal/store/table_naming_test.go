// Copyright 2015 Eryx <evorui at gmail dot com>, All rights reserved.
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

package store

import (
	"testing"
)

// md5("<module>") truncated to 12 lowercase hex chars, computed offline.
// These pin legacyModuleHash to the exact bytes the old helpers produced so a
// future change to the hash cannot silently break the reverse map.
func TestLegacyModuleHash(t *testing.T) {
	cases := map[string]string{
		"core/blog":        "d35b85a1f8d9",
		"core/general":     "d2747912c78e",
		"core/comment":     "4c7c5e656873",
		"core/gdoc":        "d72e2d753ce1",
		"ruilog/notebook":  "d2f76f1e5c87",
		"sysinner/website": "682fa7c72e2a",
	}
	for name, want := range cases {
		if got := legacyModuleHash(name); got != want {
			t.Errorf("legacyModuleHash(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestModuleHashToKey(t *testing.T) {
	got := moduleHashToKey([]string{"core/blog", "ruilog/notebook", ""})
	want := map[string]string{
		"d35b85a1f8d9": "core_blog",
		"d2f76f1e5c87": "ruilog_notebook",
	}
	if len(got) != len(want) {
		t.Fatalf("moduleHashToKey map size = %d, want %d (%+v)", len(got), len(want), got)
	}
	for h, k := range want {
		if got[h] != k {
			t.Errorf("moduleHashToKey[%q] = %q, want %q", h, got[h], k)
		}
	}
}

func TestParseLegacyModuleTable(t *testing.T) {
	hashToKey := moduleHashToKey([]string{"core/blog", "ruilog/notebook"})

	cases := []struct {
		name       string
		wantPrefix string
		wantKey    string
		wantModel  string
		wantOK     bool
	}{
		// legacy node table for a known module
		{"hpn_d35b85a1f8d9_post", "hpn", "core_blog", "post", true},
		// legacy term table for a known module; model may itself contain "_"
		{"hpt_d35b85a1f8d9_node_tag", "hpt", "core_blog", "node_tag", true},
		// legacy table for the exp module
		{"hpn_d2f76f1e5c87_entry", "hpn", "ruilog_notebook", "entry", true},
		// 12-hex hash but module unknown -> orphan, do not touch
		{"hpn_000000000000_post", "", "", "", false},
		// already on the new scheme -> regex miss, idempotent
		{"hpn_core_blog_post", "", "", "", false},
		{"hpt_ruilog_notebook_tag", "", "", "", false},
		// unrelated tables
		{"hp_modules", "", "", "", false},
		{"hp_sys_config", "", "", "", false},
		{"nx_d35b85a1f8d9_post", "", "", "", false},
		{"", "", "", "", false},
		// hash shorter/longer than 12 hex chars is not legacy
		{"hpn_d35b85a1f8d_post", "", "", "", false},
		{"hpn_d35b85a1f8d9ab_post", "", "", "", false},
		// uppercase hex is not produced by the legacy helpers
		{"hpn_D35B85A1F8D9_post", "", "", "", false},
	}

	for _, c := range cases {
		prefix, key, model, ok := parseLegacyModuleTable(c.name, hashToKey)
		if ok != c.wantOK || prefix != c.wantPrefix || key != c.wantKey || model != c.wantModel {
			t.Errorf("parseLegacyModuleTable(%q) = (%q,%q,%q,%t), want (%q,%q,%q,%t)",
				c.name, prefix, key, model, ok,
				c.wantPrefix, c.wantKey, c.wantModel, c.wantOK)
		}
	}
}
