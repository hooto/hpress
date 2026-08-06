# Copyright 2015 Eryx <evorui at gmail dot com>, All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: all build_frontend build_backend build clipper release run run-be run-fe check install-deps version clean

# Vite outDir; the Go binary embeds this directory via go:embed.
FRONTEND_DIR := websrv/mgr2/frontend
DIST_DIR     := websrv/mgr2/dist
BINARY       := bin/hooto-press
MAIN         := ./cmd/server/main.go
PKG_MGR      := pnpm

# Full production build: frontend first (→ DIST_DIR, embedded into the binary),
# then the Go binary, so the shipped binary carries the current UI.
all: build_frontend build_backend
	@echo ""
	@echo "Build complete!  Binary: $(BINARY)"
	@echo ""

build:
	@$(MAKE) all

build_frontend:
	@echo "Building frontend (mgr2 SPA → $(DIST_DIR))..."
	cd $(FRONTEND_DIR) && $(PKG_MGR) build

build_backend:
	@echo "Building backend..."
	CGO_ENABLED=0 go build -o $(BINARY) $(MAIN)

clipper:
	@echo "Building clipper..."
	CGO_ENABLED=0 go build -o bin/clipper cmd/clipper/*.go

# Stripped, static linux/amd64 build for deployment (embeds current frontend).
release: build_frontend
	@echo "Cross-compiling release (linux/amd64, stripped)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "-s -w" -o $(BINARY) $(MAIN)

# Dev: run backend (:9533) + Vite dev server (:5173) together; Ctrl+C stops both.
# The backend embeds a committed placeholder dist, so no frontend build is
# required here — the live UI is served by Vite with HMR and proxies /hp/* API
# calls to the backend.
run: build_backend
	@echo "Starting backend + frontend dev (Ctrl+C to stop both)..."
	@echo "  backend  : http://localhost:9533"
	@echo "  frontend : http://localhost:5173/hp/mgr2/"
	(trap 'kill 0' SIGINT; ./$(BINARY) -logtostderr=true & cd $(FRONTEND_DIR) && $(PKG_MGR) dev)

run-be: build_backend
	./$(BINARY) -logtostderr=true

run-fe:
	cd $(FRONTEND_DIR) && $(PKG_MGR) dev

# Frontend type-check (svelte-check + tsc).
check:
	cd $(FRONTEND_DIR) && $(PKG_MGR) check

install-deps:
	@echo "Installing frontend dependencies..."
	cd $(FRONTEND_DIR) && $(PKG_MGR) install
	@echo "Installing backend dependencies..."
	go mod download
	@echo "Dependencies installed!"

# Print config.Version/Release baked into the binary.
version: build_backend
	./$(BINARY) version

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY)
	rm -rf $(DIST_DIR)
	@echo "Clean complete!"
