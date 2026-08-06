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

// Client configuration and per-article state for the clipper publish CLI.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hooto/htoml4g/htoml"
)

// ClientConfig is the on-disk CLI config at ~/.hooto-press.toml.
type ClientConfig struct {
	Server  ClientServer  `toml:"server"`
	Auth    ClientAuth    `toml:"auth"`
	Publish ClientPublish `toml:"publish"`
	Extract ClientExtract `toml:"extract"`
	LLM     ClientLLM     `toml:"llm"`
}

type ClientServer struct {
	BaseURL string `toml:"base_url"`
}

type ClientAuth struct {
	AccessKey string `toml:"access_key"` // "ak_<id>_<secret>"
}

type ClientPublish struct {
	Module      string `toml:"module"`        // e.g. "core/blog"
	ModelID     string `toml:"model_id"`      // e.g. "entry"
	ImageOutDir string `toml:"image_out_dir"` // where extracted images live
}

// ClientExtract selects the HTML -> markdown conversion backend.
type ClientExtract struct {
	Mode string `toml:"mode"` // "classic" (default) | "llm"
}

// ClientLLM holds credentials for the DeepSeek-compatible OpenAI chat API used by
// the "llm" extract backend. Independent of [auth] (publish) credentials.
type ClientLLM struct {
	BaseURL string `toml:"base_url"` // e.g. "https://api.deepseek.com"
	APIKey  string `toml:"api_key"`
	Model   string `toml:"model"`   // e.g. "deepseek-chat"
	Timeout int    `toml:"timeout"` // overall per-request timeout in seconds (default 600)
	// EnableThinking controls the model's reasoning ("thinking") phase. false
	// (default) sends reasoning_effort="none", which skips the slow reasoning
	// phase reasoning models (deepseek-v4-flash, deepseek-reasoner) would
	// otherwise emit before the answer — ideal for HTML extraction. Set true to
	// let the model reason for hard layouts.
	EnableThinking bool `toml:"enable_thinking"`
	// Prefilter controls HTML sanitization before the LLM call: "" or "on"
	// (default) strip scripts/styles/noise attributes to save tokens; "off"
	// sends the raw HTML. Applies only to the llm backend.
	Prefilter string `toml:"prefilter"`
}

// clientConfigPath returns ~/.hooto-press.toml.
func clientConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".hooto-press.toml"), nil
}

// LoadClientConfig reads ~/.hooto-press.toml. A missing file yields a zero
// config + nil error so callers can detect "not configured" by empty fields.
func LoadClientConfig() (*ClientConfig, error) {
	path, err := clientConfigPath()
	if err != nil {
		return nil, err
	}
	cfg := &ClientConfig{}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if err := htoml.DecodeFromFile(path, cfg); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return cfg, nil
}

// SaveClientConfig writes ~/.hooto-press.toml.
func SaveClientConfig(cfg *ClientConfig) error {
	path, err := clientConfigPath()
	if err != nil {
		return err
	}
	return htoml.EncodeToFile(cfg, path, nil)
}

// ArticleState is the per-article manifest written next to the .md file
// (<file>.toml). It captures everything needed to re-publish or update the
// article, including the server-side node id.
type ArticleState struct {
	ServerBaseURL string            `toml:"server_base_url"`
	Module        string            `toml:"module"`
	ModelID       string            `toml:"model_id"`
	NodeID        string            `toml:"node_id"` // server unique id (empty until first publish)
	Title         string            `toml:"title"`
	Status        int16             `toml:"status"`
	Created       uint32            `toml:"created"`
	Updated       uint32            `toml:"updated"`
	Categories    map[string]string `toml:"categories"` // termModel name -> term id
	Tags          map[string]string `toml:"tags"`       // termModel name -> comma titles
	Images        []ArticleImage    `toml:"images"`
	Keywords      []string          `toml:"keywords"` // curated keyword list extracted from the source (llm mode); pre-publish this is the only populated field
}

type ArticleImage struct {
	Local       string `toml:"local"`        // local filesystem path
	StoragePath string `toml:"storage_path"` // "/deft/<date>/<file>"
	Ref         string `toml:"ref"`          // markdown placeholder ref
}

// articleStatePath returns "<mdPath with .md stripped>.toml".
func articleStatePath(mdPath string) string {
	ext := filepath.Ext(mdPath)
	if ext == ".md" {
		return strings.TrimSuffix(mdPath, ext) + ".toml"
	}
	return mdPath + ".toml"
}

// LoadArticleState reads the per-article state file (ok if missing → returns
// nil, nil).
func LoadArticleState(mdPath string) (*ArticleState, error) {
	path := articleStatePath(mdPath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	st := &ArticleState{}
	if err := htoml.DecodeFromFile(path, st); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return st, nil
}

// SaveArticleState writes the per-article state file.
func SaveArticleState(mdPath string, st *ArticleState) error {
	return htoml.EncodeToFile(st, articleStatePath(mdPath), nil)
}
