# hpress — module development CLI

Command-line workflow for hooto-press content modules: scaffold locally, edit
with any editor, pack into an innerstack v2 `.ipk` package, and upload it to a
hpress instance (the `mod-set/spec-upload-commit` endpoint).

```text
hpress module-init <dir>    scaffold a module (spec.json, ipk.toml, views/)
hpress module-build <dir>   pack the module dir into <name>_<version>_all_src.ipk
hpress module-push <path>   upload a .ipk (packing first when <path> is a dir)
```

## Workflow

```bash
# 1) scaffold: module name defaults to the last two dir segments (demo/hello)
hpress module-init modules/demo/hello --name demo/hello

# 2) edit spec.json / views/*.tpl with vim, an IDE, or anything else
#    (see doc/modules.md for the spec schema; modules/ruilog/notebook is a
#    larger reference; bump meta.version when uploading a change)

# 3) pack + upload (module-push on a dir does the module-build step for you)
hpress module-build modules/demo/hello
hpress module-push modules/demo/hello --server http://localhost:9533 --key ak_...

# or push an existing package file
hpress module-push modules/demo/hello/demo-hello_0.1.0_all_src.ipk
```

## Configuration

Server URL and access key are read from `~/.hooto-press.toml` (shared with the
clipper CLI — configure once with `clipper auth`):

```toml
[server]
base_url = "http://localhost:9533"

[auth]
access_key = "ak_<id>_<secret>"
```

The IAM access key must be bound to the hpress app and carry the `editor.write`
scope (`app=<app_id>` where `<app_id>` is the hpress `[iam_auth].app_id`).
`--server` / `--key` flags override the config for a single run.

## Package format

`module-build` writes the innerstack v2 IPK1 container natively (`IPK1` magic,
LE32 header length, JSON header, xz-compressed tar data block; the header
checksum is sha256 over the data block). The tar layout mirrors
`innerstack pkg-build` output — a `.` root, an `.ipk/metadata.json` descriptor,
then the `[build].include` entries — so the artifact is interchangeable with
one produced by the innerstack CLI. Uploads are capped at 8 MiB.
