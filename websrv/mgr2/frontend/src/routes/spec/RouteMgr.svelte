<script lang="ts">
  // spec Route manager (modal body). Ports spec/router/*.tpl + spec.js
  // RouteList/RouteSet/RouteSetCommit/RouteDel + the template picker.
  // Route fields: path, dataAction (module actions), template (+ picker via
  // fs-tpl-list), params (key/value), default. Save → spec-route-set,
  // Delete → spec-route-del.
  import { onMount } from 'svelte'
  import { api, ApiError } from '../../lib/api'
  import { openModal } from '../../lib/modal'
  import { innerShow } from '../../lib/alert'
  import Alert from '../../lib/Alert.svelte'
  import TemplatePicker from './TemplatePicker.svelte'
  import { routedef, namereg, objectClone, generalOnoff } from './defs'

  export let modname = ''
  let view: 'list' | 'set' = 'list'
  let items: any[] = []
  let actions: any[] = []
  let form: any = objectClone(routedef)
  let editing = false
  const alertId = 'hpm-spec-routeset-alert'

  async function load() {
    try {
      const data = await api.get<any>('mod-set/spec-entry', { name: modname })
      if (data && data.kind === 'Spec') {
        items = data.router?.routes || []
        actions = data.actions || []
      }
    } catch {
      /* ignore */
    }
  }

  function openSet(modelid?: string) {
    editing = !!modelid
    if (modelid) {
      const r = items.find((x) => x.path === modelid)
      if (r) {
        form = {
          ...objectClone(r),
          kind: 'SpecRoute',
          modname,
          _params: Object.entries(r.params || {}).map(([k, v]) => ({ key: k, value: String(v) })),
        }
        view = 'set'
      }
    } else {
      form = { ...objectClone(routedef), modname, _params: [], default: false }
      view = 'set'
    }
  }

  function addParam() {
    form._params.push({ key: '', value: '' })
    form._params = form._params
  }
  function delParam(i: number) {
    form._params.splice(i, 1)
    form._params = form._params
  }

  function pickTemplate() {
    openModal({
      title: 'Select a Template',
      width: 700,
      height: 500,
      body: TemplatePicker,
      props: { modname, onselect: (p: string) => (form.template = p) },
    })
  }

  async function save() {
    try {
      const params: Record<string, string> = {}
      for (const p of form._params) {
        if (!p.key || !p.value) continue
        if (!namereg.test(p.key)) throw 'Invalid Param Name : ' + p.key
        params[p.key] = p.value
      }
      const rsp = await api.put('mod-set/spec-route-set', {
        path: form.path,
        dataAction: form.dataAction,
        template: form.template,
        modname,
        params,
        default: form.default === true || form.default === '1' || form.default === 1,
      })
      if (!rsp || rsp.kind !== 'SpecRoute') return
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
      const rsp = await api.put('mod-set/spec-route-del', { path: form.path, modname })
      if (!rsp || rsp.kind !== 'SpecRoute') return
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
    <button class="btn btn-primary btn-sm" on:click={() => openSet()}>New Route</button>
  </div>
  <table class="table table-hover">
    <thead><tr><th>Path</th><th>Action</th><th>Template</th><th>Default</th><th></th></tr></thead>
    <tbody>
      {#each items as it (it.path)}
        <tr>
          <td>{it.path}</td>
          <td>{it.dataAction}</td>
          <td>{it.template}</td>
          <td>{it.default ? 'Yes' : ''}</td>
          <td align="right">
            <button class="btn btn-sm btn-outline-dark" on:click={() => openSet(it.path)}>Edit</button>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{:else}
  <div class="mb-2">
    <label class="form-label">Path</label>
    {#if editing}
      <input type="text" class="form-control" value={form.path} disabled />
    {:else}
      <input type="text" class="form-control" bind:value={form.path} />
    {/if}
  </div>
  <div class="mb-2">
    <label class="form-label">Data Action</label>
    <select class="form-select" bind:value={form.dataAction}>
      <option value=""></option>
      {#each actions as a}<option value={a.name}>{a.name}</option>{/each}
    </select>
  </div>
  <div class="mb-2">
    <label class="form-label">Template</label>
    <div class="input-group">
      <input type="text" class="form-control" bind:value={form.template} />
      <button class="btn btn-outline-dark" on:click={pickTemplate}>Select a Template</button>
    </div>
  </div>
  <div class="mb-2">
    <label class="form-label">Default</label>
    <select class="form-select" bind:value={form.default}>
      <option value={0}>No</option>
      <option value={1}>Yes</option>
    </select>
  </div>

  <h6>Params</h6>
  {#each form._params as p, i (i)}
    <div class="input-group mb-1">
      <input class="form-control form-control-sm" placeholder="key" bind:value={p.key} />
      <input class="form-control form-control-sm" placeholder="value" bind:value={p.value} />
      <button class="btn btn-sm btn-outline-danger" on:click={() => delParam(i)}>x</button>
    </div>
  {/each}
  <button class="btn btn-sm btn-link" on:click={addParam}>+ param</button>

  <div class="hpm-block-gap-row-sm mt-3">
    <button class="btn btn-primary" on:click={save}>Save</button>
    {#if editing}
      <button class="btn btn-danger" on:click={del}>Delete</button>
    {/if}
    <button class="btn btn-outline-primary" on:click={() => (view = 'list')}>Cancel</button>
  </div>
{/if}
