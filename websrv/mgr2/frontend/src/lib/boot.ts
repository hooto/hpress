// Boot data: site branding (sys/config-list) and the active module list
// (mod-set/spec-list) for the dynamic node nav. These mirror what the legacy
// Go shell inlined via SysConfig + what hpNode.navRefresh fetched.
import { writable, derived, get } from 'svelte/store'
import { api } from './api'
import type { Spec, SysConfigItem } from './types'

export const siteConfig = writable<Record<string, string>>({})
export const specs = writable<Spec[]>([])
export const specsReady = writable(false)

export const siteName = derived(siteConfig, (c) => c['frontend_header_site_name'] || '')
export const siteLogo = derived(siteConfig, (c) => c['frontend_header_site_logo_url'] || '')

export async function loadSiteConfig() {
  try {
    const data = await api.get<{ items?: SysConfigItem[] }>('sys/config-list')
    const map: Record<string, string> = {}
    for (const it of data.items || []) map[it.key] = it.value
    siteConfig.set(map)
  } catch {
    /* branding is non-fatal */
  }
}

export async function refreshSpecList(): Promise<Spec[]> {
  try {
    const data = await api.get<{ items?: Spec[] }>('mod-set/spec-list')
    const list = (data.items || []).filter((s) => s.status === 1)
    specs.set(list)
    specsReady.set(true)
    return list
  } catch {
    specs.set([])
    specsReady.set(true)
    return []
  }
}

export async function bootApp() {
  await Promise.all([loadSiteConfig(), refreshSpecList()])
}

// find a spec by module name
export function specByName(name: string): Spec | undefined {
  return get(specs).find((s) => s.meta?.name === name)
}
