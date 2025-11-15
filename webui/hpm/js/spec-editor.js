// Copyright 2015 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

var hpSpecEditor = {};

hpSpecEditor.Open = function (modname) {
  var topnav = $("#hp-uh-topnav");
  topnav.find("a.active").removeClass("active");

  if (topnav.find("a[modname='" + modname + "']").length > 0) {
    topnav.find("a[modname='" + modname + "']").addClass("active");
    hpSpecEditor.Index(modname);
    return;
  }

  lynkui.url.eventRegister("spec-editor/" + modname, function () {
    hpSpecEditor.Index(modname);
  });

  $("#hp-uh-topnav").append(
    '<a class="lynkui-nav-item active" modname="' +
      modname +
      '" href="#spec-editor/' +
      modname +
      '">Spec Editor (' +
      modname +
      ")</a>"
  );

  lcData.Init("speceditor", function () {
    hpSpecEditor.Index(modname);
  });
};

hpSpecEditor.Index = function (modname) {
  if (!modname) {
    return;
  }

  l9rTab.pool = {};

  lynkui.session.set("hp-speceditor-modname", modname);
  lynkui.session.set("modname_current", "/");

  hpMgr.TplCmd("spec/editor/desk", {
    callback: function (err, data) {
      if (err) {
        return alert(err);
      }

      $("#com-content").html(data);

      lynkui.template.renderFromId(
        "lcbind-proj-filenav",
        "hp-speceditor-fsnav-tpl"
      );

      lynkui.use(
        [
          "~/cm/5/lib/codemirror.css",
          "~/cm/5/lib/codemirror.js",
          "~/cm/5/theme/monokai.css",
        ],
        function () {
          lynkui.use(
            [
              "~/cm/5/mode/clike/clike.js",
              "~/cm/5/mode/javascript/javascript.js",
              "~/cm/5/mode/css/css.js",
              "~/cm/5/mode/htmlmixed/htmlmixed.js",
              "~/cm/5/mode/markdown/markdown.js",
              "~/cm/5/mode/xml/xml.js",
              "~/cm/5/addon/selection/active-line.js",
              "~/cm/5/addon/hint/show-hint.js",
              "~/cm/5/addon/hint/javascript-hint.js",
              "~/cm/5/addon/selection/active-line.js",
              "~/cm/5/addon/fold/foldcode.js",
              "~/cm/5/addon/fold/foldgutter.js",
              "~/cm/5/addon/edit/closetag.js",
              "~/cm/5/addon/edit/closebrackets.js",
            ],
            function () {
              //
              l9rProjFs.UiTreeLoad({
                path: "/",
              });

              l9rProjFs.OpenHistoryTabs();
            }
          );
        }
      );
    },
  });
};

var l9rProjFs = {
  //
};

l9rProjFs.OpenHistoryTabs = function () {
  // console.log("l9rProj.OpenHistoryTabs");

  // var last_tab_urid = lynkui.storage.set(lynkui.session.get("podid") +"."+ lynkui.session.get("proj_name") +".tab."+ item.target);

  lcData.Query(
    "files",
    "projdir",
    lynkui.session.get("hp-speceditor-modname"),
    function (ret) {
      // console.log("Query files");
      if (!ret) {
        return;
      }

      if (
        ret.value.id &&
        ret.value.projdir == lynkui.session.get("hp-speceditor-modname")
      ) {
        var icon = undefined;
        if (ret.value.icon) {
          icon = ret.value.icon;
        }

        var cab = l9rTab.frame[ret.value.cabid];
        if (!cab) {
          ret.value.cabid = l9rTab.def;
          cab = l9rTab.frame[l9rTab.def];
        }

        var tabLastActive = lynkui.storage.get(
          lynkui.session.get("hp-speceditor-modname") +
            ".cab." +
            ret.value.cabid
        );

        var titleOnly = true;
        if (!cab.actived || tabLastActive == ret.value.id) {
          l9rTab.frame[ret.value.cabid].actived = true;
          titleOnly = false;
        }

        l9rTab.Open({
          uri: ret.value.filepth,
          type: "editor",
          icon: icon,
          close: true,
          titleOnly: titleOnly,
          refreshActiveStorageOff: true,
          success: function () {
            // $('#pgtab'+ ret.value.id).addClass("current");
          },
        });

        if (
          ret.value.ctn1_sum.length > 10 &&
          ret.value.ctn1_sum != ret.value.ctn0_sum
        ) {
          $("#pgtab" + ret.value.id + " .chg").show();
          $("#pgtab" + ret.value.id + " .pgtabtitle").addClass("chglight");
        }
      }

      ret.continue();
    }
  );
};

l9rProjFs.RootRefresh = function () {
  l9rProjFs.UiTreeLoad({
    path: "/",
  });
};

