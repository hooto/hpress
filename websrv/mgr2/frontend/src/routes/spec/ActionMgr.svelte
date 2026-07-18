<script lang="ts">
  // spec Action manager (modal body). Ports spec/action/*.tpl + spec.js
  // ActionList/ActionSet/ActionSetCommit/ActionDel. datax rows: name, type
  // (list/entry), query.table (node.X / term.X), limit, order, pager,
  // cache_ttl. On save, type is prefixed node.|term. per the table selection
  // and the table is sliced to the bare name. Delete via spec-action-del.
  import { onMount } from 'svelte'
  import { api, ApiError } from '../../lib/api'
  import { innerShow } from '../../lib/alert'
  import Alert from '../../lib/Alert.svelte'
  import { dataxTypedef, actiondef, generalOnoff, namereg, objectClone } from './defs'

  export let modname = ''
  let view: 'list' | 'set' = 'list'
  let items: any[] = []
  let nodeModels: any[] = []
  let termModels: any[] = []
  let form: any = objectClone(actiondef)
  let editing = false
  const alertId = 'hpm-spec-actionset-alert'

  async function load() {
    try {
      const data = await api.get<any>('mod-set/spec-entry', { name: modname })
      if (data && data.kind === 'Spec') {
        items = data.actions || []
        nodeModels = data.nodeModels || []
        termModels = data.termModels || []
      }
    } catch {
      /* ignore */
    }
  }

  function openSet(modelid?: string) {
    editing = !!modelid
    if (modelid) {
      const a = items.find((x) => x.name === modelid)
      if (a) {
        form = {
          ...objectClone(a),
          kind: 'SpecAction',
          datax: (a.datax || []).map((d: any) => normDatax(d)),
        }
        view = 'set'
      }
    } else {
      form = { ...objectClone(actiondef), modname, datax: [normDatax({})] }
      view = 'set'
    }
  }

  function normDatax(d: any) {
    const type = d.type || 'node.list'
    const bareType = type.split('.')[1] || 'list'
    const table = type.split('.')[0] && d.query?.table ? type.split('.')[0] + '.' + d.query.table : ''
    return {
      name: d.name || '',
      type: bareType,
      _table: table,
      query: { table: d.query?.table || '', limit: d.query?.limit || 10, order: d.query?.order || '' },
      pager: d.pager === true || d.pager === 'true' ? 'true' : 'false',
      cache_ttl: d.cache_ttl || 0,
    }
  }

  function addDatax() {
    form.datax.push(normDatax({}))
    form.datax = form.datax
  }
  function delDatax(i: number) {
    form.datax.splice(i, 1)
    form.datax = form.datax
  }

  async function save() {
    try {
      if (!namereg.test(form.name)) throw 'Invalid Action Name'
      const req: any = { name: form.name, modname, datax: [] }
      for (const d of form.datax) {
        if (!d.name) continue
        if (!namereg.test(d.name)) throw 'Invalid Datax Name : ' + d.name
        let type = d.type
        const tbl = d._table
        if (tbl.startsWith('node.')) type = 'node.' + type
        else if (tbl.startsWith('term.')) type = 'term.' + type
        else throw 'Invalid Query Table Name : ' + tbl
        const bare = tbl.slice(tbl.indexOf('.') + 1)
        if (!namereg.test(bare)) throw 'Invalid Query Table Name : ' + bare
        req.datax.push({
          name: d.name,
          type,
          query: { table: bare, limit: parseInt(d.query.limit), order: d.query.order },
          pager: d.pager === 'true',
          cache_ttl: parseInt(d.cache_ttl),
        })
      }
      const rsp = await api.put('mod-set/spec-action-set', req)
      if (!rsp || rsp.kind !== 'Action') return
      innerShow(alertId, 'success', 'Successful updated')
      await load()
      view = 'list'
    } catch (e) {
      if (e instanceof ApiError) innerShow(alertId, 'danger', e.message)
      else innerShow(alertId, 'danger', String(e))
    }
  }

  async function del() {
    try {
      if (!namereg.test(form.name)) throw 'Invalid Action Name'
      const rsp = await api.put('mod-set/spec-action-del', { name: form.name, modname, datax: [] })
      if (!rsp || rsp.kind !== 'Action') return
      innerShow(alertId, 'success', 'Successful updated')
      await load()
      view = 'list'
    } catch (e) {
      if (e instanceof ApiError) innerShow(alertId, 'danger', e.message)
    }
  }

  onMount(load)
</script>

<Alert id={alertId} />

{#if view === 'list'}
  <div class="d-flex justify-content-end" style="margin-bottom:8px">
    <button class="btn btn-primary btn-sm" on:click={() => openSet()}>New Action</button>
  </div>
  <table class="table table-hover">
    <thead><tr><th>Name</th><th>Datax</th><th></th></tr></thead>
    <tbody>
      {#each items as it (it.name)}
        <tr>
          <td>{it.name}</td>
          <td>{(it.datax || []).length}</td>
          <td align="right">
            <button class="btn btn-sm btn-outline-dark" on:click={() => openSet(it.name)}>Edit</button>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{:else}
  <div class="mb-2">
    <label class="form-label">Action Name</label>
    {#if editing}
      <input type="text" class="form-control" value={form.name} disabled />
    {:else}
      <input type="text" class="form-control" bind:value={form.name} placeholder="[a-z][a-z0-9_]+" />
    {/if}
  </div>

  <h6>Datax</h6>
  {#each form.datax as d, i (i)}
    <div class="border rounded p-2 mb-2">
      <div class="row">
        <div class="col">
          <label class="form-label">Name</label>
          <input class="form-control form-control-sm" bind:value={d.name} />
        </div>
        <div class="col">
          <label class="form-label">Type</label>
          <select class="form-select form-select-sm" bind:value={d.type}>
            {#each dataxTypedef as t}<option value={t.type}>{t.name}</option>{/each}
          </select>
        </div>
        <div class="col">
          <label class="form-label">Query Table</label>
          <select class="form-select form-select-sm" bind:value={d._table}>
            {#each nodeModels as m}<option value={'node.' + m.meta.name}>node : {m.meta.name}</option>{/each}
            {#each termModels as m}<option value={'term.' + m.meta.name}>term : {m.meta.name}</option>{/each}
          </select>
        </div>
      </div>
      <div class="row mt-1">
        <div class="col">
          <label class="form-label">Limit</label>
          <input class="form-control form-control-sm" bind:value={d.query.limit} />
        </div>
        <div class="col">
          <label class="form-label">Order</label>
          <input class="form-control form-control-sm" bind:value={d.query.order} />
        </div>
        <div class="col">
          <label class="form-label">Pager</label>
          <select class="form-select form-select-sm" bind:value={d.pager}>
            {#each generalOnoff as o}<option value={o.type}>{o.name}</option>{/each}
          </select>
        </div>
        <div class="col">
          <label class="form-label">Cache TTL</label>
          <input class="form-control form-control-sm" bind:value={d.cache_ttl} />
        </div>
        <div class="col d-flex align-items-end">
          <button type="button" class="btn btn-sm btn-outline-danger" on:click={() => delDatax(i)}>Remove</button>
        </div>
      </div>
    </div>
  {/each}
  <button type="button" class="btn btn-sm btn-link" on:click={addDatax}>+ datax</button>

  <div class="hpm-block-gap-row-sm mt-3">
    <button class="btn btn-primary" on:click={save}>Save</button>
    {#if editing}
      <button class="btn btn-danger" on:click={del}>Delete</button>
    {/if}
    <button class="btn btn-outline-primary" on:click={() => (view = 'list')}>Cancel</button>
  </div>
{/if}
