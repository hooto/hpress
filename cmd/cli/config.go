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

// CLI config: the [server]+[auth] subset of ~/.hooto-press.toml — the config
// file shared with the clipper CLI (`clipper auth` writes it; sections this
// tool does not use, such as [publish] or [llm], are silently ignored).

package main

import (
	"fmt"
	"os"

	"github.com/hooto/htoml4g/htoml"

	"github.com/hooto/hpress/internal/hpclient"
)

type cliConfig struct {
	Server cliConfigServer `toml:"server"`
	Auth   cliConfigAuth   `toml:"auth"`
}

type cliConfigServer struct {
	BaseURL string `toml:"base_url"`
}

type cliConfigAuth struct {
	AccessKey string `toml:"access_key"` // "ak_<id>_<secret>"
}

// loadCliConfig reads the CLI config ($HOOTOPRESS_CONFIG_FILE, default
// ~/.hooto-press.toml; shared with the clipper CLI). A missing file yields a
// zero config + nil error so callers detect "not configured" by empty fields.
func loadCliConfig() (*cliConfig, error) {
	path, err := hpclient.ConfigFilePath()
	if err != nil {
		return nil, err
	}
	cfg := &cliConfig{}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if err := htoml.DecodeFromFile(path, cfg); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return cfg, nil
}