l9rProjFs.UiTreeLoad = function (options) {
  // console.log("l9rProjFs.UiTreeLoad"+ options.path);

  options = options || {};

  if (typeof options.success !== "function") {
    options.success = function () {};
  }

  if (typeof options.error !== "function") {
    options.error = function () {};
  }

  if (options.toggle === undefined) {
    options.toggle = false;
  }

  var ptdid = lynkui.utilx.cryptoMd5(options.path);
  if (options.path == lynkui.session.get("modname_current")) {
    ptdid = "root";
  }

  if (
    ptdid != "root" &&
    options.toggle &&
    document.getElementById("fstd" + ptdid)
  ) {
    $("#fstd" + ptdid).remove();
    return;
  }

  var req = {
    path: options.path, // lynkui.session.get("modname_current"),
  };

  // console.log("path reload "+ options.path);

  req.success = function (rs) {
    var ls = rs.items;
    var lsfs = [];

    for (var i in ls) {
      if (ls[i].name == "spec.json") {
        // TODO
        // continue;
      }

      var fspath = rs.path + "/" + ls[i].name;
      ls[i].path = fspath.replace(/\/+/g, "/");
      ls[i].fsid = lynkui.utilx.cryptoMd5(ls[i].path);

      ls[i].fstype = "none";

      var ico = "page_white";

      if (ls[i].isdir !== undefined && ls[i].isdir == true) {
        ico = "folder";
        ls[i].fstype = "dir";
      } else if (
        ls[i].mime.substring(0, 4) == "text" ||
        ls[i].name.slice(-4) == ".tpl" ||
        ls[i].mime.substring(0, 23) == "application/x-httpd-php" ||
        ls[i].mime == "application/javascript" ||
        ls[i].mime == "application/x-empty" ||
        ls[i].mime == "inode/x-empty" ||
        ls[i].mime == "application/json"
      ) {
        if (ls[i].mime == "text/x-php" || ls[i].name.slice(-4) == ".php") {
          ico = "page_white_php";
        } else if (
          ls[i].name.slice(-2) == ".h" ||
          ls[i].name.slice(-4) == ".hpp"
        ) {
          ico = "page_white_h";
        } else if (ls[i].name.slice(-2) == ".c") {
          ico = "page_white_c";
        } else if (
          ls[i].name.slice(-4) == ".cpp" ||
          ls[i].name.slice(-3) == ".cc"
        ) {
          ico = "page_white_cplusplus";
        } else if (
          ls[i].name.slice(-3) == ".js" ||
          ls[i].name.slice(-4) == ".css"
        ) {
          ico = "page_white_code";
        } else if (
          ls[i].name.slice(-5) == ".html" ||
          ls[i].name.slice(-4) == ".htm" ||
          ls[i].name.slice(-6) == ".phtml" ||
          ls[i].name.slice(-6) == ".xhtml" ||
          ls[i].name.slice(-4) == ".tpl"
        ) {
          ico = "page_white_world";
        } else if (
          ls[i].name.slice(-3) == ".sh" ||
          ls[i].mime == "text/x-shellscript"
        ) {
          ico = "application_osx_terminal";
        } else if (ls[i].name.slice(-3) == ".rb") {
          ico = "page_white_ruby";
        } else if (ls[i].name.slice(-3) == ".go") {
          ico = "ht-page_white_golang";
        } else if (ls[i].name.slice(-5) == ".java") {
          ico = "page_white_cup";
        } else if (
          ls[i].name.slice(-3) == ".py" ||
          ls[i].name.slice(-4) == ".yml" ||
          ls[i].name.slice(-5) == ".yaml" ||
          ls[i].name.slice(-3) == ".md"
        ) {
          ico = "page_white_code";
        }

        // ls[i].href = "javascript:h5cTabOpen('{$p}','w0','editor',{'img':'{$fmi}', 'close':'1'})";

        ls[i].fstype = "text";
      } else if (
        ls[i].mime.slice(-5) == "image" ||
        ls[i].name.slice(-4) == ".jpg" ||
        ls[i].name.slice(-4) == ".png" ||
        ls[i].name.slice(-4) == ".gif" ||
        ls[i].name.slice(-4) == ".svg"
      ) {
        ico = "page_white_picture";
      }

      ls[i].ico = ico;

      lsfs.push(ls[i]);
    }

    if (document.getElementById("fstd" + ptdid) == null) {
      $("#ptp" + ptdid).after(
        '<div id="fstd' + ptdid + '" style="padding-left:20px;"></div>'
      );
    }

    lynkui.template.renderFromId("fstd" + ptdid, "lcx-filenav-tree-tpl", lsfs);

    options.success();

    setTimeout(function () {
      l9rProjFs.UiTreeEventRefresh();
      l9rLayout.Resize();
    }, 10);
  };

  req.error = function (status, message) {
    // console.log(status, message);
    options.error(status, message);
  };

  l9rPodFs.List(req);
};

