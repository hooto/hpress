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
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/hooto/hcaptcha/captcha4g"
	"github.com/lessos/lessgo/types"
	"github.com/lessos/lessgo/utils"
	"github.com/lynkdb/lynkapi/go/lynktable"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/datax"
	"github.com/hooto/hpress/internal/hpapi"
	"github.com/hooto/hpress/internal/store"
	"github.com/hooto/hpress/internal/web"
)

const (
	nsModName          = "core/comment"
	errCaptchaNotMatch = "CaptchaNotMatch"
)

// CommentEmbed renders the comment embed (list + new-comment form) for a given
// refer node. Replaces httpsrv Comment.EmbedAction.
func CommentEmbed(c fiber.Ctx) error {

	if web.Param(c, "refer_modname") == "" || web.Param(c, "refer_id") == "" {
		return nil
	}

	query := datax.NewQuery(nsModName, "entry")
	query.Limit(500)
	query.Filter("status", 1)
	query.Order("created asc")
	query.Filter("field_refer_id", web.Param(c, "refer_id"))
	query.Filter("field_refer", web.Param(c, "refer_modname")+"."+web.Param(c, "refer_datax_table"))

	data := map[string]interface{}{
		"list":                       query.NodeList([]string{}, []string{}),
		"new_form_refer_id":          web.Param(c, "refer_id"),
		"new_form_refer_modname":     web.Param(c, "refer_modname"),
		"new_form_refer_datax_table": web.Param(c, "refer_datax_table"),
		"new_form_author":            "Guest",
		"LANG":                       web.ResolveLang(c),
		"URL_MOD_PATH":               "/hp/+/comment",
	}

	return web.Render(c, nsModName, "embed.tpl", data)
}

// CommentSet accepts a new comment submission. Replaces httpsrv Comment.SetAction.
func CommentSet(c fiber.Ctx) error {

	var set TypeComment

	defer func() { _ = web.JSON(c, &set) }()

	if err := web.Bind(c, &set); err != nil {
		set.Error = &types.ErrorMeta{
			Code:    "400",
			Message: err.Error(),
		}
		return nil
	}

	set.Content = strings.TrimSpace(set.Content)
	set.Author = strings.TrimSpace(set.Author)

	refModOK := false
	for _, spec := range config.Modules {

		if spec.Meta.Name == set.ReferModName {
			refModOK = true
			break
		}
	}
	if !refModOK {
		set.Error = &types.ErrorMeta{
			Code:    "400",
			Message: "Spec Not Found",
		}
		return nil
	}

	if set.ReferID == "" || set.ReferDataxTable == "" {
		set.Error = &types.ErrorMeta{
			Code:    "400",
			Message: "ReferID or ReferDataxTable Can Not be Null",
		}
		return nil
	}

	reTitle := "Re: "
	prevQuery := datax.NewQuery(set.ReferModName, set.ReferDataxTable)
	prevQuery.Filter("id", set.ReferID)
	if rs := prevQuery.NodeEntry(); rs.Error != nil || rs.Kind != "Node" {
		set.Error = &types.ErrorMeta{
			Code:    "400",
			Message: "Refer Content Not Found",
		}
		return nil
	} else {
		reTitle += rs.Title
	}

	if set.Content == "" {
		set.Error = &types.ErrorMeta{
			Code:    "400",
			Message: "Content Can Not be Null",
		}
		return nil
	}

	if err := captcha4g.Verify(set.CaptchaToken, set.CaptchaWord); err != nil {

		set.Error = &types.ErrorMeta{
			Code:    errCaptchaNotMatch,
			Message: "Word Verification do not match",
		}

		return nil
	}

	set.Meta.ID = utils.StringNewRand(16)
	set.Meta.Created = lynktable.TimeNow("datetime")

	tn := uint32(time.Now().Unix())

	//
	item := map[string]interface{}{
		"id":                  set.Meta.ID,
		"pid":                 "00",
		"title":               reTitle,
		"field_title":         reTitle,
		"status":              1,
		"userid":              utils.StringEncode16("guest", 8),
		"field_refer_id":      set.ReferID,
		"field_refer":         set.ReferModName + "." + set.ReferDataxTable,
		"field_author":        set.Author,
		"field_content":       set.Content,
		"field_address":       "",
		"created":             tn,
		"updated":             tn,
		"field_content_attrs": "[]",
	}

	if err := store.Data.Insert(hpapi.NodeTableName("core/comment", "entry"), item).Err(); err != nil {
		set.Error = &types.ErrorMeta{
			Code:    "500",
			Message: err.Error(),
		}
		return nil
	} else {
		set.Kind = "Comment"
	}

	return nil
}
