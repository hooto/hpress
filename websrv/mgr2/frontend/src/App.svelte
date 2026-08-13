<script lang="ts">
  // SPA shell. Mirrors websrv/mgr/views/index.tpl + main.js Boot/BootInit:
  // shows the topbar, renders the active hash-route section into #com-content,
  // mounts the modal stack + blocking alert, and wires the global Ctrl/Cmd+S.
  import { onMount } from 'svelte'
  import Topbar from './lib/Topbar.svelte'
  import Modal from './lib/Modal.svelte'
  import { hashRoute } from './lib/router'
  import { bootApp } from './lib/boot'
  import { blockingAlert, alertClose, alertClass } from './lib/alert'
  import { hotkeyCtrlS } from './lib/store'

  import SysSection from './routes/sys/SysSection.svelte'
  import type { Component } from 'svelte'

  // Route section dispatchers, loaded on demand. Each dynamic import becomes
  // its own dist/assets chunk named after the file (e.g. node-section-<hash>.js),
  // fetched only on first navigation to that section. CodeMirror / marked /
  // DOMPurify live in the node-section chunk, so the initial load never pulls
  // them. sys is the default/landing route, so it stays eager.
  const lazySections: Record<string, () => Promise<{ default: Component<any> }>> = {
    s2: () => import('./routes/s2/S2Section.svelte'),
    spec: () => import('./routes/spec/SpecSection.svelte'),
    node: () => import('./routes/node/NodeSection.svelte'),
  }

  onMount(() => {
    bootApp()
  })

  function onKeydown(e: KeyboardEvent) {
    // hpMgr hotkey_ctrl_s — 83 = 'S', Ctrl (non-Mac) / Cmd (Mac)
    if (e.keyCode === 83 && (navigator.platform.match('Mac') ? e.metaKey : e.ctrlKey)) {
      e.preventDefault()
      const h = $hotkeyCtrlS
      if (h) h()
    }
  }

  const route = $derived($hashRoute || 'sys/index')
  const section = $derived(route.split('/')[0])
  // The lazy section's import promise for the active route, or undefined for sys
  // (eager) / unknown routes. Recomputes only on section change, so in-section
  // sub-route navigation never re-imports.
  const sectionPromise = $derived(lazySections[section]?.())
</script>

<svelte:window onkeydown={onKeydown} />

<Topbar />

<div id="com-content" class="hpm-block-gap-frame">
  {#if section === 'sys'}
    <SysSection {route} />
  {:else if sectionPromise}
    {#await sectionPromise}
      <div class="d-flex justify-content-center py-5 text-muted">
        <div class="spinner-border spinner-border-sm" role="status" aria-label="Loading"></div>
      </div>
    {:then mod}
      {@const Component = mod.default}
      <Component {route} />
    {:catch}
      <div class="container-fluid py-4"><p class="text-muted">Failed to load section.</p></div>
    {/await}
  {:else}
    <div class="container-fluid py-4"><p class="text-muted">Route: {route}</p></div>
  {/if}
</div>

<Modal />

{#if $blockingAlert}
  <!-- Session-expired hard overlay (non-dismissable). Kept separate from the
       Modal.svelte carousel: it must sit above any open form modal and is a
       one-shot navigate-to-sign-in, not a stack step. Bootstrap modal markup
       keeps it visually consistent with the new chrome. -->
  <div
    class="modal show"
    tabindex="-1"
    role="dialog"
    aria-modal="true"
    style="display:block;z-index:2000"
  >
    <div class="modal-dialog modal-dialog-centered" style="--bs-modal-width:440px">
      <div class="modal-content">
        <div class="modal-body">
          <div class={`alert ${alertClass($blockingAlert.type)}`} style="margin-bottom:0">
            {$blockingAlert.msg}
          </div>
        </div>
        <div class="modal-footer">
          {#if ($blockingAlert.options?.buttons || []).length}
            {#each $blockingAlert.options!.buttons! as b}
              <a
                class="btn btn-primary"
                href={b.href || 'javascript:void(0)'}
                onclick={() => {
                  if (!b.href) alertClose()
                }}>{b.title}</a
              >
            {/each}
          {:else}
            <button class="btn btn-primary" onclick={alertClose}>OK</button>
          {/if}
        </div>
      </div>
    </div>
  </div>
  <div class="modal-backdrop fade show" style="z-index:1999"></div>
{/if}
