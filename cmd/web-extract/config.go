// Client configuration and per-article state for the web-extract publish CLI.

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
}

type ClientServer struct {
	BaseURL string `toml:"base_url"`
}

type ClientAuth struct {
	AccessKey string `toml:"access_key"` // "ak_<id>_<secret>"
}

type ClientPublish struct {
	Module     string `toml:"module"`      // e.g. "core/blog"
	ModelID    string `toml:"model_id"`    // e.g. "entry"
	ImageOutDir string `toml:"image_out_dir"` // where extracted images live
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
	NodeID        string            `toml:"node_id"` // server unique id
	Title         string            `toml:"title"`
	Status        int16             `toml:"status"`
	Created       uint32            `toml:"created"`
	Updated       uint32            `toml:"updated"`
	Categories    map[string]string `toml:"categories"` // termModel name -> term id
	Tags          map[string]string `toml:"tags"`       // termModel name -> comma titles
	Images        []ArticleImage    `toml:"images"`
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
