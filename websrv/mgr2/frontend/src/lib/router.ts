// Minimal hash router, mirroring lynkui.url.{eventRegister, eventHandler}.
// Routes live in location.hash (e.g. #node/index/core/blog) so the browser URL
// path stays /hp/mgr2/ — no server deep-link support needed. `hashRoute` is a
// store holding the current route (hash without '#'); App.svelte switches
// sections on it. Nav links are plain <a href="#..."> anchors.
import { writable } from 'svelte/store'
import { navLastActive } from './store'

export const hashRoute = writable<string>(readHash())

function readHash(): string {
  let h = window.location.hash.replace(/^#/, '')
  if (h.startsWith('/')) h = h.slice(1)
  return h
}

export function navigate(route: string) {
  if (route.startsWith('#')) route = route.slice(1)
  if (route.startsWith('/')) route = route.slice(1)
  if (readHash() === route) {
    // same hash won't fire hashchange — re-trigger
    hashRoute.set(route)
  } else {
    window.location.hash = route
  }
}

function onHash() {
  let h = readHash()
  if (!h) h = 'sys/index'
  hashRoute.set(h)
  navLastActive.set(h)
}

if (typeof window !== 'undefined') {
  window.addEventListener('hashchange', onHash)
  // ensure an initial route
  if (!readHash()) {
    // restore last-active section (default sys/index)
    const last = (() => {
      try {
        const v = localStorage.getItem('hpm_nav_last_active')
        return v ? JSON.parse(v) : ''
      } catch {
        return ''
      }
    })()
    window.location.hash = last || 'sys/index'
  } else {
    onHash()
  }
}

// helpers for active nav highlighting
export function routeEq(route: string): boolean {
  return readHash() === route
}
