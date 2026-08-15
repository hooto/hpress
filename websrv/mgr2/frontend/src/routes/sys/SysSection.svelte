<script lang="ts">
  // sys module shell — breadcrumb + the 3-tab sub-nav (#hpm-sys-nav) on the
  // right, the active sub-view rendered into #work-content. Mirrors
  // sys/index.tpl + sys.js Init/Index.
  import Status from './Status.svelte'
  import IamStatus from './IamStatus.svelte'
  import Config from './Config.svelte'

  let { route = 'sys/index' }: { route?: string } = $props()

  const sub = $derived(
    route === 'sys/iam-status' ? 'iam' : route === 'sys/config' ? 'config' : 'status',
  )

  const subTitles: Record<string, string> = {
    status: 'Status',
    iam: 'User Authentication',
    config: 'Settings',
  }
  const subTitle = $derived(subTitles[sub] ?? 'Status')
</script>

<div class="hpm-block-gap-column">
  <div
    class="d-flex flex-row align-items-center justify-content-between hpm-block-gap-row-sm"
    style="margin-bottom:8px"
  >
    <ol class="breadcrumb mb-0">
      <li class="breadcrumb-item">System</li>
      <li class="breadcrumb-item active">{subTitle}</li>
    </ol>
    <div id="hpm-sys-nav" class="hpm-node-nav hpm-block-gap-row">
      <a
        class={'btn btn-outline-dark lynkui-nav-item' + (sub === 'status' ? ' active' : '')}
        href="#sys/status">Status</a
      >
      <a
        class={'btn btn-outline-dark lynkui-nav-item' +
          (sub === 'iam' ? ' active' : '')}
        href="#sys/iam-status">User Authentication</a
      >
      <a
        class={'btn btn-outline-dark lynkui-nav-item' + (sub === 'config' ? ' active' : '')}
        href="#sys/config">Settings</a
      >
    </div>
  </div>

  <div id="work-content">
    {#key sub}
      {#if sub === 'iam'}
        <IamStatus />
      {:else if sub === 'config'}
        <Config />
      {:else}
        <Status />
      {/if}
    {/key}
  </div>
</div>

<style>
  .hp-sys-table {
    font-size: 10pt;
  }
  .hp-sys-table :global(td) {
    padding: 5px !important;
  }
  .hp-sys-table :global(tr.line) {
    border-top: 1px solid #ccc;
  }
</style>
