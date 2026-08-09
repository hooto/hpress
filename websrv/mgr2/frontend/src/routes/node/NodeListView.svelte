<script lang="ts">
  // node list view. Ports node/list.tpl + node.js List/ListBatch*/Del.
  // Toolbar (Back / New Content / search / batch), table with batch checkboxes,
  // Sub-Contents drill-down (ext_node_refer), pager (RangeLen 20).
  import { onMount } from 'svelte'
  import type { Snippet } from 'svelte'
  import { api, ApiError } from '../../lib/api'
  import { openModal, closeModal } from '../../lib/modal'
  import { innerShow } from '../../lib/alert'
  import { flashThen } from '../../lib/feedback'
  import { nodelsPage, nodeReferActive } from '../../lib/store'
  import Pagination from '../../lib/Pagination.svelte'
  import EmptyState from '../../lib/EmptyState.svelte'
  import { unixTimeFormat, pager as pagerCalc } from '../../lib/util'
  import { statusDef } from './defs'
  import type { Spec, Node, Pager as PagerT } from '../../lib/types'

  // A node row as rendered: created/updated are pre-formatted date strings,
  // ext_access_counter defaulted. The raw wire item is `any` (untrusted), the
  // stored list is typed so the template is checked.
  interface NodeRow {
    id: string
    title: string
    status: number
    created: string
    updated: string
    ext_access_counter: number
  }

  let {
    modname,
    modelid,
    spec,
    onnew = () => {},
    onedit = () => {},
    tabs,
  }: {
    modname: string
    modelid: string
    spec: Spec
    onnew?: () => void
    onedit?: (id: string) => void
    tabs?: Snippet
  } = $props()

  let items: NodeRow[] = $state([])
  let pg: PagerT | null = $state(null)
  let qry = $state('')
  // deep-mutated by bind:checked={checked[id]} and reassigned by toggleAll
  let checked: Record<string, boolean> = $state({})
  let checkAll = $state(false)
  let loaded = $state(false)

  const model = $derived(
    (spec.nodeModels || []).find((m) => m.meta?.name === modelid),
  )
  const ext = $derived(model?.extensions || {})
  const referback = $derived(ext.node_refer || '')
  const anyChecked = $derived(Object.values(checked).some(Boolean))

  async function load() {
    const referid = $nodeReferActive || ''
    const params: Record<string, any> = {
      modname,
      modelid,
      ext_node_refer: referid,
      page: $nodelsPage > 1 ? $nodelsPage : '',
      fields: 'no_fields',
      terms: 'no_terms',
    }
    if (qry) params.qry_text = qry
    try {
      const rsj = await api.get<{ kind?: string; items?: Node[]; meta?: any; model?: any }>(
        'node/list',
        params,
      )
      if (!rsj || rsj.kind !== 'NodeList' || !rsj.items || rsj.items.length < 1) {
        items = []
        pg = null
        loaded = true
        // empty list is shown inline (EmptyState), not as a top alert
        innerShow('hpm-node-alert', '', '')
        return
      }
      innerShow('hpm-node-alert', '', '')
      const meta = rsj.meta || {}
      const list: NodeRow[] = (rsj.items || []).map((it: any) => ({
        id: it.id,
        title: it.title,
        status: it.status,
        created: unixTimeFormat(it.created, 'Y-m-d'),
        updated: unixTimeFormat(it.updated, 'Y-m-d'),
        ext_access_counter: it.ext_access_counter || 0,
      }))
      meta.RangeLen = 20
      pg = pagerCalc(meta)
      items = list
      checked = {}
      checkAll = false
      loaded = true
    } catch (e) {
      if (!(e instanceof ApiError && e.code === 'Unauthorized')) {
        items = []
      }
      loaded = true
    }
  }

  function listPage(n: number) {
    nodelsPage.set(n)
    load()
  }

  function search() {
    nodelsPage.set(1)
    load()
  }

  function toggleAll() {
    const next: Record<string, boolean> = {}
    if (checkAll) {
      for (const it of items) next[it.id] = true
    }
    checked = next
  }

  function batchTodo() {
    const n = Object.values(checked).filter(Boolean).length
    openModal({
      title: 'Batch operation',
      width: 800,
      height: 300,
      html: `<div id="hpm-nodels-batch-set-alert" class="alert alert-info">${n} items selected</div>`,
      buttons: [
        { title: 'Confirm to delete', class: 'btn-danger', click: batchDelete },
        { title: 'Cancel', click: () => {} },
      ],
    })
  }

  async function batchDelete() {
    const ids = items.filter((it) => checked[it.id]).map((it) => it.id)
    try {
      // DELETE is a side-effecting action: use POST (api.del), never GET — a
      // prefetched GET link would silently delete content for a signed-in user.
      await api.del('node/del', { modname, modelid, id: ids.join(',') })
      flashThen(
        'hpm-nodels-batch-set-alert',
        'success',
        'Successful operation',
        () => {
          closeModal()
          load()
        },
        500,
      )
    } catch (e) {
      if (e instanceof ApiError) {
        innerShow('hpm-nodels-batch-set-alert', 'danger', e.message)
      }
    }
  }

  function del(id: string) {
    openModal({
      title: 'Delete',
      html: '<div id="hpm-node-del" class="alert alert-danger">Are you sure to delete this?</div>',
      buttons: [
        { title: 'Confirm to delete', class: 'btn-danger', click: () => delCommit(id) },
        { title: 'Cancel', click: () => {} },
      ],
    })
  }

  async function delCommit(id: string) {
    try {
      await api.del('node/del', { modname, modelid, id })
      flashThen('hpm-node-del', 'success', 'Successful deleted', () => {
        closeModal()
        load()
      }, 500)
    } catch (e) {
      if (e instanceof ApiError) innerShow('hpm-node-del', 'danger', e.message)
    }
  }

  function subContents(id: string) {
    nodeReferActive.set(id)
    nodelsPage.set(1)
    load()
  }
  function referBack() {
    nodeReferActive.set(referback || '')
    nodelsPage.set(1)
    load()
  }

  onMount(load)
