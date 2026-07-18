<script lang="ts">
  // spec InfoSet modal body. Ports spec/info-set.tpl + InfoSet/InfoSetCommit.
  import { onMount } from 'svelte'
  import { api, ApiError } from '../../lib/api'
  import { closeModal } from '../../lib/modal'
  import { innerShow } from '../../lib/alert'
  import Alert from '../../lib/Alert.svelte'
  import { objectClone, specdef, statuses } from './defs'

  export let name: string | undefined = undefined
  export let onSaved: () => void = () => {}

  let form: any = objectClone(specdef)
  let loaded = false
  const alertId = 'hpm-specset-alert'

  onMount(async () => {
    if (name) {
      try {
        const data = await api.get<any>('mod-set/spec-entry', { name })
        if (data && data.kind === 'Spec') form = data
      } catch {
        /* ignore */
      }
    }
    loaded = true
  })

  async function save() {
    try {
      const rsp = await api.put('mod-set/spec-info-set', {
        meta: { name: form.meta.name },
        srvname: form.srvname,
        title: form.title,
        status: parseInt(form.status),
        theme_config: form.theme_config || '',
      })
      if (!rsp || rsp.kind !== 'Spec') return
      innerShow(alertId, 'success', 'Successful updated')
      onSaved()
      setTimeout(closeModal, 1000)
    } catch (e) {
      if (e instanceof ApiError) innerShow(alertId, 'danger', e.message)
    }
  }
</script>

{#if loaded}
  <form id="hpm-specset" on:submit|preventDefault>
    <Alert id={alertId} />
    <div class="mb-3">
      <label class="form-label">Module Name</label>
      <input
        type="text"
        class="form-control"
        value={form.meta.name}
        on:input={(e) => (form.meta.name = e.currentTarget.value)}
        placeholder="lowercase, [a-z][a-z0-9_]+"
        disabled={!!name}
      />
    </div>
    <div class="mb-3">
      <label class="form-label">Service Name</label>
      <input type="text" class="form-control" bind:value={form.srvname} />
    </div>
    <div class="mb-3">
      <label class="form-label">Title</label>
      <input type="text" class="form-control" bind:value={form.title} />
    </div>
    {#if form.meta.name !== 'core/general'}
      <div class="mb-3">
        <label class="form-label">Status</label>
        <select class="form-select" bind:value={form.status}>
          {#each statuses as s}<option value={s.value}>{s.name}</option>{/each}
        </select>
      </div>
    {/if}
    <div class="mb-3">
      <label class="form-label">Theme Config (JSON)</label>
      <textarea class="form-control" rows="6" bind:value={form.theme_config}></textarea>
    </div>
  </form>
  <div class="text-center" style="margin-top:8px">
    <button class="btn btn-primary" on:click={save}>Save</button>
  </div>
{/if}
