# clipper

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
go build -o ./bin/clipper ./cmd/clipper/
```

> The source directory and command name are `clipper`; the packaged /
> distributed binary is named `hp-clipper` (the `hp-` prefix is added at
> packaging time).

> Requires the local iam checkout (`replace github.com/hooto/iam/v2` in
> `go.mod`), which provides the access-key request-signing support.

## Usage

```bash
# 1. extract a saved HTML page to markdown (+ images under ./var/output/<date>/)
./bin/clipper extract article.html

# 1b. extract using the LLM backend (see "Extract backend" below) instead of classic
./bin/clipper extract --mode llm article.html

# 2. preview the markdown locally (renders to HTML, serves images from ./var/output)
./bin/clipper preview article.md --open            # opens browser
./bin/clipper preview article.md --port 8080

# 3. publish a new node
./bin/clipper publish article.md

# 4. update an already-published node (reads article.toml)
./bin/clipper update article.md
```

Each operation is its own subcommand. Run `./bin/clipper --help` (or
`./bin/clipper <command> --help`) for the full flag list.

## Extract backend: classic vs LLM

Extraction has two HTML → markdown backends, selected by `--mode classic|llm`
(default `classic`, or `[extract].mode` from the config when the flag is omitted):

- **classic** (default) — the rule-based `JohannesKaufmann/html-to-markdown`
  converter. Deterministic, offline, no extra credentials.
- **llm** — sends the (site-cleaned) HTML to a DeepSeek-compatible OpenAI Chat
  Completions API, which extracts the article body and emits clean, structured
  markdown, dropping navigation/header/footer/sidebar chrome. Useful for pages
  whose layout defeats the rule-based converter. In addition to a faithful
  conversion, the model fixes obvious typos and broken formatting it encounters
  in the source and prints a list of every correction for human review (see
  "Correction review"). Non-deterministic by nature (run with `temperature: 0`);
  trades reproducibility for layout resilience.

Both backends share the **same** image download/re-encode + rewrite step, so
image handling (`./var/output/<date>/<hash>.jpg`, `{{hp_storage_service_endpoint}}`
references) is identical regardless of mode. Only the HTML → markdown text step
differs.

The LLM backend needs credentials in `~/.hooto-press.toml` (independent of the
publish `[auth]` key):

```bash
./bin/clipper auth \
  --mode llm \
  --llm-base-url https://api.deepseek.com \
  --llm-api-key sk-... \
  --llm-model deepseek-chat
```

This adds the following blocks to `~/.hooto-press.toml`:

```toml
[extract]
mode = "llm"            # "classic" (default) | "llm"

[llm]
base_url = "https://api.deepseek.com"
api_key = "sk-..."
model = "deepseek-chat"  # defaults to "deepseek-chat" if omitted
timeout = 600            # per-request timeout in seconds (default 600); 0 also means default
enable_thinking = false  # default off; set true to let reasoning models "think"
prefilter = "on"         # default on; "off" sends raw HTML to the model
```

`base_url` may be any OpenAI-compatible endpoint. With `mode = "llm"` set in the
config, a plain `./bin/clipper extract article.html` uses the LLM backend;
pass `extract --mode classic` to override per-run. If `[llm]` is unconfigured and
you run `extract --mode llm`, the tool errors with the `auth` command to run.

The `model`, `timeout`, `enable_thinking`, and `prefilter` settings persist via
the `auth` subcommand (`auth --llm-timeout 1200`, `auth --llm-thinking on|off`
default `off`, `auth --llm-prefilter on|off` default `on`). Each is also available
as a **per-run override** on the `extract` command (empty = use the config), winning
for a single run:

```bash
./bin/clipper extract --mode llm --llm-prefilter off --llm-thinking on article.html
./bin/clipper extract --mode llm --llm-model deepseek-chat --llm-timeout 900 article.html
```

### HTML pre-filtering (default on)

Before the HTML is sent to the model it is sanitized to cut token cost and help
the model focus on content: `<script>`, `<style>`, `<noscript>`, `<iframe>`,
`<svg>`, `<template>`, `<form>`, `<link>`/`<meta>`, and similar non-content
elements are dropped; HTML comments are removed; and every attribute irrelevant
to markdown (`class`, `style`, `id`, `data-*`, `on*`, `aria-*`, `srcset`,
`integrity`, ...) is stripped — keeping only the markdown-relevant ones
(`src`, `href`, `alt`, `title`, `colspan`, ...). Article text, headings, lists,
tables, code, links, and images are all preserved. This typically removes the
majority of the bytes on a real page.

This runs only for the llm backend (the classic converter is cheap and reads the
raw HTML directly). The byte reduction is logged on stderr
(`llm: pre-filter N -> M bytes (-P%)`). Turn it off if a page's content is
encoded in a way the filter drops (rare): `auth --llm-prefilter off`.

> **Note**: the LLM must preserve image URLs verbatim (the prompt enforces this).
> A very large page may exceed the model's context window; on truncation the tool
> aborts with a message to use `--mode classic` instead.

### Thinking (reasoning) is off by default

Reasoning models (`deepseek-v4-flash`, `deepseek-reasoner`, …) emit a long
"thinking" phase before the answer, which is slow and unnecessary for HTML
extraction. By default the tool sends `reasoning_effort = "none"`, skipping that
phase — the model returns the markdown directly. This is accepted by both
reasoning and non-reasoning DeepSeek models, so it is safe regardless of model.

The response is **streamed**, with progress on stderr (`model=..., thinking=...` →
`reasoning...` → `receiving response...` → `done`). The model returns a JSON
envelope `{"markdown", "changes"}`; only the decoded markdown is kept, and the
reported corrections are printed for review (see "Correction review"). Any
reasoning the model still emits (when thinking is on) is discarded.

Turn thinking back on only if a messy layout needs the extra understanding:

```bash
./bin/clipper auth --llm-thinking on     # enable (slower; raise --llm-timeout if needed)
./bin/clipper auth --llm-thinking off    # disable (default, faster)
```

### Correction review

Besides a faithful conversion, the LLM backend is instructed to fix **obvious**
defects in the source body text and to report each one:

- **Typos / common misspellings** — unambiguous errors only (e.g. `teh`→`the`,
  `recieve`→`receive`). It will not touch valid domain terms, proper names,
  identifiers, code, or British/American spelling variants, and it leaves any
  "fix" that could change meaning alone.
- **Broken formatting** — malformed/unclosed tags, broken list/table/code
  structure, a missing or wrong code-fence language, doubled or stray
  characters, broken HTML entity encoding. It does not reflow already-correct
  content.

After writing the `.md`, the tool prints the correction list to **stderr** for
human review before publishing — for example:

```
llm: corrections applied (2) — review before publishing:
  1. typo: 'recieve' -> 'receive' (paragraph 3)
  2. repaired broken <ul> nesting in section 2
