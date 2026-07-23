<script lang="ts">
  // spec Action edit form (modal body / carousel pagelet). Extracted from
  // ActionMgr so list↔form navigation goes through the modal carousel (slide +
  // resize) instead of an in-place {#if} swap. Ports spec/action/set.tpl +
  // spec.js ActionSet/ActionSetCommit/ActionDel. Each datax row: name, type
  // (list/entry), query table (node.X / term.X), limit, order, pager, cache_ttl.
  // On save, type is prefixed node.|term. per the table selection and the table
  // is sliced to the bare name. Save → spec-action-set; Delete → spec-action-del.
  import { onMount } from "svelte";
  import { api, ApiError } from "../../lib/api";
  import { closeModal, patchTopModal, type ModalButton } from "../../lib/modal";
  import { innerShow } from "../../lib/alert";
  import Alert from "../../lib/Alert.svelte";
  import {
    dataxTypedef,
    actiondef,
    generalOnoff,
    namereg,
    objectClone,
    withStableScroll,
  } from "./defs";

  let {
    modname = "",
    action = undefined,
    nodeModels = [],
    termModels = [],
    onSaved = () => {},
  }: {
    modname?: string;
    action?: any;
    nodeModels?: any[];
    termModels?: any[];
    onSaved?: () => void;
  } = $props();

  // $state makes `form` deeply reactive — push/splice on the datax array and
  // every bind:value update automatically (no more form.x = form.x re-trigger).
  let form: any = $state({
    ...objectClone(actiondef),
    modname,
    datax: [normDatax({})],
  });
  let editing = $state(false);
  const alertId = "hpm-spec-actionset-alert";

  // Split the stored "node.list"/"term.entry" type into a bare type (list/entry)
  // and a qualified _table (node.X / term.X) for the two dropdowns, and coerce
  // pager to the generalOnoff 'true'/'false' string so the ON/OFF select always
  // reflects the real value.
  function normDatax(d: any) {
    const type = d.type || "node.list";
    const bareType = type.split(".")[1] || "list";
    const table =
      type.split(".")[0] && d.query?.table
        ? type.split(".")[0] + "." + d.query.table
        : "";
    return {
      name: d.name || "",
      type: bareType,
      _table: table,
      query: {
        table: d.query?.table || "",
        limit: d.query?.limit || 10,
        order: d.query?.order || "",
      },
      pager: d.pager === true || d.pager === "true" ? "true" : "false",
      cache_ttl: d.cache_ttl || 0,
    };
  }

  onMount(() => {
    innerShow(alertId, "", ""); // clear stale banner from a prior instance
    const buttons: ModalButton[] = [
      { title: "Save", class: "btn-primary", click: save, dismiss: false },
    ];
    if (action) {
      editing = true;
      form = {
        ...objectClone(action),
        kind: "SpecAction",
        datax: (action.datax || []).map((d: any) => normDatax(d)),
      };
      buttons.push({
        title: "Delete",
        class: "btn-danger",
        click: del,
        dismiss: false,
      });
    }
    buttons.push({ title: "Cancel", class: "btn-outline-primary" });
    patchTopModal({ buttons });
  });

  function addDatax() {
    form.datax.push(normDatax({}));
  }
  function delDatax(i: number) {
    withStableScroll(() => form.datax.splice(i, 1));
  }

  async function save() {
    try {
      if (!namereg.test(form.name)) throw "Invalid Action Name";
      const req: any = { name: form.name, modname, datax: [] };
      for (const d of form.datax) {
        if (!d.name) continue;
        if (!namereg.test(d.name)) throw "Invalid Datax Name : " + d.name;
        let type = d.type;
        const tbl = d._table;
        if (tbl.startsWith("node.")) type = "node." + type;
        else if (tbl.startsWith("term.")) type = "term." + type;
        else throw "Invalid Query Table Name : " + tbl;
        const bare = tbl.slice(tbl.indexOf(".") + 1);
        if (!namereg.test(bare)) throw "Invalid Query Table Name : " + bare;
        req.datax.push({
          name: d.name,
          type,
          query: {
            table: bare,
            limit: parseInt(d.query.limit),
            order: d.query.order,
          },
          pager: d.pager === "true",
          cache_ttl: parseInt(d.cache_ttl),
        });
      }
      const rsp = await api.put("mod-set/spec-action-set", req);
      if (!rsp || rsp.kind !== "Action") return;
      innerShow(alertId, "success", "Successful updated");
      onSaved();
      setTimeout(closeModal, 600);
    } catch (e) {
      if (e instanceof ApiError) innerShow(alertId, "danger", e.message);
      else innerShow(alertId, "danger", String(e));
    }
  }

  async function del() {
    try {
      if (!namereg.test(form.name)) throw "Invalid Action Name";
      const rsp = await api.put("mod-set/spec-action-del", {
        name: form.name,
        modname,
        datax: [],
      });
      if (!rsp || rsp.kind !== "Action") return;
      innerShow(alertId, "success", "Successful updated");
      onSaved();
      setTimeout(closeModal, 600);
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
  <div class="row mb-3 align-items-center">
    <div class="col col-2">
      <label class="form-label">Action Name</label>
    </div>
    <div class="col">
      {#if editing}
        <input type="text" class="form-control" value={form.name} disabled />
      {:else}
        <input
          type="text"
          class="form-control"
          bind:value={form.name}
          placeholder="[a-z][a-z0-9_]+"
        />
      {/if}
    </div>
  </div>

  <div class="row mb-2 align-items-center">
    <div class="col-2">
      <label class="form-label">Datax</label>
    </div>
    <div class="col">
      {#each form.datax as d (d)}
        <div class="border rounded p-2 mb-2 hpm-module-datax-wrap">
          <div class="hpm-module-datax-grid">
            <div class="">
              <label class="form-label">Name</label>
              <input class="form-control form-control-sm" bind:value={d.name} />
            </div>
            <div class="">
              <label class="form-label">Type</label>
              <select class="form-select form-select-sm" bind:value={d.type}>
                {#each dataxTypedef as t (t.type)}<option value={t.type}
                    >{t.name}</option
                  >{/each}
              </select>
            </div>
            <div class="">
              <label class="form-label">Query Table</label>
              <select class="form-select form-select-sm" bind:value={d._table}>
                {#each nodeModels as m (m.meta.name)}<option
                    value={"node." + m.meta.name}>node : {m.meta.name}</option
                  >{/each}
                {#each termModels as m (m.meta.name)}<option
                    value={"term." + m.meta.name}>term : {m.meta.name}</option
                  >{/each}
              </select>
            </div>
            <div class="">
              <label class="form-label">Limit</label>
              <input
                class="form-control form-control-sm"
                bind:value={d.query.limit}
              />
            </div>
            <div class="">
              <label class="form-label">Order</label>
              <input
                class="form-control form-control-sm"
                bind:value={d.query.order}
              />
            </div>
            <div class="">
              <label class="form-label">Pager</label>
              <select class="form-select form-select-sm" bind:value={d.pager}>
                {#each generalOnoff as o (o.type)}<option value={o.type}
                    >{o.name}</option
                  >{/each}
              </select>
            </div>
            <div class="">
              <label class="form-label">Cache TTL</label>
              <input
                class="form-control form-control-sm"
                bind:value={d.cache_ttl}
              />
            </div>
          </div>
          <div class="hpm-module-datax-action">
            <button
              type="button"
              class="hpm-datax-close"
              title="Remove"
              aria-label="Remove"
              onclick={() => delDatax(form.datax.indexOf(d))}
            >
              <i class="bi bi-x-lg"></i>
            </button>
          </div>
        </div>
      {/each}
      <button
        type="button"
        class="btn hpm-btn-sm btn-outline-dark"
        onclick={addDatax}><i class="bi bi-plus-lg"></i> Datax</button
      >
    </div>
  </div>
</form>

<style>
  /* Each datax row is a flex row: the responsive grid fills the width and the
     Remove control sits in its own slot at the top-right corner — a small
     window-close × (muted by default, danger red on hover). Kept in flow
     (rather than absolute) so it can never overlap a wrapped grid cell. */
  .hpm-module-datax-wrap {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
  }
  .hpm-module-datax-grid {
    flex: 1 1 0;
    min-width: 0;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 200px), 1fr));
    gap: 0.5rem 1rem;
  }
  .hpm-module-datax-action {
    flex: 0 0 auto;
  }
  .hpm-datax-close {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1.75rem;
    height: 1.75rem;
    padding: 0;
    border: 0;
    border-radius: 0.375rem;
    background: transparent;
    color: var(--bs-secondary-color, #6c757d);
    line-height: 1;
    cursor: pointer;
    transition:
      background-color 0.15s ease,
      color 0.15s ease;
  }
  .hpm-datax-close:hover {
    background: rgba(220, 53, 69, 0.12);
    color: #dc3545;
  }
</style>
