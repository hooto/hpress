<script lang="ts">
  // Bootstrap-native modal with a lynkui-style multi-step pagelet carousel.
  //
  // Mirrors the legacy lynkui.modal (assets/lynkui/main.js ~1815-2230): ONE
  // modal whose body is a horizontal track of "pagelets"; navigating the store
  // stack slides the track (translateX by the active index) and morphs the
  // dialog size to the active step. Header/footer render once, driven by the
  // top (active) entry — exactly how lynkui morphed header/footer on switch.
  //
  // Driven entirely by the `modals` store (lib/modal.ts). We deliberately do
  // NOT instantiate bootstrap.Modal: its singleton show()/hide() lifecycle
  // races both the in-place carousel swaps and the setTimeout(closeModal, N)
  // pattern used by InfoSet/Upload/S2Upload/NodeListView. Bootstrap CSS classes
  // provide the chrome; a small Svelte action + a few handlers cover body-scroll
  // lock, scrollbar compensation, ESC, backdrop click, and focus save/restore.
  //
  // Dynamic height: a step whose `height` is 'max' / 'auto' / omitted is sized
  // to its real content (measured from .hpm-pagelet-inner) and capped at the
  // viewport; the dialog height morphs via the .hpm-dialog transition. A numeric
  // `height` keeps the legacy definite-height path (body scrolls at that height).
  // This is what makes the spec form modals (NodeSet/ActionSet/TermSet/RouteSet)
  // grow/shrink with their rows — and avoids the blank-after-mutation repaint
  // bug a permanent full-viewport scroll container caused.
  import { modals, closeModal, prevModal, type ModalSpec, type ModalButton } from './modal'

  // Viewport gutter so the dialog never touches the browser edges (always leave
  // outer margin). Bump this for a more generous frame.
  const VIEWPORT_MARGIN = 24

  const stack = $derived($modals)
  const active = $derived(stack.length - 1)
  // NOTE: a reactive declaration cannot carry an inline TS type annotation
  // on its declared variable — the colon is read as a label separator and
  // breaks the script block. Rely on the store element type instead.
  const top = $derived(stack[active])
  const isOpen = $derived(stack.length > 0)

  // The slide offset, as a concrete percent string, applied inline on the track.
  // active 0 → "0%", 1 → "-100%", 2 → "-200%" … push slides left (current page
  // exits left, new page enters from the right); pop/Back reverses.
  const trackShiftPct = $derived(`${active * -100}%`)

  let trackEl: HTMLElement | undefined
  let dialogEl: HTMLElement | undefined
  let savedFocus: HTMLElement | null = null
  // Plain let (NOT $state): read-then-written inside the focus $effect, so it
  // must not be a tracked dependency or the effect would loop.
  let lastFocusedActive = -1

  // Measured dialog content height (px) for the dynamic path. null = use the
  // fixed/numeric heightCss path.
  let dynHeight = $state<number | null>(null)

  // A step is "dynamic" (content-measured height) when its height is 'max',
  // 'auto', or omitted. Numeric heights stay on the definite-height path.
  const useDynamic = (m: ModalSpec | undefined): boolean =>
    !!m && (m.height === 'max' || m.height === 'auto' || m.height === undefined)

  // width → --bs-modal-width (Bootstrap maps it to the dialog max-width; the
  // scoped style below also forces `width` so the dialog matches the spec).
  function widthCss(m: ModalSpec | undefined): string {
    const gap = VIEWPORT_MARGIN * 2
    if (!m) return ''
    if (m.width === 'max') return `--bs-modal-width:calc(100vw - ${gap}px);`
    if (typeof m.width === 'number')
      return `--bs-modal-width:min(${m.width}px, calc(100vw - ${gap}px));`
    return `--bs-modal-width:500px;` // Bootstrap default
  }

  // height → the dynamic path emits only the viewport max-height cap; the
  // measured pixel height is appended by dialogStyle(). The numeric path keeps
  // a definite height (required for the flex chain below to scroll). 'auto'
  // historically meant "content-sized, capped" — now it is truly content-measured.
  function heightCss(m: ModalSpec | undefined): string {
    const gap = VIEWPORT_MARGIN * 2
    if (!m) return ''
    if (useDynamic(m)) return `max-height:calc(100vh - ${gap}px);`
    if (typeof m.height === 'number')
      return `height:min(${m.height}px, calc(100vh - ${gap}px));max-height:calc(100vh - ${gap}px);`
    return `max-height:calc(100vh - ${gap}px);`
  }

  // Inline style for the dialog: width + height per the active spec, plus the
  // measured pixel height on the dynamic path. Kept as a function (mirrors the
  // legacy `style={boxStyle(m)}` pattern). The track's translateX is set
  // separately and concretely on .hpm-track so its transition actually animates.
  function dialogStyle(m: ModalSpec | undefined): string {
    let s = `${widthCss(m)}${heightCss(m)}`
    // Only the dynamic path gets a measured height — never fight a numeric one.
    if (useDynamic(m) && dynHeight != null) s += `height:${dynHeight}px;`
    s += `--hp-modal-gap:${VIEWPORT_MARGIN}px;`
    return s
  }

  function focusFirstInActive() {
    if (!trackEl) return
    const page = trackEl.querySelector('.hpm-pagelet.is-active') as HTMLElement | null
    if (!page) return
    const sel =
      'input,select,textarea,button:not([disabled]),a[href],[tabindex]:not([tabindex="-1"])'
    const focusable = page.querySelector(sel) as HTMLElement | null
    ;(focusable || page).focus()
  }

  function lockScroll() {
    const sw = window.innerWidth - document.documentElement.clientWidth
    if (sw > 0) document.body.style.paddingRight = sw + 'px'
    document.body.classList.add('modal-open') // Bootstrap CSS: overflow:hidden
  }
  function unlockScroll() {
    document.body.classList.remove('modal-open')
    document.body.style.paddingRight = ''
  }

  // Svelte action bound to the root .modal — runs on mount/unmount of the
  // {#if isOpen} block, so it captures open/close transitions precisely
  // (pushing/popping pagelets keeps the modal mounted → no re-lock).
  function modalLifecycle(_node: HTMLElement) {
    savedFocus = document.activeElement as HTMLElement | null
    lockScroll()
    setTimeout(focusFirstInActive, 0)
    return {
      destroy() {
        unlockScroll()
        savedFocus?.focus?.()
        savedFocus = null
        lastFocusedActive = -1
      },
    }
  }

  // After each slide (push/pop/prev) move focus into the newly active pagelet,
  // waiting for the 300ms transition to finish. Initial open is focused by the
  // action above, so we skip when prev < 0.
  $effect(() => {
    // tracked deps: isOpen, active
    if (!isOpen) {
      lastFocusedActive = -1
      return
    }
    if (active !== lastFocusedActive) {
      const prev = lastFocusedActive
      lastFocusedActive = active
      if (prev >= 0) setTimeout(focusFirstInActive, 320)
    }
  })

  // Dynamic-height measurement. Re-keyed on `top` (patchTopModal changes top
  // without changing active, and the footer height can change when a body adds
  // its Save/Delete/Cancel buttons). Measures the active pagelet's content
  // (.hpm-pagelet-inner) + header + footer, caps at the viewport, and writes
  // dynHeight so .hpm-dialog morphs. A ResizeObserver on the content catches
  // row add/del and async onMount fetch growth; a resize listener recomputes
  // the cap. Cleanup disconnects both on active/top change and on unmount.
  //
  // No feedback loop: dynHeight drives dialog→body→pagelet height, but the
  // observed .hpm-pagelet-inner has no height constraint, so its offsetHeight
  // depends only on content — the height transition never re-fires the RO.
  $effect(() => {
    // tracked deps: top, active, isOpen
    const t = top
    const open = isOpen
    if (!open || !t || !useDynamic(t)) {
      dynHeight = null
      return
    }
    const inner = trackEl?.querySelector(
      '.hpm-pagelet.is-active .hpm-pagelet-inner',
    ) as HTMLElement | null
    // Body not flushed yet (first paint) — a later top/active change re-runs us.
    if (!inner) return

    let lastH = -1
    const measure = () => {
      const header = dialogEl?.querySelector('.hpm-header') as HTMLElement | null
      const footer = dialogEl?.querySelector('.modal-footer') as HTMLElement | null
      const hh = header?.offsetHeight ?? 0
      const ff = footer?.offsetHeight ?? 0
      const cap = window.innerHeight - VIEWPORT_MARGIN * 2
      const h = Math.min(hh + inner.offsetHeight + ff, cap)
      // Suppress identical-height writes (the 200ms width morph re-wraps content
      // and would otherwise fire the RO many times, restarting the transition).
      if (h !== lastH) {
        lastH = h
        dynHeight = h
      }
    }
    measure() // synchronous, pre-paint on step change
    const ro = new ResizeObserver(measure)
    ro.observe(inner)
    const onWin = () => measure()
    window.addEventListener('resize', onWin)
    return () => {
      ro.disconnect()
      window.removeEventListener('resize', onWin)
    }
  })

  // ESC closes the top modal (Bootstrap convention; legacy lynkui had none).
  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && $modals.length > 0) {
      e.preventDefault()
      closeModal()
    }
  }

  // Backdrop click closes only when the top entry opts in (ModalSpec.backdrop
  // === true; defaults false, matching legacy "click backdrop does NOT close").
  function onBackdropClick(e: MouseEvent) {
    if (e.target !== e.currentTarget) return
    if ($modals[$modals.length - 1]?.backdrop === true) closeModal()
  }

  function btnClick(b: ModalButton) {
    if (b.click) b.click()
    if (b.dismiss !== false) closeModal()
  }
