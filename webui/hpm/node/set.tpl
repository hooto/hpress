<div class="hpm-block-gap-column">
  <div id="hpm-nodeset-layout" class="hpm-block-gap-row">
    <div id="hpm-nodeset-laymain" style="width: 75%">loading</div>
    <div id="hpm-nodeset-layside" style="width: 25%"></div>
  </div>

  <div calss="hpm-block-gap-row-sm">
    <button class="btn btn-primary" onclick="hpNode.SetCommit()">Save</button>
    <button class="btn btn-outline-primary" onclick="hpNode.List()">
      Cancel
    </button>
  </div>
</div>

<div id="hpm-node-set-opts" class="d-none">
  <div id="hpm-node-set-opts-label">Content</div>
</div>

<script id="hpm-nodeset-tpl" type="text/html">
  <input type="hidden" name="id" value="{[=it.id]}" />
  <div id="hpm-nodeset-top-title"></div>
  <div id="hpm-nodeset-tops"></div>
  <div id="hpm-nodeset-fields"></div>
</script>

<script id="hpm-nodeset-tplstatus" type="text/html">
  <div class="hpm-nodeset-tplx">
    <label class="form-label">Status</label>
    <select name="status" class="form-select">
      {[~it._status_def :sv]}
      <option
        value="{[=sv.type]}"
        {[?
        hpMgr.Equal(sv.type,
        it.status)]}
        selected{[?]}
      >
        {[=sv.name]}
      </option>
      {[~]}
    </select>
  </div>
</script>

<script id="hpm-nodeset-tpltext" type="text/html">
  <div class="hpm-nodeset-tplx hpm-nodeset-tpltext">
    <label class="form-label">
      <span>{[=it.title]}</span>

      {[? it.attr_lang_list]}
      <select
        id="field_{[=it.name]}_langs"
        class="field-nav-lang form-select"
        onchange="hpNode.SetFieldLang('{[=it.name]}')"
      >
        {[~it.attr_lang_list :v]}
        <option value="{[=v.id]}">{[=v.name]}</option>
        {[~]}
      </select>
      {[?]}
    </label>

    <input
      type="hidden"
      id="field_{[=it.name]}_attr_format"
      name="field_{[=it.name]}_attr_format"
      value="{[=it.attr_format]}"
    />

    <div class="editor-outbox">
      <div id="field_{[=it.name]}_inner_toolbar" class="editor-inner-toolbar">
        <span id="field_{[=it.name]}_editor_nav" class="editor-nav ">
          {[~it._formats :v]}
          <button
            class="tpltext-editor-item editor-nav-{[=v.name]} btn btn-sm btn-light"
            onclick="hpEditor.Open('{[=it.name]}', '{[=v.name]}')"
          >
            {[=v.value]}
          </button>
          {[~]}
        </span>

        <span class="vline"></span>

        <span id="field_{[=it.name]}_editor_mdr" class="" style="display:none">
          <button
            class="btn btn-sm btn-light preview_open"
            onclick="hpEditor.PreviewOpen('{[=it.name]}')"
            style="display:none"
          >
            Preview
          </button>
          <button
            class="btn btn-sm btn-light preview_close"
            onclick="hpEditor.PreviewClose('{[=it.name]}')"
            style="display:none"
          >
            Close Preview
          </button>
          <button
            class="btn btn-sm btn-light storage-image-insert"
            onclick="hpEditor.StorageImageSelector('{[=it.name]}')"
          >
            Image
          </button>
        </span>
      </div>

      <div id="field_{[=it.name]}_layout" class="editor-fra">
        <div id="field_{[=it.name]}_editor" class="editor-fra-item">
          <textarea
            class="form-control"
            id="field_{[=it.name]}"
            name="field_{[=it.name]}"
            rows="{[if (it.attr_ui_rows) {]}{[=it.attr_ui_rows]}{[} else {]}6{[}]}"
          >
{[=it.value]}</textarea
          >
        </div>
        <div
          id="field_{[=it.name]}_colpreview"
          style="display:none"
          classs="editor-fra-item"
        >
          <div
            class="hp-content hp-scroll"
            id="field_{[=it.name]}_preview"
            style="padding:10px"
          ></div>
        </div>
      </div>
    </div>
  </div>
