<script lang="ts">
  // spec NodeModel list (modal body / carousel pagelet). Ports spec/node/list.tpl
  // + spec.js NodeList. The edit form lives in NodeSet.svelte and is opened as a
  // second carousel pagelet via openModal, so list↔form navigation slides (and
  // the modal resizes) like the legacy lynkui pagelets — no in-place {#if} swap.
  import { onMount } from "svelte";
  import { api } from "../../lib/api";
  import { openModal } from "../../lib/modal";
  import { innerShow } from "../../lib/alert";
  import Alert from "../../lib/Alert.svelte";
  import NodeSet from "./NodeSet.svelte";

  export let modname = "";
  let items: any[] = [];
  const alertId = "hpm-spec-nodelist-alert";

  async function load() {
    try {
      const data = await api.get<any>("mod-set/spec-entry", { name: modname });
      if (data && data.kind === "Spec") {
        items = (data.nodeModels || []).map((m: any) => ({
          ...m,
          _fieldsNum: (m.fields || []).length,
          _termsNum: (m.terms || []).length,
        }));
      }
    } catch {
      /* ignore */
    }
  }

  function openSet(modelid?: string) {
    openModal({
      title: modelid ? "Edit Node" : "New Node",
      width: 1200,
      height: "max",
      body: NodeSet,
      props: { modname, modelid, onSaved: load },
    });
  }

  onMount(load);
</script>

<Alert id={alertId} />

<div class="d-flex justify-content-end" style="margin-bottom:8px">
  <button class="btn btn-primary btn-sm" on:click={() => openSet()}
    >New Node</button
  >
</div>
<table class="table table-hover">
  <thead
    ><tr><th>Name</th><th>Title</th><th>Fields</th><th>Terms</th><th></th></tr
    ></thead
  >
  <tbody>
    {#each items as it (it.meta.name)}
      <tr>
        <td>{it.meta.name}</td>
        <td>{it.title}</td>
        <td>{it._fieldsNum}</td>
        <td>{it._termsNum}</td>
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
