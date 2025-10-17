<table
  id="lcbind-layout"
  border="0"
  cellpadding="0"
  cellspacing="0"
  class=""
  width="100%"
>
  <tr>
    <td id="lcbind-proj-filenav" class="lcx-lay-colbg" width="220px"></td>

    <td class="lclay-colsep lclay-col-resize" lc-layid="lclay-colmain"></td>
    <td id="lclay-colmain" class="lcx-lay-colbg"></td>
  </tr>
</table>

<div id="lctab-tpl" style="display: none">
  <table
    id="lctab-box{[=it.tabid]}"
    class="lctab-box"
    width="100%"
    height="100%"
  >
    <tr>
      <td class="" valign="top">
        <div id="lctab-nav{[=it.tabid]}" class="lctab-nav">
          <div class="lctab-navm">
            <div id="lctab-navtabs{[=it.tabid]}" class="lctab-navs"></div>
          </div>
          <div class="lctab-navr">
            <div class="lcpg-tab-more" href="#{[=it.tabid]}"></div>
          </div>
        </div>

        <div id="lctab-bar{[=it.tabid]}" class="lctab-bar"></div>
        <div id="lctab-body{[=it.tabid]}" class="lctab-body less_scroll"></div>
      </td>
    </tr>
  </table>
</div>

<div id="lctab-openfiles-ol" class="less_scroll"></div>

<script id="hp-speceditor-fsnav-tpl" type="text/html">
  <div class="lcx-fsnav">
    <span class="lfn-title">Files</span>

    <ul class="lfn-menus">
      <li class="lfnm-item d-flex flex-items-center justify-content-center">
        <div>
          <svg class="bi" width="16" height="16" fill="currentColor">
            <use
              xlink:href="/hp/lynkui/~/bi/v1/bootstrap-icons.svg#cloud-arrow-up"
            />
          </svg>
        </div>

        <ul class="lfnm-item-submenu">
          <li>
            <a
              href="#proj/fs/file-new"
              onclick="l9rProjFs.FileNew('file', null, '')"
            >
              <img src="~/hpm/img/page_white_add.png" class="h5c_icon" />
              New File
            </a>
          </li>
          <li>
            <a href="#proj/fs/file-upl" onclick="l9rProjFs.FileUpload(null)">
              <img src="~/hpm/img/page_white_get.png" class="h5c_icon" />
              Upload
            </a>
          </li>
        </ul>
      </li>

      <li
        class="lfnm-item d-flex flex-items-center justify-content-center"
        onclick="l9rProjFs.RootRefresh()"
      >
        <div>
          <svg class="bi" width="16" height="16" fill="currentColor">
            <use
              xlink:href="/hp/lynkui/~/bi/v1/bootstrap-icons.svg#arrow-repeat"
            />
          </svg>
        </div>

        <!-- <a
          href="#fs/refresh"
          class="icon-refresh icon-white lfnm-item-ico"
          title="Refresh"
        >
		</a>-->
      </li>
    </ul>
  </div>

  <!-- Project Files Tree -->
  <div id="lcbind-fsnav-fstree" class="less_scroll">
    <div id="fstdroot" class="lcx-fstree">loading</div>
  </div>
</script>

<!--- TPL: File Item -->
<script id="lcx-filenav-tree-tpl" type="text/html">
  {[~it :v]}
  <div
    id="ptp{[=v.fsid]}"
    class="lcx-fsitem"
    lc-fspath="{[=v.path]}"
    lc-fstype="{[=v.fstype]}"
    lc-fsico="{[=v.ico]}"
  >
    <img
      src="{[=hpMgr.frtbase]}~/hpm/img/mtypes/{[=v.ico]}.png"
      align="absmiddle"
    />
    <a href="#" class="anoline">{[=v.name]}</a>
  </div>
  {[~]}
</script>

