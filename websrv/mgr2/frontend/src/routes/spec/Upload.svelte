<script lang="ts">
  // spec package upload modal body / carousel pagelet. Ports spec/upload.tpl +
  // UploadCommit. Upload/Cancel live in the modal's fixed footer (only the body
  // scrolls) via patchTopModal — matches NodeSet. base64 data-URL POST to
  // spec-upload-commit (8 MiB cap, 600s timeout). On success → onDone() +
  // closeModal.
  import { onMount } from "svelte";
  import { api, ApiError } from "../../lib/api";
  import { closeModal, patchTopModal } from "../../lib/modal";
  import { innerShow } from "../../lib/alert";
import { flashThen } from "../../lib/feedback";
  import Alert from "../../lib/Alert.svelte";
  import { readFileAsDataURL } from "../../lib/s2/upload";

  let {
    onDone = () => {},
  }: { onDone?: () => void } = $props();

  const alertId = "hpm-spec-upload-alert";
  let fileInput: HTMLInputElement;

  onMount(() => {
    innerShow(alertId, "", ""); // clear stale banner from a prior instance
    patchTopModal({
      buttons: [
        { title: "Upload", class: "btn-primary", click: upload, dismiss: false },
        { title: "Cancel", class: "btn-outline-primary" },
      ],
    });
  });

  async function upload() {
    const files = fileInput.files;
    if (!files || !files.length) {
      innerShow(alertId, "danger", "Please select a file");
      return;
    }
    const file = files[0];
    if (file.size > 8 * 1024 * 1024) {
      innerShow(
        alertId,
        "danger",
        "The file is too large to upload (less than 8MB)",
      );
      return;
    }
    try {
      const data = await readFileAsDataURL(file);
      const rsp = await api.post("mod-set/spec-upload-commit", {
        kind: "SpecUploadCommit",
        size: file.size,
        name: file.name,
        data,
      });
      if (!rsp || rsp.kind !== "Spec") return;
      onDone();
      flashThen(alertId, "success", "Successfully commit", closeModal, 1000);
    } catch (e) {
      if (e instanceof ApiError) innerShow(alertId, "danger", e.message);
      else innerShow(alertId, "danger", String(e));
    }
  }
</script>

<Alert id={alertId} />
<div class="mb-2">
  <label class="form-label" for="specupload-file">Select package file (.txz / .tgz)</label>
  <input id="specupload-file" type="file" class="form-control" bind:this={fileInput} />
</div>