</script>

<div class="hpm-block-gap-column">
  <div class="d-flex flex-row align-items-center justify-content-between hpm-block-gap-row-sm" style="margin-bottom:8px">
    <div class="d-flex flex-row align-items-center hpm-block-gap-row-sm">
      {#if referback}
        <button class="btn btn-primary" onclick={referBack}>Back</button>
      {/if}
      <button class="btn btn-primary" onclick={onnew}>New {model?.title || 'Content'}</button>
      <form onsubmit={(e) => { e.preventDefault(); search() }} class="d-inline-block">
        <input type="text" class="form-control hpm-query-input d-inline-block" style="width:240px"
          placeholder="Press Enter to Search" bind:value={qry} />
      </form>
      {#if anyChecked}
        <button class="btn btn-outline-primary" onclick={batchTodo}>Select Contents todo ...</button>
      {/if}
    </div>
    {#if tabs}{@render tabs()}{/if}
  </div>

  <div class="hpm-table-std">
    {#if items.length}
      <table class="table table-hover align-middle" style="margin:0">
        <thead>
          <tr>
            <th style="width:20px">
              <input class="hpm-nodels-chk-all" type="checkbox" bind:checked={checkAll}
                onchange={toggleAll} />
            </th>
            <th>Title</th>
            {#if ext.node_sub_refer}<th></th>{/if}
            <th>Status</th>
            {#if ext.access_counter}<th>Access</th>{/if}
            <th>Created</th>
            <th>Updated</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {#each items as v (v.id)}
            <tr>
              <td>
                <input class="hpm-nodels-chk-item" type="checkbox" bind:checked={checked[v.id]}
                  value={v.id} />
              </td>
              <td>
                <button type="button" class="hp-link-btn" onclick={() => onedit(v.id)}>{v.title}</button>
              </td>
              {#if ext.node_sub_refer}
                <td>
                  <button class="btn btn-sm btn-outline-dark" onclick={() => subContents(v.id)}
                    >Sub Contents</button
                  >
                </td>
              {/if}
              <td>{statusDef.find((s) => s.type === v.status)?.name}</td>
              {#if ext.access_counter}<td>{v.ext_access_counter}</td>{/if}
              <td>{v.created}</td>
              <td>{v.updated}</td>
              <td align="right">
                <button class="btn btn-sm btn-outline-dark" onclick={() => del(v.id)}>Delete</button>
                <button class="btn btn-sm btn-outline-dark" onclick={() => onedit(v.id)}>Edit</button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {:else if loaded}
      <EmptyState />
    {/if}
  </div>

  {#if pg}<Pagination pg={pg} onpage={listPage} />{/if}
</div>
