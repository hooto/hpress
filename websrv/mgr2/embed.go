// Copyright 2015 Eryx <evorui at gmail dot com>, All rights reserved.
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

// Package mgr2 serves the Svelte-based management backend at /hp/mgr2.
// The frontend is a Vite + Svelte 5 SPA whose built output (dist/) is embedded
// into the binary via go:embed. It reuses the existing /hp/v1 REST API and the
// IAM management-permission gate, and leaves the legacy /hp/mgr (lynkui/jQuery)
// backend untouched — both coexist.
package mgr2

import "embed"

// distFS holds the Vite build output (index.html + hashed assets/). The dist
// directory is produced by `pnpm build` in frontend/. A committed placeholder
// dist/index.html keeps `go build` working before the first frontend build.
//
//go:embed all:dist
var distFS embed.FS
