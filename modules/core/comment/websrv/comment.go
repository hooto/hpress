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

	"github.com/hooto/hcaptcha/captcha4g"
	"github.com/hooto/httpsrv"
	"github.com/lessos/lessgo/types"
	"github.com/lessos/lessgo/utils"
	"github.com/lynkdb/iomix/rdb"

	"github.com/hooto/hpress/config"
	"github.com/hooto/hpress/datax"
	"github.com/hooto/hpress/store"
)

const (
	nsModName          = "core/comment"
	errCaptchaNotMatch = "CaptchaNotMatch"
)

type Comment struct {
	*httpsrv.Controller
}

func (c Comment) EmbedAction() {

	c.AutoRender = false

	if c.Params.Value("refer_modname") == "" || c.Params.Value("refer_id") == "" {
		return
	}

	qry := datax.NewQuery(nsModName, "entry")
	qry.Limit(500)
	qry.Filter("status", 1)
	qry.Order("created asc")
	qry.Filter("field_refer_id", c.Params.Value("refer_id"))
	qry.Filter("field_refer", c.Params.Value("refer_modname")+"."+c.Params.Value("refer_datax_table"))

	c.Data["list"] = qry.NodeList([]string{}, []string{})

	c.Data["new_form_refer_id"] = c.Params.Value("refer_id")
	c.Data["new_form_refer_modname"] = c.Params.Value("refer_modname")
	c.Data["new_form_refer_datax_table"] = c.Params.Value("refer_datax_table")

	c.Data["new_form_author"] = "Guest"

	c.Render(nsModName, "embed.tpl")
}

func (c Comment) SetAction() {

	var set TypeComment

	defer c.RenderJson(&set)

	if err := c.Request.JsonDecode(&set); err != nil {
		set.Error = &types.ErrorMeta{
			Code:    "400",
			Message: err.Error(),
		}
		return
	}

	set.Content = strings.TrimSpace(set.Content)
	set.Author = strings.TrimSpace(set.Author)

	ref_mod_ok := false
	for _, spec := range config.Modules {

		if spec.Meta.Name == set.ReferModName {
			ref_mod_ok = true
			break
		}
	}
	if !ref_mod_ok {
		set.Error = &types.ErrorMeta{
			Code:    "400",
			Message: "Spec Not Found",
		}
		return
	}

	if set.ReferID == "" || set.ReferDataxTable == "" {
		set.Error = &types.ErrorMeta{
			Code:    "400",
			Message: "ReferID or ReferDataxTable Can Not be Null",
		}
		return
	}

	re_title := "Re: "
	prevq := datax.NewQuery(set.ReferModName, set.ReferDataxTable)
	prevq.Filter("id", set.ReferID)
	if rs := prevq.NodeEntry(); rs.Error != nil || rs.Kind != "Node" {
		set.Error = &types.ErrorMeta{
			Code:    "400",
			Message: "Refer Content Not Found",
		}
		return
	} else {
		re_title += rs.Title
	}

	if set.Content == "" {
		set.Error = &types.ErrorMeta{
			Code:    "400",
			Message: "Content Can Not be Null",
		}
		return
	}

	if err := captcha4g.Verify(set.CaptchaToken, set.CaptchaWord); err != nil {

		set.Error = &types.ErrorMeta{
			Code:    errCaptchaNotMatch,
			Message: "Word Verification do not match",
		}

		return
	}

	set.Meta.ID = utils.StringNewRand(16)
	set.Meta.Created = rdb.TimeNow("datetime")

	tn := uint32(time.Now().Unix())

	//
	item := map[string]interface{}{
		"id":                  set.Meta.ID,
		"pid":                 "00",
		"title":               re_title,
		"field_title":         re_title,
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

	if _, err := store.Data.Insert("hpn_"+utils.StringEncode16("core/comment", 12)+"_entry", item); err != nil {
		set.Error = &types.ErrorMeta{
			Code:    "500",
			Message: err.Error(),
		}
		return
	} else {
		set.Kind = "Comment"
	}
}
