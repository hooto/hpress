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

package websrv

import (
	"github.com/gofiber/fiber/v3"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/web"
)

// Register mounts the comment module routes on a fiber router. The caller mounts
// the router at "/hp/+/comment" (note the literal "+", which the caller escapes).
// Routes reproduce the httpsrv "Comment" controller paths
// {prefix}/comment/embed, {prefix}/comment/set plus the static asset prefix /~.
func Register(router fiber.Router) {
	router.All("/comment/embed", CommentEmbed)
	router.All("/comment/set", CommentSet)
	router.Get("/~/*", web.DiskStatic(config.Prefix+"/modules/core/comment/static/"))
}