var _fsItemPath = "";
l9rProjFs.UiTreeEventRefresh = function () {
  // console.log("l9rProjFs.UiTreeEventRefresh");
  $(".lcx-fsitem").unbind();
  $(".lcx-fsitem").bind("contextmenu", function (e) {
    var h = $("#lcbind-fsnav-rcm").height();
    // h = $(this).find(".hdev-rcmenu").height();
    var t = e.pageY;
    var bh = $("body").height() - 20;
    if (t + h > bh) {
      t = bh - h;
    }

    var bw = $("body").width() - 20;
    var l = e.pageX;
    if (l > bw - 200) {
      l = bw - 200;
    }
    // console.log("pos"+ t +"x"+ l);
    $("#lcbind-fsnav-rcm")
      .css({
        top: t + "px",
        left: l + "px",
      })
      .show(10);

    _fsItemPath = $(this).attr("lc-fspath");

    var fstype = $(this).attr("lc-fstype");
    if (fstype == "dir") {
      $(".fsrcm-isdir").show();
    } else {
      $(".fsrcm-isdir").hide();
    }

    return false;
  });
  $(".lcx-fsitem").bind("click", function () {
    var fstype = $(this).attr("lc-fstype");
    var fspath = $(this).attr("lc-fspath");
    var fsicon = $(this).attr("lc-fsico");

    switch (fstype) {
      case "dir":
        l9rProjFs.UiTreeLoad({
          path: fspath,
          toggle: true,
        });
        break;
      case "text":
        l9rTab.Open({
          uri: fspath,
          // colid : "lclay-colmain",
          type: "editor",
          icon: fsicon,
          close: true,
        });
        break;
      default:
      //
    }
  });

  //
  $(".lcbind-fsrcm-item").unbind();
  $(".lcbind-fsrcm-item").bind("click", function () {
    var action = $(this).attr("lc-fsnav");

    // var ppath = path.slice(0, path.lastIndexOf("/"));
    // var fname = path.slice(path.lastIndexOf("/") + 1);
    // console.log("right click: "+ action);
    switch (action) {
      case "new-file":
        l9rProjFs.FileNew("file", _fsItemPath, "");
        break;
      case "new-dir":
        l9rProjFs.FileNew("dir", _fsItemPath, "");
        break;
      case "upload":
        l9rProjFs.FileUpload(_fsItemPath);
        break;
      case "rename":
        l9rProjFs.FileRename(_fsItemPath);
        break;
      case "file-del":
        l9rProjFs.FileDel(_fsItemPath);
        break;
      default:
        break;
    }

    $("#lcbind-fsnav-rcm").hide();
  });

  $(document).click(function () {
    $("#lcbind-fsnav-rcm").hide();
  });
};

l9rProjFs.FileNew = function (type, path, file) {
  if (path === undefined || path === null) {
    path = lynkui.session.get("modname_current");
  }

  var formid = Math.random().toString(36).slice(2);

  var req = {
    title: type == "dir" ? "New Folder" : "New File",
    position: "cursor",
    width: 600,
    height: 200,
    tplid: "lcbind-fstpl-filenew",
    data: {
      formid: formid,
      file: file,
      path: path,
      type: type,
    },
    buttons: [
      {
        onclick: 'l9rProjFs.FileNewSave("' + formid + '")',
        title: "Create",
        style: "btn-primary",
      },
      {
        onclick: "lynkui.modal.close()",
        title: "Close",
      },
    ],
  };

  req.success = function () {
    $("#" + formid + " :input[name=name]").focus();
  };

  lynkui.modal.open(req);
};

var l9r = {};

l9r.HeaderAlert = function (type, message) {
  alert(message);
};

l9rProjFs.FileNewSave = function (formid) {
  var path = $("#" + formid + " :input[name=path]").val();
  var name = $("#" + formid + " :input[name=name]").val();
  if (name === undefined || name.length < 1) {
    alert("Filename can not be null"); // TODO
    return;
  }

  l9rPodFs.Post({
    path: path + "/" + name,
    data: "\n",
    success: function (rsp) {
      // hdev_header_alert('success', "{{T . "Successfully Done"}}");

      // if (typeof _plugin_yaf_cvlist == 'function') {
      //     _plugin_yaf_cvlist();
      // }

      l9rProjFs.UiTreeLoad({
        path: path,
      });

      lynkui.modal.close();
    },
    error: function (status, message) {
      // console.log(status, message);
      // hdev_header_alert('error', textStatus+' '+xhr.responseText);
    },
  });

  return false;
};

// html5 file uploader
var _fsUploadRequestId = "";
var _fsUploadAreaId = "";
var _fsUploadBind = null;

function _fsUploadTraverseTree(reqid, item, path) {
  path = path || "";

  if (item.isFile) {
    // Get file
    item.file(function (file) {
      //console.log("File:", path + file.name);
      if (file.size > 10 * 1024 * 1024) {
        $("#" + reqid + " .state")
          .show()
          .append(
            "<div>" + path + " Failed: File is too large to upload</div>"
          );
        return;
      }

      _fsUploadCommit(reqid, file);
    });
  } else if (item.isDirectory) {
    // Get folder contents
    var dirReader = item.createReader();
    dirReader.readEntries(function (entries) {
      for (var i = 0; i < entries.length; i++) {
        _fsUploadTraverseTree(reqid, entries[i], path + item.name + "/");
      }
    });
  }
}

