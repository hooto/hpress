<style>
  .hpm-spec-node-field-attr-item td {
    padding: 0 2px 4px;
  }
</style>

<div id="hpm-spec-nodeset" class="hpm-modal-formset">
  <div id="hpm-spec-nodeset-alert" class=""></div>

  <input type="hidden" name="modname" value="{[=it._modname]}" />

  {[? it.meta.name]}
  <input type="hidden" name="name" value="{[=it.meta.name]}" />
  {[??]}
  <div class="d-flex flex-sm-row">
    <label class="form-label">Name</label>
    <div class="flex-fill">
      <input
        type="text"
        class="form-control form-control-sm"
        name="name"
        placeholder="Node Name"
        value="{[=it.meta.name]}"
      />
    </div>
  </div>
  {[?]}

  <div class="d-flex flex-sm-row">
    <label class="form-label">Title</label>
    <div class="flex-fill">
      <input
        type="text"
        class="form-control form-control-sm"
        name="title"
        placeholder="Title"
        value="{[=it.title]}"
      />
    </div>
  </div>

  <div class="d-flex flex-sm-row">
    <label class="form-label">Fields</label>
    <div class="flex-fill">
      <table class="table table-condensed" width="100%">
        <thead>
          <tr>
            <th>Name</th>
            <th>Title</th>
            <th>Type</th>
            <th>Length</th>
            <th>Index Type</th>
            <th>Extended attributes</th>
            <th></th>
          </tr>
        </thead>
        <tbody id="hpm-spec-node-fields">
          {[~it.fields :v]}
          <tr id="field-seq-{[=v._seqid]}" class="hpm-spec-node-field-item">
            <td>
              <input
                type="text"
                class="form-control form-control-sm"
                name="field_name"
                size="10"
                value="{[=v.name]}"
                readonly
              />
            </td>
            <td>
              <input
                type="text"
                class="form-control form-control-sm"
                name="field_title"
                size="20"
                value="{[=v.title]}"
              />
            </td>
            <td>
              <select class="form-select form-select-sm" name="field_type">
                {[~it._field_typedef :fv]}
                <option
                  value="{[=fv.type]}"
                  {[?
                  hpMgr.Equal(fv.type,v.type)]}
                  selected{[?]}
                >
                  {[=fv.name]}
                </option>
                {[~]}
              </select>
            </td>
            <td>
              <input
                class="form-control form-control-sm"
                type="text"
                name="field_length"
                size="3"
                value="{[=v.length]}"
              />
            </td>
            <td>
              <select
                class="form-select form-select-sm"
                name="field_index_type"
              >
                {[~it._field_idx_typedef :fv]}
                <option
                  value="{[=fv.type]}"
                  {[?
                  hpMgr.Equal(fv.type,v.indexType)]}
                  selected{[?]}
                >
                  {[=fv.name]}
                </option>
                {[~]}
              </select>
            </td>
            <td>
              <table>
                <tbody class="hpm-spec-node-field-attrs">
                  {[~v.attrs :atv]}
                  <tr class="hpm-spec-node-field-attr-item">
                    <td>
                      <input
                        type="text"
                        class="form-control form-control-sm"
                        name="field_attr_key"
                        size="8"
                        value="{[=atv.key]}"
                      />
                    </td>
                    <td>
                      <input
                        type="text"
                        class="form-control form-control-sm"
                        name="field_attr_value"
                        size="16"
                        value="{[=atv.value]}"
                      />
                    </td>
                  </tr>
                  {[~]}
                </tbody>
              </table>
            </td>
            <td>
              <button
                class="btn btn-outline-dark btn-sm"
                onclick="hpSpec.NodeSetFieldAttrAppend('{[=v._seqid]}')"
              >
                + Attribute
              </button>
            </td>
          </tr>
          {[~]}
        </tbody>
      </table>
      <div class="d-flex justify-content-start" style="margin-top: 8px">
        <button
          class="btn btn-sm btn-outline-dark"
          onclick="hpSpec.NodeSetFieldAppend()"
        >
          New Field
        </button>
      </div>
    </div>
  </div>

  <div class="d-flex flex-sm-row">
    <label class="form-label">Terms</label>
    <div class="flex-fill">
      <table class="table table-condensed" width="100%">
        <thead>
          <tr>
            <th>Name</th>
            <th>Title</th>
            <th>Type</th>
          </tr>
        </thead>
        <tbody id="hpm-spec-node-terms">
          {[~it.terms :v]}
          <tr id="field-seq-{[=v._seqid]}" class="hpm-spec-node-term-item">
            <td>
              <input
                type="text"
                class="form-control form-control-sm"
                name="term_name"
                size="20"
                value="{[=v.meta.name]}"
                readonly
              />
            </td>
            <td>
              <input
                type="text"
                class="form-control form-control-sm"
                name="term_title"
                size="30"
                value="{[=v.title]}"
              />
            </td>
            <td>
              <select class="form-select form-select-sm" name="term_type">
                {[~it._term_typedef :fv]}
                <option
                  value="{[=fv.type]}"
                  {[?
                  hpMgr.Equal(fv.type,v.type)]}
                  selected{[?]}
                >
                  {[=fv.name]}
                </option>
                {[~]}
              </select>
            </td>
            {[~]}
          </tr>
        </tbody>
      </table>
      <div class="d-flex justify-content-start" style="margin-top: 8px">
        <button
          class="btn btn-sm btn-outline-dark"
          onclick="hpSpec.NodeSetTermAppend()"
        >
          New Term
        </button>
      </div>
    </div>
  </div>

  <div class="d-flex flex-sm-row">
    <label class="form-label">Extensions</label>
    <div class="flex-fill">
      <table class="table table-condensed align-middle" width="100%">
        <thead>
          <tr>
            <th>Option</th>
            <th>Attributes</th>
          </tr>
        </thead>
        <tbody id="hpm-spec-node-exts">
          <tr>
            <td>Access Counter</td>
            <td>
              <select
                class="form-select form-select-sm"
                name="ext_access_counter"
              >
                {[~it._general_onoff :gv]}
                <option
                  value="{[=gv.type]}"
                  {[?
                  hpMgr.Equal(it.extensions.access_counter,gv.type)]}
                  selected{[?]}
                >
                  {[=gv.name]}
                </option>
                {[~]}
              </select>
            </td>
          </tr>
          <tr>
            <td>Comment Enable</td>
            <td>
              <select
                class="form-select form-select-sm"
                name="ext_comment_enable"
              >
                {[~it._general_onoff :gv]}
                <option
                  value="{[=gv.type]}"
                  {[?
                  hpMgr.Equal(it.extensions.comment_enable,gv.type)]}
                  selected{[?]}
                >
                  {[=gv.name]}
                </option>
                {[~]}
              </select>
            </td>
          </tr>
          <tr>
            <td>Comment On/Off Per Entry</td>
            <td>
              <select
                class="form-select form-select-sm"
                name="ext_comment_perentry"
              >
                {[~it._general_onoff :gv]}
                <option
                  value="{[=gv.type]}"
                  {[?
                  hpMgr.Equal(it.extensions.comment_perentry,gv.type)]}
                  selected{[?]}
                >
                  {[=gv.name]}
                </option>
                {[~]}
              </select>
            </td>
          </tr>
          <tr>
            <td>Permalink Settings</td>
            <td>
              <select class="form-select form-select-sm" name="ext_permalink">
                {[~it._permalink_def :gv]}
                <option
                  value="{[=gv.type]}"
                  {[?
                  hpMgr.Equal(it.extensions.permalink,gv.type)]}
                  selected{[?]}
                >
                  {[=gv.name]}
                </option>
                {[~]}
              </select>
            </td>
          </tr>
          {[if (it.extensions.node_sub_refer) {]}
          <tr>
            <td>Node Sub Refer</td>
            <td>{[=it.extensions.node_sub_refer]}</td>
          </tr>
          {[} else {]}
          <tr>
            <td>Refer to Node Name</td>
            <td>
              <input
                type="text"
                class="form-control form-control-sm"
                name="ext_node_refer"
                value="{[=it.extensions.node_refer]}"
              />
            </td>
          </tr>
          {[}]}
          <tr>
            <td>Full Text Search Enable</td>
            <td>
              <select class="form-select form-select-sm" name="ext_text_search">
                {[~it._general_onoff :gv]}
                <option
                  value="{[=gv.type]}"
                  {[?
                  hpMgr.Equal(it.extensions.text_search,gv.type)]}
                  selected{[?]}
                >
                  {[=gv.name]}
                </option>
                {[~]}
              </select>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</div>

