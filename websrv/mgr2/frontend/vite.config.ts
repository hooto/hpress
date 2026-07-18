import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

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

export default defineConfig({
  plugins: [svelte()],
  base: '/hp/mgr2/',
  build: {
    outDir: '../dist',
    emptyOutDir: true,
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
