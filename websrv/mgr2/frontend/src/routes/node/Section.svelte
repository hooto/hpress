<script lang="ts">
  // node section orchestrator. Route: node/index/<modname>. Renders the
  // node-model / term-model nav, and switches the workspace between the node
  // list, node set (create/edit), and the term manager (for term models).
  // Ports node/index.tpl + node.js Index/List/Set wiring. Per-module active
  // model ids persist to localStorage (hpm_snm_<mod>/hpm_stm_<mod>).
  import { onMount } from 'svelte'
  import { specs, refreshSpecList, specByName } from '../../lib/boot'
  import { api } from '../../lib/api'
  import { nodelsPage, termlsPage } from '../../lib/store'
  import Alert from '../../lib/Alert.svelte'
  import NodeListView from './NodeListView.svelte'
  import NodeSetView from './NodeSetView.svelte'
  import TermManager from '../../lib/term/TermManager.svelte'
  import type { Spec } from '../../lib/types'

  export let route = 'node/index'

  $: modname = route.startsWith('node/index/') ? route.slice('node/index/'.length) : ''

  let spec: Spec | undefined
  let activeType: 'node' | 'term' = 'node'
  let activeNodeId: string | null = null
  let subview: 'list' | 'set' = 'list'
  let ready = false
  let activeNodeModel = ''
  let activeTermModel = ''

  function lsGet(key: string): string {
    try {
      const v = localStorage.getItem(key)
      return v ? JSON.parse(v) : ''
    } catch {
      return ''
    }
  }
  function lsSet(key: string, val: string) {
    try {
      localStorage.setItem(key, JSON.stringify(val))
    } catch {
      /* ignore */
    }
  }

  $: nodeModels = ((spec?.nodeModels as any[]) || []).filter((m) => !m.extensions?.node_refer)
  $: termModels = (spec?.termModels as any[]) || []

  async function loadSpec() {
    spec = specByName(modname)
    if (!spec || !spec.nodeModels) {
      try {
        spec = await api.get<Spec>('mod-set/spec-entry', { name: modname })
      } catch {
        spec = undefined
      }
    }
    ready = true
    const stored = lsGet('hpm_snm_' + modname)
    activeNodeModel = stored || nodeModels[0]?.meta?.name || ''
    activeTermModel = lsGet('hpm_stm_' + modname) || termModels[0]?.meta?.name || ''
    activeType = 'node'
    subview = 'list'
  }

  $: if (modname) loadSpec()

  function selectNodeModel(name: string) {
    activeNodeModel = name
    lsSet('hpm_snm_' + modname, name)
    nodelsPage.set(1)
    termlsPage.set(1)
    activeType = 'node'
    subview = 'list'
  }
  function selectTermModel(name: string) {
    activeTermModel = name
    lsSet('hpm_stm_' + modname, name)
    nodelsPage.set(1)
    termlsPage.set(1)
    activeType = 'term'
  }
  function newContent() {
    activeNodeId = null
    subview = 'set'
  }
  function editNode(id: string) {
    activeNodeId = id
    subview = 'set'
  }
  function backToList() {
    subview = 'list'
  }

  onMount(() => {
    if (!$specs.length) refreshSpecList()
  })
</script>

{#if ready && spec}
  <div class="hpm-block-gap-column">
    <div class="d-flex flex-row justify-content-between hpm-block-gap-row">
      <div class="d-flex flex-row align-self-center hpm-block-gap-row-sm"></div>
      <div class="d-flex flex-row hpm-block-gap-row-sm">
        <div class="hpm-block-gap-row-sm">
          {#each nodeModels as m (m.meta.name)}
            <button
              class={'node-item btn btn-outline-dark' +
                (activeType === 'node' && activeNodeModel === m.meta.name ? ' active' : '')}
              on:click={() => selectNodeModel(m.meta.name)}>{m.title || m.meta.name}</button
            >
          {/each}
        </div>
        <div class="hpm-block-gap-row-sm">
          {#each termModels as m (m.meta.name)}
            <button
              class={'term-item btn btn-outline-dark' +
                (activeType === 'term' && activeTermModel === m.meta.name ? ' active' : '')}
              on:click={() => selectTermModel(m.meta.name)}>{m.title || m.meta.name}</button
            >
          {/each}
        </div>
      </div>
    </div>

    <div class="hpm-block-gap-column">
      <Alert id="hpm-node-alert" />
      {#if activeType === 'term'}
        {#if activeTermModel}
          <TermManager modname={modname} modelid={activeTermModel} alertId="hpm-node-alert" />
        {/if}
      {:else if subview === 'list'}
        {#if activeNodeModel}
          <NodeListView
            modname={modname}
            modelid={activeNodeModel}
            spec={spec}
            onnew={newContent}
            onedit={editNode}
          />
        {/if}
      {:else}
        <NodeSetView
          modname={modname}
          modelid={activeNodeModel}
          nodeid={activeNodeId}
          oncancel={backToList}
          onsaved={backToList}
        />
      {/if}
    </div>
  </div>
{:else}
  <div class="text-muted p-3">loading</div>
{/if}
