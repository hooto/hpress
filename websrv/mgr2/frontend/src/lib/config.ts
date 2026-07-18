// Base-path resolution for the mgr2 SPA. The app is always mounted at
// `*/hp/mgr2/`; derive the `/hp` prefix from the current URL so it works even
// if config.Config.UrlBasePath is set (the Vite `base` is fixed to /hp/mgr2/,
// so asset URLs only resolve when UrlBasePath is empty — the default).
function computeBase(): string {
  const p = window.location.pathname.replace(/\/+$/, '')
  const idx = p.indexOf('/hp/mgr2')
  if (idx >= 0) return p.slice(0, idx) + '/hp'
  return '/hp'
}

const BASE = computeBase()

export const paths = {
  base: BASE, // "/hp" (or "<urlbase>/hp")
  api: BASE + '/v1/', // "/hp/v1/"
  signOut: BASE + '/user-auth/sign-out',
  signIn: BASE + '/user-auth/sign-in',
  s2: BASE + '/s2', // image serving /hp/s2/<path>
  /** hp_storage_service_endpoint token used in markdown image DSL */
  storageEndpoint: BASE + '/s2',
}
