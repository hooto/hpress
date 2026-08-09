<script lang="ts">
  // Inline alert bound to an element id (lynkui.alert.innerShow). Place
  // <Alert id="some-id" /> where the legacy tpl had <span id="some-id-alert">.
  import { inlineAlerts, alertClass } from './alert'

  let { id }: { id: string } = $props()
  const entry = $derived($inlineAlerts[id])
  const cls = $derived(entry ? alertClass(entry.type) : '')
</script>

{#if entry}
  <div class={`alert ${cls} hpm-alert-inner`} role="alert" style="display: block">{entry.msg}</div>
{/if}
