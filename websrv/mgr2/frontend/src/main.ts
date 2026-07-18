import { mount } from 'svelte'
import 'bootstrap/dist/css/bootstrap.min.css'
import 'bootstrap/dist/js/bootstrap.bundle.min.js'
import 'bootstrap-icons/font/bootstrap-icons.css'
import { marked } from 'marked'

import './styles/hp-main-v2.css'
import './styles/hpm-main.css'
import './styles/hpm-defx.css'
import App from './App.svelte'

// marked config mirrors webui/hpm/js/main.js:98 (gfm, smartLists, smartypants).
// marked v14 dropped `tables` (gfm enables tables) and the legacy `sanitize`
// option (the editor preview sanitizes via DOMPurify directly); the typed
// MarkedOptions also no longer lists smartLists/smartypants, so cast to keep
// the legacy toggles in play.
marked.setOptions({
  gfm: true,
  breaks: false,
  smartLists: true,
  smartypants: true,
} as any)

const target = document.getElementById('app')!

// The boot loader is rendered inline in index.html and shown while the bundle
// downloads/parses. Keep it visible for at least minBootMs so it never flashes
// for just a few ms on a fast connection. The start timestamp is captured by
// an inline script right after the loader markup (≈ first paint); fall back to
// performance.now() if that marker is absent (e.g. HMR, where it is stale and
// bootDelay goes negative → no wait).
const bootStart = Number((globalThis as any).__hpMgr2BootStart) || performance.now()
const minBootMs = 200
const bootDelay = minBootMs - (performance.now() - bootStart)

// Svelte 5's mount() appends into the target without removing its existing
// children, so the loader in index.html would otherwise stay visible
// alongside the mounted app. Drop it first, then mount.
;(async () => {
  if (bootDelay > 0) {
    await new Promise((r) => setTimeout(r, bootDelay))
  }
  target.replaceChildren()
  mount(App, { target })
})()
