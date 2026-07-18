<script lang="ts">
  // spec TermModel manager (modal body). Ports spec/term/list.tpl +
  // spec/term/set.tpl + spec.js TermList/TermSet/TermSetCommit. List + set
  // views switched internally (New/Edit).
  import { onMount } from 'svelte'
  import { api, ApiError } from '../../lib/api'
  import { closeModal } from '../../lib/modal'
  import { innerShow } from '../../lib/alert'
  import Alert from '../../lib/Alert.svelte'
  import { termTypedef, termdef, namereg, objectClone } from './defs'

  export let modname = ''
  let view: 'list' | 'set' = 'list'
  let items: any[] = []
  let form: any = objectClone(termdef)
  let editing = false
  const alertId = 'hpm-spec-termset-alert'

  async function load() {
    try {
      const data = await api.get<any>('mod-set/spec-entry', { name: modname })
      if (data && data.kind === 'Spec') items = data.termModels || []
    } catch {
      /* ignore */
    }
  }

  function openSet(modelid?: string) {
    editing = !!modelid
    if (modelid) {
      api.get<any>('term-model/entry', { modname, modelid }).then((data) => {
        if (data && data.kind === 'TermModel') {
          form = { ...data, _modname: modname }
          view = 'set'
        }
      })
    } else {
      form = { ...objectClone(termdef), modname, _modname: modname }
      view = 'set'
    }
  }

  async function save() {
    try {
      const rsp = await api.put('mod-set/spec-term-set', {
        meta: { name: form.meta.name },
        type: form.type,
        title: form.title,
        modname,
      })
      if (!rsp || rsp.kind !== 'TermModel') return
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
    <button class="btn btn-primary btn-sm" on:click={() => openSet()}>New Term</button>
  </div>
  <table class="table table-hover">
    <thead><tr><th>Name</th><th>Title</th><th>Type</th><th></th></tr></thead>
    <tbody>
      {#each items as it (it.meta.name)}
        <tr>
          <td>{it.meta.name}</td>
          <td>{it.title}</td>
          <td>{it.type}</td>
          <td align="right">
            <button class="btn btn-sm btn-outline-dark" on:click={() => openSet(it.meta.name)}>Edit</button>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{:else}
  <form on:submit|preventDefault={save}>
    <div class="mb-3">
      <label class="form-label">Name</label>
      {#if editing}
        <input type="text" class="form-control" value={form.meta.name} disabled />
      {:else}
        <input type="text" class="form-control" bind:value={form.meta.name} placeholder="[a-z][a-z0-9_]+" />
      {/if}
    </div>
    <div class="mb-3">
      <label class="form-label">Title</label>
      <input type="text" class="form-control" bind:value={form.title} />
    </div>
    <div class="mb-3">
      <label class="form-label">Type</label>
      <select class="form-select" bind:value={form.type}>
        {#each termTypedef as t}<option value={t.type}>{t.name}</option>{/each}
      </select>
    </div>
    <div class="hpm-block-gap-row-sm">
      <button class="btn btn-primary" on:click={save}>Save</button>
      <button class="btn btn-outline-primary" on:click={() => (view = 'list')}>Cancel</button>
    </div>
  </form>
{/if}