```

When the model made no changes it prints
`llm: no corrections applied (body copied verbatim)`.

The list is the model's **self-report**; treat it as a prompt to eyeball the
diff, not as ground truth. The fidelity rules (verbatim text, byte-for-byte
image URLs, exact heading levels, dropped chrome) still hold — correction is the
only permitted deviation. If the model ever ignores the JSON contract and
returns plain markdown, the conversion still succeeds; only the correction list
is lost (logged as `json=false` on the `done` line).

### Keyword extraction (llm backend)

When the source page itself carries a **curated keyword/tag list** — a compact,
explicitly delimited set of short index terms the author set apart from the
prose (recognized by structure, not by an exact label) — the model copies it
verbatim into the response and the tool writes it to the per-article state file
next to the markdown (`<file>.toml`), for example:

```toml
node_id = ""                              # empty until you publish
keywords = ["distributed cache", "go", "concurrency"]
```

The body markdown is left untouched — the list is captured *in addition*, not
removed. The list is matched by shape (a delimited run of short index terms set
off from the prose — e.g. `Tags: go, cache, concurrency`), under whatever label
the site uses: `关键词` / `关键字` / `核心词` / `主题词` / `标签` / `索引词`,
`Keywords` / `Key words` / `Tags` / `Labels` / `Subjects` / `Topics` /
`Index terms`, and so on. If the page has no such list, nothing is written and
the tool prints `llm: no curated keyword list found`. The model is instructed
not to mine, summarize, or invent keywords from body prose; it only lifts a list
the author deliberately assembled.

This is the **same** `<file>.toml` that `publish`/`update` use:

- On a first extract `node_id` is empty (no publish yet); keywords is the only
  populated field.
- Publishing fills in `node_id` and **preserves** the keywords.
- On `publish`, the extracted keywords are offered as the default for the
  module's **tag** term model, so they flow straight into the article's tags —
  edit or clear at the prompt as needed.
- Re-extracting a page refreshes the keywords while keeping any already-published
  `node_id`.

### Model choice

For plain HTML → markdown extraction a **non-reasoning** model is fastest and
cheapest and produces identical structure — prefer `deepseek-chat`. Use a
reasoning model (with thinking on) only for hard layouts.

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
./bin/clipper auth \
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
./bin/clipper publish article.md
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
./bin/clipper update article.md
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
| `main.go` | cobra command tree (`extract`, `preview`, `publish`, `update`, `auth`) |
| `extract.go` | HTML → markdown pipeline: site cleanup, classic converter, image download/re-encode (UTC+8 date dirs) |
| `llm.go` | LLM (DeepSeek-compatible) HTML → markdown backend + prompt |
| `preview.go` | local markdown → HTML preview server |
| `client.go` | access-key-token hpress REST client (spec/term/node/s2-obj) |
| `publish.go` | interactive spec-driven publish + update flow |
| `config.go` | `~/.hooto-press.toml` + per-article `*.toml` state |
