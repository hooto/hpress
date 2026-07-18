<script lang="ts">
  // route template picker modal body. Lists the module's view templates
  // (mod-set/fs-tpl-list); Select sets the route template via onselect and
  // closes this modal (popping back to the RouteSet view).
  import { onMount } from 'svelte'
  import { api } from '../../lib/api'
  import { closeModal } from '../../lib/modal'

  export let modname = ''
  export let onselect: (path: string) => void = () => {}

  let items: { path: string }[] = []

  onMount(async () => {
    try {
      const data = await api.get<{ items?: { path: string }[] }>('mod-set/fs-tpl-list', { modname })
      items = data?.items || []
    } catch {
      items = []
    }
  })

  function pick(p: string) {
    onselect(p)
    closeModal()
  }
</script>

<table class="table table-hover">
  <thead><tr><th>Template</th><th></th></tr></thead>
  <tbody>
    {#each items as it (it.path)}
      <tr>
        <td>{it.path}</td>
        <td align="right">
          <button class="btn btn-sm btn-outline-dark" on:click={() => pick(it.path)}>Select</button>
        </td>
      </tr>
    {/each}
  </tbody>
</table>
