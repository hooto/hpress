<div class="hpm-block-gap-column">
  <div id="hpm-termset" class="hpm-table-std">loading</div>

  <div class="hpm-block-gap-row-sm">
    <button class="btn btn-primary" onclick="hpTerm.SetCommit()">Save</button>
    <button class="btn btn-outline-primary" onclick="hpTerm.List()">
      Cancel
    </button>
  </div>
</div>

<script id="hpm-termset-tpl" type="text/html">
  <input type="hidden" name="model_type" value="{[=it.model.type]}" />
  <input type="hidden" name="id" value="{[=it.id]}" />
  <input type="hidden" name="status" value="{[=it.status]}" />

  <div class="mb-3">
    <label class="form-label">Title</label>
    <input
      name="title"
      type="text"
      value="{[=it.title]}"
      class="form-control"
    />
  </div>

  {[? hpMgr.Equal(it.model.type,"taxonomy")]}
  <div class="mb-3">
    <label class="form-label">Relations</label>
    <select name="pid" class="form-select">
      <option value="0" {[? hpMgr.Equal(it.pid, 0)]} selected{[?]}>ROOT</option>
      {[~it._taxonomy_ls.items :v]}
      <!-- -->
      {[? ((hpMgr.Equal(v.pid,0) && hpMgr.NotEqual(v.id,it.id)))]}
      <option value="{[=v.id]}" {[? hpMgr.Equal(it.pid,v.id)]} selected{[?]}>
        {[=v.title]}
      </option>
      <!-- -->
      {[? v._subs]}
      <!-- -->
      {[~v._subs :v2]}
      <!-- -->
      {[? hpMgr.NotEqual(v2.id,it.id)]}
      <option value="{[=v2.id]}" {[? hpMgr.Equal(it.pid,v2.id)]} selected{[?]}>
        {[=hpTerm.Sprint(v2._dp)]}{[=v2.title]}
      </option>
      <!-- -->
      {[?]}
      <!-- -->
      {[~]}
      <!-- -->
      {[?]}
      <!-- -->
      {[?]}
      <!-- -->
      {[~]}
    </select>
  </div>
  {[?]}
  <!-- -->
  {[? hpMgr.Equal(it.model.type,"taxonomy")]}
  <div class="mb-3">
    <label class="form-label">Weight</label>
    <input
      name="weight"
      type="text"
      value="{[=it.weight]}"
      class="form-control"
    />
  </div>
  {[?]}
</script>
