<script lang="ts">
  // spec Action list (modal body / carousel pagelet). Ports spec/action/list.tpl
  // + spec.js ActionList. The edit form lives in ActionSet.svelte and is opened
  // as a second carousel pagelet via openModal, so list↔form navigation slides
  // (and the modal resizes) like the legacy lynkui pagelets — no in-place {#if}
  // swap. The query-table dropdown options (node/term models) and the edited
  // row are passed straight to the form; no extra fetch.
  import { onMount } from "svelte";
  import { api } from "../../lib/api";
  import { openModal } from "../../lib/modal";
  import Alert from "../../lib/Alert.svelte";
  import ActionSet from "./ActionSet.svelte";

  export let modname = "";
  let items: any[] = [];
  let nodeModels: any[] = [];
  let termModels: any[] = [];
  const alertId = "hpm-spec-actionlist-alert";

  async function load() {
    try {
      const data = await api.get<any>("mod-set/spec-entry", { name: modname });
      if (data && data.kind === "Spec") {
        items = data.actions || [];
        nodeModels = data.nodeModels || [];
        termModels = data.termModels || [];
      }
    } catch {
      /* ignore */
    }
  }

  function openSet(name?: string) {
    const action = name ? items.find((x) => x.name === name) : undefined;
    openModal({
      title: action ? "Edit Action" : "New Action",
      width: 1100,
      height: "max",
      body: ActionSet,
      props: { modname, action, nodeModels, termModels, onSaved: load },
    });
  }

  onMount(load);
</script>

<Alert id={alertId} />

<div class="d-flex justify-content-end" style="margin-bottom:8px">
  <button class="btn btn-primary btn-sm" on:click={() => openSet()}
    >New Action</button
  >
</div>
<table class="table table-hover">
  <thead><tr><th>Name</th><th>Datax</th><th></th></tr></thead>
  <tbody>
    {#each items as it (it.name)}
      <tr>
        <td>{it.name}</td>
        <td>{(it.datax || []).length}</td>
        <td align="right">
          <button
            class="btn btn-sm btn-outline-dark"
            on:click={() => openSet(it.name)}>Edit</button
          >
        </td>
      </tr>
    {/each}
  </tbody>
</table>
