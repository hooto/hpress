<script lang="ts">
  // sys/config — key/value editor. Ports sys/config.tpl + Config/ConfigSetCommit
  // in sys.js. Values only (no add/delete keys); textarea when type=="text".
  import { onMount } from 'svelte'
  import { api, ApiError } from '../../lib/api'
  import { innerShow } from '../../lib/alert'
  import Alert from '../../lib/Alert.svelte'
  import type { SysConfigItem } from '../../lib/types'

  // items are reassigned and deep-mutated through bind:value={it.value}.
  let items: SysConfigItem[] = $state([])
  let loaded = $state(false)
  const alertId = 'hpm-sys-configset-alert'

  onMount(async () => {
    try {
      const data = await api.get<{ items?: SysConfigItem[] }>('sys/config-list')
      items = (data.items || []).map((it) => ({ ...it, comment: it.comment || '' }))
      loaded = true
    } catch (e) {
      if (!(e instanceof ApiError && e.code === 'Unauthorized')) {
        alert('Error: Please try again later')
      }
    }
  })

  async function save() {
    try {
      const req = { items: items.map((it) => ({ key: it.key, value: it.value })) }
      await api.put('sys/config-set', req)
      innerShow(alertId, 'success', 'Successful updated')
    } catch (e) {
      if (e instanceof ApiError) {
        innerShow(alertId, 'danger', e.message || 'Network Connection Exception')
      }
    }
  }
</script>

<div class="card">
  <div class="card-header"><div>System Config</div></div>
  <div id="hpm-sys-configset" class="card-body">
    <Alert id={alertId} />

    {#if loaded}
      <table width="100%" class="table hpm-table-middle table-striped">
        <thead>
          <tr>
            <th>Key</th>
            <th>Value</th>
            <th>Comment</th>
          </tr>
        </thead>
        <tbody>
          {#each items as it (it.key)}
            <tr>
              <td width="30%">{it.key}</td>
              <td width="40%">
                {#if it.type === 'text'}
                  <textarea
                    class="form-control"
                    rows="3"
                    bind:value={it.value}></textarea
                  >
                {:else}
                  <input type="text" class="form-control" bind:value={it.value} />
                {/if}
              </td>
              <td>{it.comment}</td>
            </tr>
          {/each}
        </tbody>
      </table>

      <button class="btn btn-primary" onclick={save}>Save</button>
    {/if}
  </div>
</div>
