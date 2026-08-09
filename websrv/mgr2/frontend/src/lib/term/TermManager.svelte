<script lang="ts">
  // Term taxonomy/tag sub-editor. Reached only from inside the node section
  // (term-model buttons). Ports term.js (hpTerm.List/Set/SetCommit) + the
  // term/list.tpl and term/set.tpl templates. Manages its own list↔set view
  // switching and the persisted page (hpm_termls_page).
  import { onMount } from 'svelte'
  import type { Snippet } from 'svelte'
  import { api, ApiError } from '../api'
  import { innerShow } from '../alert'
  import { flashThen } from '../feedback'
  import Pagination from '../Pagination.svelte'
  import EmptyState from '../EmptyState.svelte'
  import { unixTimeFormat, pager as pagerCalc } from '../util'
  import { termlsPage } from '../store'
  import type { Term, TermList, TermModel, Pager as PagerT } from '../types'

  interface TermRow {
    id: string
    pid: number | string
    title: string
    weight: number
    status: number
    created: string
    updated: string
    _subs?: TermRow[]
    _dp?: number
  }

  let {
    modname,
    modelid,
    alertId = 'hpm-node-alert',
    tabs,
  }: {
    modname: string
    modelid: string
    alertId?: string
    tabs?: Snippet
  } = $props()

  let view: 'list' | 'set' = $state('list')
  let model: { title?: string; type?: string } = $state({})
  let items: TermRow[] = $state([])
  // meta is only read inside loadList (a closure), never rendered, so plain let.
  let meta: any = {}
  let pg: PagerT | null = $state(null)
  let qry = $state('')
  let loaded = $state(false)

  // set-view state (deep-mutated by setCommit: form.id = ...)
  let form: any = $state({})

  const isTaxonomy = $derived(model.type === 'taxonomy')

  async function loadList() {
    try {
      const params: Record<string, any> = { modname, modelid }
      if (!isTaxonomy || meta.itemsPerList === undefined) {
        // tags are paginated; taxonomy is non-paginated (server returns full)
      }
      if ($termlsPage > 1) params.page = $termlsPage
      if (qry) params.qry_text = qry
      const rsj = await api.get<TermList>('term/list', params)
      if (!rsj || rsj.kind !== 'TermList' || !rsj.items || rsj.items.length < 1) {
        items = []
        meta = rsj?.meta || {}
        loaded = true
        // empty list is shown inline (EmptyState), not as a top alert
        innerShow(alertId, '', '')
        pg = null
        return
      }
      innerShow(alertId, '', '')
      model = rsj.model || {}
      meta = rsj.meta || {}
      const list: TermRow[] = (rsj.items || []).map((it: any) => ({
        id: it.id,
        pid: it.pid || 0,
        title: it.title,
        weight: it.weight || 0,
        status: it.status || 0,
        created: unixTimeFormat(it.created, 'Y-m-d'),
        updated: unixTimeFormat(it.updated, 'Y-m-d H:i:s'),
      }))
      // build taxonomy tree (tag stays flat)
      if (model.type === 'taxonomy') {
        for (const it of list) {
          if (it.pid == 0) it._subs = listSubRange(list, it.id, 0)
        }
        pg = null
      } else {
        meta.RangeLen = 20
        pg = pagerCalc(meta)
      }
      items = list
      loaded = true
    } catch (e) {
      if (!(e instanceof ApiError && e.code === 'Unauthorized')) {
        alert('SpecListRefresh error, Please try again later (EC:app-termlist)')
      }
      loaded = true
    }
  }

  // hpTerm.ListSubRange — flatten children of pid with depth
  function listSubRange(ls: TermRow[], pid: string, dpnum: number): TermRow[] {
    const rs: TermRow[] = []
    dpnum++
    for (const it of ls) {
      if (it.pid == pid) {
        it._dp = dpnum
        rs.push(it)
        const subs = listSubRange(ls, it.id, dpnum)
        rs.push(...subs)
      }
    }
    return rs
  }

  function sprint(n: number): string {
    return '    '.repeat(n)
  }

  function listPage(n: number) {
    termlsPage.set(n)
    loadList()
  }

  function search() {
    termlsPage.set(1)
    loadList()
  }

  async function openSet(termid?: string) {
    view = 'set'
    try {
      if (termid) {
        const data = await api.get<Term>('term/entry', { modname, modelid, id: termid })
        if (!data || data.kind !== 'Term') {
          innerShow(alertId, 'info', 'Item Not Found')
          view = 'list'
          return
        }
        form = {
          model,
          id: data.id || '0',
          pid: data.pid || 0,
          title: data.title || '',
          status: data.status || 1,
          weight: data.weight || 0,
          _taxonomy_ls: { items },
        }
      } else {
        const m = await api.get<TermModel>('term-model/entry', { modname, modelid })
        form = {
          model: m,
          id: '0',
          pid: 0,
          title: '',
          status: 1,
          weight: 0,
          _taxonomy_ls: { items },
        }
        model = m
      }
      innerShow(alertId, '', '')
    } catch (e) {
      if (!(e instanceof ApiError && e.code === 'Unauthorized')) {
        alert('SpecListRefresh error, Please try again later (EC:app-termlist)')
      }
    }
  }

  async function setCommit() {
    try {
      const req: any = {
        kind: 'Term',
        id: parseInt(form.id),
        title: form.title,
        status: parseInt(form.status),
      }
      // legacy term.js:295 uses `=` not `==`, so weight/pid are ALWAYS sent
      // (taxonomy: real values; tag: parsed from non-existent fields → null).
      req.weight = parseInt(form.weight)
      req.pid = parseInt(form.pid)
      const data = await api.post<Term>('term/set', req, { modname, modelid })
      if (!data || data.kind !== 'Term') {
        return
      }
      form.id = data.id
      flashThen(alertId, 'success', 'Successful operation', () => {
        view = 'list'
        loadList()
      }, 500)
    } catch (e) {
      if (e instanceof ApiError) {
        innerShow(alertId, 'danger', e.message)
      }
    }
  }

  function cancel() {
    view = 'list'
  }

  onMount(loadList)
