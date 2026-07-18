<script lang="ts">
  // S2 upload modal body. Single-file input + drag-and-drop area (recursive
  // folder traversal). Ports the hpm-s2-objnew-tpl + ObjNew/_objNewUpload +
  // drag handlers in s2.js. 10 MiB/file cap, base64 POST, status lines,
  // auto-close after success.
  import { uploadS2Object, collectDroppedFiles } from './upload'
  import { closeModal } from '../modal'

  export let path = '/'
  export let onDone: () => void = () => {}

  let ppath = path
  let statusLines: { ok: boolean; msg: string }[] = []
  let dragging = false
  let fileInput: HTMLInputElement

  async function handleFiles(files: File[]) {
    for (const file of files) {
      const r = await uploadS2Object(ppath, file)
      statusLines = [...statusLines, { ok: r.ok, msg: r.msg }]
    }
    onDone()
    setTimeout(closeModal, 1000)
  }

  function onInputChange() {
    if (fileInput.files && fileInput.files.length) {
      handleFiles(Array.from(fileInput.files))
    }
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
<div class="mb-3">
  <label class="form-label">Select a single file to upload</label>
  <input type="file" class="form-control" bind:this={fileInput} on:change={onInputChange} />
</div>
<div class="mb-3">
  <label class="form-label">Select multifile to upload</label>
  <div
    class="_hpm_s2_fsupload_area"
    class:dashed={dragging}
    on:dragenter={() => (dragging = true)}
    on:dragleave={() => (dragging = false)}
    on:dragover={onDragOver}
    on:drop={onDrop}
  >
    Drag and Drop your files or folders to here
  </div>
</div>

{#if statusLines.length}
  <div class="alert alert-success" style="display:block">
    {#each statusLines as s}
      <div>{s.msg}</div>
    {/each}
  </div>
{/if}

<style>
  ._hpm_s2_fsupload_area {
    width: 100%;
    color: #333;
    font-size: 18px;
    padding: 20px;
    border: 3px dashed rgb(0, 120, 231);
    border-radius: 10px;
    text-align: center;
    box-sizing: border-box;
  }
</style>
