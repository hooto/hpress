<script lang="ts">
  // Bootstrap pagination driven by a Pager (util.pager). Replaces the inline
  // hpm-*-pager-tpl templates.
  import type { Pager } from './types'

  let {
    pg,
    onpage,
  }: { pg: Pager; onpage: (n: number) => void } = $props()

  const show = $derived(pg.PageCount > 1)
</script>

{#if show}
  <ul class="pagination pagination-sm">
    {#if pg.FirstPageNumber}
      <li class="page-item">
        <button type="button" class="page-link" onclick={() => onpage(pg.FirstPageNumber)}
          >&laquo;</button
        >
      </li>
    {/if}
    {#if pg.PrevPageNumber}
      <li class="page-item">
        <button type="button" class="page-link" onclick={() => onpage(pg.PrevPageNumber)}
          >&lsaquo;</button
        >
      </li>
    {/if}
    {#each pg.RangePages as p (p)}
      <li class={'page-item' + (p === pg.CurrentPageNumber ? ' active' : '')}>
        <button type="button" class="page-link" onclick={() => onpage(p)}>{p}</button>
      </li>
    {/each}
    {#if pg.NextPageNumber}
      <li class="page-item">
        <button type="button" class="page-link" onclick={() => onpage(pg.NextPageNumber)}
          >&rsaquo;</button
        >
      </li>
    {/if}
    {#if pg.LastPageNumber}
      <li class="page-item">
        <button type="button" class="page-link" onclick={() => onpage(pg.LastPageNumber)}
          >&raquo;</button
        >
      </li>
    {/if}
  </ul>
{/if}
