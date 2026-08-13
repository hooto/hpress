// Modal stack, mirroring lynkui.modal.{open, close, prev}. The legacy admin
// drives most spec/sub-editor UI through a modal stack where prev(cb) pops the
// top modal and refreshes the one beneath. Each entry renders a Svelte component
// (or raw HTML) as its body; buttons are real handlers, replacing the old
// eval'd onclick strings.
import { writable } from 'svelte/store'
import type { Component } from 'svelte'

export interface ModalButton {
  title: string
  class?: string // btn-primary | btn-danger | btn-dark | btn-inverse
  click?: () => void
  dismiss?: boolean // close the modal after click (default true)
}

export interface ModalSpec {
  id?: string
  title?: string
  body?: Component // Svelte component rendered as the modal body
  props?: Record<string, any>
  html?: string // raw HTML body (simple confirm/alerts)
  width?: number | 'max'
  height?: number | 'max' | 'auto'
  buttons?: ModalButton[]
  backEnable?: boolean // show a Back button (calls prevModal)
  onPrev?: () => void // callback run after this modal is popped via prev
  backdrop?: boolean // click backdrop to close (default false, matches legacy)
}

export const modals = writable<ModalSpec[]>([])

export function openModal(spec: ModalSpec) {
  modals.update((s) => [...s, spec])
  return spec
}

// replace the top modal in-place (refresh content without growing the stack)
export function setTopModal(spec: ModalSpec) {
  modals.update((s) => {
    if (s.length === 0) return [spec]
    const n = s.slice(0, -1)
    n.push(spec)
    return n
  })
}

// Merge `patch` (e.g. { buttons }) into the TOP modal so a body component can
// populate the fixed footer after mount. Creates a new top object (so the
// footer, which reads top.buttons, re-renders) without growing the stack. The
// Modal.svelte {#each} is keyed by index, so the body component stays mounted.
export function patchTopModal(patch: Partial<ModalSpec>) {
  modals.update((s) => {
    if (!s.length) return s
    const n = s.slice(0, -1)
    n.push({ ...s[s.length - 1], ...patch })
    return n
  })
}

export function closeModal() {
  modals.update((s) => s.slice(0, -1))
}

// pop the top modal (revealing the previous one) then run cb
export function prevModal(cb?: () => void) {
  modals.update((s) => s.slice(0, -1))
  if (cb) cb()
}

export function closeAllModals() {
  modals.set([])
}
