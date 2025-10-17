<style>
  .hpm-spec-action-datax-attr-item td {
    padding: 0 2px 4px;
  }
</style>

<div id="hpm-spec-actionset" class="hpm-modal-formset">
  <div id="hpm-spec-actionset-alert"></div>

  <input type="hidden" name="modname" value="{[=it._modname]}" />

  <div class="d-flex flex-sm-row">
    <label class="form-label">Name</label>
    <div class="flex-fill">
      <input
        type="text"
        class="form-control form-control-sm"
        name="name"
        placeholder="Action Name"
        value="{[=it.name]}"
        {[?
        it.name]}
        readonly{[?]}
      />
    </div>
  </div>

  <div class="d-flex flex-sm-row">
    <label class="form-label">Datax</label>
    <div class="flex-fill">
      <table class="table table-condensed" width="100%">
        <thead>
          <tr>
            <th>Name</th>
            <th>Query Table</th>
            <th>Pager</th>
            <th>Type</th>
            <th>Limit</th>
            <th>Order</th>
            <th>Cache TTL (ms)</th>
            <th></th>
          </tr>
        </thead>
        <tbody id="hpm-spec-action-dataxs">
          {[~it.datax :v]}
          <tr id="datax-seq-{[=v._seqid]}" class="hpm-spec-action-datax-item">
            <td>
              <input
                type="text"
                class="form-control form-control-sm"
                name="datax_name"
                size="10"
                value="{[=v.name]}"
                readonly
              />
            </td>
            <td>
              <select
                class="form-control form-control-sm"
                name="datax_query_table"
              >
                {[~v.query_table_items :v2]}
                <option value="{[=v2.value]}" {[? v2._selected ]} selected{[?]}>
                  {[=v2.display_name]}
                </option>
                {[~]}
              </select>
            </td>
            <td>
              <select class="form-control form-control-sm" name="datax_pager">
                <option value="true" {[ if (v.pager) { ]}selected{[ } ]}>
                  YES
                </option>
                <option value="false" {[ if (!v.pager) { ]}selected{[ } ]}>
                  NO
                </option>
              </select>
            </td>
            <td>
              <select class="form-control form-control-sm" name="datax_type">
                {[~it._datax_typedef :fv]}
                <option
                  value="{[=fv.type]}"
                  {[?
                  hpMgr.Equal(fv.type,v.type.slice(5))]}
                  selected{[?]}
                >
                  {[=fv.name]}
                </option>
                {[~]}
              </select>
            </td>
            <td>
              <input
                type="text"
                class="form-control form-control-sm"
                name="datax_query_limit"
                size="4"
                value="{[=v.query.limit]}"
              />
            </td>
            <td>
              <input
                type="text"
                class="form-control form-control-sm"
                name="datax_query_order"
                size="4"
                value="{[=v.query.order]}"
              />
            </td>
            <td>
              <input
                type="text"
                class="form-control form-control-sm"
                name="datax_cache_ttl"
                size="4"
                value="{[=v.cache_ttl]}"
              />
            </td>
            <td align="right">
              <button
                class="btn btn-outline-dark btn-sm"
                onclick="hpSpec.ActionSetDataxDel(this)"
              >
                &times;
              </button>
            </td>
          </tr>
          {[~]}
        </tbody>
      </table>
    </div>
  </div>
</div>

<script id="hpm-spec-action-datax-item-tpl" type="text/html">
  <tr id="datax-seq-{[=it._seqid]}" class="hpm-spec-action-datax-item">
    <td><input type="text" class="form-control form-control-sm" name="datax_name" size="10" value=""></td>
    <td>
      <select class="form-control form-control-sm" name="datax_query_table">
      {[~it._nodeModels :nmv]}
        <option value="node.{[=nmv.meta.name]}" >node : {[=nmv.meta.name]}
        </option>
      {[~]}
      {[~it._termModels :tmv]}
        <option value="term.{[=tmv.meta.name]}" >term : {[=tmv.meta.name]}
        </option>
      {[~]}
      </select>
    </td>
    <td>
      <select class="form-control form-control-sm" name="datax_pager">
        <option value="true">YES</option>
        <option value="false" selected>NO</option>
      </select>
    </td>
    <td>
      <select class="form-control form-control-sm" name="datax_type">
      {[~it._datax_typedef :fv]}
        <option value="{[=fv.type]}" {[? hpMgr.Equal(fv.type,"list")]} selected{[?]}>{[=fv.name]}</option>
      {[~]}
      </select>
    </td>
    <td>
      <input type="text" class="form-control form-control-sm" name="datax_query_limit" size="4" value="1">
    </td>
    <td>
      <input type="text" class="form-control form-control-sm" name="datax_query_order" size="4" value="">
    </td>
    <td>
      <input type="text" class="form-control form-control-sm" name="datax_cache_ttl" size="4" value="0">
    </td>
  </tr>
</script>
