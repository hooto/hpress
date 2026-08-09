<script lang="ts">
  // Inline empty-state for list cards — sits inside .hpm-table-std where the
  // table would be, muted and centered. Replaces the per-list
  // `{:else if loaded}<div class="hpm-empty text-muted">...</div>` copy.
  let { text = 'Item Not Found' }: { text?: string } = $props()
</script>

<div class="hpm-empty text-muted">{text}</div>
