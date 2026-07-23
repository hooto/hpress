// Utility helpers — direct ports of lynkui.utilx.* and hpMgr.* used across the
// legacy admin JS (webui/hpm/js/main.js, *.js and the .tpl templates).
import SparkMD5 from 'spark-md5'
import { tick } from 'svelte'
import type { Pager } from './types'

// lynkui.utilx.cryptoMd5 — hex md5 of a string (used to mint stable element /
// cache keys). Returns hex string.
export function md5(str: string): string {
  return SparkMD5.hash(str)
}

// lynkui.utilx.objectClone — deep clone (data is always JSON-serializable).
export function objectClone<T>(obj: T): T {
  if (obj === null || obj === undefined) return obj
  return JSON.parse(JSON.stringify(obj))
}

// Removing a row inside a scrolling modal pagelet can leave it scrolled past
// the now-shorter content (looks blank). Capture scrollTop before the mutation,
// run it, then restore (clamped to the valid range) after Svelte flushes. Used
// by the spec form bodies (NodeSet/RouteSet/ActionSet) where add/remove rows
// change body height.
export async function withStableScroll(mutate: () => void): Promise<void> {
  const pagelet = document.querySelector(
    '.hpm-pagelet.is-active',
  ) as HTMLElement | null
  const before = pagelet?.scrollTop ?? 0
  ;(document.activeElement as HTMLElement | null)?.blur?.()
  mutate()
  await tick()
  if (!pagelet) return
  const max = Math.max(0, pagelet.scrollHeight - pagelet.clientHeight)
  pagelet.scrollTop = Math.min(before, max)
}

// trim arbitrary leading/trailing chars; default whitespace.
export function trim(str: string, chars?: string): string {
  if (!str) return ''
  if (!chars) return String(str).replace(/^\s+|\s+$/g, '')
  const set = chars.replace(/([.*+?^${}()|[\]\\])/g, '\\$1')
  return str.replace(new RegExp('^[' + set + ']+|[' + set + ']+$', 'g'), '')
}

// ---- phpjs-style date formatting (tokens Y y m n d j H G i s) ----
function pad2(n: number): string {
  return (n < 10 ? '0' : '') + n
}

export function formatDate(d: Date, fmt: string): string {
  const map: { [k: string]: string } = {
    Y: '' + d.getFullYear(),
    y: ('' + d.getFullYear()).slice(-2),
    m: pad2(d.getMonth() + 1),
    n: '' + (d.getMonth() + 1),
    d: pad2(d.getDate()),
    j: '' + d.getDate(),
    H: pad2(d.getHours()),
    G: '' + d.getHours(),
    i: pad2(d.getMinutes()),
    s: pad2(d.getSeconds()),
  }
  let out = ''
  for (let i = 0; i < fmt.length; i++) {
    out += map[fmt[i]] !== undefined ? map[fmt[i]] : fmt[i]
  }
  return out
}

// lynkui.utilx.unixTimeFormat(epochSec, fmt) — e.g. unixTimeFormat(created,"Y-m-d")
export function unixTimeFormat(epoch: number, fmt: string): string {
  if (!epoch) return ''
  return formatDate(new Date(epoch * 1000), fmt)
}

// lynkui.utilx.timeParseFormat(dateStr, fmt) — parse a date/datetime string then
// reformat. Accepts epoch numbers or strings like "2024-01-05 12:30:00".
export function timeParseFormat(value: any, fmt: string): string {
  if (!value && value !== 0) return ''
  let d: Date
  if (typeof value === 'number') {
    d = new Date(value * 1000)
  } else {
    const n = Number(value)
    if (!isNaN(n) && /^\d+$/.test(String(value).trim())) {
      d = new Date(n * 1000)
    } else {
      d = new Date(String(value).replace(/-/g, '/'))
    }
  }
  if (isNaN(d.getTime())) return String(value)
  return formatDate(d, fmt)
}

// ---- last mouse position (lynkui.utilx.pos) used by tablet popover ----
let lastPos = { left: 0, top: 0 }
if (typeof window !== 'undefined') {
  document.addEventListener('mousemove', (e) => {
    lastPos = { left: e.pageX, top: e.pageY }
  })
}
export function pos(): { left: number; top: number } {
  return { left: lastPos.left, top: lastPos.top }
}

// hpMgr.Equal / NotEqual / GreaterThan — used inside doT templates; kept for
// parity when porting conditional class logic.
export const Equal = (a: any, b: any) => a == b
export const NotEqual = (a: any, b: any) => a != b
export const GreaterThan = (a: any, b: any) => a > b

