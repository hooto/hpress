<script lang="ts">
  // spec NodeModel edit form (modal body / carousel pagelet). Extracted from
  // NodeMgr so list↔form navigation goes through the modal carousel (slide +
  // resize) instead of an in-place {#if} swap. Ports spec/node/set.tpl +
  // spec.js NodeSet/NodeSetCommit. Edits fields, attached terms, extensions.
  // On Save → spec-node-set, then onSaved() + closeModal (slides back to list).
  import { onMount } from "svelte";
  import { api, ApiError } from "../../lib/api";
  import { closeModal, patchTopModal } from "../../lib/modal";
  import { innerShow } from "../../lib/alert";
  import { flashThen } from "../../lib/feedback";
  import Alert from "../../lib/Alert.svelte";
  import {
    nodedef,
    fieldTypedef,
    fieldIdxTypedef,
    generalOnoff,
    permalinkDef,
    namereg,
    objectClone,
    withStableScroll,
  } from "./defs";

  let {
    modname = "",
    modelid = undefined,
    onSaved = () => {},
  }: { modname?: string; modelid?: string; onSaved?: () => void } = $props();

  // $state makes `form` deeply reactive — push/splice on the nested arrays and
  // every bind:value update automatically.
  let form: any = $state(objectClone(nodedef));
  let editing = $state(false);
  const alertId = "hpm-spec-nodeset-alert";

  function normNode(d: any) {
    d.fields = (d.fields || []).map((f: any) => ({
      ...f,
      length: f.length || 0,
      indexType: f.indexType || 0,
      attrs: f.attrs || [],
    }));
    d.terms = d.terms || [];
    d.extensions = d.extensions || {};
    const ext = d.extensions;
    // generalOnoff options are the strings 'true'/'false'; coerce so an empty
    // extension shows OFF. save() maps these back to booleans.
    ext.access_counter = ext.access_counter ? "true" : "false";
    ext.comment_enable = ext.comment_enable ? "true" : "false";
    ext.comment_perentry = ext.comment_perentry ? "true" : "false";
    ext.text_search = ext.text_search ? "true" : "false";
    ext.node_refer = ext.node_refer || "";
    ext.permalink = ext.permalink || "";
    return d;
  }

  onMount(async () => {
    innerShow(alertId, "", ""); // clear stale banner from a prior instance
    // Save/Cancel live in the modal's fixed footer (only the body scrolls).
    patchTopModal({
      buttons: [
        { title: "Save", class: "btn-primary", click: save, dismiss: false },
        { title: "Cancel", class: "btn-outline-primary" },
      ],
    });
    if (modelid) {
      editing = true;
      try {
        const data = await api.get<any>("node-model/entry", {
          modname,
          modelid,
        });
        if (data && data.kind === "NodeModel") {
          form = normNode({ ...data, _modname: modname });
        }
      } catch {
        /* ignore */
      }
    } else {
      form = normNode({ ...objectClone(nodedef), modname, _modname: modname });
    }
  });

  function addField() {
    form.fields.push({
      name: "",
      title: "",
      type: "string",
      length: 0,
      indexType: 0,
      attrs: [],
    });
  }
  function addAttr(f: any) {
    f.attrs.push({ key: "", value: "" });
  }
  function addTerm() {
    form.terms.push({ meta: { name: "" }, title: "", type: "taxonomy" });
  }

  // Removing a row can leave the scrollable pagelet scrolled past the now-
  // shorter content (looks blank). withStableScroll keeps the scroll position
  // stable across the mutation (shared helper in lib/util.ts).
  function delField(i: number) {
    withStableScroll(() => form.fields.splice(i, 1));
  }
  function delAttr(f: any, i: number) {
    withStableScroll(() => f.attrs.splice(i, 1));
  }
  function delTerm(i: number) {
    withStableScroll(() => form.terms.splice(i, 1));
  }

  async function save() {
    try {
      const req: any = {
        meta: { name: form.meta.name },
        title: form.title,
        modname,
        fields: [],
        terms: [],
        extensions: {
          access_counter:
            form.extensions.access_counter === true ||
            form.extensions.access_counter === "true",
          comment_enable:
            form.extensions.comment_enable === true ||
            form.extensions.comment_enable === "true",
          comment_perentry:
            form.extensions.comment_perentry === true ||
            form.extensions.comment_perentry === "true",
          node_refer: form.extensions.node_refer || "",
          text_search:
            form.extensions.text_search === true ||
            form.extensions.text_search === "true",
          permalink: form.extensions.permalink || "",
        },
      };
      for (const f of form.fields) {
        if (!f.name) continue;
        if (!namereg.test(f.name)) throw "Invalid Field Name : " + f.name;
        const attrs = [];
        for (const a of f.attrs) {
          if (a.key) {
            if (!namereg.test(a.key))
              throw "Invalid Field Attribute Key : " + a.key;
            attrs.push({ key: a.key, value: a.value });
          }
        }
        req.fields.push({
          name: f.name,
          title: f.title,
          type: f.type,
          length: f.length,
          indexType: parseInt(f.indexType),
          attrs,
        });
      }
      for (const t of form.terms) {
        if (!t.meta.name) continue;
        if (!namereg.test(t.meta.name))
          throw "Invalid Term Name : " + t.meta.name;
        req.terms.push({
          meta: { name: t.meta.name },
          title: t.title,
          type: t.type,
        });
      }
      const rsp = await api.put("mod-set/spec-node-set", req);
      if (!rsp || rsp.kind !== "NodeModel") return;
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
      <span class="form-label">Metadata</span>
    </div>
    <div class="col">
      <div class="row">
        <div class=" col-6">
          <label class="form-label" for="nodeset-name">Node Model Name</label>
          {#if editing}
            <input
              id="nodeset-name"
              type="text"
              class="form-control"
              value={form.meta.name}
              disabled
            />
          {:else}
            <input
              id="nodeset-name"
              type="text"
              class="form-control"
              bind:value={form.meta.name}
              placeholder="[a-z][a-z0-9_]+"
            />
          {/if}
        </div>
        <div class=" col-6">
          <label class="form-label" for="nodeset-title">Title</label>
          <input id="nodeset-title" type="text" class="form-control" bind:value={form.title} />
        </div>
      </div>
    </div>
  </div>

  <div class="row mb-2 align-items-center">
    <div class="col-2">
      <span class="form-label">Fields</span>
    </div>
    <div class="col">
      <table class="table hpm-table-std">
        <thead>
          <tr>
            <th>Name</th>
            <th>Title</th>
            <th>Type</th>
            <th>Length</th>
            <th>Index</th>
            <th style="min-width: 12rem">Attributes</th>
            <th style="min-width: 4rem"></th>
          </tr>
        </thead>
        <tbody>
          {#each form.fields as f (f)}
            <tr>
              <td
                ><input
                  class="form-control form-control-sm"
                  bind:value={f.name}
                /></td
              >
              <td
                ><input
                  class="form-control form-control-sm"
                  bind:value={f.title}
                /></td
              >
              <td>
                <select class="form-select form-select-sm" bind:value={f.type}>
                  {#each fieldTypedef as t}<option value={t.type}
                      >{t.name}</option
                    >{/each}
                </select>
              </td>
              <td
                ><input
                  class="form-control form-control-sm"
                  style="width:70px"
                  bind:value={f.length}
                /></td
              >
              <td>
                <select
                  class="form-select form-select-sm"
                  bind:value={f.indexType}
                >
                  {#each fieldIdxTypedef as t}<option value={t.type}
                      >{t.name}</option
                    >{/each}
                </select>
              </td>
              <td>
                {#each f.attrs as a (a)}
                  <div class="d-flex mb-1">
                    <input
                      class="form-control form-control-sm me-1"
                      style="width:70px"
                      placeholder="key"
                      bind:value={a.key}
                    />
                    <input
                      class="form-control form-control-sm me-1"
                      style="width:120px"
                      placeholder="value"
                      bind:value={a.value}
                    />
                    <button
                      type="button"
                      class="btn btn-sm btn-link text-danger"
                      title="Remove"
                      aria-label="Remove attribute"
                      onclick={() => delAttr(f, f.attrs.indexOf(a))}
                      ><i class="bi bi-x-lg"></i></button
                    >
                  </div>
                {/each}
                <button
                  type="button"
                  class="btn hpm-btn-sm btn-outline-dark"
                  onclick={() => addAttr(f)}
                  ><i class="bi bi-plus-lg"></i> Attr</button
                >
              </td>
              <td class="text-end"
                ><button
                  type="button"
                  class="btn btn-sm btn-outline-danger"
                  onclick={() => delField(form.fields.indexOf(f))}>Del</button
                ></td
              >
            </tr>
          {/each}
        </tbody>
      </table>
      <button
        type="button"
        class="btn hpm-btn-sm btn-outline-dark"
        onclick={addField}><i class="bi bi-plus-lg"></i> Add Field</button
      >
    </div>
  </div>

  <div class="row mb-2 align-items-center">
    <div class="col-2">
      <span class="form-label">Terms</span>
    </div>
    <div class="col">
      <table class="table table-sm hpm-table-std">
        <thead>
          <tr>
            <th>Name</th>
            <th>Title</th>
            <th>Type</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {#each form.terms as t (t)}
            <tr>
              <td
                ><input
                  class="form-control form-control-sm"
                  style="width:140px"
                  bind:value={t.meta.name}
                /></td
              >
              <td
                ><input
                  class="form-control form-control-sm"
                  style="width:160px"
                  bind:value={t.title}
                /></td
              >
              <td>
                <select
                  class="form-select form-select-sm"
                  style="width:130px"
                  bind:value={t.type}
                >
                  <option value="taxonomy">Categories</option>
                  <option value="tag">Tags</option>
                </select>
              </td>
              <td class="text-end"
                ><button
                  type="button"
                  class="btn btn-sm btn-outline-danger"
                  onclick={() => delTerm(form.terms.indexOf(t))}>Del</button
                ></td
              >
            </tr>
          {/each}
        </tbody>
      </table>
      <button
        type="button"
        class="btn hpm-btn-sm btn-outline-dark"
        onclick={addTerm}><i class="bi bi-plus-lg"></i> Add Term</button
      >
    </div>
  </div>

  <div class="row mb-2 align-items-center">
    <div class="col-2">
      <span class="form-label">External</span>
    </div>
    <div class="col">
      <div class="hpm-ext-grid">
        <div>
          <label class="form-label" for="nodeset-access-counter">Access Counter</label>
          <select
            id="nodeset-access-counter"
            class="form-select form-select-sm"
            bind:value={form.extensions.access_counter}
          >
            {#each generalOnoff as o}<option value={o.type}>{o.name}</option
              >{/each}
          </select>
        </div>
        <div>
          <label class="form-label" for="nodeset-text-search">Text Search</label>
          <select
            id="nodeset-text-search"
            class="form-select form-select-sm"
            bind:value={form.extensions.text_search}
          >
            {#each generalOnoff as o}<option value={o.type}>{o.name}</option
              >{/each}
          </select>
        </div>
        <div>
          <label class="form-label" for="nodeset-comment-enable">Comment Enable</label>
          <select
            id="nodeset-comment-enable"
            class="form-select form-select-sm"
            bind:value={form.extensions.comment_enable}
          >
            {#each generalOnoff as o}<option value={o.type}>{o.name}</option
              >{/each}
          </select>
        </div>
        <div>
          <label class="form-label" for="nodeset-comment-perentry">Comment Per-Entry</label>
          <select
            id="nodeset-comment-perentry"
            class="form-select form-select-sm"
            bind:value={form.extensions.comment_perentry}
          >
            {#each generalOnoff as o}<option value={o.type}>{o.name}</option
              >{/each}
          </select>
        </div>
        <div>
          <label class="form-label" for="nodeset-permalink">Permalink</label>
          <select
            id="nodeset-permalink"
            class="form-select form-select-sm"
            bind:value={form.extensions.permalink}
          >
            {#each permalinkDef as p}<option value={p.type}>{p.name}</option
              >{/each}
          </select>
        </div>
      </div>
    </div>
  </div>
  <div class="row mb-2 align-items-center">
    <div class="col-2">
      <label class="form-label" for="nodeset-node-refer">Node Refer</label>
    </div>
    <div class="col">
      <input
        id="nodeset-node-refer"
        type="text"
        class="form-control"
        bind:value={form.extensions.node_refer}
      />
    </div>
  </div>
</form>

<style>
  /* Responsive Extensions grid. The max MUST be 1fr (not a fixed px) so the
     cells stretch to FILL the row — a fixed max (e.g. 350px) leaves the
     remainder blank when a row can't fit another whole cell. min 150px makes
     cells wrap once the column would drop below that; auto-fit collapses empty
     trailing tracks so a partial last row still fills. min(100%, 150px) keeps a
     single cell from overflowing on very narrow modals. */
  .hpm-ext-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 200px), 1fr));
    gap: 0.5rem 1rem;
  }
</style>
