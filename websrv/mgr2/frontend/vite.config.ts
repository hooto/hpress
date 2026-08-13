import { createHash } from 'node:crypto'
import fs from 'node:fs'
import path from 'node:path'
import { defineConfig, type Plugin } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import purgeCss from 'vite-plugin-purgecss'
// @ts-expect-error subset-font ships no type declarations
import subsetFont from 'subset-font'

// The built SPA is embedded into the Go binary (websrv/mgr2/dist) and served
// at /hp/mgr2. `base` must match the fiber mount so generated asset URLs
// resolve. outDir is outside the frontend root, so emptyOutDir is explicit.
//
// Dev mode (`pnpm dev`): Vite serves the SPA at http://localhost:5173/hp/mgr2/
// and proxies every other /hp/* backend route to the Go server started
// separately on :9533 (REST API /hp/v1, object storage /hp/s2, IAM auth
// /hp/user-auth). Override the backend address with HPRESS_BACKEND.
// /hp/mgr2/* stays on Vite (the app shell + HMR), via bypass.
const BACKEND = process.env.HPRESS_BACKEND || 'http://localhost:9533'

// Purge unused CSS from the final bundle. The biggest contributors are
// bootstrap.min.css (~228 KB) and bootstrap-icons.css (~100 KB, ~2000 icons
// of which the SPA uses ~20). vite-plugin-purgecss runs in generateBundle
// (build only; dev/HMR is untouched) and scans the whole emitted JS + the
// index.html as content, so Svelte's hashed scoped classes and every class
// literal in the bundle are seen and kept.
//
// extractor: the plugin labels the bundle content extension:"html", which
// makes PurgeCSS use its HTML extractor (reads class="..." only). That misses
// classes injected via JS (e.g. classList.add('modal-open') in lib/Modal.svelte).
// Register a plain identifier extractor so every class token in the JS string
// pool is captured. Costs a few KB of over-keeping but removes the whole
// "JS-only class got purged" failure mode.
//
// safelist: :root / html / body host Bootstrap's --bs-* custom properties.
// They are selectors, not content tokens, so no extractor sees them; without
// an explicit keep PurgeCSS would strip every Bootstrap CSS variable.
const purgeCssOptions = {
  extractors: [
    {
      extractor: (content: string) => content.match(/[A-Za-z0-9_-]+/g) ?? [],
      extensions: ['html'],
    },
  ],
  safelist: {
    standard: [/^:root$/, /^html$/, /^body$/],
  },
}

// kebab-case a chunk/asset stem: S2Section -> s2-section,
// NodeSection -> node-section. Already-lowercase names (codemirror,
// node-editor, fonts, images) pass through unchanged.
const toKebab = (s: string) =>
  s
    .replace(/([a-z0-9])([A-Z])/g, '$1-$2')
    .replace(/([A-Z])([A-Z][a-z])/g, '$1-$2')
    .toLowerCase()

