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

// module-push: upload a module package to a hpress instance via
// /hp/v1/mod-set/spec-upload-commit. A module directory is packed first
// (module-build, the artifact is kept next to the module); a .ipk file is
// uploaded as-is.

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hooto/hpress/internal/hpclient"
)

// uploadSizeMax mirrors specUploadSizeMax in internal/api/modset-spec-upload.go.
const uploadSizeMax = 8 * 1024 * 1024

// pushScopes are the IAM access-key scopes a module-push key must carry.
var pushScopes = []string{
	"app=<app_id>", // cross-app binding; introspect rejects without it
	"editor.write", // mod-set/spec-upload-commit route gate
}

func newModulePushCmd() *cobra.Command {
	var server, key string
	cmd := &cobra.Command{
		Use:   "module-push <dir-or-file.ipk>",
		Short: "Build (if dir) and upload a module package to a hpress instance",
		Long: "Upload a module .ipk package to mod-set/spec-upload-commit. When " +
			"<path> is a module directory it is packed first (module-build); when " +
			"it is a .ipk file it is uploaded as-is. Credentials come from the CLI " +
			"config ($HOOTOPRESS_CONFIG_FILE, default ~/.hooto-press.toml; " +
			"[server].base_url + [auth].access_key, shared with the clipper CLI) " +
			"or the --server/--key flags.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModulePush(args[0], server, key)
		},
	}
	f := cmd.Flags()
	f.StringVar(&server, "server", "", "hpress base URL, e.g. http://localhost:9533")
	f.StringVar(&key, "key", "", `access key export "ak_<id>_<secret>" (from IAM)`)
	return cmd
}

func runModulePush(path, serverFlag, keyFlag string) error {

	cfg, err := loadCliConfig()
	if err != nil {
		return err
	}
	baseURL := serverFlag
	if baseURL == "" {
		baseURL = cfg.Server.BaseURL
	}
	accessKey := keyFlag
	if accessKey == "" {
		accessKey = cfg.Auth.AccessKey
	}
	if baseURL == "" || accessKey == "" {
		return fmt.Errorf("not configured: run `clipper auth --key ak_... --server <url>` or pass --server/--key")
	}

	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	var ipkName string
	var ipkData []byte
	if fi.IsDir() {
		fmt.Println("building package from", path)
		data, name, err := pkgBuild(path)
		if err != nil {
			return err
		}
		// keep the built artifact next to the module for inspection / manual re-push
		artifact := filepath.Join(path, name)
		if err := os.WriteFile(artifact, data, 0644); err != nil {
			return err
		}
		fmt.Println("built", artifact)
		ipkName, ipkData = name, data
	} else {
		if !strings.HasSuffix(path, ".ipk") {
			return fmt.Errorf("%s: expected a module directory or a .ipk file", path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		ipkName, ipkData = filepath.Base(path), data
	}

	if len(ipkData) > uploadSizeMax {
		return fmt.Errorf("package %s is %d bytes (max %d)", ipkName, len(ipkData), uploadSizeMax)
	}

	client, err := hpclient.NewClient(baseURL, accessKey)
	if err != nil {
		return err
	}

	fmt.Printf("uploading %s (%d bytes) to %s\n", ipkName, len(ipkData), baseURL)
	if err := client.SpecUploadCommit(ipkName, ipkData); err != nil {
		var ae *hpclient.ApiError
		if errors.As(err, &ae) && (ae.Code == "Unauthorized" || ae.Code == "AccessDenied") {
			return fmt.Errorf("%w\n  hint: IAM access key needs scopes: %s\n  (app=<app_id> = hpress [iam_auth].app_id)",
				err, strings.Join(pushScopes, ", "))
		}
		return err
	}
	fmt.Println("module package committed")
	return nil
}
