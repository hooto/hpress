// Typed fetch client for /hp/v1/. Mirrors hpMgr.ApiCmd semantics:
//  - HTTP 401 (from web.Auth middleware) triggers the session-expired login
//    alert (the trigger is the HTTP status, not the JSON error.code).
//  - A 200 response may still carry error.code (e.g. AccessDenied) — surface it
//    to the caller as a thrown ApiError so the component can show an inline msg.
//  - Same-origin: the IAM session cookie (x-inauth2) is sent automatically.
import { paths } from './config'
import { alertLogin, authExpiredMessage } from './alert'
import type { TypeMeta } from './types'

export class ApiError extends Error {
  code: string
  constructor(code: string, message: string) {
    super(message || code)
    this.code = code
    this.name = 'ApiError'
  }
}

function buildUrl(path: string, params?: Record<string, any>): string {
  let url = paths.api + path.replace(/^\//, '')
  if (params) {
    const sp = new URLSearchParams()
    for (const [k, v] of Object.entries(params)) {
      if (v === undefined || v === null || v === '') continue
      sp.append(k, String(v))
    }
    const qs = sp.toString()
    if (qs) url += (url.indexOf('?') >= 0 ? '&' : '?') + qs
  }
  return url
}

async function request<T = any>(
  path: string,
  opts: { method?: string; params?: Record<string, any>; body?: any } = {},
): Promise<T> {
  const { method = 'GET', params, body } = opts
  const init: RequestInit = {
    method,
    credentials: 'same-origin',
    headers: {} as Record<string, string>,
  }
  if (body !== undefined && body !== null) {
    ;(init.headers as Record<string, string>)['Content-Type'] = 'application/json'
    init.body = typeof body === 'string' ? body : JSON.stringify(body)
  }
  const res = await fetch(buildUrl(path, params), init)

  if (res.status === 401) {
    // Read the auth-error detail the backend surfaces (web.Auth returns the
    // real IAM failure, e.g. "auth-denied : iat expired") so the session-expired
    // modal can show an accurate message.
    let detail = ''
    try {
      const t = await res.text()
      if (t) {
        const d = JSON.parse(t)
        detail = d?.error?.message || ''
      }
    } catch {
      /* ignore non-JSON / empty bodies */
    }
    alertLogin(authExpiredMessage(detail))
    throw new ApiError('Unauthorized', detail || 'Unauthorized')
  }

  let data: any = null
  const text = await res.text()
  if (text) {
    try {
      data = JSON.parse(text)
    } catch {
      data = text
    }
  }

  if (data && typeof data === 'object' && data.error) {
    throw new ApiError(data.error.code || 'Error', data.error.message || data.error.code || 'Error')
  }
  return data as T
}

export const api = {
  get: <T = any>(path: string, params?: Record<string, any>) =>
    request<T>(path, { method: 'GET', params }),
  post: <T = any>(path: string, body?: any, params?: Record<string, any>) =>
    request<T>(path, { method: 'POST', body, params }),
  put: <T = any>(path: string, body?: any, params?: Record<string, any>) =>
    request<T>(path, { method: 'PUT', body, params }),
  del: <T = any>(path: string, params?: Record<string, any>) =>
    request<T>(path, { method: 'POST', params }), // server is method-agnostic
}

// Convenience: assert a successful response's kind, returns the data unchanged.
export function ensure<T extends TypeMeta>(data: T, expectKind?: string): T {
  if (expectKind && data.kind && data.kind !== expectKind) {
    throw new ApiError('Error', 'unexpected response kind: ' + data.kind)
  }
  return data
}