<!--- TPL: File Right Click Menu -->
<div id="lcbind-fsnav-rcm" style="display: none">
  <div class="lcbind-fsrcm-item fsrcm-isdir" lc-fsnav="new-file">
    <div class="rcico">
      <img src="~/hpm/img/page_white_add.png" align="absmiddle" />
    </div>
    <a href="#" class="rcctn">New File</a>
  </div>
  <div class="lcbind-fsrcm-item fsrcm-isdir" lc-fsnav="upload">
    <div class="rcico">
      <img src="~/hpm/img/page_white_get.png" align="absmiddle" />
    </div>
    <a href="#" class="rcctn">Upload</a>
  </div>

  <div class="rcm-sepline fsrcm-isdir"></div>

  <div class="lcbind-fsrcm-item" lc-fsnav="rename">
    <div class="rcico">
      <img src="~/hpm/img/page_white_copy.png" align="absmiddle" />
    </div>
    <a href="#" class="rcctn">Rename</a>
  </div>
  <div class="lcbind-fsrcm-item" lc-fsnav="file-del">
    <div class="rcico">
      <img src="~/hpm/img/delete.png" align="absmiddle" />
    </div>
    <a href="#" class="rcctn">Delete</a>
  </div>
</div>

<!-- TPL : File New -->
<script id="lcbind-fstpl-filenew" type="text/html">
  <form
    id="{[=it.formid]}"
    action="#"
    onsubmit="l9rProjFs.FileNewSave('{[=it.formid]}');return false;"
  >
    <div class="input-group" style="padding:16px">
      <span class="input-group-text">
        <img
          src="{[=hpMgr.frtbase]}~/hpm/img/folder_add.png"
          class="h5c_icon"
        />
        <span>{[=it.path]}</span>
      </span>
      <input
        type="text"
        name="name"
        value="{[=it.file]}"
        class="form-control"
      />
    </div>
    <input type="hidden" name="path" value="{[=it.path]}" />
    <input type="hidden" name="type" value="{[=it.type]}" />
  </form>
</script>

<!-- TPL : File Rename -->
<script id="lcbind-fstpl-filerename" type="text/html">
  <form
    id="{[=it.formid]}"
    action="#"
    onsubmit="l9rProjFs.FileRenameSave('{[=it.formid]}');return false;"
  >
    <div class="input-group" style="padding:16px">
      <span class="input-group-text">
        <img
          src="{[=hpMgr.frtbase]}~/hpm/img/folder_edit.png"
          class="h5c_icon"
        />
      </span>
      <input
        type="text"
        name="pathset"
        value="{[=it.path]}"
        class="form-control"
      />
    </div>
    <input type="hidden" name="path" value="{[=it.path]}" />
  </form>
</script>

<!-- TPL : File Delete -->
<script id="lcbind-fstpl-filedel" type="text/html">
  <form
    id="{[=it.formid]}"
    action="#"
    onsubmit="l9rProjFs.FileDelSave('{[=it.formid]}');return false;"
    style="padding:16px"
  >
    <input type="hidden" name="path" value="{[=it.path]}" />
    <div class="alert alert-danger" role="alert">
      <p>Are you sure to delete this file or folder?</p>
      <p><strong>{[=it.path]}</strong></p>
      <p></p>
    </div>
  </form>
</script>

<!-- TPL : File Upload -->
<script id="lcbind-fstpl-fileupload" type="text/html">
  <style type="text/css">
    .lsarea {
      margin: 10px 0;
      display: inline-block;
      height: 100px;
      width: 100%;
      color: #333;
      font-size: 18px;
      padding: 10px;
      border: 3px dashed #5cb85c;
      border-radius: 10px;
      text-align: center;
      vertical-align: middle;
      -webkit-box-sizing: border-box;
      -moz-box-sizing: border-box;
      box-sizing: border-box;
    }

    .lsstate {
      margin-top: 10px;
    }
  </style>
  <div id="{[=it.reqid]}" style="padding:16px">
    <div>The target of Upload directory</div>
    <div class="input-group">
      <span class="input-group-text"
        ><img
          src="{[=hpMgr.frtbase]}~/hpm/img/page_white_get.png"
          align="absmiddle"
      /></span>
      <input
        class="form-control"
        name="path"
        type="text"
        value="{[=it.path]}"
      />
      <button
        class="btn btn-outline-secondary"
        type="button"
        onclick="_fs_upl_chgdir()"
      >
        Change directory
      </button>
    </div>
    <div id="{[=it.areaid]}" class="lsarea">
      Drag and Drop your files or folders to here
    </div>
    <div class="alert alert-info lsstate" style="display:none"></div>
  </div>
</script>
