<script lang="ts">
  // spec InfoSet modal body / carousel pagelet. Ports spec/info-set.tpl +
  // InfoSet/InfoSetCommit (edit only; module creation moved to the hpress
  // CLI module-init). Save/Cancel live in the modal's fixed footer (only
  // the body scrolls) via patchTopModal — matches NodeSet. On Save →
  // spec-info-set, then onSaved() + closeModal.
  import { onMount, untrack } from "svelte";
  import { api, ApiError } from "../../lib/api";
  import { closeModal, patchTopModal } from "../../lib/modal";
  import { innerShow } from "../../lib/alert";
  import { flashThen } from "../../lib/feedback";
  import Alert from "../../lib/Alert.svelte";
  import { statuses } from "./defs";

  let {
    name = "",
    onSaved = () => {},
  }: { name?: string; onSaved?: () => void } = $props();

  // skeleton holding the module name until spec-entry loads; the prop is
  // read once at mount (modal bodies are fresh instances per open)
  let form: any = $state({
    meta: { name: untrack(() => name) },
    srvname: "",
    title: "",
    status: 1,
    theme_config: "",
  });
  const alertId = "hpm-specset-alert";

  onMount(async () => {
    innerShow(alertId, "", ""); // clear stale banner from a prior instance
    patchTopModal({
      buttons: [
        { title: "Save", class: "btn-primary", click: save, dismiss: false },
        { title: "Cancel", class: "btn-outline-primary" },
      ],
    });
    try {
      const data = await api.get<any>("mod-set/spec-entry", { name });
      if (!data || data.kind !== "Spec") {
        innerShow(alertId, "danger", "Spec Not Found");
        return;
      }
      // theme_config is edited as text in a textarea; if the server returns
      // a JSON object, stringify it so the field doesn't render [object Object].
      const tc = data.theme_config;
      form = {
        ...data,
        theme_config:
          tc && typeof tc === "object"
            ? JSON.stringify(tc, null, 2)
            : tc || "",
      };
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : String(e);
      innerShow(alertId, "danger", msg);
    }
  });

  async function save() {
    try {
      const rsp = await api.put("mod-set/spec-info-set", {
        meta: { name: form.meta.name },
        srvname: form.srvname,
        title: form.title,
        status: parseInt(form.status),
        theme_config: form.theme_config || "",
      });
      if (!rsp || rsp.kind !== "Spec") return;
      onSaved();
      flashThen(alertId, "success", "Successful updated", closeModal, 600);
    } catch (e) {
      if (e instanceof ApiError) innerShow(alertId, "danger", e.message);
      else innerShow(alertId, "danger", String(e));
    }
  }
</script>

<form id="hpm-specset" onsubmit={(e) => e.preventDefault()} class="hpm-form-rows">
  <Alert id={alertId} />
  <div class="row mb-2 align-items-center">
    <div class="col col-2">
      <label class="form-label" for="infoset-modname">Module Name</label>
    </div>
    <div class="col">
      <input
        id="infoset-modname"
        type="text"
        class="form-control"
        bind:value={form.meta.name}
        disabled
      />
    </div>
  </div>
  <div class="row mb-2 align-items-center">
    <div class="col-2">
      <label class="form-label" for="infoset-srvname">Service Name</label>
    </div>
    <div class="col">
      <input id="infoset-srvname" type="text" class="form-control" bind:value={form.srvname} />
    </div>
  </div>
  <div class="row mb-2 align-items-center">
    <div class="col-2">
      <label class="form-label" for="infoset-title">Title</label>
    </div>
    <div class="col">
      <input id="infoset-title" type="text" class="form-control" bind:value={form.title} />
    </div>
  </div>
  {#if form.meta.name !== "core/general"}
    <div class="row mb-2 align-items-center">
      <div class="col-2">
        <label class="form-label" for="infoset-status">Status</label>
      </div>
      <div class="col">
        <select id="infoset-status" class="form-select" bind:value={form.status}>
          {#each statuses as s (s.value)}<option value={s.value}>{s.name}</option>{/each}
        </select>
      </div>
    </div>
  {/if}
  <div class="row mb-2 align-items-start">
    <div class="col-2">
      <label class="form-label" for="infoset-theme">Theme Config (JSON)</label>
    </div>
    <div class="col">
      <textarea id="infoset-theme" class="form-control" rows="6" bind:value={form.theme_config
      }></textarea>
    </div>
  </div>
</form>
