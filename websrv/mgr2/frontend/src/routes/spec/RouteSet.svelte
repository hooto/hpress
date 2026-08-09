<script lang="ts">
  // spec Route edit form (modal body / carousel pagelet). Extracted from
  // RouteMgr so list↔form navigation goes through the modal carousel (slide +
  // resize) instead of an in-place {#if} swap. Ports spec/router/set.tpl +
  // spec.js RouteSet/RouteSetCommit/RouteDel + the template picker. Route
  // fields: path, dataAction, template (+ picker), params (key/value), default.
  // On Save → spec-route-set, then onSaved() + closeModal (slides back to list).
  // On Delete (edit only) → spec-route-del.
  import { onMount, untrack } from "svelte";
  import { api, ApiError } from "../../lib/api";
  import { closeModal, openModal, patchTopModal, type ModalButton } from "../../lib/modal";
  import { innerShow } from "../../lib/alert";
import { flashThen } from "../../lib/feedback";
  import Alert from "../../lib/Alert.svelte";
  import TemplatePicker from "./TemplatePicker.svelte";
  import { routedef, namereg, objectClone, generalOnoff, withStableScroll } from "./defs";

  let {
    modname = "",
    route = undefined,
    actions = [],
    onSaved = () => {},
  }: {
    modname?: string;
    route?: any;
    actions?: any[];
    onSaved?: () => void;
  } = $props();

  // $state makes `form` deeply reactive — push/splice on _params and every
  // bind:value update automatically.
  let form: any = $state(formFrom(objectClone(routedef), untrack(() => modname)));
  let editing = $state(false);
  const alertId = "hpm-spec-routeset-alert";

  // Flatten route.params ({k:v}) into an editable _params array of {key,value},
  // and normalize `default` to the generalOnoff 'true'/'false' string so the
  // ON/OFF select always reflects the real value (the old code bound a server
  // boolean against 0/1 number options, so a default route wrongly showed OFF).
  function formFrom(d: any, mod: string) {
    const params = d.params || {};
    return {
      ...d,
      kind: "SpecRoute",
      modname: mod,
      _params: Object.entries(params).map(([k, v]) => ({
        key: k,
        value: String(v),
      })),
      default: d.default ? "true" : "false",
    };
  }

  onMount(() => {
    innerShow(alertId, "", ""); // clear stale banner from a prior instance
    const buttons: ModalButton[] = [
      { title: "Save", class: "btn-primary", click: save, dismiss: false },
    ];
    if (route) {
      editing = true;
      form = formFrom(objectClone(route), modname);
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

  function addParam() {
    form._params.push({ key: "", value: "" });
  }
  function delParam(i: number) {
    withStableScroll(() => form._params.splice(i, 1));
  }

  function pickTemplate() {
    openModal({
      title: "Select a Template",
      width: 700,
      height: 500,
      body: TemplatePicker,
      props: { modname, onselect: (p: string) => (form.template = p) },
      buttons: [{ title: "Cancel", class: "btn-outline-primary" }],
    });
  }

  async function save() {
    try {
      const params: Record<string, string> = {};
      for (const p of form._params) {
        if (!p.key || !p.value) continue;
        if (!namereg.test(p.key)) throw "Invalid Param Name : " + p.key;
        params[p.key] = p.value;
      }
      const rsp = await api.put("mod-set/spec-route-set", {
        path: form.path,
        dataAction: form.dataAction,
        template: form.template,
        modname,
        params,
        default: form.default === true || form.default === "true",
      });
      if (!rsp || rsp.kind !== "SpecRoute") return;
      onSaved();
      flashThen(alertId, "success", "Successful updated", closeModal, 600);
    } catch (e) {
      if (e instanceof ApiError) innerShow(alertId, "danger", e.message);
      else innerShow(alertId, "danger", String(e));
    }
  }

  async function del() {
    try {
      const rsp = await api.put("mod-set/spec-route-del", {
        path: form.path,
        modname,
      });
      if (!rsp || rsp.kind !== "SpecRoute") return;
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
      <label class="form-label" for="routeset-path">Path</label>
    </div>
    <div class="col">
      {#if editing}
        <input id="routeset-path" type="text" class="form-control" value={form.path} disabled />
      {:else}
        <input id="routeset-path" type="text" class="form-control" bind:value={form.path} />
      {/if}
    </div>
  </div>

  <div class="row mb-2 align-items-center">
    <div class="col-2">
      <label class="form-label" for="routeset-action">Data Action</label>
    </div>
    <div class="col">
      <select id="routeset-action" class="form-select" bind:value={form.dataAction}>
        <option value=""></option>
        {#each actions as a (a.name)}<option value={a.name}>{a.name}</option>{/each}
      </select>
    </div>
  </div>

  <div class="row mb-2 align-items-center">
    <div class="col-2">
      <label class="form-label" for="routeset-template">Template</label>
    </div>
    <div class="col">
      <div class="input-group">
        <input id="routeset-template" type="text" class="form-control" bind:value={form.template} />
        <button
          type="button"
          class="btn btn-outline-dark"
          onclick={pickTemplate}>Select a Template</button
        >
      </div>
    </div>
  </div>

  <div class="row mb-2 align-items-center">
    <div class="col-2">
      <label class="form-label" for="routeset-default">Default</label>
    </div>
    <div class="col">
      <select id="routeset-default" class="form-select" bind:value={form.default}>
        {#each generalOnoff as o (o.type)}<option value={o.type}>{o.name}</option>{/each}
      </select>
    </div>
  </div>

  <div class="row mb-2 align-items-start">
    <div class="col-2">
      <span class="form-label">Params</span>
    </div>
    <div class="col">
      {#each form._params as p, i (i)}
        <div class="input-group mb-1">
          <input
            class="form-control form-control-sm"
            placeholder="key"
            bind:value={p.key}
          />
          <input
            class="form-control form-control-sm"
            placeholder="value"
            bind:value={p.value}
          />
          <button
            type="button"
            class="btn btn-sm btn-outline-danger"
            onclick={() => delParam(i)}>x</button
          >
        </div>
      {/each}
      <button type="button" class="btn hpm-btn-sm btn-outline-dark" onclick={addParam}
        ><i class="bi bi-plus-lg"></i> Param</button
      >
    </div>
  </div>
</form>
