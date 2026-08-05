// web-extract: HTML → markdown extraction, local preview, and hpress publish.
//
// Usage:
//
//	web-extract <file.html>                         # extract to <file>.md + images
//	web-extract --preview <file.md> [--port N] [--open]
//	web-extract --publish <file.md>                 # publish (create)
//	web-extract --update  <file.md>                 # update an existing node
//	web-extract auth --key ak_<id>_<secret> --server <url> [--module <mod>] [--out <dir>]
//
// Flags are parsed before the positional file. The access key is generated in
// IAM (user access-key CRUD); store its "ak_<id>_<secret>" export in
// ~/.hooto-press.toml via the "auth" subcommand.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const defaultOutDir = "./var/output"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {

	// "auth" subcommand: configure ~/.hooto-press.toml
	if len(os.Args) > 1 && os.Args[1] == "auth" {
		return runAuth(os.Args[2:])
	}

	fs := flag.NewFlagSet("web-extract", flag.ContinueOnError)
	preview := fs.Bool("preview", false, "render <file.md> to HTML and serve locally")
	publish := fs.Bool("publish", false, "publish <file.md> to hpress (create)")
	update := fs.Bool("update", false, "update an already-published node from its state file")
	out := fs.String("out", defaultOutDir, "image output directory (extract/preview)")
	port := fs.Int("port", 9599, "preview server port")
	openBrowser := fs.Bool("open", false, "open the preview in the default browser")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}
	file := fs.Arg(0)
	if file == "" {
		fs.Usage()
		return fmt.Errorf("no input file")
	}

	switch {
	case *preview:
		return runPreview(file, *out, *port, *openBrowser)
	case *publish:
		return runPublish(file, false)
	case *update:
		return runPublish(file, true)
	default:
		return runExtract(file, *out)
	}
}

// runAuth writes/updates ~/.hooto-press.toml with the provided fields.
func runAuth(args []string) error {
	fs := flag.NewFlagSet("auth", flag.ContinueOnError)
	key := fs.String("key", "", `access key export "ak_<id>_<secret>" (from IAM)`)
	server := fs.String("server", "", "hpress base URL, e.g. http://localhost:9533")
	module := fs.String("module", "", "default target module, e.g. core/blog")
	modelID := fs.String("model", "", "default node model id, e.g. entry")
	out := fs.String("out", "", "image output directory")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := LoadClientConfig()
	if err != nil {
		return err
	}
	if *key != "" {
		cfg.Auth.AccessKey = *key
	}
	if *server != "" {
		cfg.Server.BaseURL = *server
	}
	if *module != "" {
		cfg.Publish.Module = *module
	}
	if *modelID != "" {
		cfg.Publish.ModelID = *modelID
	}
	if *out != "" {
		cfg.Publish.ImageOutDir = *out
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
}
