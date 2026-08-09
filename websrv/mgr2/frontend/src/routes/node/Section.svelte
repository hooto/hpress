<script lang="ts">
  // node section orchestrator. Route: node/index/<modname>. Renders the
  // node-model / term-model nav, and switches the workspace between the node
  // list, node set (create/edit), and the term manager (for term models).
  // Ports node/index.tpl + node.js Index/List/Set wiring. Per-module active
  // model ids persist to localStorage (hpm_snm_<mod>/hpm_stm_<mod>).
  import { onMount } from 'svelte'
  import { specsReady, refreshSpecList, specByName } from '../../lib/boot'
  import { api } from '../../lib/api'
  import { nodelsPage, termlsPage } from '../../lib/store'
  import Alert from '../../lib/Alert.svelte'
  import NodeListView from './NodeListView.svelte'
  import NodeSetView from './NodeSetView.svelte'
  import ModelTabs from './ModelTabs.svelte'
  import TermManager from '../../lib/term/TermManager.svelte'
  import type { NodeModel, TermModel, Spec } from '../../lib/types'

  let { route = 'node/index' }: { route?: string } = $props()

  const modname = $derived(
    route.startsWith('node/index/') ? route.slice('node/index/'.length) : '',
  )

  let spec: Spec | undefined = $state(undefined)
  let activeType: 'node' | 'term' = $state('node')
  let activeNodeId: string | null = $state(null)
  let subview: 'list' | 'set' = $state('list')
  let ready = $state(false)
  let activeNodeModel = $state('')
  let activeTermModel = $state('')

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

  // node/term model lists are plain state, set atomically with `spec` inside
  // loadSpec (not a separate $derived). A derived reactive would only recompute
  // for the new `spec` later in the reactive cycle, so on a fresh mount it
  // lagged behind `spec`/`ready` and the module bar rendered empty.
  let nodeModels: NodeModel[] = $state([])
  let termModels: TermModel[] = $state([])

  async function loadSpec() {
    // Capture modname up front: the await below resumes in a later microtask,
    // by which point `modname` may have been reassigned by a newer navigation.
    const name = modname
    let s: Spec | undefined = specByName(name)
    if (!s || !s.nodeModels) {
      try {
        s = await api.get<Spec>('mod-set/spec-entry', { name })
      } catch {
        s = undefined
      }
    }
    // A newer navigation started while we were fetching; drop this stale result.
    if (name !== modname) return

    // Compute the filtered model lists into locals FIRST, assign them to state,
    // and derive the default active model from the LOCAL list -- not from the
    // nodeModels/termModels $state. Reading those $state values here would
    // register them as $effect deps while this same function also writes them
    // (it runs inside the modname/specsReady $effect): a write-read cycle that
    // throws effect_update_depth_exceeded on mount.
    const nms = ((s?.nodeModels as NodeModel[]) || []).filter((m) => !m.extensions?.node_refer)
    const tms = (s?.termModels as TermModel[]) || []

    spec = s
    nodeModels = nms
    termModels = tms
    ready = true
    activeNodeModel = lsGet('hpm_snm_' + name) || nms[0]?.meta?.name || ''
    activeTermModel = lsGet('hpm_stm_' + name) || tms[0]?.meta?.name || ''
    activeType = 'node'
    subview = 'list'
  }

  // Re-run when the module changes OR once the spec list finishes loading, so an
  // early click (before boot) recovers instead of falling through to a failing
  // spec-entry fetch.
  $effect(() => {
    if (modname && $specsReady) loadSpec()
  })

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
    if (!$specsReady) refreshSpecList()
  })
</script>

{#if ready && spec}
  <div class="hpm-block-gap-column">
    <Alert id="hpm-node-alert" />
    {#if activeType === 'term'}
      {#if activeTermModel}
        {#key modname + '|' + activeTermModel}
          <TermManager modname={modname} modelid={activeTermModel} alertId="hpm-node-alert">
            {#snippet tabs()}
              <ModelTabs
                {nodeModels}
                {termModels}
                {activeType}
                {activeNodeModel}
                {activeTermModel}
                onSelectNode={selectNodeModel}
                onSelectTerm={selectTermModel}
              />
            {/snippet}
          </TermManager>
        {/key}
      {/if}
    {:else if subview === 'list'}
      {#if activeNodeModel}
        {#key modname + '|' + activeNodeModel}
          <NodeListView
            modname={modname}
            modelid={activeNodeModel}
            spec={spec}
            onnew={newContent}
            onedit={editNode}
          >
            {#snippet tabs()}
              <ModelTabs
                {nodeModels}
                {termModels}
                {activeType}
                {activeNodeModel}
                {activeTermModel}
                onSelectNode={selectNodeModel}
                onSelectTerm={selectTermModel}
              />
            {/snippet}
          </NodeListView>
        {/key}
      {/if}
    {:else}
      <div class="d-flex flex-row align-items-center justify-content-between hpm-block-gap-row-sm">
        <div class="d-flex flex-row align-self-center hpm-block-gap-row-sm">
          <div id="hpm-node-set-opts-label">
            {activeNodeId ? 'Editing' : 'Create new Content'}
          </div>
        </div>
        <ModelTabs
          {nodeModels}
          {termModels}
          {activeType}
          {activeNodeModel}
          {activeTermModel}
          onSelectNode={selectNodeModel}
          onSelectTerm={selectTermModel}
        />
      </div>
      {#key modname + '|' + activeNodeModel + '|' + (activeNodeId ?? '')}
        <NodeSetView
          modname={modname}
          modelid={activeNodeModel}
          nodeid={activeNodeId}
          oncancel={backToList}
          onsaved={backToList}
        />
      {/key}
    {/if}
  </div>
{:else}
  <div class="text-muted p-3">loading</div>
{/if}
