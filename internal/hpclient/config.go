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

package hpclient

import (
	"os"
	"path/filepath"
	"strings"
)

// EnvConfigFile overrides the CLI config file location, so multiple hpress
// servers can be addressed by switching the environment instead of rewriting
// the shared config.
const EnvConfigFile = "HOOTOPRESS_CONFIG_FILE"

// defaultConfigFile is used when EnvConfigFile is unset.
const defaultConfigFile = "~/.hooto-press.toml"

// ConfigFilePath returns the CLI config file path: $HOOTOPRESS_CONFIG_FILE if
// set, else ~/.hooto-press.toml. A leading "~/" in either is expanded to the
// user home directory.
func ConfigFilePath() (string, error) {
	p := strings.TrimSpace(os.Getenv(EnvConfigFile))
	if p == "" {
		p = defaultConfigFile
	}
	if p != "~" && !strings.HasPrefix(p, "~/") {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, strings.TrimPrefix(p, "~")), nil
}
