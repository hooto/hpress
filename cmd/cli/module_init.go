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

// module-init: scaffold a content module directory (spec.json, ipk.toml,
// views/). The skeleton mirrors modules/ruilog/notebook — the smallest fully
// working module — trimmed to one node model, two term models, list/view
// actions, and a two-route router.
//
// spec.json and ipk.toml are rendered through text/template (name/title
// substitution); the views are written verbatim — they are hpress server-side
// templates ({{pagelet ...}} and friends) that must NOT be expanded locally.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

// modNameRe mirrors modNamePattern in internal/modset/modset.go — the server
// applies the same rule on upload. Duplicated here because importing modset
// would drag the whole store/config stack into the CLI.
var modNameRe = regexp.MustCompile(`^[0-9a-z/]{3,30}$`)

func newModuleInitCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "module-init <dir>",
		Short: "Scaffold a content module directory (spec.json, ipk.toml, views/)",
		Long: "Scaffold a content module at <dir>: spec.json (one entry node model " +
			"with tags/categories terms, list/view actions, router), ipk.toml " +
			"(innerstack package spec), and minimal views/. The module name " +
			"defaults to the last two dir path segments (modules/demo/hello -> " +
			"demo/hello); pass --name to override.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModuleInit(args[0], name)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "module name, e.g. demo/hello (default: last two dir path segments)")
	return cmd
}

func runModuleInit(dir, nameOverride string) error {

	name, err := moduleNameDerive(dir, nameOverride)
	if err != nil {
		return err
	}

	// refuse to clobber an existing module (idempotent safety, not overwrite)
	if _, err := os.Stat(filepath.Join(dir, "spec.json")); err == nil {
		return fmt.Errorf("%s already contains a spec.json — refusing to overwrite", dir)
	}

	srvName := name[strings.LastIndex(name, "/")+1:]

	vars := map[string]string{
		"Name":    name,
		"SrvName": srvName,
		"PkgName": strings.ReplaceAll(name, "/", "-"),
		"Title":   titleCase(srvName),
	}

	// name-substituted skeletons
	templates := map[string]string{
		"spec.json": tplSpec,
		"ipk.toml":  tplIpkToml,
	}
	// verbatim hpress server-side templates
	assets := map[string]string{
		"views/entry.tpl":           tplViewEntry,
		"views/list.tpl":            tplViewList,
		"views/term/categories.tpl": tplViewTermCategories,
	}

	for path, body := range templates {
		full := filepath.Join(dir, filepath.FromSlash(path))
		if err := renderTemplate(full, body, vars); err != nil {
			return err
		}
		fmt.Println("created", full)
	}
	for path, body := range assets {
		full := filepath.Join(dir, filepath.FromSlash(path))
		if err := writeFileNew(full, body); err != nil {
			return err
		}
		fmt.Println("created", full)
	}

	fmt.Printf("\nmodule %s scaffolded\n", name)
	fmt.Printf("next: edit %s, then `hpress module-build %s` and `hpress module-push %s`\n",
		filepath.Join(dir, "spec.json"), dir, dir)
	return nil
}

// moduleNameDerive resolves the module name: an explicit override wins;
// otherwise the last two segments of the cleaned dir path (modules/demo/hello
// -> demo/hello; a bare path segment is used as-is).
func moduleNameDerive(dir, override string) (string, error) {

	name := strings.TrimSpace(override)
	if name != "" {
		name = strings.ToLower(filepath.ToSlash(filepath.Clean(name)))
		if !modNameRe.MatchString(name) {
			return "", fmt.Errorf("invalid module name %q (want lowercase digits and '/', 3-30 chars)", name)
		}
		return name, nil
	}

	var segs []string
	for s := range strings.SplitSeq(strings.ToLower(filepath.ToSlash(filepath.Clean(dir))), "/") {
		if s == "" || s == "." || s == ".." {
			continue
		}
		segs = append(segs, s)
	}
	switch {
	case len(segs) == 0:
		return "", fmt.Errorf("cannot derive a module name from %q (pass --name)", dir)
	case len(segs) == 1:
		name = segs[0]
	default:
		name = segs[len(segs)-2] + "/" + segs[len(segs)-1]
	}

	if !modNameRe.MatchString(name) {
		return "", fmt.Errorf("invalid module name %q (want lowercase digits and '/', 3-30 chars)", name)
	}
	return name, nil
}

