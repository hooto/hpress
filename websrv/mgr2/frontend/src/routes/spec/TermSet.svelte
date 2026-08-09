<script lang="ts">
  // spec TermModel edit form (modal body / carousel pagelet). Extracted from
  // TermMgr so list↔form navigation goes through the modal carousel (slide +
  // resize) instead of an in-place {#if} swap. Ports spec/term/set.tpl +
  // spec.js TermSet/TermSetCommit. Fields: name, title, type (taxonomy/tag).
  // On edit, re-fetches the term via term-model/entry (mirrors NodeSet). On
  // Save → spec-term-set, then onSaved() + closeModal (slides back to list).
  import { onMount } from "svelte";
  import { api, ApiError } from "../../lib/api";
  import { closeModal, patchTopModal } from "../../lib/modal";
  import { innerShow } from "../../lib/alert";
import { flashThen } from "../../lib/feedback";
  import Alert from "../../lib/Alert.svelte";
  import { termTypedef, termdef, namereg, objectClone } from "./defs";

  let {
    modname = "",
    modelid = undefined,
    onSaved = () => {},
  }: { modname?: string; modelid?: string; onSaved?: () => void } = $props();

  let form: any = $state(objectClone(termdef));
  let editing = $state(false);
  const alertId = "hpm-spec-termset-alert";

  onMount(async () => {
    innerShow(alertId, "", ""); // clear stale banner from a prior instance
    patchTopModal({
      buttons: [
        { title: "Save", class: "btn-primary", click: save, dismiss: false },
        { title: "Cancel", class: "btn-outline-primary" },
      ],
    });
    if (modelid) {
      editing = true;
      try {
        const data = await api.get<any>("term-model/entry", {
          modname,
          modelid,
        });
        if (data && data.kind === "TermModel") {
          form = { ...data, _modname: modname };
        }
      } catch {
        /* ignore */
      }
    } else {
      form = { ...objectClone(termdef), modname, _modname: modname };
    }
  });

  async function save() {
    try {
      if (!namereg.test(form.meta.name))
        throw "Invalid Term Name : " + form.meta.name;
      const rsp = await api.put("mod-set/spec-term-set", {
        meta: { name: form.meta.name },
        type: form.type,
        title: form.title,
        modname,
      });
      if (!rsp || rsp.kind !== "TermModel") return;
      onSaved();
      flashThen(alertId, "success", "Successful updated", closeModal, 600);
    } catch (e) {
      if (e instanceof ApiError) innerShow(alertId, "danger", e.message);
      else innerShow(alertId, "danger", String(e));
    }
  }
</script>

<Alert id={alertId} />

<form
  onsubmit={(e) => {
    e.preventDefault();
    save();
  }}
  class="hpm-form-rows"
>
  <div class="row mb-2 align-items-center">
    <div class="col col-2">
      <label class="form-label" for="termset-name">Name</label>
    </div>
    <div class="col">
      {#if editing}
        <input
          id="termset-name"
          type="text"
          class="form-control"
          value={form.meta.name}
          disabled
        />
      {:else}
        <input
          id="termset-name"
          type="text"
          class="form-control"
          bind:value={form.meta.name}
          placeholder="[a-z][a-z0-9_]+"
        />
      {/if}
    </div>
  </div>

  <div class="row mb-2 align-items-center">
    <div class="col-2">
      <label class="form-label" for="termset-title">Title</label>
    </div>
    <div class="col">
      <input id="termset-title" type="text" class="form-control" bind:value={form.title} />
    </div>
  </div>

  <div class="row mb-2 align-items-center">
    <div class="col-2">
      <label class="form-label" for="termset-type">Type</label>
    </div>
    <div class="col">
      <select id="termset-type" class="form-select" bind:value={form.type}>
        {#each termTypedef as t (t.type)}<option value={t.type}>{t.name}</option>{/each}
      </select>
    </div>
  </div>
</form>
