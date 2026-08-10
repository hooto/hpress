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

// Publish / update flow: fetch the module spec, interactively gather metadata,
// upload images, and create (or update) the node.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/lessos/lessgo/types"

	"github.com/hooto/hpress/internal/hpapi"
)

var (
	// mdImageRefRe matches the markdown image placeholder written by extract:
	//   ![alt]({{hp_storage_service_endpoint}}/2026/08/03/<hash>.jpg?ipn=s800x)
	//   ![alt]({{hp_storage_service_endpoint}}/2026/08/03/<hash>.svg)
	// capturing the <date>/<file> tail (storage path under /deft). Raster refs
	// carry the resize query; SVG (vector) does not.
	mdImageRefRe = regexp.MustCompile(`\{\{hp_storage_service_endpoint\}\}/([0-9]{4}/[0-9]{2}/[0-9]{2}/[0-9a-fA-F]+\.(?:jpg|svg))`)
)

// publishScopes are the IAM access-key scopes a publish-capable key must carry.
// Shown on setup (auth) and whenever the server rejects a publish call on
// auth/access grounds, so the required scopes are discoverable without the docs.
// "app=<app_id>" is the cross-app binding (hpress instance app_id =
// [iam_auth].app_id / instance_id); without it introspect returns inactive and
// Auth fails with "Unauthorized: auth-denied". The editor.* scopes are the
// hpress route gates the publish flow actually hits.
var publishScopes = []string{
	"app=<app_id>", // cross-app binding; introspect rejects without it
	"editor.read",  // mod-set/spec-entry — fetch module spec
	"editor.list",  // term/list — fetch categories/taxonomy
	"editor.write", // node/set + s2-obj/put — create node + upload images
}

// scopeHint is a one-line summary of publishScopes for error messages.
func scopeHint() string {
	return "IAM access key needs scopes: " + strings.Join(publishScopes, ", ")
}

// authScopeSuffix appends the scope hint when err is an auth/access denial —
// "Unauthorized" from the Auth gate or "AccessDenied" from a route permission
// gate — so the user sees exactly which scopes to grant. Returns "" for
// unrelated errors to keep them terse.
func authScopeSuffix(err error) string {
	var ae *apiError
	if errors.As(err, &ae) && (ae.Code == "Unauthorized" || ae.Code == "AccessDenied") {
		return "\n  hint: " + scopeHint() + "\n  (app=<app_id> = hpress [iam_auth].app_id)"
	}
	return ""
}

