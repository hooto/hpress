# web-extract

A command-line authoring → publish pipeline for hpress:

1. **Extract** a local `.html` to markdown, downloading and re-encoding its
   images (saved under `YYYY/MM/DD` in UTC+8).
2. **Preview** the resulting `.md` rendered to HTML with the local images.
3. **Publish** the `.md` + images to a hpress module via the REST API, with
   interactive category selection driven by the module's live spec.
4. **Update** an already-published article from its saved state file.

## Build

From the hpress repo root:

```bash
go build -o ./bin/web-extract ./cmd/web-extract/
```

> Requires the local iam checkout (`replace github.com/hooto/iam/v2` in
> `go.mod`), which provides the access-key request-signing support.

## Usage

```bash
# 1. extract a saved HTML page to markdown (+ images under ./var/output/<date>/)
./bin/web-extract article.html

# 2. preview the markdown locally (renders to HTML, serves images from ./var/output)
./bin/web-extract --preview article.md --open            # opens browser
./bin/web-extract --preview article.md --port 8080

# 3. publish a new node
./bin/web-extract --publish article.md

# 4. update an already-published node (reads article.toml)
./bin/web-extract --update article.md
```

Flags are parsed before the positional file.

## One-time setup

### 1. Create an IAM access key

In IAM, create a **user access key** (user access-key CRUD). You get an export
string of the form `ak_<id>_<secret>` **once** — copy it.

The key must carry **scopes** that grant both (a) the cross-app binding and (b)
the hpress operations it needs. When IAM introspects the key it copies the key's
scopes into the session's permissions, and hpress's route gates check those. So
grant the key these scopes:

- `app=<hpress app_id>` — the hpress instance's `app_id` (= `[iam_auth].app_id`
  in `etc/config.toml`, also the `instance_id`). Required: a key minted for one
  app cannot be used against another.
- `editor.read` — fetch the module spec.
- `editor.list` — fetch categories/terms.
- `editor.write` — create/update the node **and** upload images (`s2-obj/put`
  is gated on `editor.write`).

So a publish-capable key's scopes look like:
`app=<app_id>, editor.read, editor.list, editor.write`.

### 2. Configure the CLI

```bash
./bin/web-extract auth \
  --key ak_<id>_<secret> \
  --server http://localhost:9533 \
  --module core/blog \
  --model entry \
  --out ./var/output
```

This writes `~/.hooto-press.toml`. Re-run with any flag to update individual
fields.

### 3. Publish

```bash
./bin/web-extract --publish article.md
```

The flow:
- fetches the module spec (`core/blog`) to discover fields + term models;
- for each **taxonomy** term model (e.g. `categories`) it fetches the term tree
  and prompts you to pick one; for each **tag** term model it prompts for a
  comma-separated list;
- uploads every image referenced in the markdown to hpress storage
  (`/deft/<date>/<hash>.jpg`);
- creates the node (`status=1`, content as a `format:md` text field) and prints
  the server-assigned node id.

On success it writes `article.toml` next to the markdown, capturing the module,
model, node id, title, categories, tags, and the uploaded image manifest.

### 4. Update

Edit `article.md`, then:

```bash
./bin/web-extract --update article.md
```

This re-reads `article.toml` for the node id, re-uploads any new/changed
images, and updates the node in place.

## How authentication works

The CLI holds an IAM **user access key** (`ak_<id>_<secret>`). For each request
it mints a short-lived access-token JWT signed with the key's secret (HS256,
`Kid = <access-key id>`, no `Sub` claim) and sends it as
`Authorization: Bearer <token>`. The access-key **secret never leaves the CLI**.

hpress resolves the request via `iamserver.AppVerifier.Resolve` (in
`websrv/web/middleware.go`), which routes a token without a `Sub` claim to IAM's
`/v2/open/app-auth/introspect`. IAM verifies the token against its key store (the
only holder of the secret), checks the key's `app=<app_id>` scope binds it to
this app, and returns the owner's identity with the key's scopes as permissions.
The resolved identity is cached by access-key id for ~60s. Browser session
cookies (which carry a `Sub` claim) take the existing session path unchanged.

> Build note: hpress `go.mod` uses a local `replace` of `github.com/hooto/iam/v2`
> so the `Resolve`/introspect support is picked up.

## Files

| File | Purpose |
|---|---|
| `main.go` | flag/subcommand dispatch (`auth`, `--preview/--publish/--update`) |
| `extract.go` | HTML → markdown + image download/re-encode (UTC+8 date dirs) |
| `preview.go` | local markdown → HTML preview server |
| `client.go` | access-key-token hpress REST client (spec/term/node/s2-obj) |
| `publish.go` | interactive spec-driven publish + update flow |
| `config.go` | `~/.hooto-press.toml` + per-article `*.toml` state |
