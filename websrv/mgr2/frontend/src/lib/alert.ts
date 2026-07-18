// Alert surfaces, mirroring lynkui.alert.{innerShow, open, error} and the
// hpMgr.AlertUserLogin flow.
//  - innerShow(id, type, msg): non-blocking inline banner bound to an element
//    id (the dominant in-form feedback). Rendered by <Alert id="..."/>.
//  - open(type, msg, options): blocking Bootstrap modal — used only for the
//    session-expired login prompt. Rendered by <BlockingAlert/> in App.svelte.
import { writable } from 'svelte/store'
import { paths } from './config'

export type AlertType = 'info' | 'success' | 'danger' | 'warn' | 'error' | ''

interface InlineAlert {
  type: AlertType
  msg: string
}

export const inlineAlerts = writable<Record<string, InlineAlert>>({})
export const blockingAlert = writable<null | {
  type: AlertType
  msg: string
  options?: { title?: string; buttons?: { title: string; href?: string }[]; close?: boolean }
}>(null)

// maps lynkui "warn"/"error" → bootstrap alert classes
const classFor: Record<string, string> = {
  info: 'alert-info',
  success: 'alert-success',
  danger: 'alert-danger',
  warn: 'alert-warning',
  error: 'alert-danger',
}

export function alertClass(type: string): string {
  return classFor[type] || 'alert-info'
}

export function innerShow(id: string, type: AlertType, msg: string) {
  if (!type) {
    inlineAlerts.update((a) => {
      const n = { ...a }
      delete n[id]
      return n
    })
    return
  }
  inlineAlerts.update((a) => ({ ...a, [id]: { type, msg } }))
}

export function innerHide(id: string) {
  inlineAlerts.update((a) => {
    const n = { ...a }
    delete n[id]
    return n
  })
}

export function alertOpen(
  type: AlertType,
  msg: string,
  options?: { title?: string; buttons?: { title: string; href?: string }[]; close?: boolean },
) {
  blockingAlert.set({ type, msg, options })
}

export function alertError(msg: string) {
  alertOpen('error', msg)
}

export function alertClose() {
  blockingAlert.set(null)
}

const DEFAULT_LOGIN_MSG =
  'You are not logged in, or your login session has expired. Please sign in again'

// Build the IAM sign-in URL that returns to `returnTo` (defaults to the current
// page) after a successful login. /user-auth/sign-in stores this in the
// current-url cookie, which /user-auth/callback redirects back to.
export function signInUrl(returnTo?: string): string {
  const target = returnTo ?? (typeof window !== 'undefined' ? window.location.href : '')
  if (!target) return paths.signIn
  const sep = paths.signIn.indexOf('?') >= 0 ? '&' : '?'
  return paths.signIn + sep + 'current_url=' + encodeURIComponent(target)
}

// Compose the session-expired modal message, calling out the specific IAM
// "iat expired" / auth-denied failure when the backend surfaces the detail.
export function authExpiredMessage(detail?: string): string {
  const d = (detail || '').toLowerCase()
  if (d.includes('iat expired') || d.includes('auth-denied') || d.includes('invalid access token')) {
    return 'Your login session has expired (IAM access token is no longer valid). Please sign in again.'
  }
  if (d.includes('not authenticated')) {
    return DEFAULT_LOGIN_MSG
  }
  if (detail) {
    return `Your login session is no longer valid (${detail}). Please sign in again.`
  }
  return DEFAULT_LOGIN_MSG
}

// hpMgr.AlertUserLogin — session expired; non-dismissable, single "SIGN IN"
// button that navigates (full page) to the IAM login entry and returns to the
// current URL after a successful login.
export function alertLogin(msg?: string) {
  alertOpen('warn', msg || DEFAULT_LOGIN_MSG, {
    close: false,
    buttons: [{ title: 'SIGN IN', href: signInUrl() }],
  })
}
