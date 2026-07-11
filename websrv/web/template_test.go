// Copyright 2015 Eryx <evorui at gmail dot com>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package web

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplateLoaderSetRenderClean(t *testing.T) {
	dir := t.TempDir()

	// a template that exercises a builtin (upper), a registered func, a sub-template
	// include, and dot data.
	if err := os.WriteFile(filepath.Join(dir, "index.tpl"), []byte(
		`{{upper .Name}}|{{shout .Name}}|{{template "_part.tpl" .}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "_part.tpl"), []byte(`[{{.Name}}]`), 0o644); err != nil {
		t.Fatal(err)
	}

	l := NewTemplateLoader()
	l.RegisterFunc("shout", func(s string) string { return strings.ToUpper(s) + "!" })

	l.Set("core/general", []string{dir})

	var sb strings.Builder
	if err := l.Render(&sb, "core/general", "index.tpl", map[string]string{"Name": "hpress"}); err != nil {
		t.Fatalf("render: %v", err)
	}
	got := sb.String()
	want := "HPRESS|HPRESS!|[hpress]"
	if got != want {
		t.Fatalf("render output = %q, want %q", got, want)
	}

	// case-insensitive lookup
	sb.Reset()
	if err := l.Render(&sb, "core/general", "INDEX.TPL", map[string]string{"Name": "x"}); err != nil {
		t.Fatalf("render uppercase lookup: %v", err)
	}

	// idempotent Set does not blow away the set
	l.Set("core/general", []string{dir})
	sb.Reset()
	if err := l.Render(&sb, "core/general", "index.tpl", map[string]string{"Name": "hpress"}); err != nil {
		t.Fatalf("render after re-set: %v", err)
	}

	// Clean removes the set
	l.Clean("core/general")
	if err := l.Render(&sb, "core/general", "index.tpl", nil); err == nil {
		t.Fatal("expected error after Clean, got nil")
	}
}
