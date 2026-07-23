<script lang="ts">
  // spec Route list (modal body / carousel pagelet). Ports spec/router/list.tpl
  // + spec.js RouteList. The edit form lives in RouteSet.svelte and is opened
  // as a second carousel pagelet via openModal, so list↔form navigation slides
  // (and the modal resizes) like the legacy lynkui pagelets — no in-place {#if}
  // swap. The dataAction dropdown options (module actions) and the edited row
  // are passed straight to the form; no extra fetch.
  import { onMount } from "svelte";
  import { api } from "../../lib/api";
  import { openModal } from "../../lib/modal";
  import Alert from "../../lib/Alert.svelte";
  import RouteSet from "./RouteSet.svelte";

  export let modname = "";
  let items: any[] = [];
  let actions: any[] = [];
  const alertId = "hpm-spec-routelist-alert";

  async function load() {
    try {
      const data = await api.get<any>("mod-set/spec-entry", { name: modname });
      if (data && data.kind === "Spec") {
        items = data.router?.routes || [];
        actions = data.actions || [];
      }
    } catch {
      /* ignore */
    }
  }

  function openSet(path?: string) {
    const route = path ? items.find((x) => x.path === path) : undefined;
    openModal({
      title: route ? "Edit Route" : "New Route",
      width: 1000,
      height: "max",
      body: RouteSet,
      props: { modname, route, actions, onSaved: load },
    });
  }

  onMount(load);
</script>

<Alert id={alertId} />

<div class="d-flex justify-content-end" style="margin-bottom:8px">
  <button class="btn btn-primary btn-sm" on:click={() => openSet()}
    >New Route</button
  >
</div>
<table class="table table-hover">
  <thead
    ><tr
      ><th>Path</th><th>Action</th><th>Template</th><th>Default</th><th></th></tr
    ></thead
  >
  <tbody>
    {#each items as it (it.path)}
      <tr>
        <td>{it.path}</td>
        <td>{it.dataAction}</td>
        <td>{it.template}</td>
        <td>{it.default ? "Yes" : ""}</td>
        <td align="right">
          <button
            class="btn btn-sm btn-outline-dark"
            on:click={() => openSet(it.path)}>Edit</button
          >
        </td>
      </tr>
    {/each}
  </tbody>
</table>