function _fsUploadHanderDragEnter(evt) {
  this.setAttribute("style", "border-style:dashed;");
}

function _fsUploadHanderDragLeave(evt) {
  this.setAttribute("style", "");
}

function _fsUploadHanderDragOver(evt) {
  evt.stopPropagation();
  evt.preventDefault();
}

function _fsUploadCommit(reqid, file) {
  var reader = new FileReader();

  reader.onload = (function (file) {
    return function (e) {
      if (e.target.readyState != FileReader.DONE) {
        return;
      }

      var ppath = $("#" + reqid + " :input[name=path]").val();
      // console.log("upload path: "+ ppath);

      l9rPodFs.Post({
        path: ppath + "/" + file.name,
        size: file.size,
        data: e.target.result,
        encode: "base64",
        success: function (rsp) {
          $("#" + reqid)
            .find(".alert")
            .show()
            .append("<div>" + file.name + " OK</div>");

          // console.log(rsp);
          // hdev_header_alert('success', "{{T . "Successfully Done"}}");

          // if (typeof _plugin_yaf_cvlist == 'function') {
          //     _plugin_yaf_cvlist();
          // }

          // l9rProjFs.UiTreeLoad({path: ppath});
          // lynkui.modal.close();
        },
        error: function (status, message) {
          // console.log(message);

          $("#" + reqid)
            .find(".alert")
            .show()
            .append("<div>" + file.name + " Failed</div>");
          // console.log(status, message);
          // hdev_header_alert('error', textStatus+' '+xhr.responseText);
        },
      });
    };
  })(file);

  reader.readAsDataURL(file);
}

function _fsUploadHander(evt) {
  evt.stopPropagation();
  evt.preventDefault();

  var items = evt.dataTransfer.items;
  for (var i = 0; i < items.length; i++) {
    // webkitGetAsEntry is where the magic happens
    var item = items[i].webkitGetAsEntry();
    if (item) {
      _fsUploadTraverseTree(_fsUploadRequestId, item);
    }
  }
}

l9rProjFs.FileUpload = function (path) {
  if (path === undefined || path === null) {
    path = lynkui.session.get("modname_current");
    // alert("Path can not be null"); // TODO
    // return;
  }

  // Check for the various File API support.
  if (window.File && window.FileReader && window.FileList && window.Blob) {
    // Great success! All the File APIs are supported.
  } else {
    alert("The File APIs are not fully supported in this browser");
    return;
  }

  var reqid = Math.random().toString(36).slice(2);
  var areaid = Math.random().toString(36).slice(2);

  // console.log("ids 1: "+ reqid +", "+ areaid);

  var req = {
    title: "Upload File From Location",
    position: "cursor",
    width: 700,
    height: 400,
    tplid: "lcbind-fstpl-fileupload",
    data: {
      areaid: areaid,
      reqid: reqid,
      path: path,
    },
    buttons: [
      // {
      //     onclick : "l9rProjFs.FileUploadSave(\""+ reqid +"\",\""+ areaid +"\")",
      //     title   : "Commit",
      //     style   : "btn-primary"
      // },
      {
        onclick: "lynkui.modal.close()",
        title: "Close",
      },
    ],
  };

  req.success = function () {
    _fsUploadRequestId = reqid;

    // console.log("ids: "+ _fsUploadRequestId +", "+ areaid);

    if (_fsUploadBind != null) {
      _fsUploadBind.removeEventListener(
        "dragenter",
        _fsUploadHanderDragEnter,
        false
      );
      _fsUploadBind.removeEventListener(
        "dragover",
        _fsUploadHanderDragOver,
        false
      );
      _fsUploadBind.removeEventListener("drop", _fsUploadHander, false);
      _fsUploadBind.removeEventListener(
        "dragleave",
        _fsUploadHanderDragLeave,
        false
      );

      _fsUploadBind = null;
    }

    // console.log("id:"+ areaid);

    _fsUploadBind = document.getElementById(areaid);

    // console.log(_fsUploadBind);

    _fsUploadBind.addEventListener(
      "dragenter",
      _fsUploadHanderDragEnter,
      false
    );
    _fsUploadBind.addEventListener("dragover", _fsUploadHanderDragOver, false);
    _fsUploadBind.addEventListener("drop", _fsUploadHander, false);
    _fsUploadBind.addEventListener(
      "dragleave",
      _fsUploadHanderDragLeave,
      false
    );
  };

  lynkui.modal.open(req);
};

l9rProjFs.FileRename = function (path) {
  if (path === undefined || path === null) {
    alert("Path can not be null"); // TODO
    return;
  }

  var formid = Math.random().toString(36).slice(2);

  var req = {
    title: "Rename File/Folder",
    position: "cursor",
    width: 600,
    height: 200,
    tplid: "lcbind-fstpl-filerename",
    data: {
      formid: formid,
      path: path,
    },
    buttons: [
      {
        onclick: 'l9rProjFs.FileRenameSave("' + formid + '")',
        title: "Rename",
        style: "btn-primary",
      },
      {
        onclick: "lynkui.modal.close()",
        title: "Close",
      },
    ],
  };

  req.success = function () {
    $("#" + formid + " :input[name=pathset]").focus();
  };

  lynkui.modal.open(req);
};

