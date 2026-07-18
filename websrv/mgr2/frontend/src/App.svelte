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

  import SysSection from './routes/sys/Section.svelte'
  import S2Section from './routes/s2/Section.svelte'
  import SpecSection from './routes/spec/Section.svelte'
  import SpecEditorSection from './routes/spec-editor/Section.svelte'
  import NodeSection from './routes/node/Section.svelte'

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

  $: route = $hashRoute || 'sys/index'
  $: section = route.split('/')[0]
</script>

<svelte:window on:keydown={onKeydown} />

<Topbar />

<div id="com-content" class="hpm-block-gap-frame">
  {#if section === 'sys'}
    <SysSection {route} />
  {:else if section === 's2'}
    <S2Section {route} />
  {:else if section === 'spec-editor'}
    <SpecEditorSection {route} />
  {:else if section === 'spec'}
    <SpecSection {route} />
  {:else if section === 'node'}
    <NodeSection {route} />
  {:else}
    <div class="container-fluid py-4"><p class="text-muted">Route: {route}</p></div>
  {/if}
</div>

<Modal />

{#if $blockingAlert}
  <div class="hpm-modal-overlay" style="z-index:2000">
    <div class="hpm-modal-box" style="width:440px;height:auto">
      <div class="hpm-modal-body">
        <div class={`alert ${alertClass($blockingAlert.type)}`} style="margin-bottom:0">
          {$blockingAlert.msg}
        </div>
      </div>
      <div class="hpm-modal-foot">
        {#if ($blockingAlert.options?.buttons || []).length}
          {#each $blockingAlert.options!.buttons! as b}
            <a
              class="btn btn-primary"
              href={b.href || 'javascript:void(0)'}
              on:click={() => {
                if (!b.href) alertClose()
              }}>{b.title}</a
            >
          {/each}
        {:else}
          <button class="btn btn-primary" on:click={alertClose}>OK</button>
        {/if}
      </div>
    </div>
  </div>
{/if}