</script>

{#if view === 'list'}
  <div class="hpm-block-gap-column">
    <div class="d-flex flex-row align-items-center justify-content-between hpm-block-gap-row-sm" style="margin-bottom:8px">
      <div class="d-flex flex-row align-items-center hpm-block-gap-row-sm">
        <button class="btn btn-primary" onclick={() => openSet()}>
          New {model.title || 'Term'}
        </button>
        <form onsubmit={(e) => { e.preventDefault(); search() }} class="d-inline-block ms-2">
          <input
            type="text"
            class="form-control hpm-query-input d-inline-block"
            style="width:240px"
            placeholder="Press Enter to Search"
            bind:value={qry}
          />
        </form>
      </div>
      {#if tabs}{@render tabs()}{/if}
    </div>

    <div class="hpm-table-std">
      {#if items.length}
        <table class="table table-hover align-middle" style="margin:0">
          <thead>
            <tr>
              <th style="width:80px">ID</th>
              <th>Title</th>
              {#if isTaxonomy}<th>Weight</th>{/if}
              <th>Created</th>
              <th>Updated</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {#each items as v (v.id)}
              {#if v.pid == 0}
                <tr>
                  <td>{v.id}</td>
                  <td>{v.title}</td>
                  {#if isTaxonomy}<td>{v.weight}</td>{/if}
                  <td>{v.created}</td>
                  <td>{v.updated}</td>
                  <td align="right">
                    <button class="btn btn-sm btn-outline-dark" onclick={() => openSet(v.id)}
                      >Edit</button
                    >
                  </td>
                </tr>
                {#if v._subs}
                  {#each v._subs as v2 (v2.id)}
                    <tr>
                      <td>{v2.id}</td>
                      <td>{@html sprint(v2._dp || 0)}{v2.title}</td>
                      <td>{@html sprint(v2._dp || 0)}{v2.weight}</td>
                      <td>{v2.created}</td>
                      <td>{v2.updated}</td>
                      <td align="right">
                        <button class="btn btn-sm btn-outline-dark" onclick={() => openSet(v2.id)}
                          >Edit</button
                        >
                      </td>
                    </tr>
                  {/each}
                {/if}
              {/if}
            {/each}
          </tbody>
        </table>
      {:else if loaded}
        <EmptyState />
      {/if}
    </div>
    {#if pg}<Pagination pg={pg} onpage={listPage} />{/if}
  </div>
{:else}
  <div class="hpm-block-gap-column">
    <div class="hpm-table-std">
      {#if form.model}
        <input type="hidden" value={form.model.type || ''} />
        <input type="hidden" value={form.id} />
        <input type="hidden" value={form.status} />

        <div class="mb-3">
          <label class="form-label" for="termmgr-title">Title</label>
          <input id="termmgr-title" type="text" class="form-control" bind:value={form.title} />
        </div>

        {#if isTaxonomy}
          <div class="mb-3">
            <label class="form-label" for="termmgr-relations">Relations</label>
            <select id="termmgr-relations" class="form-select" bind:value={form.pid}>
              <option value={0}>ROOT</option>
              {#each form._taxonomy_ls.items || [] as v (v.id)}
                {#if v.pid == 0 && v.id != form.id}
                  <option value={v.id}>{v.title}</option>
                {/if}
                {#if v._subs}
                  {#each v._subs as v2 (v2.id)}
                    {#if v2.id != form.id}
                      <option value={v2.id}>{@html sprint(v2._dp || 0)}{v2.title}</option>
                    {/if}
                  {/each}
                {/if}
              {/each}
            </select>
          </div>
          <div class="mb-3">
            <label class="form-label" for="termmgr-weight">Weight</label>
            <input id="termmgr-weight" type="text" class="form-control" bind:value={form.weight} />
          </div>
        {/if}
      {/if}
    </div>

    <div class="hpm-block-gap-row-sm">
      <button class="btn btn-primary" onclick={setCommit}>Save</button>
      <button class="btn btn-outline-primary" onclick={cancel}>Cancel</button>
    </div>
  </div>
{/if}