l9rProjFs.FileRenameSave = function (formid) {
  var path = $("#" + formid + " :input[name=path]").val();
  var pathset = $("#" + formid + " :input[name=pathset]").val();
  if (pathset === undefined || pathset.length < 1) {
    alert("Path can not be null"); // TODO
    return;
  }

  if (path == pathset) {
    lynkui.modal.close();
    return;
  }

  l9rPodFs.Rename({
    path: path,
    pathset: pathset,
    success: function (rsp) {
      // hdev_header_alert('success', "{{T . "Successfully Done"}}");

      // if (typeof _plugin_yaf_cvlist == 'function') {
      //     _plugin_yaf_cvlist();
      // }
      var ppath = path.slice(0, path.lastIndexOf("/"));
      if (!ppath || ppath == "") {
        ppath = "/";
      }
      // console.log(ppath);

      l9rProjFs.UiTreeLoad({
        path: ppath,
      });
      lynkui.modal.close();
    },
    error: function (status, message) {
      console.log(status, message);
      // hdev_header_alert('error', textStatus+' '+xhr.responseText);
    },
  });
};

l9rProjFs.FileDel = function (path) {
  if (path === undefined || path === null) {
    alert("Path can not be null"); // TODO
    return;
  }

  var formid = Math.random().toString(36).slice(2);

  var req = {
    title: "Delete File or Folder",
    position: "cursor",
    width: 600,
    height: 260,
    tplid: "lcbind-fstpl-filedel",
    data: {
      formid: formid,
      path: path,
    },
    buttons: [
      {
        onclick: 'l9rProjFs.FileDelSave("' + formid + '")',
        title: "Confirm and Delete",
        style: "btn-danger",
      },
      {
        onclick: "lynkui.modal.close()",
        title: "Cancel",
      },
    ],
  };

  lynkui.modal.open(req);
};

l9rProjFs.FileDelSave = function (formid) {
  var path = $("#" + formid + " :input[name=path]").val();
  if (path === undefined || path.length < 1) {
    alert("Path can not be null"); // TODO
    return;
  }

  l9rPodFs.Del({
    path: path,
    success: function (rsp) {
      var fsid = "ptp" + lynkui.utilx.cryptoMd5(path);
      $("#" + fsid).remove();

      lynkui.modal.close();
    },
    error: function (status, message) {
      alert(message);
      // console.log(status, message);
      // hdev_header_alert('error', textStatus+' '+xhr.responseText);
    },
  });
};

var l9rPodFs = {};

l9rPodFs.Get = function (options) {
  // console.log(options);
  // Force options to be an object
  options = options || {};

  if (options.path === undefined) {
    // console.log("undefined");
    return;
  }

  if (typeof options.success !== "function") {
    options.success = function () {};
  }

  if (typeof options.error !== "function") {
    options.error = function () {};
  }

  var url =
    "mod-set-fs/get?modname=" + lynkui.session.get("hp-speceditor-modname");
  // url += "?access_token="+ lynkui.cookie.get("access_token");
  url += "&path=" + options.path;

  // console.log("box refresh:"+ url);
  hpMgr.ApiCmd(url, {
    callback: function (err, data) {
      if (!data) {
        options.error(500, "Networking Error");
      } else if (data.kind == "FsFile") {
        options.success(data);
      } else if (data.error) {
        options.error(data.error.code, data.error.message);
      } else {
        options.error(500, "Networking Error");
      }
    },
    error: function (xhr, textStatus, error) {
      options.error(textStatus, error);
    },
  });
};

l9rPodFs.Post = function (options) {
  options = options || {};

  if (typeof options.success !== "function") {
    options.success = function () {};
  }

  if (typeof options.error !== "function") {
    options.error = function () {};
  }

  if (options.path === undefined) {
    options.error(400, "path can not be null");
    return;
  }

  if (options.data === undefined) {
    options.error(400, "data can not be null");
    return;
  }

  if (options.encode === undefined) {
    options.encode = "text";
  }

  var req = {
    path: options.path,
    body: options.data,
    encode: options.encode,
    sumcheck: options.sumcheck,
  };

  var url =
    "mod-set-fs/put?modname=" + lynkui.session.get("hp-speceditor-modname");

  hpMgr.ApiCmd(url, {
    method: "POST",
    timeout: 30000,
    data: JSON.stringify(req),
    callback: function (err, data) {
      if (!data) {
        options.error(500, "Networking Error");
      } else if (data.kind == "FsFile") {
        options.success(data);
      } else if (data.error) {
        options.error(data.error.code, data.error.message);
      } else {
        options.error(500, "Networking Error");
      }
    },
    error: function (xhr, textStatus, error) {
      options.error(textStatus, error);
    },
  });
};

