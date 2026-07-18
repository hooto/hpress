<script lang="ts">
  // node list view. Ports node/list.tpl + node.js List/ListBatch*/Del.
  // Toolbar (Back / New Content / search / batch), table with batch checkboxes,
  // Sub-Contents drill-down (ext_node_refer), pager (RangeLen 20).
  import { onMount } from 'svelte'
  import { api, ApiError } from '../../lib/api'
  import { openModal, closeModal } from '../../lib/modal'
  import { innerShow } from '../../lib/alert'
  import { nodelsPage, nodeReferActive } from '../../lib/store'
  import Pagination from '../../lib/Pagination.svelte'
  import { unixTimeFormat, pager as pagerCalc } from '../../lib/util'
  import type { Spec, Node, Pager as PagerT } from '../../lib/types'

  export let modname: string
  export let modelid: string
  export let spec: Spec
  export let onnew: () => void = () => {}
  export let onedit: (id: string) => void = () => {}

  const statusDef = [
    { type: 1, name: 'Publish' },
    { type: 2, name: 'Draft' },
    { type: 3, name: 'Private' },
  ]

  let items: any[] = []
  let meta: any = {}
  let pg: PagerT | null = null
  let qry = ''
  let checked: Record<string, boolean> = {}
  let checkAll = false

  $: model = (spec.nodeModels || []).find((m: any) => m.meta?.name === modelid) || ({} as any)
  $: ext = model.extensions || {}
  $: referback = ext.node_refer || ''

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
        innerShow('hpm-node-alert', 'info', 'Item Not Found')
        return
      }
      innerShow('hpm-node-alert', '', '')
      meta = rsj.meta || {}
      const list = rsj.items.map((it: any) => ({
        ...it,
        created: unixTimeFormat(it.created, 'Y-m-d'),
        updated: unixTimeFormat(it.updated, 'Y-m-d'),
        ext_access_counter: it.ext_access_counter || 0,
      }))
      meta.RangeLen = 20
      pg = pagerCalc(meta)
      items = list
      checked = {}
      checkAll = false
    } catch (e) {
      if (!(e instanceof ApiError && e.code === 'Unauthorized')) {
        items = []
      }
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

  $: anyChecked = Object.values(checked).some(Boolean)

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
      await api.get('node/del', { modname, modelid, id: ids.join(',') })
      innerShow('hpm-nodels-batch-set-alert', 'success', 'Successful operation')
      setTimeout(() => {
        closeModal()
        load()
      }, 500)
    } catch (e) {
      if (e instanceof ApiError) {
        innerShow('hpm-nodels-batch-set-alert', 'danger', e.message)
      }
    }
  }

  function del(id: string) {
    openModal({
      title: 'Delete',
      height: 200,
      html: '<div id="hpm-node-del" class="alert alert-danger">Are you sure to delete this?</div>',
      buttons: [
        { title: 'Confirm to delete', class: 'btn-danger', click: () => delCommit(id) },
        { title: 'Cancel', click: () => {} },
      ],
    })
  }

  async function delCommit(id: string) {
    try {
      await api.get('node/del', { modname, modelid, id })
      innerShow('hpm-node-del', 'success', 'Successful deleted')
      setTimeout(() => {
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
  <div class="d-flex flex-row align-items-center hpm-block-gap-row-sm" style="margin-bottom:8px">
    {#if referback}
      <button class="btn btn-primary" on:click={referBack}>Back</button>
    {/if}
    <button class="btn btn-primary" on:click={onnew}>New {model.title || 'Content'}</button>
    <form on:submit|preventDefault={search} class="d-inline-block">
      <input type="text" class="form-control hpm-query-input d-inline-block" style="width:240px"
        placeholder="Press Enter to Search" bind:value={qry} />
    </form>
    {#if anyChecked}
      <button class="btn btn-outline-primary" on:click={batchTodo}>Select Contents todo ...</button>
    {/if}
  </div>

  <div class="hpm-table-std">
    {#if items.length}
      <table class="table table-hover align-middle" style="margin:0">
        <thead>
          <tr>
            <th style="width:20px">
              <input class="hpm-nodels-chk-all" type="checkbox" bind:checked={checkAll}
                on:change={toggleAll} />
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
                <a href="javascript:void(0)" on:click={() => onedit(v.id)}>{v.title}</a>
              </td>
              {#if ext.node_sub_refer}
                <td>
                  <button class="btn btn-sm btn-outline-dark" on:click={() => subContents(v.id)}
                    >Sub Contents</button
                  >
                </td>
              {/if}
              <td>{statusDef.find((s) => s.type === v.status)?.name}</td>
              {#if ext.access_counter}<td>{v.ext_access_counter}</td>{/if}
              <td>{v.created}</td>
              <td>{v.updated}</td>
              <td align="right">
                <button class="btn btn-sm btn-outline-dark" on:click={() => del(v.id)}>Delete</button>
                <button class="btn btn-sm btn-outline-dark" on:click={() => onedit(v.id)}>Edit</button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>

  {#if pg}<Pagination pg={pg} onpage={listPage} />{/if}
</div>
