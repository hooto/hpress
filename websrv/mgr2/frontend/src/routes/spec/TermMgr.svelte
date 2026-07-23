<script lang="ts">
  // spec TermModel list (modal body / carousel pagelet). Ports spec/term/list.tpl
  // + spec.js TermList. The edit form lives in TermSet.svelte and is opened as a
  // second carousel pagelet via openModal, so list↔form navigation slides (and
  // the modal resizes) like the legacy lynkui pagelets — no in-place {#if} swap.
  import { onMount } from "svelte";
  import { api } from "../../lib/api";
  import { openModal } from "../../lib/modal";
  import Alert from "../../lib/Alert.svelte";
  import TermSet from "./TermSet.svelte";

  export let modname = "";
  let items: any[] = [];
  const alertId = "hpm-spec-termlist-alert";

  async function load() {
    try {
      const data = await api.get<any>("mod-set/spec-entry", { name: modname });
      if (data && data.kind === "Spec") items = data.termModels || [];
    } catch {
      /* ignore */
    }
  }

  function openSet(modelid?: string) {
    openModal({
      title: modelid ? "Edit Term" : "New Term",
      width: 700,
      height: "auto",
      body: TermSet,
      props: { modname, modelid, onSaved: load },
    });
  }

  onMount(load);
</script>

<Alert id={alertId} />

<div class="d-flex justify-content-end" style="margin-bottom:8px">
  <button class="btn btn-primary btn-sm" on:click={() => openSet()}
    >New Term</button
  >
</div>
<table class="table table-hover">
  <thead><tr><th>Name</th><th>Title</th><th>Type</th><th></th></tr></thead>
  <tbody>
    {#each items as it (it.meta.name)}
      <tr>
        <td>{it.meta.name}</td>
        <td>{it.title}</td>
        <td>{it.type}</td>
        <td align="right">
          <button
            class="btn btn-sm btn-outline-dark"
            on:click={() => openSet(it.meta.name)}>Edit</button
          >
        </td>
      </tr>
    {/each}
  </tbody>
</table>