</script>

<svelte:window on:keydown={onKeydown} />

{#if isOpen && top}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- The click only dismisses when backdrop===true; Escape (window handler)
       is the keyboard equivalent. -->
  <div
    class="modal show hpm-modal"
    tabindex="-1"
    role="dialog"
    aria-modal="true"
    aria-labelledby={top.title ? 'hpm-modal-title' : undefined}
    use:modalLifecycle
    on:click={onBackdropClick}
  >
    <div class="modal-dialog hpm-dialog" bind:this={dialogEl} style={dialogStyle(top)}>
      <div class="modal-content hpm-content">
        <div class="modal-header hpm-header">
          {#if active > 0 && top.backEnable !== false}
            <button type="button" class="btn btn-dark btn-sm me-2" on:click={() => prevModal(top.onPrev)}>
              Back
            </button>
          {/if}
          {#if top.title}
            <h5 class="modal-title flex-grow-1" id="hpm-modal-title">{top.title}</h5>
          {:else}
            <span class="flex-grow-1"></span>
          {/if}
          <button type="button" class="btn-close" aria-label="Close" on:click={() => closeModal()}></button>
        </div>

        <div class="modal-body hpm-body">
          <div class="hpm-track" bind:this={trackEl} style={`transform: translateX(${trackShiftPct})`}>
            {#each stack as m, i (i)}
              {@const Body = m.body}
              <div class="hpm-pagelet" class:is-active={i === active} aria-hidden={i !== active}>
                <div class="hpm-pagelet-inner">
                  {#if Body}
                    <Body {...m.props} />
                  {:else if m.html}
                    {@html m.html}
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        </div>

        {#if top.buttons && top.buttons.length}
          <div class="modal-footer">
            {#each top.buttons as b}
              <button type="button" class={`btn ${b.class || 'btn-dark'}`} on:click={() => btnClick(b)}>
                {b.title}
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </div>
  <div class="modal-backdrop fade show"></div>
{/if}

<style>
  /* Override Bootstrap's .modal (display:block + overflow-y:auto): center the
     dialog ourselves and never page-scroll. Bootstrap's .modal-dialog-centered
     forces min-height:calc(100% - margin*2) on the dialog, which makes an
     auto-height dialog fill the viewport and breaks the definite-height flex
     chain needed for body scrolling — so we center via this flex container and
     drop modal-dialog-centered. */
  .hpm-modal {
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
  }
  .hpm-dialog {
    margin: 0;
    width: var(--bs-modal-width);
    max-width: var(--bs-modal-width);
    /* Height source: numeric steps get a definite height via heightCss; dynamic
       ('max'/'auto') steps get a measured pixel height via the inline dynHeight.
       Both animate through this transition so list(numeric)↔form(dynamic) morph
       the same property smoothly. The height MUST be definite for the flex chain
       below to distribute space so the body shrinks and the pagelet scrolls —
       an indefinite/auto height leaves the chain indefinite and content clips. */
    transition: width 200ms ease, max-width 200ms ease, height 200ms ease;
  }
  /* Concrete viewport-bound cap (not max-height:inherit) so the flex column has
     a definite upper bound: when content exceeds it, the body shrinks
     (min-height:0) and the active pagelet scrolls — for fixed, max, AND dynamic
     heights. --hp-modal-gap is set inline on the dialog (dialogStyle). */
  .hpm-content {
    height: 100%;
    max-height: calc(100vh - var(--hp-modal-gap, 24px) * 2);
    overflow: hidden;
  }
  .hpm-header {
    flex-shrink: 0;
  }
  .hpm-body {
    flex: 1 1 auto;
    min-height: 0;
    overflow: hidden;
    padding: 0;
  }
  .hpm-track {
    display: flex;
    flex-flow: row nowrap;
    width: 100%;
    height: 100%;
    /* transform is set inline (translateX by the active step) so the transition
       below animates; a var()-derived transform does not transition reliably. */
    transition: transform 300ms ease-out;
    /* Do NOT add `will-change: transform` here. It permanently promotes the
       track to a composited layer, and Chrome then fails to repaint the
       scrollable pagelets (overflow-y:auto) inside it after their content
       mutates (add/remove rows in NodeSet/ActionSet/RouteSet/TermSet) — the
       body looks blank even though the DOM is correct, until a forced reflow
       (e.g. opening DevTools) repaints it. The track is composited only for
       the ~300ms slide (the transition above); once it settles Blink drops
       the layer and content mutations repaint normally. */
  }
  .hpm-pagelet {
    flex: 0 0 100%;
    min-width: 100%;
    height: 100%;
    overflow-y: auto;
    overflow-x: hidden;
    scrollbar-width: thin;
    /* padding moved to .hpm-pagelet-inner so the inner's offsetHeight is the
       natural content height the dynamic path measures. */
  }
  .hpm-pagelet-inner {
    padding: var(--bs-modal-padding, 1rem);
  }
  @media (prefers-reduced-motion: reduce) {
    .hpm-dialog,
    .hpm-track {
      transition: none;
    }
  }
</style>