<script id="hpm-spec-node-field-item-tpl" type="text/html">
  <tr id="field-seq-{[=it._seqid]}" class="hpm-spec-node-field-item">
    <td>
      <input
        type="text"
        class="form-control form-control-sm"
        name="field_name"
        size="10"
        value=""
      />
    </td>
    <td>
      <input
        type="text"
        class="form-control form-control-sm"
        name="field_title"
        size="16"
        value=""
      />
    </td>
    <td>
      <select class="form-select form-select-sm" name="field_type">
        {[~it._field_typedef :fv]}
        <option
          value="{[=fv.type]}"
          {[?
          hpMgr.Equal(fv.type,it._type)]}
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
        name="field_length"
        size="5"
        value="0"
      />
    </td>
    <td>
      <select class="form-select form-select-sm" name="field_index_type">
        {[~it._field_idx_typedef :fv]}
        <option
          value="{[=fv.type]}"
          {[?
          hpMgr.Equal(fv.type,it._indexType)]}
          selected{[?]}
        >
          {[=fv.name]}
        </option>
        {[~]}
      </select>
    </td>
    <td>
      <table>
        <tbody class="hpm-spec-node-field-attrs"></tbody>
      </table>
    </td>
    <td>
      <button
        class="btn btn-outline-dark btn-sm"
        onclick="hpSpec.NodeSetFieldAttrAppend('{[=it._seqid]}')"
      >
        + Attribute
      </button>
    </td>
  </tr>
</script>

<script id="hpm-spec-node-field-attr-item-tpl" type="text/html">
  <tr class="hpm-spec-node-field-attr-item">
    <td>
      <input
        type="text"
        class="form-control form-control-sm"
        name="field_attr_key"
        size="8"
        value=""
      />
    </td>
    <td>
      <input
        type="text"
        class="form-control form-control-sm"
        name="field_attr_value"
        size="12"
        value=""
      />
    </td>
  </tr>
</script>

<script id="hpm-spec-node-term-item-tpl" type="text/html">
  <tr id="field-seq-{[=it._seqid]}" class="hpm-spec-node-term-item">
    <td>
      <input
        type="text"
        class="form-control form-control-sm"
        name="term_name"
        size="20"
        value=""
      />
    </td>
    <td>
      <input
        type="text"
        class="form-control form-control-sm"
        name="term_title"
        size="30"
        value=""
      />
    </td>
    <td>
      <select class="form-select form-select-sm" name="term_type">
        {[~it._term_typedef :fv]}
        <option
          value="{[=fv.type]}"
          {[?
          hpMgr.Equal(fv.type,it._type)]}
          selected{[?]}
        >
          {[=fv.name]}
        </option>
        {[~]}
      </select>
    </td>
  </tr>
</script>
