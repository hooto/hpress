<script lang="ts">
  // The per-module model selector ([Blog] node-model + [Tags][Categories]
  // term-models). Rendered by Section into each sub-view's toolbar via a named
  // snippet so it shares the same row as that view's New/Search actions (left)
  // while staying visible across list / set / term views.
  import type { NodeModel, TermModel } from '../../lib/types'

  let {
    nodeModels = [],
    termModels = [],
    activeType = 'node',
    activeNodeModel = '',
    activeTermModel = '',
    onSelectNode = () => {},
    onSelectTerm = () => {},
  }: {
    nodeModels: NodeModel[]
    termModels: TermModel[]
    activeType?: 'node' | 'term'
    activeNodeModel?: string
    activeTermModel?: string
    onSelectNode?: (name: string) => void
    onSelectTerm?: (name: string) => void
  } = $props()
</script>

<div class="d-flex flex-row align-items-center hpm-block-gap-row-sm">
  <div class="hpm-block-gap-row-sm">
    {#each nodeModels as m (m.meta?.name)}
      <button
        class={'node-item btn btn-outline-dark' +
          (activeType === 'node' && activeNodeModel === m.meta?.name ? ' active' : '')}
        onclick={() => onSelectNode(m.meta!.name)}>{m.title || m.meta?.name}</button
      >
    {/each}
  </div>
  <div class="hpm-block-gap-row-sm">
    {#each termModels as m (m.meta?.name)}
      <button
        class={'term-item btn btn-outline-dark' +
          (activeType === 'term' && activeTermModel === m.meta?.name ? ' active' : '')}
        onclick={() => onSelectTerm(m.meta!.name)}>{m.title || m.meta?.name}</button
      >
    {/each}
  </div>
</div>
