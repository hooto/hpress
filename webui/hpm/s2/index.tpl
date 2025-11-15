<div class="hpm-block-gap-column">
  <div
    id="hpm-s2-objls-navbar"
    class="d-flex flex-row justify-content-between hpm-block-gap-row"
  >
    <div
      id="hpm-s2-objls-dirnav"
      class="align-self-center hpm-breadcrumb"
    ></div>
    <div id="hpm-s2-objls-optools" class="hpm-node-nav hpm-nav-right">
      <div class="btn btn-primary" onclick="hpS2.ObjNew('file')">
        Upload New File
      </div>
    </div>
  </div>

  <div id="" class="hpm-table-std">
    <table class="table table-hover align-middle">
      <thead>
        <tr>
          <th width="64px"></th>
          <th>Name</th>
          <th style="text-align: right">Size</th>
          <th></th>
          <th></th>
        </tr>
      </thead>
      <tbody id="hpm-s2-objls"></tbody>
    </table>
  </div>
</div>

<script id="hpm-s2-objls-dirnav-tpl" type="text/html">
  {[~it.items :v]}
  <li>
    <a href="#{[=v.path]}" onclick="hpS2.ObjList('{[=v.path]}')">{[=v.name]}</a>
  </li>
  {[~]}
</script>

<script id="hpm-s2-objls-tpl" type="text/html">
  {[~it.items :v]}
  <tr id="obj{[=v._id]}">
    <td>
      {[ if (v.isdir) { ]}
      <svg class="bi" width="16" height="16" fill="currentColor">
        <use xlink:href="/hp/lynkui/~/bi/v1/bootstrap-icons.svg#folder" />
      </svg>
      {[ } else if (v._isimg) { ]}
      <a href="{[=v.self_link]}" target="_blank"
        ><img src="{[=v.self_link]}?ipl=w64,h64,c" width="64" height="64"
      /></a>
      {[ } ]}
    </td>
    <td class="ts3-fontmono">
      {[ if (v.isdir) { ]}
      <a class="obj-item-dir" href="#objs" path="{[=v._abspath]}"
        >{[=v.name]}</a
      >
      {[ } else { ]}
      <a class="obj-item-file" href="{[=v.self_link]}" target="_blank"
        >{[=v.name]}</a
      >
      {[ } ]}
    </td>
    <td align="right">
      {[?!v.isdir]} {[=hpS2.UtilResourceSizeFormat(v.size)]}
    </td>
    {[?]}
    <td align="right">
      {[=lynkui.utilx.timeParseFormat(v.modtime, "Y-m-d H:i:s")]}
    </td>
    <td align="right">
      {[ if (!v.isdir) { ]}
      <a
        class="obj-item-del btn btn-outline-dark btn-sm"
        href="#obj-del"
        obj="{[=v._abspath]}"
      >
        Delete
      </a>
      {[ } ]}
    </td>
  </tr>
  {[~]}
</script>

<!-- TPL : File New -->
<style type="text/css">
  ._hpm_s2_fsupload_area {
    margin: 0;
    display: inline-block;
    width: 100%;
    color: #333;
    font-size: 18px;
    padding: 20px;
    border: 3px dashed rgb(0, 120, 231);
    border-radius: 10px;
    text-align: center;
    vertical-align: middle;
    -webkit-box-sizing: border-box;
    -moz-box-sizing: border-box;
    box-sizing: border-box;
  }
</style>

<script id="hpm-s2-objnew-tpl" type="text/html">
  <form
    id="{[=it.formid]}"
    action="#"
    onsubmit="hpS2.ObjNewSave('{[=it.formid]}');return false;"
  >
    <input type="hidden" name="type" value="{[=it.type]}" />
    <div class="mb-3">
      <label class="form-label">The target upload directory</label>
      <input
        type="text"
        name="path"
        class="form-control"
        placeholder="Folder Path"
        value="{[=it.path]}"
      />
    </div>
    <div class="mb-3">
      <label class="form-label">Select a single file to upload</label>
      <input
        id="hpm-s2-objnew-files"
        type="file"
        name="file"
        class="form-control"
        placeholder="File Path"
        value=""
      />
    </div>
    <div class="mb-3">
      <label class="form-label">Select multifile to upload</label>
      <div id="hpm-s2-fsupload-area" class="_hpm_s2_fsupload_area">
        Drag and Drop your files or folders to here
      </div>
    </div>
  </form>
  <div
    id="{[=it.formid]}-alert"
    class="alert alert-success"
    style="display:none"
  ></div>
</script>

<!-- TPL : File Rename -->
<script id="hpm-s2-objrename-tpl" type="text/html">
  <form
    id="{[=it.formid]}"
    action="#"
    onsubmit="hpS2.ObjRenameSave('{[=it.formid]}');return false;"
  >
    <div class="input-prepend" style="margin-left:2px">
      <span class="add-on">
        <img src="{[=hpMgr.base]}-/img/folder_edit.png" class="h5c_icon" />
      </span>
      <input
        type="text"
        name="pathset"
        value="{[=it.path]}"
        style="width:500px;"
      />
      <input type="hidden" name="path" value="{[=it.path]}" />
    </div>
  </form>
</script>

<script type="text/javascript">
  $("#hpm-s2-objls").on("click", ".obj-item-dir", function () {
    hpS2.ObjList($(this).attr("path"));
  });
  $("#hpm-s2-objls").on("click", ".obj-item-del", function () {
    var r = confirm("This file will be deleted, Confirm?");
    if (r == true) {
      hpS2.ObjDel($(this).attr("obj"));
    }
  });
  $("#hpS2-object-dirnav").on("click", ".obj-item-dir", function () {
    hpS2.ObjList($(this).attr("path"));
  });
</script>
