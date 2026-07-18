<script lang="ts">
  // spec NodeModel manager (modal body). Ports spec/node/list.tpl +
  // spec/node/set.tpl + spec.js NodeList/NodeSet/NodeSetCommit. The NodeSet
  // form edits fields (name/title/type/length/indexType/attrs), attached
  // terms, and extensions. Regex name validation; no delete for node-models.
  import { onMount } from 'svelte'
  import { api, ApiError } from '../../lib/api'
  import { innerShow } from '../../lib/alert'
  import Alert from '../../lib/Alert.svelte'
  import {
    nodedef,
    fieldTypedef,
    fieldIdxTypedef,
    generalOnoff,
    permalinkDef,
    namereg,
    objectClone,
  } from './defs'

  export let modname = ''
  let view: 'list' | 'set' = 'list'
  let items: any[] = []
  let form: any = objectClone(nodedef)
  let editing = false
  const alertId = 'hpm-spec-nodeset-alert'

  async function load() {
    try {
      const data = await api.get<any>('mod-set/spec-entry', { name: modname })
      if (data && data.kind === 'Spec') {
        items = (data.nodeModels || []).map((m: any) => ({
          ...m,
          _fieldsNum: (m.fields || []).length,
          _termsNum: (m.terms || []).length,
        }))
      }
    } catch {
      /* ignore */
    }
  }

  function normNode(d: any) {
    d.fields = (d.fields || []).map((f: any) => ({ ...f, length: f.length || 0, indexType: f.indexType || 0, attrs: f.attrs || [] }))
    d.terms = d.terms || []
    d.extensions = d.extensions || {}
    const ext = d.extensions
    ext.access_counter = !!ext.access_counter
    ext.comment_enable = !!ext.comment_enable
    ext.comment_perentry = !!ext.comment_perentry
    ext.text_search = !!ext.text_search
    ext.node_refer = ext.node_refer || ''
    ext.permalink = ext.permalink || ''
    return d
  }

  function openSet(modelid?: string) {
    editing = !!modelid
    if (modelid) {
      api.get<any>('node-model/entry', { modname, modelid }).then((data) => {
        if (data && data.kind === 'NodeModel') {
          form = normNode({ ...data, _modname: modname })
          view = 'set'
        }
      })
    } else {
      form = normNode({ ...objectClone(nodedef), modname, _modname: modname })
      view = 'set'
    }
  }

  function addField() {
    form.fields.push({ name: '', title: '', type: 'string', length: 0, indexType: 0, attrs: [] })
    form.fields = form.fields
  }
  function delField(i: number) {
    form.fields.splice(i, 1)
    form.fields = form.fields
  }
  function addAttr(f: any) {
    f.attrs.push({ key: '', value: '' })
    f.attrs = f.attrs
  }
  function delAttr(f: any, i: number) {
    f.attrs.splice(i, 1)
    f.attrs = f.attrs
  }
  function addTerm() {
    form.terms.push({ meta: { name: '' }, title: '', type: 'taxonomy' })
    form.terms = form.terms
  }
  function delTerm(i: number) {
    form.terms.splice(i, 1)
    form.terms = form.terms
  }

  async function save() {
    try {
      const req: any = {
        meta: { name: form.meta.name },
        title: form.title,
        modname,
        fields: [],
        terms: [],
        extensions: {
          access_counter: form.extensions.access_counter === true || form.extensions.access_counter === 'true',
          comment_enable: form.extensions.comment_enable === true || form.extensions.comment_enable === 'true',
          comment_perentry: form.extensions.comment_perentry === true || form.extensions.comment_perentry === 'true',
          node_refer: form.extensions.node_refer || '',
          text_search: form.extensions.text_search === true || form.extensions.text_search === 'true',
          permalink: form.extensions.permalink || '',
        },
      }
      for (const f of form.fields) {
        if (!f.name) continue
        if (!namereg.test(f.name)) throw 'Invalid Field Name : ' + f.name
        const attrs = []
        for (const a of f.attrs) {
          if (a.key) {
            if (!namereg.test(a.key)) throw 'Invalid Field Attribute Key : ' + a.key
            attrs.push({ key: a.key, value: a.value })
          }
        }
        req.fields.push({
          name: f.name,
          title: f.title,
          type: f.type,
          length: f.length,
          indexType: parseInt(f.indexType),
          attrs,
        })
      }
      for (const t of form.terms) {
        if (!t.meta.name) continue
        if (!namereg.test(t.meta.name)) throw 'Invalid Term Name : ' + t.meta.name
        req.terms.push({ meta: { name: t.meta.name }, title: t.title, type: t.type })
      }
      const rsp = await api.put('mod-set/spec-node-set', req)
      if (!rsp || rsp.kind !== 'NodeModel') return
      innerShow(alertId, 'success', 'Successful updated')
      await load()
      view = 'list'
    } catch (e) {
      if (e instanceof ApiError) innerShow(alertId, 'danger', e.message)
      else innerShow(alertId, 'danger', String(e))
    }
  }

  onMount(load)
</script>

<Alert id={alertId} />