// hpMgr.Pager (main.js:222) — verbatim port. `metalist` is a TypeListMeta plus
// an optional RangeLen (client window length; default 10, hpm uses 20).
export function pager(metalist: {
  startIndex?: number
  totalResults?: number
  itemsPerList?: number
  RangeLen?: number
}): Pager {
  if (!metalist.startIndex) metalist.startIndex = 0
  if (!metalist.totalResults) metalist.totalResults = 0
  if (!metalist.itemsPerList || metalist.itemsPerList < 1) metalist.itemsPerList = 10
  if (!metalist.RangeLen) metalist.RangeLen = 10
  else if (metalist.RangeLen < 1) metalist.RangeLen = 1

  const pg: Pager = {
    ItemCount: metalist.totalResults,
    CountPerPage: metalist.itemsPerList,
    PageCount: 0,
    CurrentPageNumber: 0,
    FirstPageNumber: 0,
    PrevPageNumber: 0,
    NextPageNumber: 0,
    LastPageNumber: 0,
    RangeLen: metalist.RangeLen,
    RangeStartNumber: 1,
    RangeEndNumber: 0,
    RangePages: [],
  }

  if (metalist.startIndex > 0) {
    pg.CurrentPageNumber = Math.floor(metalist.startIndex / metalist.itemsPerList) + 1
  }

  pg.PageCount = Math.floor(pg.ItemCount / pg.CountPerPage)
  if (pg.ItemCount % pg.CountPerPage > 0) pg.PageCount++

  if (pg.CurrentPageNumber < 1) pg.CurrentPageNumber = 1
  else if (pg.CurrentPageNumber > pg.PageCount) pg.CurrentPageNumber = pg.PageCount

  if (pg.CurrentPageNumber > pg.RangeLen / 2) {
    pg.RangeStartNumber = pg.CurrentPageNumber - Math.floor(pg.RangeLen / 2)
  }
  pg.RangeEndNumber = pg.PageCount
  if (pg.RangeStartNumber + pg.RangeLen < pg.PageCount) {
    pg.RangeEndNumber = pg.RangeStartNumber + pg.RangeLen - 1
  }

  if (pg.CurrentPageNumber > 1) pg.PrevPageNumber = pg.CurrentPageNumber - 1
  if (pg.CurrentPageNumber < pg.PageCount) pg.NextPageNumber = pg.CurrentPageNumber + 1

  for (let i = pg.RangeStartNumber; i <= pg.RangeEndNumber; i++) pg.RangePages.push(i)
  if (pg.RangeStartNumber > 1) pg.FirstPageNumber = 1
  if (pg.RangeEndNumber < pg.PageCount) pg.LastPageNumber = pg.PageCount

  return pg
}

// humanize byte size (s2/index renders sizes). lynkui.utilx.byteSizeFormat is
// unused by hpm, but the s2 template calls hpS2.UtilResourceSizeFormat — ported
// here for the s2 module.
export function byteSizeFormat(n: number): string {
  if (!n || n < 1) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  let v = n
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i++
  }
  return (i < 2 ? Math.round(v) : v.toFixed(2).replace(/\.?0+$/, '')) + ' ' + units[i]
}

// hpSys.UtilResourceSizeFormat — verbatim port. Returns "N UNIT" (legacy
// emitted an HTML <span> around the unit; we render plain text — visually
// equivalent, no rule styled that span).
export function fmtResourceSize(size: number): string {
  const ms: [number, string][] = [
    [6, 'EB'],
    [5, 'PB'],
    [4, 'TB'],
    [3, 'GB'],
    [2, 'MB'],
    [1, 'KB'],
  ]
  for (const [exp, unit] of ms) {
    if (size > Math.pow(1024, exp)) {
      return (size / Math.pow(1024, exp)).toFixed(0) + ' ' + unit
    }
  }
  if (size === 0) return '0'
  return size + ' B'
}

// hpSys.UtilDurationFormat — verbatim port. `fix` is an optional divisor.
export function fmtDuration(timems: number, fix?: number): string {
  const ms: [number, string][] = [
    [86400000, 'day'],
    [3600000, 'hour'],
    [60000, 'minute'],
    [1000, 'second'],
  ]
  if (!timems) timems = 0
  timems = fix ? Math.floor(timems / fix) : Math.floor(timems)
  let ts = ''
  for (const [unit, name] of ms) {
    if (timems >= unit) {
      const t = Math.floor(timems / unit)
      if (t > 0) {
        ts += t + ' ' + name
        if (t > 1) ts += 's'
        ts += ', '
        timems = Math.floor(timems % unit)
      }
    }
  }
  if (ts.length > 2) ts = ts.slice(0, ts.length - 2)
  else if (timems > 0) ts = timems + ' microseconds'
  else ts = '0'
  return ts
}