// Subset the bootstrap-icons font to only the glyphs actually rendered.
// PurgeCSS already trims the .bi-* CSS rules, but the font binary still carries
// all ~2000 icons (~314 KB woff2+woff). This build-only plugin runs AFTER
// purgecss (both enforce:post, array order), scans the already-purged CSS for
// the PUA codepoints of surviving .bi-* rules, subsets the woff2 to just those
// glyphs (134 KB -> ~1.5 KB), and drops the legacy .woff fallback (woff2 is
// universally supported). The subset is re-emitted under a content-hashed name
// so the filename tracks its much smaller bytes for cache-busting.
function bootstrapIconsSubset(): Plugin {
  let outDir = ''
  const orphans: string[] = []
  return {
    name: 'hpress:subset-bootstrap-icons',
    apply: 'build',
    enforce: 'post',
    configResolved(config) {
      outDir = path.resolve(config.root, config.build.outDir)
    },
    async generateBundle(_options, bundle) {
      const files = Object.keys(bundle)

      // 1. Collect the PUA chars referenced by surviving .bi-* rules.
      const chars = new Set<string>()
      for (const f of files) {
        const a = bundle[f]
        if (f.endsWith('.css') && a.type === 'asset') {
          for (const m of String(a.source).matchAll(
            /\.bi-[a-z0-9-]+:before\{content:"([^"]*)"/g,
          )) {
            for (const ch of m[1]) chars.add(ch)
          }
        }
      }
      if (chars.size === 0) return
      const text = [...chars].join('')

      // 2. Subset the woff2 to exactly those glyphs.
      const woff2Key = files.find(
        (f) => f.includes('bootstrap-icons') && f.endsWith('.woff2'),
      )
      if (!woff2Key) return
      const woff2 = bundle[woff2Key]
      if (woff2.type !== 'asset') return
      const subset = Buffer.from(
        await subsetFont(Buffer.from(woff2.source as Uint8Array), text, {
          targetFormat: 'woff2',
        }),
      )

      // 3. Emit the subset under a content-hashed name. Rolldown forbids adding
      //    or removing bundle keys in generateBundle, so the original font
      //    assets are neutralized by emptying their source (a 0-byte file
      //    embeds nothing into the Go binary) rather than by deletion.
      const hash = createHash('sha1').update(subset).digest('hex').slice(0, 8)
      this.emitFile({
        type: 'asset',
        fileName: `assets/bootstrap-icons-${hash}.woff2`,
        source: subset,
      })
      ;(bundle[woff2Key] as { source: Uint8Array }).source = Buffer.alloc(0)
      orphans.push(woff2Key)

      // 4. Empty the .woff fallback too (woff2 is universally supported).
      const woffKey = files.find(
        (f) =>
          f.includes('bootstrap-icons') &&
          f.endsWith('.woff') &&
          !f.endsWith('.woff2'),
      )
      if (woffKey) {
        ;(bundle[woffKey] as { source: Uint8Array }).source = Buffer.alloc(0)
        orphans.push(woffKey)
      }

      // 5. Rewrite CSS: drop the .woff src entry and repoint the woff2 url at
      //    the new hashed filename.
      for (const f of files) {
        const a = bundle[f]
        if (!f.endsWith('.css') || a.type !== 'asset') continue
        let css = String(a.source)
        css = css.replace(/,url\([^)]+\)\s*format\("woff"\)/g, '')
        css = css.replace(/(bootstrap-icons-)[\w-]+(\.woff2)/g, `$1${hash}$2`)
        a.source = css
      }
    },
    writeBundle() {
      // Rolldown forbids deleting bundle keys, so the original fonts were
      // emptied (0 bytes) in generateBundle rather than removed. Now that all
      // files are written, delete those placeholder files from disk.
      for (const f of orphans) {
        fs.rmSync(path.join(outDir, f), { force: true })
      }
    },
  }
}

export default defineConfig({
  plugins: [svelte(), purgeCss(purgeCssOptions), bootstrapIconsSubset()],
  base: '/hp/mgr2/',
  build: {
    outDir: '../dist',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        // CodeMirror core (lib/codemirror.js) is imported by the node editor
        // (HpEditor). By default it lands tangled with marked/DOMPurify in the
        // node-section chunk; extract the core into its own chunk so the big
        // editor lib is cached independently and node-section stays lean.
        // Modes/addons stay with the editor. Matching only node_modules vendor
        // leaves avoids the manualChunks-on-app-code trap.
        manualChunks(id) {
          if (id.includes('/node_modules/codemirror/lib/')) return 'codemirror'
        },
        // chunkFileNames governs emitted JS chunk filenames. "markdown" is an
        // auto-chunk Rolldown names after the markdown mode; rename it to its
        // functional role (node-editor), then kebab-case every name so
        // S2Section -> s2-section, NodeSection -> node-section. The entry chunk
        // (index) uses entryFileNames, so it is unaffected. Pure filename
        // mapping — no chunk restructuring, no graph change.
        chunkFileNames: (chunkInfo) =>
          `assets/${toKebab(chunkInfo.name === 'markdown' ? 'node-editor' : chunkInfo.name)}-[hash].js`,
        // assetFileNames governs extracted CSS (and fonts/images). CSS takes its
        // name from the JS chunk, so kebab it too — s2-section.css sits beside
        // s2-section.js. Fonts/images are already lowercase, so unaffected.
        assetFileNames: (assetInfo) => {
          const stem = (assetInfo.name ?? '').replace(/\.[^.]+$/, '')
          return `assets/${toKebab(stem)}-[hash][extname]`
        },
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/hp': {
        target: BACKEND,
        changeOrigin: true,
        ws: false, // leave Vite's own HMR socket alone
        // Serve the SPA shell + assets from Vite, proxy everything else.
        bypass(req) {
          if (req.url?.startsWith('/hp/mgr2')) return req.url
        },
      },
    },
  },
})