// runPublish publishes (isUpdate=false) or updates (isUpdate=true) the markdown
// at mdPath to the configured hpress module.
func runPublish(mdPath string, isUpdate bool) error {

	cfg, err := LoadClientConfig()
	if err != nil {
		return err
	}
	if cfg.Server.BaseURL == "" || cfg.Auth.AccessKey == "" {
		return fmt.Errorf("not configured: run `%s auth --key <ak_...> --server <url> --module <mod>` first",
			os.Args[0])
	}
	if cfg.Publish.Module == "" {
		return fmt.Errorf("no default module configured (set [publish].module or pass --module)")
	}

	md, err := os.ReadFile(mdPath)
	if err != nil {
		return err
	}
	mdContent := string(md)

	client, err := NewClient(cfg.Server.BaseURL, cfg.Auth.AccessKey)
	if err != nil {
		return err
	}

	outDir := cfg.Publish.ImageOutDir
	if outDir == "" {
		outDir = defaultOutDir
	}

	// load prior state (required for update)
	state, _ := LoadArticleState(mdPath)
	if isUpdate && (state == nil || state.NodeID == "") {
		return fmt.Errorf("no saved state for %s — publish first (run without --update)", mdPath)
	}

	reader := bufio.NewReader(os.Stdin)

	// 1) module spec + node model
	spec, err := client.SpecGet(cfg.Publish.Module)
	if err != nil {
		return fmt.Errorf("fetch spec: %w%s", err, authScopeSuffix(err))
	}
	if spec.Error != nil {
		return fmt.Errorf("spec fetch failed: %s", spec.Error.Message)
	}
	model := pickNodeModel(reader, spec, cfg.Publish.ModelID, state)
	if model == nil {
		return fmt.Errorf("no node model selected")
	}
	fmt.Printf("\nPublishing to module %q, model %q\n", cfg.Publish.Module, model.Meta.Name)

	// 2) title (default: prior state → first H1 → filename)
	defTitle := ""
	if state != nil {
		defTitle = state.Title
		if state.Tags == nil {
			state.Tags = map[string]string{}
		}
		if state.Categories == nil {
			state.Categories = map[string]string{}
		}
	}
	if defTitle == "" {
		defTitle = defaultTitle(mdContent, mdPath)
	}
	title := prompt(reader, "Title", defTitle)

	// 3) terms (taxonomy → interactive tree pick; tag → comma list)
	terms := []hpapi.NodeTerm{}
	stateCats := map[string]string{}
	stateTags := map[string]string{}
	var keywords []string
	if state != nil {
		stateCats, stateTags = state.Categories, state.Tags
		keywords = state.Keywords
	}

	for _, tm := range model.Terms {
		switch tm.Type {
		case "taxonomy":
			id, err := pickTermInteractive(reader, client, cfg.Publish.Module, tm, stateCats[tm.Meta.Name])
			if err != nil {
				fmt.Fprint(os.Stderr, "skip term: ", err, authScopeSuffix(err), "\n")
				continue
			}
			terms = append(terms, hpapi.NodeTerm{Name: tm.Meta.Name, Value: id, Type: "taxonomy"})
			stateCats[tm.Meta.Name] = id
		case "tag":
			def := stateTags[tm.Meta.Name]
			// Offer the extracted keyword list as the starting point for a tag
			// field that has no prior value; the operator can edit or clear it.
			if def == "" && len(keywords) > 0 {
				def = strings.Join(keywords, ", ")
			}
			val := prompt(reader, "Tags ("+tm.Title+", comma-separated)", def)
			if strings.TrimSpace(val) != "" {
				terms = append(terms, hpapi.NodeTerm{Name: tm.Meta.Name, Value: val, Type: "tag"})
				stateTags[tm.Meta.Name] = val
			}
		}
	}

	// 4) other declared fields (skip title/content — handled explicitly)
	extraFields, err := promptExtraFields(reader, model)
	if err != nil {
		return err
	}

	// 5) upload images referenced in the markdown. On re-publish / update the
	//    prior state manifest records what is already on the server, so unchanged
	//    images are skipped instead of re-uploaded.
	var priorImages []ArticleImage
	if state != nil {
		priorImages = state.Images
	}
	images, err := uploadReferencedImages(client, mdContent, outDir, priorImages)
	if err != nil {
		return fmt.Errorf("image upload: %w%s", err, authScopeSuffix(err))
	}

	// 6) build + send the node
	fields := []*hpapi.NodeField{
		{Name: "title", Value: title},
		{Name: "content", Value: mdContent, Attrs: types.KvPairs{&types.KvPair{Key: "format", Value: "md"}}},
	}
	fields = append(fields, extraFields...)

	node := &hpapi.Node{
		Status: 1,
		Title:  title,
		Fields: fields,
		Terms:  terms,
	}
	if isUpdate && state != nil {
		node.ID = state.NodeID
	}

	resp, err := client.NodeSet(cfg.Publish.Module, model.Meta.Name, node)
	if err != nil {
		return fmt.Errorf("node set: %w%s", err, authScopeSuffix(err))
	}
	if resp.Error != nil {
		return fmt.Errorf("node set failed: %s", resp.Error.Message)
	}
	if resp.ID == "" {
		return fmt.Errorf("node set returned no id")
	}

	verb := "published"
	if isUpdate {
		verb = "updated"
	}
	fmt.Printf("\n✓ %s node %s  (module %s)\n", verb, resp.ID, cfg.Publish.Module)

	// 7) persist state
	newState := &ArticleState{
		ServerBaseURL: cfg.Server.BaseURL,
		Module:        cfg.Publish.Module,
		ModelID:       model.Meta.Name,
		NodeID:        resp.ID,
		Title:         title,
		Status:        resp.Status,
		Created:       resp.Created,
		Updated:       resp.Updated,
		Categories:    stateCats,
		Tags:          stateTags,
		Images:        images,
	}
	// carry the extracted keyword list forward so update/re-publish keeps it
	if state != nil {
		newState.Keywords = state.Keywords
	}
	if err := SaveArticleState(mdPath, newState); err != nil {
		return fmt.Errorf("save state: %w", err)
	}
	fmt.Printf("✓ state saved to %s\n", articleStatePath(mdPath))
	return nil
}

