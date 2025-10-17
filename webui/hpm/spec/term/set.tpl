<div id="hpm-spec-termset-alert"></div>

<div id="hpm-spec-termset">

  <input type="hidden" name="modname" value="{[=it._modname]}">
  {[? it.meta.name]}
  <input type="hidden" name="name" value="{[=it.meta.name]}">
  {[??]}
  <div class="form-group">
    <label>Name</label>
    <input type="text" class="form-control" name="name" 
      placeholder="Term Name" value="{[=it.meta.name]}">
  </div>
  {[?]}

  <div class="form-group">
    <label>Title</label>
    <input type="text" class="form-control" name="title" 
      placeholder="Title" value="{[=it.title]}">
  </div>

  <div class="form-group">
    <label>Type</label>
    <select class="form-control" name="type">
      <option value="taxonomy" {[? hpMgr.Equal(it.type,'taxonomy')]} selected{[?]}>Categories</option>
      <option value="tag" {[? hpMgr.Equal(it.type,'tag')]} selected{[?]}>Tags</option>
    </select>
  </div>
</div>
