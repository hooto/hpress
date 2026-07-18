<script lang="ts">
  // spec package upload modal body. Ports spec/upload.tpl + UploadCommit.
  // base64 data-URL POST to spec-upload-commit (8 MiB cap, 600s timeout).
  import { api, ApiError } from '../../lib/api'
  import { closeModal } from '../../lib/modal'
  import { innerShow } from '../../lib/alert'
  import Alert from '../../lib/Alert.svelte'
  import { readFileAsDataURL } from '../../lib/s2/upload'

  export let onDone: () => void = () => {}
  const alertId = 'hpm-spec-upload-alert'
  let fileInput: HTMLInputElement

  async function upload() {
    const files = fileInput.files
    if (!files || !files.length) {
      innerShow(alertId, 'danger', 'Please select a file')
      return
    }
    const file = files[0]
    if (file.size > 8 * 1024 * 1024) {
      innerShow(alertId, 'danger', 'The file is too large to upload (less than 8MB)')
      return
    }
    try {
      const data = await readFileAsDataURL(file)
      const rsp = await api.post('mod-set/spec-upload-commit', {
        kind: 'SpecUploadCommit',
        size: file.size,
        name: file.name,
        data,
      })
      if (!rsp || rsp.kind !== 'Spec') return
      innerShow(alertId, 'success', 'Successfully commit')
      onDone()
      setTimeout(closeModal, 1000)
    } catch (e) {
      if (e instanceof ApiError) innerShow(alertId, 'danger', e.message)
    }
  }
</script>

<Alert id={alertId} />
<div class="mb-3">
  <label class="form-label">Select package file (.txz / .tgz)</label>
  <input type="file" class="form-control" bind:this={fileInput} />
</div>
<div class="text-center">
  <button class="btn btn-primary" on:click={upload}>Upload</button>
</div>
