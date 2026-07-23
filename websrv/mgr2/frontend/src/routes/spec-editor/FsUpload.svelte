<script lang="ts">
  // spec-editor drag & drop batch upload modal body. Ports the legacy
  // lcbind-fstpl-fileupload modal (spec-editor.js FileUpload): an editable
  // target directory + a drop zone that recursively traverses dropped files &
  // folders (webkitGetAsEntry) and base64-PUTs each file to mod-set-fs/put.
  // Status line per file; onDone refreshes the tree.
  import { collectDroppedFiles } from '../../lib/s2/upload'
  import { uploadModFsFile, joinFsPath } from '../../lib/fsupload'

  export let modname = ''
  export let path = '/'
  export let onDone: () => void = () => {}

  let ppath = path
  let statusLines: { ok: boolean; msg: string }[] = []
  let dragging = false

  async function handleFiles(files: File[]) {
    for (const file of files) {
      const r = await uploadModFsFile(modname, joinFsPath(ppath, file.name), file)
      statusLines = [...statusLines, { ok: r.ok, msg: r.msg }]
    }
    onDone()
  }

  function onDrop(e: DragEvent) {
    e.preventDefault()
    e.stopPropagation()
    dragging = false
    if (e.dataTransfer) {
      collectDroppedFiles(e.dataTransfer.items).then(handleFiles)
    }
  }

  function onDragOver(e: DragEvent) {
    e.preventDefault()
    e.stopPropagation()
  }
</script>

<div class="mb-3">
  <label class="form-label">The target upload directory</label>
  <input type="text" class="form-control" placeholder="Folder Path" bind:value={ppath} />
</div>

<div class="lcide-dropzone" class:dashed={dragging}
  on:dragenter={() => (dragging = true)}
  on:dragleave={() => (dragging = false)}
  on:dragover={onDragOver}
  on:drop={onDrop}
>
  Drag and Drop your files or folders to here
</div>

{#if statusLines.length}
  <div class="alert alert-info mt-3 mb-0">
    {#each statusLines as s}
      <div class={s.ok ? '' : 'text-danger'}>{s.msg}</div>
    {/each}
  </div>
{/if}

<style>
  .lcide-dropzone {
    width: 100%;
    color: #333;
    font-size: 18px;
    padding: 28px 10px;
    border: 3px dashed #5cb85c;
    border-radius: 10px;
    text-align: center;
    box-sizing: border-box;
  }
  .lcide-dropzone.dashed {
    background: rgba(92, 184, 92, 0.12);
  }
</style>