// renderTemplate expands body (a text/template over the vars map) and writes it
// to path, creating parent directories as needed. O_EXCL guards against
// clobbering an existing file.
func renderTemplate(path, body string, vars map[string]string) error {
	tpl, err := template.New(filepath.Base(path)).Parse(body)
	if err != nil {
		return fmt.Errorf("parse template %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()
	if err := tpl.Execute(fp, vars); err != nil {
		return fmt.Errorf("render %s: %w", path, err)
	}
	return nil
}

// writeFileNew writes body verbatim to a new file at path.
func writeFileNew(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()
	if _, err := fp.WriteString(body); err != nil {
		return err
	}
	return nil
}

// titleCase upper-cases the first rune ("hello" -> "Hello").
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

const tplSpec = `{
  "meta": {
    "name": "{{.Name}}",
    "version": "0.1.0"
  },
  "srvname": "{{.SrvName}}",
  "status": 1,
  "title": "{{.Title}}",
  "nodeModels": [
    {
      "meta": {
        "name": "entry"
      },
      "modname": "{{.Name}}",
      "title": "{{.Title}}",
      "fields": [
        {
          "name": "title",
          "type": "string",
          "length": "100",
          "title": "Title"
        },
        {
          "name": "content",
          "type": "text",
          "length": "0",
          "attrs": [
            {
              "key": "ui_rows",
              "value": "20"
            },
            {
              "key": "formats",
              "value": "text,md,html"
            }
          ],
          "title": "Content"
        }
      ],
      "terms": [
        {
          "meta": {
            "name": "tags"
          },
          "type": "tag",
          "title": "Tags"
        },
        {
          "meta": {
            "name": "categories"
          },
          "type": "taxonomy",
          "title": "Categories"
        }
      ],
      "extensions": {
        "access_counter": true,
        "permalink": "name",
        "text_search": true
      }
    }
  ],
  "termModels": [
    {
      "meta": {
        "name": "tags"
      },
      "type": "tag",
      "title": "Tags"
    },
    {
      "meta": {
        "name": "categories"
      },
      "type": "taxonomy",
      "title": "Categories"
    }
  ],
  "actions": [
    {
      "name": "list",
      "datax": [
        {
          "name": "list",
          "type": "node.list",
          "pager": true,
          "query": {
            "table": "entry",
            "limit": 10
          },
          "cache_ttl": 600000
        },
        {
          "name": "categories",
          "type": "term.list",
          "query": {
            "table": "categories",
            "limit": 100
          },
          "cache_ttl": 3600000
        }
      ]
    },
    {
      "name": "view",
      "datax": [
        {
          "name": "entry",
          "type": "node.entry",
          "query": {
            "table": "entry",
            "limit": 1
          },
          "cache_ttl": 600000
        }
      ]
    }
  ],
  "router": {
    "routes": [
      {
        "path": "view/:id",
        "dataAction": "view",
        "template": "entry.tpl"
      },
      {
        "path": "list",
        "dataAction": "list",
        "template": "list.tpl",
        "default": true
      }
    ]
  }
}
`

const tplIpkToml = `# ipk.toml - Package Specification v2
# Innerstack Package Build Configuration

[metadata]
name        = "{{.PkgName}}"
version     = "0.1.0"
authors     = ["hooto.com"]
description = "{{.Title}} content module for hooto-press"
homepage    = "https://github.com/hooto/hpress"
license     = "Apache-2.0"
categories  = ["app/cms"]

[build]
# Pure content package: spec.json + templates. No build script needed.
# Bare directory names are copied recursively by the builder.
include = [
  "spec.json",
  "views",
]
`

const tplViewEntry = `<!doctype html>
<html lang="en">
  {{pagelet . "core/general" "v3/html-header.tpl"}}
  <body id="hp-body" class="hp-theme-paper">
    {{pagelet . "core/general" "v3/nav-header.tpl" "topnav"}}

    <div class="container hp-block-gap-column">
      <div class="hp-block-gap">
        <div class="col">
          <div class="hp-ctn-title">Entry</div>
        </div>
      </div>

      <div class="hp-block-gap-column">
        <div class="hp-node-view">
          <div class="hp-header">{{FieldStringPrint .entry "title" .LANG}}</div>

          <div class="hp-info">
            <span class="info-item">
              Published: {{UnixtimeFormat .entry.Created "Y-m-d"}}
            </span>

            {{range $term := .entry.Terms}} {{if eq $term.Name "categories"}}
            {{if $term.Items}}
            <span class="info-item">
              Categories: {{range $term_item := $term.Items}}
              <a
                href='{{$.baseuri}}/list?term_categories={{printf "%d" $term_item.ID}}'
                >{{$term_item.Title}}</a
              >
              {{end}}
            </span>
            {{end}} {{end}} {{end}} {{range $term := .entry.Terms}} {{if eq
            $term.Name "tags"}} {{if $term.Items}}
            <span class="info-item">
              Tags: {{range $term_item := $term.Items}}
              <a
                href="{{$.baseuri}}/list?term_tags={{$term_item.Title}}"
                class="info-tag-item"
                >{{$term_item.Title}}</a
              >
              {{end}}
            </span>
            {{end}} {{end}} {{end}}
          </div>

          <div class="content hp-content">
            {{FieldHtmlPrint .entry "content" .LANG}}
          </div>
        </div>
      </div>
    </div>

    {{pagelet . "core/general" "v3/footer.tpl"}}
    {{pagelet . "core/general" "html-footer.tpl"}}
  </body>
</html>
`

const tplViewList = `<!doctype html>
<html lang="en">
  {{pagelet . "core/general" "v3/html-header.tpl"}}
  <body id="hp-body" class="hp-theme-paper">
    {{pagelet . "core/general" "v3/nav-header.tpl" "topnav"}}

    <div class="container hp-block-gap-column">
      <div class="hp-block-gap">
        <div class="col">
          <div class="hp-ctn-title">Entries</div>
        </div>
      </div>

      <div class="hp-block-gap-column">
        <div class="row hp-block-gap">
          <div class="col col-11 col-lg-8 hp-block-gap-column">
            <div class="hp-node-list d-flex flex-column hp-block-gap-column">
              {{range $v := .list.Items}}
              <div class="hp-node-list-item clearfix">
                <h4 class="hp-node-list-heading">
                  <a href="{{$.baseuri}}/view/{{$v.ID}}.html"
                    >{{FieldStringPrint $v "title" $.LANG}}</a
                  >
                </h4>
                <div class="hp-node-list-info">
                  <span class="info-item">
                    Published: {{UnixtimeFormat $v.Created "Y-m-d"}}
                  </span>
                </div>

                <div class="hp-node-list-text">
                  {{FieldHtmlSubPrint $v "content" 100 $.LANG}}
                </div>
              </div>
              {{end}}
            </div>

            {{if .list_pager}}
            <nav>
              <ul class="pagination justify-content-center">
                {{range $page := .list_pager.RangePages}}
                <li
                  class="page-item {{if eq $page $.list_pager.CurrentPageNumber}}active{{end}}"
                >
                  <a class="page-link" href="{{$.baseuri}}/list?page={{$page}}"
                    >{{$page}}</a
                  >
                </li>
                {{end}}
              </ul>
            </nav>
            {{end}}
          </div>

          <div class="col d-none d-lg-block col-3 col-lg-3">
            {{pagelet . .modname "term/categories.tpl"}}
          </div>
        </div>
      </div>
    </div>

    {{pagelet . "core/general" "v3/footer.tpl"}}
    {{pagelet . "core/general" "html-footer.tpl"}}
  </body>
</html>
`

const tplViewTermCategories = `{{if .categories}}
<div class="hp-sidebar-section">
  <div class="header">{{.categories.Model.Title}}</div>
  <div class="list-group term-taxonomy-group">
    <div class="list-group-item">
      <a class="term-taxonomy-item" href="{{$.baseuri}}/list">All</a>
    </div>
    {{range $v := .categories.Items}} {{if ne $v.PID 0}}{{continue}}{{end}}
    <div class="list-group-item">
      <a
        class="term-taxonomy-item {{if eq $.term_categories $v.ID}} active{{end}}"
        href="{{$.baseuri}}/list?term_{{$.categories.Model.Meta.Name}}={{$v.ID}}"
        >{{$v.Title}}</a
      >
    </div>
    {{end}}
  </div>
</div>
{{end}}
`
