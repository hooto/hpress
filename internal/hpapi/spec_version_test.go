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

package hpapi

import (
	"testing"
)

type spec_version_valid_entry struct {
	v  string
	ok bool
}

func TestSpecVersion(t *testing.T) {

	v1 := SpecVersion("2.0.12")
	v2 := SpecVersion("2.0.20")
	v3 := SpecVersion("2.0.20")

	if v1.Compare(&v2) != -1 || v2.Compare(&v3) != 0 || v2.Compare(&v1) != 1 {
		t.Fatal("Failed on Compare")
	}

	if s1 := v1.Add(0, 0, 0).String(); s1 != "2.0.12" {
		t.Fatal("Failed on String")
	}

	vs := []spec_version_valid_entry{
		{"1.0.0", true},
		{"1", true},
		{"100", true},
		{"", false},
		{"-", false},
		{".", false},
		{" ", false},
		{"!", false},
		{".0", false},
		{".0.", false},
	}
	for _, v := range vs {
		vv := SpecVersion(v.v)
		if vv.Valid() != v.ok {
			t.Fatal("Failed on Valid " + v.v)
		}
	}
}

func Benchmark_SpecVersion_Valid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v1 := SpecVersion("10.10.10")
		v1.Valid()
	}
}
