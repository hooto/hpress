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
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

// Bind decodes the JSON request body into out, replacing httpsrv Request.JsonDecode.
func Bind(c fiber.Ctx, out any) error {
	return json.Unmarshal(c.Body(), out)
}

// RawBody returns the raw request body bytes, replacing httpsrv Request.RawBody.
func RawBody(c fiber.Ctx) []byte {
	return c.Body()
}