l9rPodFs.Rename = function (options) {
  options = options || {};

  if (typeof options.success !== "function") {
    options.success = function () {};
  }

  if (typeof options.error !== "function") {
    options.error = function () {};
  }

  if (options.path === undefined) {
    options.error(400, "path can not be null");
    return;
  }

  if (options.pathset === undefined) {
    options.error(400, "file can not be null");
    return;
  }

  var req = {
    path: options.path,
    pathset: options.pathset,
  };

  var url =
    "mod-set-fs/rename?modname=" + lynkui.session.get("hp-speceditor-modname");
  hpMgr.ApiCmd(url, {
    method: "POST",
    timeout: 10000,
    data: JSON.stringify(req),
    callback: function (err, data) {
      if (!data) {
        options.error(500, "Networking Error");
      } else if (data.kind == "FsFile") {
        options.success(data);
      } else if (data.error) {
        options.error(data.error.code, data.error.message);
      } else {
        options.error(500, "Networking Error");
      }
    },
    error: function (xhr, textStatus, error) {
      options.error(textStatus, error);
    },
  });
};

l9rPodFs.Del = function (options) {
  options = options || {};

  if (typeof options.success !== "function") {
    options.success = function () {};
  }

  if (typeof options.error !== "function") {
    options.error = function () {};
  }

  if (options.path === undefined) {
    options.error(400, "path can not be null");
    return;
  }

  var req = {
    path: options.path,
  };

  var url =
    "mod-set-fs/del?modname=" + lynkui.session.get("hp-speceditor-modname");

  hpMgr.ApiCmd(url, {
    method: "POST",
    timeout: 10000,
    data: JSON.stringify(req),
    callback: function (err, data) {
      if (!data) {
        options.error(500, "Networking Error");
      } else if (data.kind == "FsFile") {
        options.success(data);
      } else if (data.error) {
        options.error(data.error.code, data.error.message);
      } else {
        options.error(500, "Networking Error");
      }
    },
    error: function (xhr, textStatus, error) {
      options.error(textStatus, error);
    },
  });
};

l9rPodFs.List = function (options) {
  // Force options to be an object
  options = options || {};

  if (options.path === undefined) {
    return;
  }

  if (typeof options.success !== "function") {
    options.success = function () {};
  }

  if (typeof options.error !== "function") {
    options.error = function () {};
  }

  var url =
    "mod-set-fs/list?modname=" + lynkui.session.get("hp-speceditor-modname");
  url += "&path=" + options.path;

  hpMgr.ApiCmd(url, {
    method: "GET",
    timeout: 30000,
    callback: function (err, data) {
      if (!data) {
        options.error(500, "Networking Error");
      } else if (data.kind == "FsFileList") {
        options.success(data);
      } else if (data.error) {
        options.error(data.error.code, data.error.message);
      } else {
        options.error(500, "Networking Error");
      }
    },
    error: function (xhr, textStatus, error) {
      options.error(textStatus, error);
    },
  });
};

var l9rLayout = {
  init: false,
  colsep: 0,
  width: 0,
  height: 0,
  postop: 0,
  cols: [
    {
      id: "lcbind-proj-filenav",
      width: 15,
      minWidth: 200,
    },
    {
      id: "lclay-colmain",
      width: 85,
    },
  ],
};

l9rLayout.Initialize = function () {
  if (l9rLayout.init) {
    return;
  }

  for (var i in l9rLayout.cols) {
    var wl = lynkui.storage.get(
      lynkui.session.get("hp-speceditor-modname") +
        "_laysize_" +
        l9rLayout.cols[i].id
    );

    if (wl !== undefined && parseInt(wl) > 0) {
      l9rLayout.cols[i].width = parseInt(wl);
    } else {
      var ws = lynkui.session.get("laysize_" + l9rLayout.cols[i].id);
      if (ws !== undefined && parseInt(ws) > 0) {
        l9rLayout.cols[i].width = parseInt(ws);
      }
    }
  }
};