{#if view === 'list'}
  <div class="d-flex justify-content-end" style="margin-bottom:8px">
    <button class="btn btn-primary btn-sm" on:click={() => openSet()}>New Node</button>
  </div>
  <table class="table table-hover">
    <thead><tr><th>Name</th><th>Title</th><th>Fields</th><th>Terms</th><th></th></tr></thead>
    <tbody>
      {#each items as it (it.meta.name)}
        <tr>
          <td>{it.meta.name}</td>
          <td>{it.title}</td>
          <td>{it._fieldsNum}</td>
          <td>{it._termsNum}</td>
          <td align="right">
            <button class="btn btn-sm btn-outline-dark" on:click={() => openSet(it.meta.name)}>Edit</button>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{:else}
  <form on:submit|preventDefault={save}>
    <div class="row mb-2">
      <div class="col">
        <label class="form-label">Node Model Name</label>
        {#if editing}
          <input type="text" class="form-control" value={form.meta.name} disabled />
        {:else}
          <input type="text" class="form-control" bind:value={form.meta.name} placeholder="[a-z][a-z0-9_]+" />
        {/if}
      </div>
      <div class="col">
        <label class="form-label">Title</label>
        <input type="text" class="form-control" bind:value={form.title} />
      </div>
    </div>

    <h6 class="mt-3">Fields</h6>
    <table class="table table-sm">
      <thead><tr><th>Name</th><th>Title</th><th>Type</th><th>Length</th><th>Index</th><th>Attributes</th><th></th></tr></thead>
      <tbody>
        {#each form.fields as f (form.fields.indexOf(f))}
          <tr>
            <td><input class="form-control form-control-sm" bind:value={f.name} /></td>
            <td><input class="form-control form-control-sm" bind:value={f.title} /></td>
            <td>
              <select class="form-select form-select-sm" bind:value={f.type}>
                {#each fieldTypedef as t}<option value={t.type}>{t.name}</option>{/each}
              </select>
            </td>
            <td><input class="form-control form-control-sm" style="width:70px" bind:value={f.length} /></td>
            <td>
              <select class="form-select form-select-sm" bind:value={f.indexType}>
                {#each fieldIdxTypedef as t}<option value={t.type}>{t.name}</option>{/each}
              </select>
            </td>
            <td>
              {#each f.attrs as a (f.attrs.indexOf(a))}
                <div class="d-flex mb-1">
                  <input class="form-control form-control-sm me-1" style="width:90px" placeholder="key" bind:value={a.key} />
                  <input class="form-control form-control-sm me-1" style="width:90px" placeholder="value" bind:value={a.value} />
                  <button type="button" class="btn btn-sm btn-outline-danger" on:click={() => delAttr(f, f.attrs.indexOf(a))}>x</button>
                </div>
              {/each}
              <button type="button" class="btn btn-sm btn-link" on:click={() => addAttr(f)}>+ attr</button>
            </td>
            <td><button type="button" class="btn btn-sm btn-outline-danger" on:click={() => delField(form.fields.indexOf(f))}>x</button></td>
          </tr>
        {/each}
      </tbody>
    </table>
    <button type="button" class="btn btn-sm btn-link" on:click={addField}>+ field</button>

    <h6 class="mt-3">Attached Terms</h6>
    <table class="table table-sm">
      <thead><tr><th>Name</th><th>Title</th><th>Type</th><th></th></tr></thead>
      <tbody>
        {#each form.terms as t (form.terms.indexOf(t))}
          <tr>
            <td><input class="form-control form-control-sm" style="width:140px" bind:value={t.meta.name} /></td>
            <td><input class="form-control form-control-sm" style="width:160px" bind:value={t.title} /></td>
            <td>
              <select class="form-select form-select-sm" style="width:130px" bind:value={t.type}>
                <option value="taxonomy">Categories</option>
                <option value="tag">Tags</option>
              </select>
            </td>
            <td><button type="button" class="btn btn-sm btn-outline-danger" on:click={() => delTerm(form.terms.indexOf(t))}>x</button></td>
          </tr>
        {/each}
      </tbody>
    </table>
    <button type="button" class="btn btn-sm btn-link" on:click={addTerm}>+ term</button>

    <h6 class="mt-3">Extensions</h6>
    <div class="row">
      <div class="col">
        <label class="form-label">Access Counter</label>
        <select class="form-select form-select-sm" bind:value={form.extensions.access_counter}>
          {#each generalOnoff as o}<option value={o.type}>{o.name}</option>{/each}
        </select>
      </div>
      <div class="col">
        <label class="form-label">Text Search</label>
        <select class="form-select form-select-sm" bind:value={form.extensions.text_search}>
          {#each generalOnoff as o}<option value={o.type}>{o.name}</option>{/each}
        </select>
      </div>
      <div class="col">
        <label class="form-label">Comment Enable</label>
        <select class="form-select form-select-sm" bind:value={form.extensions.comment_enable}>
          {#each generalOnoff as o}<option value={o.type}>{o.name}</option>{/each}
        </select>
      </div>
      <div class="col">
        <label class="form-label">Comment Per-Entry</label>
        <select class="form-select form-select-sm" bind:value={form.extensions.comment_perentry}>
          {#each generalOnoff as o}<option value={o.type}>{o.name}</option>{/each}
        </select>
      </div>
      <div class="col">
        <label class="form-label">Permalink</label>
        <select class="form-select form-select-sm" bind:value={form.extensions.permalink}>
          {#each permalinkDef as p}<option value={p.type}>{p.name}</option>{/each}
        </select>
      </div>
    </div>
    <div class="mt-2">
      <label class="form-label">Node Refer (sub-content model name)</label>
      <input type="text" class="form-control" bind:value={form.extensions.node_refer} />
    </div>

    <div class="hpm-block-gap-row-sm mt-3">
      <button type="button" class="btn btn-primary" on:click={save}>Save</button>
      <button type="button" class="btn btn-outline-primary" on:click={() => (view = 'list')}>Cancel</button>
    </div>
  </form>
{/if}
