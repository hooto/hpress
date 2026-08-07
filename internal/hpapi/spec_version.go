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
	"fmt"
	"strings"
)

type SpecVersion string

func NewSpecVersion(vstr string) *SpecVersion {
	v := SpecVersion(vstr)
	return &v
}

func (v *SpecVersion) Valid() bool {
	if *v == "" {
		return false
	}

	vs := strings.Split(string(*v), ".")
	if len(vs) > 3 {
		return false
	}

	for _, vsv := range vs {
		if vsv == "" {
			return false
		}

		for i := 0; i < len(vsv); i++ {
			if vsv[i] < '0' || vsv[i] > '9' {
				return false
			}
		}
	}

	return true
}

// Compare compares this version to another version. This
// returns -1, 0, or 1 if this version is smaller, equal,
// or larger than the compared version, respectively.
func (v *SpecVersion) Compare(other *SpecVersion) int {

	if v.String() != other.String() {
		vs, vs2 := v.parse(), other.parse()
		for i := 0; i < 3; i++ {
			if lg := vs[i] - vs2[i]; lg > 0 {
				return 1
			} else if lg < 0 {
				return -1
			}
		}
	}

	return 0
}

func (v *SpecVersion) Add(major, minor, patch int32) *SpecVersion {

	vs := v.parse()

	if major > 0 {
		vs[0] += major
		vs[1] = 0
		vs[2] = 0
	}
	if minor > 0 {
		vs[1] += minor
		vs[2] = 0
	}
	if patch > 0 {
		vs[2] += patch
	}

	*v = SpecVersion(fmt.Sprintf("%d.%d.%d", vs[0], vs[1], vs[2]))

	return v
}

func (v *SpecVersion) String() string {
	return string(*v)
}

func (v *SpecVersion) parse() []int32 {

	var (
		segments = []int32{}
		num      = int32(0)
	)

	for _, char := range *v {
		if char >= '0' && char <= '9' {
			if num > 0 {
				num = 10 * num
			}
			num += char - '0'
		} else {
			if len(segments) < 3 {
				segments = append(segments, num)
			}
			num = 0
		}
	}

	if num > 0 && len(segments) < 3 {
		segments = append(segments, num)
	}

	for len(segments) < 3 {
		segments = append([]int32{0}, segments...)
	}

	return segments
}