</script>

<script id="hpm-nodeset-tplint" type="text/html">
  <div class="hpm-nodeset-tplx">
    <label class="form-label">{[=it.title]}</label>
    <input
      type="text"
      name="field_{[=it.name]}"
      class="form-control"
      value="{[=it.value]}"
    />
  </div>
</script>

<script id="hpm-nodeset-tplstring" type="text/html">
  <div class="hpm-nodeset-tplx hpm-nodeset-tplstring">
    <label class="form-label">
      <span>{[=it.title]}</span>
      {[? it.attr_lang_list]}
      <select
        id="field_{[=it.name]}_langs"
        class="field-nav-lang form-select"
        onchange="hpNode.SetFieldLang('{[=it.name]}')"
      >
        {[~it.attr_lang_list :v]}
        <option value="{[=v.id]}">{[=v.name]}</option>
        {[~]}
      </select>
      {[?]}
    </label>
    <input
      type="text"
      id="field_{[=it.name]}"
      name="field_{[=it.name]}"
      class="form-control"
      value="{[=it.value]}"
    />
  </div>
</script>

<script id="hpm-nodeset-tplterm_tag" type="text/html">
  <div class="hpm-nodeset-tplx">
    <label class="form-label">{[=it.title]}</label>
    <input
      type="text"
      name="term_{[=it.meta.name]}"
      class="form-control"
      value="{[=it.value]}"
    />
  </div>
</script>

<script id="hpm-nodeset-tplterm_taxonomy" type="text/html">
  <div class="hpm-nodeset-tplx">
    <label class="form-label">{[=it.model.title]}</label>
    <select class="form-select" name="term_{[=it.model.meta.name]}">
      {[~it.items :v]} {[ if (v.pid == 0) { ]}
      <option
        value="{[=v.id]}"
        {[?
        hpMgr.Equal(it.item.value,
        v.id)]}
        selected{[?]}
      >
        {[=v.title]}
      </option>
      {[? v._subs]} {[~v._subs :v2]}
      <option
        value="{[=v2.id]}"
        {[?
        hpMgr.Equal(it.item.value,
        v2.id)]}
        selected{[?]}
      >
        {[=hpTerm.Sprint(v2._dp)]}{[=v2.title]}
      </option>
      {[~]} {[}]} {[ } ]} {[~]}
    </select>
  </div>
</script>

<script id="hpm-nodeset-tplext_comment_perentry" type="text/html">
  <div class="hpm-nodeset-tplx">
    <label class="form-label">Comment On/Off</label>
    <select class="form-select" name="ext_comment_perentry">
      {[~it._general_onoff :gv]}
      <option
        value="{[=gv.type]}"
        {[?
        hpMgr.Equal(it.ext_comment_perentry,
        gv.type)]}
        selected{[?]}
      >
        {[=gv.name]}
      </option>
      {[~]}
    </select>
  </div>
</script>

<script id="hpm-nodeset-tplext_permalink" type="text/html">
  <div class="hpm-nodeset-tplx">
    <label class="form-label">Permalink Name</label>
    <input
      type="text"
      name="ext_permalink_name"
      class="form-control"
      value="{[=it.ext_permalink_name]}"
    />
  </div>
</script>

<script id="hpm-nodeset-tplext_node_refer" type="text/html">
  <div class="hpm-nodeset-tplx">
    <label class="form-label">Refer ID</label>
    <input
      type="text"
      name="ext_node_refer"
      class="form-control"
      value="{[=it.ext_node_refer]}"
    />
  </div>
</script>
