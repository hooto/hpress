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

// hpress: module development CLI for hooto-press.
//
//	hpress module-init <dir>   scaffold a content module (spec.json, ipk.toml, views/)
//	hpress module-build <dir>  pack the module dir into an innerstack v2 .ipk package
//	hpress module-push <path>  upload a module package (building it first when
//	                           <path> is a module directory)
//
// Server credentials live in the CLI config file ($HOOTOPRESS_CONFIG_FILE,
// default ~/.hooto-press.toml; shared with the clipper CLI — configure them
// once via `clipper auth --key ak_... --server <url>`).
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

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
		Use:   "hpress",
		Short: "hpress module development CLI (init, package, upload)",
		Long: "A command-line module workflow for hooto-press: module-init " +
			"generates a module skeleton (spec.json, ipk.toml, views/), " +
			"module-build packs it into an innerstack v2 .ipk package, and " +
			"module-push uploads the package to a hpress instance.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: "  hpress module-init modules/demo/hello\n" +
			"  hpress module-build modules/demo/hello\n" +
			"  hpress module-push modules/demo/hello --server http://localhost:9533 --key ak_...",
	}
	root.AddCommand(
		newModuleInitCmd(),
		newModuleBuildCmd(),
		newModulePushCmd(),
	)
	return root
}
