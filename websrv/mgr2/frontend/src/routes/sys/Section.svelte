<script lang="ts">
  // sys module shell — the 3-tab sub-nav (#hpm-sys-nav) + the active sub-view
  // rendered into #work-content. Mirrors sys/index.tpl + sys.js Init/Index.
  import Status from './Status.svelte'
  import IamStatus from './IamStatus.svelte'
  import Config from './Config.svelte'

  export let route = 'sys/index'

  $: sub =
    route === 'sys/iam-status' ? 'iam' : route === 'sys/config' ? 'config' : 'status'
</script>

<div class="hpm-block-gap-column">
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
