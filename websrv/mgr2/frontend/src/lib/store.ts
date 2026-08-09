// Persistent Svelte stores backed by localStorage/sessionStorage, keeping the
// SAME key names the legacy admin used (lynkui.storage / lynkui.session) so any
// shared state survives across the two versions.
import { writable, type Writable } from 'svelte/store'

function persister<T>(key: string, def: T, session = false): Writable<T> {
  const storage = session ? sessionStorage : localStorage
  let initial: T = def
  try {
    const raw = storage.getItem(key)
    if (raw !== null && raw !== '') initial = JSON.parse(raw) as T
  } catch {
    /* ignore */
  }
  const store = writable<T>(initial)
  store.subscribe((v) => {
    try {
      storage.setItem(key, JSON.stringify(v))
    } catch {
      /* ignore quota */
    }
  })
  return store
}

// core UI-state keys (module-specific keys added in their modules via persister)
export const navLastActive = persister<string>('hpm_nav_last_active', 'sys/index')
export const specActive = persister<string>('hpm_spec_active', '')
export const nodelsPage = persister<number>('hpm_nodels_page', 1)
export const termlsPage = persister<number>('hpm_termls_page', 1)
export const nodeReferActive = persister<string>('hpm_node_refer_active', '')
export const s2ObjPathActive = persister<string>('hpm_s2_obj_path_active', '/deft')

// Global Ctrl/Cmd+S handler (hpMgr.hotkey_ctrl_s). Only the node editor binds
// it (save-and-stay); null means no handler. Modules set it on mount, clear on
// destroy.
export const hotkeyCtrlS = writable<null | (() => void)>(null)