l9rLayout.BindRefresh = function () {
  $(".lclay-col-resize").bind("mousedown", function (e) {
    var layid = $(this).attr("lc-layid");

    // console.log("lclay-col-resize mousedown: "+ layid);

    var leftLayId = "",
      rightLayId = "";
    var leftIndexId = 0,
      rightIndexId = 1;
    var leftWidth = 0,
      rightWidth = 0;
    var leftMinWidth = 0,
      rightMinWidth = 0;
    for (var i in l9rLayout.cols) {
      rightLayId = l9rLayout.cols[i].id;
      rightWidth = l9rLayout.cols[i].width;
      rightMinWidth = (100 * 200) / l9rLayout.width;
      rightIndexId = i;
      if (l9rLayout.cols[i].minWidth !== undefined) {
        rightMinWidth = (100 * l9rLayout.cols[i].minWidth) / l9rLayout.width;
      }

      if (rightLayId == layid) {
        break;
      }

      leftLayId = rightLayId;
      leftWidth = rightWidth;
      leftMinWidth = rightMinWidth;
      leftIndexId = rightIndexId;
    }

    var leftStart = $("#" + leftLayId).position().left;

    // $("#lcbind-col-rsline").remove();
    // $("body").append("<div id='lcbind-col-rsline'></div>");
    // $("#lcbind-col-rsline").css({
    //     height : l9rLayout.height,
    //     left   : e.pageX,
    //     bottom : 10
    // }).show();

    var posLast = e.pageX;

    $("#lcbind-layout").bind("mousemove", function (e) {
      // console.log("lcbind-layout mousemove: "+ e.pageX);

      // $("#lcbind-col-rsline").css({left: e.pageX});

      if (Math.abs(posLast - e.pageX) < 4) {
        return;
      }
      posLast = e.pageX;

      var leftWidthNew = (100 * (e.pageX - 5 - leftStart)) / l9rLayout.width;
      // var fixWidthRate = leftWidthNew - leftWidth;
      var rightWidthNew = rightWidth - leftWidthNew + leftWidth;

      if (leftWidthNew <= leftMinWidth || rightWidthNew <= rightMinWidth) {
        return;
      }

      l9rLayout.cols[leftIndexId].width = leftWidthNew;
      l9rLayout.cols[rightIndexId].width = rightWidthNew;

      lynkui.storage.set(
        lynkui.session.get("hp-speceditor-modname") + "_laysize_" + leftLayId,
        leftWidthNew
      );
      lynkui.session.set("laysize_" + leftLayId, leftWidthNew);
      lynkui.storage.set(
        lynkui.session.get("hp-speceditor-modname") + "_laysize_" + rightLayId,
        rightWidthNew
      );
      lynkui.session.set("laysize_" + rightLayId, rightWidthNew);

      setTimeout(function () {
        l9rLayout.Resize();
      }, 0);
    });
  });

  $(document).bind("mouseup", function () {
    $("#lcbind-layout").unbind("mousemove");
    // $("#lcbind-col-rsline").remove();

    l9rLayout.Resize();

    setTimeout(function () {
      l9rLayout.Resize();
    }, 10);
  });
};

l9rLayout.ColumnSet = function (options) {
  options = options || {};

  if (typeof options.success !== "function") {
    options.success = function () {};
  }

  if (typeof options.error !== "function") {
    options.error = function () {};
  }

  if (options.id === undefined) {
    options.error(400, "ID can not be null");
    return;
  }

  var exist = false;
  for (var i in l9rLayout.cols) {
    if (l9rLayout.cols[i].id == options.id) {
      exist = true;

      if (
        options.hook !== undefined &&
        options.hook != l9rLayout.cols[i].hook
      ) {
        l9rLayout.cols[i].hook = options.hook;
      }
    }
  }

  if (!exist) {
    colSet = {
      id: options.id, // Math.random().toString(36).slice(2),
      width: 15,
    };

    if (options.width !== undefined) {
      colSet.width = options.width;
    }

    if (options.minWidth !== undefined) {
      colSet.minWidth = options.minWidth;
    }

    l9rLayout.cols.push(colSet);

    l9rLayout.BindRefresh();
  }
};

l9rLayout.ViewportHeight = function () {
  return (
    window.innerHeight ||
    document.documentElement.clientHeight || // 标准模式下的 <html> 元素
    document.body.clientHeight
  ); // 怪异模式下的 <body> 元素
};

