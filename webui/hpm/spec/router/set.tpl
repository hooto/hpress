<style type="text/css">
  #hpm-spec-route-params td {
    padding: 0 10px 10px 0;
  }
</style>

<div>
  <div id="hpm-spec-routeset-alert"></div>
  <div id="hpm-spec-routeset" class="hpm-modal-formset"></div>
</div>

<script id="hpm-spec-routeset-tpl" type="text/html">
  <input type="hidden" name="modname" value="{[=it._modname]}" />

  <div class="d-flex flex-sm-row">
    <label class="form-label">Path</label>
    <div class="flex-fill">
      <input
        type="text"
        class="form-control form-control-sm"
        name="path"
        placeholder="Route Path"
        value="{[=it.path]}"
        {[?
        it.path]}
        readonly
        {[?]}
      />
    </div>
  </div>

  <div class="d-flex flex-sm-row">
    <label class="form-label">Data Action</label>
    <div class="flex-fill">
      <select class="form-control form-control-sm" name="data_action">
        {[~it._actions :v]}
        <option
          value="{[=v.name]}"
          {[?
          hpMgr.Equal(v.name,it.dataAction)]}
          selected{[?]}
        >
          {[=v.name]}
        </option>
        {[~]}
      </select>
    </div>
  </div>

  <div class="d-flex flex-sm-row">
    <label class="form-label">Template</label>
    <div class="flex-fill">
      <div class="input-group">
        <input
          id="hpm-spec-routeset-template"
          type="text"
          class="form-control form-control-sm"
          name="template"
          placeholder="Template Path"
          value="{[=it.template]}"
        />
        <button
          class="btn btn-sm btn-outline-secondary"
          type="button"
          onclick="hpSpec.RouteSetTemplateSelect('{[=it._modname]}')"
        >
          Select a Template
        </button>
      </div>
    </div>
  </div>

  <div class="d-flex flex-sm-row">
    <label class="form-label">Params</label>
    <div class="flex-fill">
      <table id="hpm-spec-route-params" width="100%"></table>
      <div class="d-flex justify-content-start" style="margin-top: 8px">
        <button
          class="btn btn-sm btn-outline-dark"
          onclick="hpSpec.RouteSetParamAppend()"
        >
          New Param
        </button>
      </div>
    </div>
  </div>

  <div class="d-flex flex-sm-row">
    <label class="form-label">Default</label>
    <div class="flex-fill">
      <select class="form-control form-control-sm" name="default">
        <option value="1" {[if (it.default) { ]}selected{[ } ]}>Yes</option>
        <option value="0" {[if (!it.default) { ]}selected{[ } ]}>No</option>
      </select>
    </div>
  </div>
</script>

<script id="hpm-spec-route-param-item-tpl" type="text/html">
  <tr class="hpm-spec-route-param-item">
    <td>
      <input
        type="text"
        class="form-control form-control-sm "
        name="param_key"
        size="16"
        placeholder="Param Name"
        value="{[=it._key]}"
      />
    </td>
    <td>
      <input
        type="text"
        class="form-control form-control-sm "
        name="param_value"
        size="32"
        placeholder="Param Value"
        value="{[=it._value]}"
      />
    </td>
  </tr>
</script>
