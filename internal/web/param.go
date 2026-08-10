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

package web

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// Param returns the value of key, replicating httpsrv Params.Value precedence:
// path parameter → URL query → form value. An empty string is returned when the
// key is absent in any source.
func Param(c fiber.Ctx, key string) string {
	if v := c.Params(key); v != "" {
		return v
	}
	if v := c.Query(key); v != "" {
		return v
	}
	if v := c.FormValue(key); v != "" {
		return v
	}
	return ""
}

// ParamInt returns the int64 value of key (0 when empty or non-numeric),
// replicating httpsrv Params.IntValue.
func ParamInt(c fiber.Ctx, key string) int64 {
	v := Param(c, key)
	if v == "" {
		return 0
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return n
}