// pickNodeModel resolves the target node model: configured/prior id → single
// available → interactive numbered pick.
func pickNodeModel(reader *bufio.Reader, spec *hpapi.Spec, configured string, state *ArticleState) *hpapi.NodeModel {
	if state != nil && state.ModelID != "" {
		if m := spec.NodeModelGet(state.ModelID); m != nil {
			return m
		}
	}
	if configured != "" {
		if m := spec.NodeModelGet(configured); m != nil {
			return m
		}
	}
	if len(spec.NodeModels) == 1 {
		return spec.NodeModels[0]
	}
	if len(spec.NodeModels) == 0 {
		return nil
	}
	fmt.Println("\nAvailable node models:")
	for i, m := range spec.NodeModels {
		t := m.Title
		if t == "" {
			t = m.Meta.Name
		}
		fmt.Printf("  [%d] %s (%s)\n", i+1, t, m.Meta.Name)
	}
	for {
		s := prompt(reader, "Select model", "")
		n, err := strconv.Atoi(s)
		if err == nil && n >= 1 && n <= len(spec.NodeModels) {
			return spec.NodeModels[n-1]
		}
		fmt.Println("invalid selection")
	}
}

// pickTermInteractive fetches a taxonomy's terms, prints the tree, and returns
// the selected term id. Pressing Enter keeps the defaultID (used on update).
func pickTermInteractive(reader *bufio.Reader, client *Client, modname string, tm hpapi.TermModel, defaultID string) (string, error) {
	tl, err := client.TermList(modname, tm.Meta.Name)
	if err != nil {
		return "", err
	}
	if tl.Error != nil {
		return "", fmt.Errorf("%s", tl.Error.Message)
	}
	if len(tl.Items) == 0 {
		return "", fmt.Errorf("no %s found", tm.Meta.Name)
	}

	flat := flattenTerms(tl.Items)
	title := tm.Title
	if title == "" {
		title = tm.Meta.Name
	}
	fmt.Printf("\n%s:\n", title)
	curIdx := -1
	for i, ft := range flat {
		mark := "  "
		if defaultID != "" && strconv.FormatUint(uint64(ft.term.ID), 10) == defaultID {
			curIdx = i
			mark = "▶ "
		}
		fmt.Printf("%s[%d] %s%s\n", mark, i+1, strings.Repeat("  ", ft.depth), ft.term.Title)
	}
	if curIdx >= 0 {
		fmt.Printf("(current selection: [%d]; Enter to keep)\n", curIdx+1)
	}
	s := prompt(reader, "Select", "")
	if s == "" {
		if defaultID != "" {
			return defaultID, nil
		}
		// no default: require a choice
		s = prompt(reader, "Select (required)", "")
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > len(flat) {
		if defaultID != "" {
			return defaultID, nil
		}
		return "", fmt.Errorf("invalid selection")
	}
	return strconv.FormatUint(uint64(flat[n-1].term.ID), 10), nil
}

type flatTerm struct {
	term  hpapi.Term
	depth int
}

func flattenTerms(items []hpapi.Term) []flatTerm {
	byID := map[uint32]*flatTerm{}
	order := []uint32{}
	for i := range items {
		byID[items[i].ID] = &flatTerm{term: items[i]}
		order = append(order, items[i].ID)
	}
	// children attach to parents; roots are PID==0 or unknown PID
	roots := []uint32{}
	children := map[uint32][]uint32{}
	for _, id := range order {
		pid := byID[id].term.PID
		if pid == 0 || byID[pid] == nil {
			roots = append(roots, id)
		} else {
			children[pid] = append(children[pid], id)
		}
	}
	var out []flatTerm
	var walk func(id uint32, depth int)
	walk = func(id uint32, depth int) {
		out = append(out, flatTerm{term: byID[id].term, depth: depth})
		for _, c := range children[id] {
			walk(c, depth+1)
		}
	}
	for _, r := range roots {
		walk(r, 0)
	}
	return out
}

// promptExtraFields prompts for every declared field that is not title/content.
func promptExtraFields(reader *bufio.Reader, model *hpapi.NodeModel) ([]*hpapi.NodeField, error) {
	var fields []*hpapi.NodeField
	for _, f := range model.Fields {
		if f.Name == "title" || f.Name == "content" {
			continue
		}
		label := f.Title
		if label == "" {
			label = f.Name
		}
		val := prompt(reader, label, "")
		if val == "" {
			continue
		}
		nf := &hpapi.NodeField{Name: f.Name, Value: val}
		if f.Type == "text" {
			nf.Attrs = types.KvPairs{&types.KvPair{Key: "format", Value: "md"}}
		}
		fields = append(fields, nf)
	}
	return fields, nil
}

// uploadReferencedImages finds every {{hp_storage_service_endpoint}}/<date>/<file>
// reference in md, uploads the matching local file, and returns the manifest.
// Any reference already present in prior (uploaded in a previous run) is skipped
// — its manifest entry is carried over unchanged and flagged "(skip)" on the
// terminal — so updates don't re-push images that are already on the server.
func uploadReferencedImages(client *Client, md, outDir string, prior []ArticleImage) ([]ArticleImage, error) {
	matches := mdImageRefRe.FindAllStringSubmatch(md, -1)
	seen := map[string]bool{}
	priorByPath := make(map[string]ArticleImage, len(prior))
	for _, im := range prior {
		priorByPath[im.StoragePath] = im
	}
	var images []ArticleImage
	uploaded, skipped := 0, 0
	for _, m := range matches {
		rel := m[1] // <date>/<file>
		storagePath := "/deft/" + rel
		if seen[storagePath] {
			continue
		}
		seen[storagePath] = true

		// already on the server per the prior manifest — skip the upload
		if im, ok := priorByPath[storagePath]; ok {
			skipped++
			images = append(images, im)
			fmt.Printf("  (skip) %s\n", storagePath)
			continue
		}

		localPath := filepath.Join(outDir, rel)
		data, err := os.ReadFile(localPath)
		if err != nil {
			return images, fmt.Errorf("read local image %s: %w (run extract first)", localPath, err)
		}
		if err := client.S2Put(storagePath, data); err != nil {
			return images, fmt.Errorf("upload %s: %w", storagePath, err)
		}
		uploaded++
		images = append(images, ArticleImage{
			Local:       localPath,
			StoragePath: storagePath,
			Ref:         "{{hp_storage_service_endpoint}}/" + rel + imageRefSuffix(rel),
		})
		fmt.Printf("  uploaded %s\n", storagePath)
	}
	if uploaded > 0 || skipped > 0 {
		fmt.Printf("uploaded %d image(s), skipped %d (already uploaded)\n", uploaded, skipped)
	}
	return images, nil
}

// defaultTitle returns the first "# heading" in md, else the file base name.
func defaultTitle(md, mdPath string) string {
	for line := range strings.SplitSeq(md, "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "# "); ok {
			return strings.TrimSpace(rest)
		}
	}
	base := filepath.Base(mdPath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// prompt prints label [def] and reads a trimmed line, returning def on empty.
func prompt(reader *bufio.Reader, label, def string) string {
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}
