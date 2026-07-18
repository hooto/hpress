/** @type {import("@sveltejs/vite-plugin-svelte").SvelteConfig} */
export default {
  // This UI is a faithful port of a legacy admin that was never a11y-audited.
  // Suppress a11y diagnostics (kept on for the real type/error signal).
  onwarn(warning, defaultHandler) {
    if (warning.code && warning.code.startsWith('a11y-')) return
    defaultHandler(warning)
  },
}
