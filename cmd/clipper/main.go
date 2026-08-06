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

// clipper: HTML -> markdown extraction, local preview, and hpress publish.
//
//	clipper extract <file.html> [--mode classic|llm] [--out DIR] [--llm-* ...]
//	clipper preview <file.md>  [--port N] [--open] [--out DIR]
//	clipper publish <file.md>  # publish (create) a node
//	clipper update  <file.md>  # update an existing node from its state file
//	clipper auth [flags]       # configure ~/.hooto-press.toml
//
// The access key ("ak_<id>_<secret>") is generated in IAM (user access-key CRUD);
// store it in ~/.hooto-press.toml via the "auth" subcommand.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const defaultOutDir = "./var/output"

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// newRootCmd builds the top-level command tree. SilenceUsage/SilenceErrors are
// set on the root so a failed RunE prints only the one-line error (cobra would
// otherwise dump usage on every runtime error); main() prints the error itself.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "clipper",
		Short: "HTML -> markdown extraction, local preview, and hpress publish",
		Long: "A command-line authoring -> publish pipeline for hpress: extract a " +
			"local HTML page to markdown (downloading and re-encoding its images), " +
			"preview it locally, then publish or update it to a hpress module.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: "  clipper extract article.html\n" +
			"  clipper extract --mode llm article.html\n" +
			"  clipper preview --open article.md\n" +
			"  clipper publish article.md\n" +
			"  clipper update  article.md\n" +
			"  clipper auth --key ak_<id>_<secret> --server http://localhost:9533 --module core/blog",
	}
	root.AddCommand(
		newExtractCmd(),
		newPreviewCmd(),
		newPublishCmd(),
		newUpdateCmd(),
		newAuthCmd(),
	)
	return root
}

// newExtractCmd: HTML -> markdown (+ downloaded images). The --mode flag selects
// the backend ('classic' default, or 'llm'); the --llm-* flags are per-run
// overrides on top of the [llm] config block, winning for a single run only.
func newExtractCmd() *cobra.Command {
	var (
		mode         string
		out          string
		llmModel     string
		llmTimeout   int
		llmThinking  string
		llmPrefilter string
	)
	cmd := &cobra.Command{
		Use:   "extract <file.html>",
		Short: "Convert an HTML file to markdown (+ downloaded images)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadClientConfig()
			if err != nil {
				return err
			}

			// resolve mode: explicit flag -> config -> "classic".
			effMode := mode
			if effMode == "" {
				effMode = cfg.Extract.Mode
			}
			if effMode == "" {
				effMode = "classic"
			}
			if effMode != "classic" && effMode != "llm" {
				return fmt.Errorf("invalid extract mode %q (want 'classic' or 'llm')", effMode)
			}

			// apply per-run overrides on top of the configured [llm] block.
			llm := cfg.LLM
			if llmModel != "" {
				llm.Model = llmModel
			}
			if llmTimeout > 0 {
				llm.Timeout = llmTimeout
			}
			if llmThinking == "on" || llmThinking == "off" {
				llm.EnableThinking = (llmThinking == "on")
			}
			if llmPrefilter == "on" || llmPrefilter == "off" {
				llm.Prefilter = llmPrefilter
			}

			return runExtract(args[0], out, effMode, llm)
		},
	}
	f := cmd.Flags()
	f.StringVar(&mode, "mode", "", "html->md backend: 'classic' or 'llm' (default: config [extract].mode, else 'classic')")
	f.StringVar(&out, "out", defaultOutDir, "image output directory")
	f.StringVar(&llmModel, "llm-model", "", "override [llm].model for this run")
	f.IntVar(&llmTimeout, "llm-timeout", 0, "override [llm].timeout (seconds) for this run")
	f.StringVar(&llmThinking, "llm-thinking", "", "override [llm].enable_thinking for this run: 'on' or 'off'")
	f.StringVar(&llmPrefilter, "llm-prefilter", "", "override [llm].prefilter for this run: 'on' or 'off'")
	return cmd
}

// newPreviewCmd renders a markdown file to HTML and serves it locally, mapping
// the {{hp_storage_service_endpoint}} placeholder to the local image directory.
func newPreviewCmd() *cobra.Command {
	var (
		out         string
		port        int
		openBrowser bool
	)
	cmd := &cobra.Command{
		Use:   "preview <file.md>",
		Short: "Render markdown to HTML and serve it locally with its images",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPreview(args[0], out, port, openBrowser)
		},
	}
	f := cmd.Flags()
	f.StringVar(&out, "out", defaultOutDir, "image output directory")
	f.IntVar(&port, "port", 9599, "preview server port")
	f.BoolVar(&openBrowser, "open", false, "open the preview in the default browser")
	return cmd
}