l9rLayout.Resize = function () {
  // console.log("l9rLayout.Resize");

  l9rLayout.Initialize();

  var colSep = 20;

  //
  var bodyHeight = l9rLayout.ViewportHeight(); // $("body").height();
  var bodyWidth = $("body").width() - 20;
  if (bodyWidth != l9rLayout.width) {
    l9rLayout.width = bodyWidth;
    //// use-flex-layout $("#lcbind-layout").width(l9rLayout.width);
  }

  //
  var lyo_p = $("#lcbind-layout").position();
  if (!lyo_p) {
    return;
  }
  var lyo_h = bodyHeight - lyo_p.top - colSep;
  l9rLayout.postop = lyo_p.top;
  if (lyo_h < 400) {
    lyo_h = 400;
  }
  if (lyo_h != l9rLayout.height) {
    l9rLayout.height = lyo_h;
    $("#lcbind-layout").height(l9rLayout.height);
  }

  //
  var colSep1 = 100 * (colSep / l9rLayout.width);
  if (colSep1 != l9rLayout.colsep) {
    l9rLayout.colsep = colSep1;
    $(".lclay-colsep").width(l9rLayout.colsep + "%");
  }
  // console.log("colSep1: "+ colSep1);

  //
  // console.log("l9rLayout.cols.length: "+ l9rLayout.cols.length)
  var colSepAll = (l9rLayout.cols.length - 1) * colSep1;

  var rangeUsed = 0.0;
  for (var i in l9rLayout.cols) {
    if (l9rLayout.cols[i].minWidth) {
      if (
        (l9rLayout.cols[i].width * l9rLayout.width) / 100 <
        l9rLayout.cols[i].minWidth
      ) {
        l9rLayout.cols[i].width =
          100 * ((l9rLayout.cols[i].minWidth + 50) / l9rLayout.width);
      }
    }

    if (l9rLayout.cols[i].width < 10) {
      l9rLayout.cols[i].width = 15;
    } else if (l9rLayout.cols[i].width > 90) {
      l9rLayout.cols[i].width = 80;
    }

    rangeUsed += l9rLayout.cols[i].width;
  }
  // console.log("rangeUsed: "+ rangeUsed);
  // for (var i in l9rLayout.cols) {
  //     console.log("2 id: "+ l9rLayout.cols[i].id +", width: "+ l9rLayout.cols[i].width);
  // }

  var fixRate = (100 - colSepAll) / 100;
  var fixRateSpace = rangeUsed / 100;

  for (var i in l9rLayout.cols) {
    l9rLayout.cols[i].width =
      (l9rLayout.cols[i].width / fixRateSpace) * fixRate;

    $("#" + l9rLayout.cols[i].id).width(l9rLayout.cols[i].width + "%");

    if (typeof l9rLayout.cols[i].hook === "function") {
      l9rLayout.cols[i].hook(l9rLayout.cols[i]);
    }
  }

  // console.log(l9rLayout.cols[0]);

  // for (var i in l9rLayout.cols) {
  //     console.log("3 id: "+ l9rLayout.cols[i].id +", width: "+ l9rLayout.cols[i].width);
  // }

  var fsp = $("#lcbind-fsnav-fstree").position();
  if (fsp) {
    $("#lcbind-fsnav-fstree").width(
      (l9rLayout.width * l9rLayout.cols[0].width) / 100
    );
    $("#lcbind-fsnav-fstree").height(
      l9rLayout.height - (fsp.top - l9rLayout.postop)
    );
    $("#fstdroot").height(l9rLayout.height - (fsp.top - l9rLayout.postop));
  }
};

//
var lcData = {};
lcData.db = null;
lcData.version = 11;
lcData.schema = [
  {
    name: "files",
    pri: "id",
    idx: ["projdir"],
  },
  {
    name: "config",
    pri: "id",
    idx: ["type"],
  },
];
lcData.Init = function (dbname, cb) {
  var req = indexedDB.open(dbname, lcData.version);

  req.onsuccess = function (event) {
    lcData.db = event.target.result;
    cb(true);
  };

  req.onerror = function (event) {
    //console.log("IndexedDB error: " + event.target.errorCode);
    cb(true);
  };

  req.onupgradeneeded = function (event) {
    lcData.db = event.target.result;

    for (var i in lcData.schema) {
      var tbl = lcData.schema[i];

      if (lcData.db.objectStoreNames.contains(tbl.name)) {
        lcData.db.deleteObjectStore(tbl.name);
      }

      var objectStore = lcData.db.createObjectStore(tbl.name, {
        keyPath: tbl.pri,
      });

      for (var j in tbl.idx) {
        objectStore.createIndex(tbl.idx[j], tbl.idx[j], {
          unique: false,
        });
      }
    }
    cb(true);
  };
};

lcData.Put = function (tbl, entry, cb) {
  if (lcData.db == null) {
    return;
  }

  //console.log("put: "+ entry.id);

  var req = lcData.db
    .transaction([tbl], "readwrite")
    .objectStore(tbl)
    .put(entry);

  req.onsuccess = function (event) {
    if (cb != null && cb != undefined) {
      cb(true);
    }
  };

  req.onerror = function (event) {
    if (cb != null && cb != undefined) {
      cb(false);
    }
  };
};

lcData.Get = function (tbl, key, cb) {
  if (lcData.db == null) {
    return;
  }

  var req = lcData.db.transaction([tbl]).objectStore(tbl).get(key);

  req.onsuccess = function (event) {
    cb(req.result);
  };

  req.onerror = function (event) {
    cb(req.result);
  };
};

lcData.Query = function (tbl, column, value, cb) {
  if (lcData.db == null) {
    //console.log("lcData is NULL");
    return;
  }
  var req = lcData.db
    .transaction([tbl])
    .objectStore(tbl)
    .index(column)
    .openCursor();

  req.onsuccess = function (event) {
    cb(event.target.result);
  };

  req.onerror = function (event) {
    //
  };
};

lcData.Del = function (tbl, key, cb) {
  if (lcData.db == null) {
    return;
  }

  var req = lcData.db
    .transaction([tbl], "readwrite")
    .objectStore(tbl)
    .delete(key);

  req.onsuccess = function (event) {
    cb(true);
  };

  req.onerror = function (event) {
    cb(false);
  };
};

lcData.List = function (tbl, cb) {
  if (lcData.db == null) {
    return;
  }

  var req = lcData.db
    .transaction([tbl], "readwrite")
    .objectStore(tbl)
    .openCursor();

  req.onsuccess = function (event) {
    var cursor = event.target.result;
    if (cursor) {
      cb(cursor);
    }
  };

  req.onerror = function (event) {};
};
