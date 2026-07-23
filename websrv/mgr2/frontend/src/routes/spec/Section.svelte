<script lang="ts">
  // spec module dashboard. Ports spec/index.tpl + spec/list.tpl + spec.js
  // Index/List. Lists modules with counts; row actions: Develop (file IDE),
  // Setting (InfoSet). Count buttons open resource Manager modals. Toolbar:
  // Upload (package install/upgrade), New Module (InfoSet). On InfoSet save,
  // refreshes the global spec list (so the node nav updates).
  import { onMount } from "svelte";
  import { api, ApiError } from "../../lib/api";
  import { openModal } from "../../lib/modal";
  import { innerShow } from "../../lib/alert";
  import { refreshSpecList } from "../../lib/boot";
  import { navigate } from "../../lib/router";
  import { unixTimeFormat } from "../../lib/util";
  import Alert from "../../lib/Alert.svelte";
  import InfoSet from "./InfoSet.svelte";
  import Upload from "./Upload.svelte";
  import TermMgr from "./TermMgr.svelte";
  import NodeMgr from "./NodeMgr.svelte";
  import ActionMgr from "./ActionMgr.svelte";
  import RouteMgr from "./RouteMgr.svelte";
  import type { Spec } from "../../lib/types";

  export let route = "spec/index";

  let items: any[] = [];

  async function load() {
    try {
      const rsj = await api.get<{ kind?: string; items?: Spec[] }>(
        "mod-set/spec-list",
      );
      if (
        !rsj ||
        rsj.kind !== "SpecList" ||
        !rsj.items ||
        rsj.items.length < 1
      ) {
        items = [];
        innerShow("hpm-specls-alert", "info", "Item Not Found");
        return;
      }
      innerShow("hpm-specls-alert", "", "");
      items = rsj.items.map((s: any) => ({
        ...s,
        _nodeModelsNum: (s.nodeModels || []).length,
        _actionsNum: (s.actions || []).length,
        _routesNum: (s.router?.routes || []).length,
        meta: { ...s.meta, created: s.meta.created || s.meta.updated },
      }));
    } catch (e) {
      if (!(e instanceof ApiError && e.code === "Unauthorized")) items = [];
    }
  }

  function infoSet(name?: string) {
    openModal({
      title: name ? "Info Settings" : "New Module",
      width: 1200,
      height: 800,
      body: InfoSet,
      props: {
        name,
        onSaved: () => {
          load();
          refreshSpecList();
        },
      },
    });
  }

  function upload() {
    openModal({
      title: "Upload Package to Install or Upgrade Module",
      width: 700,
      height: 350,
      body: Upload,
      props: { onDone: load },
    });
  }

  function develop(name: string) {
    navigate("spec-editor/" + name);
  }

  function openMgr(which: "node" | "term" | "action" | "route", name: string) {
    const map = {
      node: NodeMgr,
      term: TermMgr,
      action: ActionMgr,
      route: RouteMgr,
    };
    const titles = {
      node: "Node List",
      term: "Term List",
      action: "Action List",
      route: "Route List",
    };
    openModal({
      title: titles[which],
      width: which === "node" ? 1000 : 900,
      height: 550,
      body: map[which],
      props: { modname: name },
    });
  }

  onMount(load);
</script>

<div class="hpm-block-gap-column">
  <div
    class="d-flex flex-row align-items-center hpm-block-gap-row-sm"
    style="margin-bottom:8px"
  >
    <button class="btn btn-primary" on:click={() => infoSet()}
      >New Module</button
    >
    <button class="btn btn-outline-primary ms-auto" on:click={upload}>
      Install or Upgrade from Package
    </button>
  </div>

  <Alert id="hpm-specls-alert" />

  <div class="hpm-table-std">
    {#if items.length}
      <table class="table table-hover align-middle">
        <thead>
          <tr>
            <th>Title</th>
            <th>Name</th>
            <th>Service Name</th>
            <th>Version</th>
            <th>Nodes</th>
            <th>Actions</th>
            <th>Routes</th>
            <th>Status</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {#each items as v (v.meta.name)}
            <tr>
              <td>{v.title}</td>
              <td>{v.meta.name}</td>
              <td>{v.srvname}</td>
              <td>{v.meta.version}</td>
              <td>
                <button
                  class="btn btn-sm btn-outline-dark"
                  on:click={() => openMgr("node", v.meta.name)}
                  >{v._nodeModelsNum}</button
                >
              </td>
              <td>
                <button
                  class="btn btn-sm btn-outline-dark"
                  on:click={() => openMgr("action", v.meta.name)}
                  >{v._actionsNum}</button
                >
              </td>
              <td>
                <button
                  class="btn btn-sm btn-outline-dark"
                  on:click={() => openMgr("route", v.meta.name)}
                  >{v._routesNum}</button
                >
              </td>
              <td>
                {#if v.status}
                  <span class="badge text-bg-success">Enable</span>
                {:else}
                  <span class="badge text-bg-secondary">Disable</span>
                {/if}
              </td>
              <td align="right">
                <button
                  class="btn btn-sm btn-outline-dark"
                  on:click={() => develop(v.meta.name)}>Develop</button
                >
                <button
                  class="btn btn-sm btn-outline-dark"
                  on:click={() => infoSet(v.meta.name)}>Setting</button
                >
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</div>
