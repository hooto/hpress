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
	"bytes"
	"errors"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// TemplateLoader loads and parses html/template sets keyed by module name,
// replacing httpsrv.TemplateLoader. Templates are read from on-disk view
// directories (hpress never uses an http.FileSystem for templates).
type TemplateLoader struct {
	mu      sync.RWMutex
	funcMap template.FuncMap
	sets    map[string]*template.Template
	paths   map[string]string
}

// NewTemplateLoader creates an empty loader seeded with the framework builtins.
func NewTemplateLoader() *TemplateLoader {
	fm := make(template.FuncMap, len(builtinFuncs))
	for k, v := range builtinFuncs {
		fm[k] = v
	}
	return &TemplateLoader{
		funcMap: fm,
		sets:    map[string]*template.Template{},
		paths:   map[string]string{},
	}
}

// RegisterFunc adds (or overrides) a template function. Must be called before
// the relevant module's templates are parsed (all registration happens at
// package init, before runtime Set calls).
func (l *TemplateLoader) RegisterFunc(name string, fn any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.funcMap[name] = fn
}

// FuncMap returns a copy of the loader's function map (builtins plus any funcs
// added via RegisterFunc). It replaces direct reads of httpsrv.TemplateFuncs for
// ad-hoc template parsing, e.g. validating an uploaded .tpl file with the full
// set of registered funcs before saving it.
func (l *TemplateLoader) FuncMap() template.FuncMap {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make(template.FuncMap, len(l.funcMap))
	for k, v := range l.funcMap {
		out[k] = v
	}
	return out
}

// Clean drops the template set and path index for a module, replacing
// httpsrv TemplateLoader.Clean.
func (l *TemplateLoader) Clean(module string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.sets, module)
	for k := range l.paths {
		if strings.HasPrefix(k, module+".") {
			delete(l.paths, k)
		}
	}
}

// Set walks each view directory, parsing .tpl/.html files into a single
// associated template set keyed by module. It is idempotent: a second call for
// an already-loaded module is a no-op, matching httpsrv TemplateLoader.Set.
func (l *TemplateLoader) Set(module string, viewpaths []string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	set, ok := l.sets[module]
	if ok {
		return
	}

	addTemplate := func(templateFile, fileStr string) {
		name := strings.Trim(templateFile, "/")
		nameL := strings.ToLower(name)

		if _, ok := l.paths[module+"."+nameL]; ok {
			return
		}

		var err error
		func() {
			defer func() {
				if e := recover(); e != nil {
					err = errors.New("panic (template loader)")
				}
			}()
			if set == nil {
				set = template.New(name).Funcs(l.funcMap)
				if _, err = set.Parse(fileStr); err == nil {
					l.sets[module] = set
				}
			} else {
				_, err = set.New(name).Parse(fileStr)
			}
		}()

		if err != nil {
			slog.Warn("web template parse err", "module", module, "template", templateFile, "err", err.Error())
			return
		}

		if nameL != name {
			set.New(nameL).Parse(fileStr)
		}
		l.paths[module+"."+nameL] = templateFile
		slog.Info("web module template added", "module", module, "template", templateFile)
	}

	for _, baseDir := range viewpaths {
		_ = filepath.WalkDir(baseDir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(p, ".tpl") && !strings.HasSuffix(p, ".html") {
				return nil
			}
			b, rerr := os.ReadFile(p)
			if rerr != nil {
				return nil
			}
			rel, _ := filepath.Rel(baseDir, p)
			addTemplate(path.Join("/", filepath.ToSlash(rel)), string(b))
			return nil
		})
	}
}

// Render executes the named template for module into wr, replacing httpsrv
// TemplateLoader.Render. Lookups are case-insensitive (lowercase fallback).
func (l *TemplateLoader) Render(wr io.Writer, module, tplPath string, arg any) error {
	defer func() {
		if e := recover(); e != nil {
			slog.Debug("web template render err", "module", module, "template", tplPath, "err", e)
		}
	}()

	l.mu.RLock()
	set, ok := l.sets[module]
	l.mu.RUnlock()
	if !ok || set == nil {
		return errors.New("module " + module + " not found")
	}

	tpl := set.Lookup(tplPath)
	if tpl == nil {
		if tplPathL := strings.ToLower(tplPath); tplPathL != tplPath {
			tpl = set.Lookup(tplPathL)
		}
		if tpl == nil {
			return errors.New("template " + module + "/" + tplPath + " not found")
		}
	}
	return tpl.Execute(wr, arg)
}

// rawRender parses and executes an inline template string (with a small cache),
// replacing httpsrv TemplateLoader.rawRender / Controller.RenderHTML.
func (l *TemplateLoader) rawRender(wr io.Writer, txt string, arg any) error {
	defer func() {
		if e := recover(); e != nil {
			slog.Debug("web raw-render err", "err", e)
		}
	}()

	t, err := template.New("raw").Funcs(l.funcMap).Parse(txt)
	if err != nil {
		return err
	}
	return t.Execute(wr, arg)
}

// renderToBuffer is a convenience used by the response helpers.
func (l *TemplateLoader) renderToBuffer(module, tplPath string, arg any) ([]byte, error) {
	var buf bytes.Buffer
	if err := l.Render(&buf, module, tplPath, arg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