// newPublishCmd publishes a markdown file as a new node.
func newPublishCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "publish <file.md>",
		Short: "Publish markdown (+ images) to hpress as a new node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPublish(args[0], false)
		},
	}
}

// newUpdateCmd updates an already-published node from its saved state file.
func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update <file.md>",
		Short: "Update an already-published node from its saved state file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPublish(args[0], true)
		},
	}
}

// newAuthCmd writes/updates ~/.hooto-press.toml with the provided fields. Only
// the flags passed are applied; the rest of the config is preserved. --key and
// --server are required (a prior config may already satisfy them).
func newAuthCmd() *cobra.Command {
	var (
		key          string
		server       string
		module       string
		modelID      string
		out          string
		mode         string
		llmBaseURL   string
		llmAPIKey    string
		llmModel     string
		llmTimeout   int
		llmThinking  string
		llmPrefilter string
	)
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Configure ~/.hooto-press.toml (server, access key, module, llm backend)",
		Long: "Configure ~/.hooto-press.toml. Only the flags passed are changed; the " +
			"rest of the config is preserved. --key and --server are required (a prior " +
			"config may already satisfy them).",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadClientConfig()
			if err != nil {
				return err
			}
			if key != "" {
				cfg.Auth.AccessKey = key
			}
			if server != "" {
				cfg.Server.BaseURL = server
			}
			if module != "" {
				cfg.Publish.Module = module
			}
			if modelID != "" {
				cfg.Publish.ModelID = modelID
			}
			if out != "" {
				cfg.Publish.ImageOutDir = out
			}
			if mode != "" {
				cfg.Extract.Mode = mode
			}
			if llmBaseURL != "" {
				cfg.LLM.BaseURL = llmBaseURL
			}
			if llmAPIKey != "" {
				cfg.LLM.APIKey = llmAPIKey
			}
			if llmModel != "" {
				cfg.LLM.Model = llmModel
			}
			if llmTimeout > 0 {
				cfg.LLM.Timeout = llmTimeout
			}
			if llmThinking == "on" || llmThinking == "off" {
				cfg.LLM.EnableThinking = (llmThinking == "on")
			}
			if llmPrefilter == "on" || llmPrefilter == "off" {
				cfg.LLM.Prefilter = llmPrefilter
			}

			if cfg.Auth.AccessKey == "" || cfg.Server.BaseURL == "" {
				return fmt.Errorf("--key and --server are required")
			}
			if cfg.Publish.ImageOutDir == "" {
				cfg.Publish.ImageOutDir = defaultOutDir
			}

			if err := SaveClientConfig(cfg); err != nil {
				return err
			}
			path, _ := clientConfigPath()
			fmt.Println("saved config to", path)
			fmt.Println("publish requires IAM access-key scopes:", strings.Join(publishScopes, ", "))
			fmt.Println("  (app=<app_id> = hpress [iam_auth].app_id; bind them to the key in IAM)")
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&key, "key", "", `access key export "ak_<id>_<secret>" (from IAM)`)
	f.StringVar(&server, "server", "", "hpress base URL, e.g. http://localhost:9533")
	f.StringVar(&module, "module", "", "default target module, e.g. core/blog")
	f.StringVar(&modelID, "model", "", "default node model id, e.g. entry")
	f.StringVar(&out, "out", "", "image output directory")
	f.StringVar(&mode, "mode", "", "extract backend: 'classic' or 'llm'")
	f.StringVar(&llmBaseURL, "llm-base-url", "", "LLM API base URL, e.g. https://api.deepseek.com")
	f.StringVar(&llmAPIKey, "llm-api-key", "", "LLM API key")
	f.StringVar(&llmModel, "llm-model", "", "LLM model id, e.g. deepseek-chat")
	f.IntVar(&llmTimeout, "llm-timeout", 0, "LLM per-request timeout in seconds (default 600; reasoning models may need more)")
	f.StringVar(&llmThinking, "llm-thinking", "", "enable model reasoning phase: 'on' or 'off' (default off, recommended for HTML extraction)")
	f.StringVar(&llmPrefilter, "llm-prefilter", "", "strip noise HTML (scripts/styles/attrs) before the LLM: 'on' or 'off' (default on)")
	return cmd
}
