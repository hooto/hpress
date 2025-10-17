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

var hpS2 = {
  bucket: "/deft",
};

hpS2.Init = function () {
  l4i.UrlEventRegister("s2/index", hpS2.Index, "hpm-topbar");
};

hpS2.Index = function () {
  l4iStorage.Set("hpm_nav_last_active", "s2/index");

  hpMgr.TplCmd("s2/index", {
    callback: function (err, data) {
      $("#com-content").html(data);
      hpS2.ObjList();
    },
  });
};

hpS2.selector_cb = null;
hpS2.selector_opts = null;

hpS2.ObjListSelector = function (cb, options) {
  hpS2.selector_cb = null;
  hpS2.selector_opts = null;

  if (cb) {
    hpS2.selector_cb = cb;
  }
  if (!options) {
    options = {};
  }
  hpS2.selector_opts = options;

  hpS2.ObjListSelectorRefresh();
};

hpS2.ObjListSelectorRefreshRender = function (path, data) {
  l4iTemplate.Render({
    dstid: "hpm-s2-objls",
    tplid: "hpm-s2-objls-tpl",
    data: data,
  });

  var dirnav = [];

  //
  path = l4i.StringTrim(path.replace(/\/+/g, "/"), "/");

  if (path.length > 0) {
    var prs = path.split("/");
    var ppath = "";
    for (var i in prs) {
      ppath += "/" + prs[i];
      dirnav.push({
        path: ppath,
        name: prs[i],
      });
    }
    dirnav[0].name = "Bucket: deft";
  }

  l4iTemplate.Render({
    dstid: "hpm-s2-objls-dirnav",
    tplid: "hpm-s2-objls-dirnav-tpl",
    data: {
      items: dirnav,
    },
  });
};

hpS2.ObjListSelectorRefresh = function (path) {
  if (!path) {
    path = l4iStorage.Get("hpm_s2_obj_path_active");
  } else {
    path = path.replace(/\/+/g, "/");
    l4iStorage.Set("hpm_s2_obj_path_active", path);
  }
  if (path) {
    path = path.replace(/\/+/g, "/");
  }
  if (path.indexOf(hpS2.bucket) != 0) {
    path = hpS2.bucket;
    l4iStorage.Set("hpm_s2_obj_path_active", path);
  }

  seajs.use(["ep"], function (EventProxy) {
    var ep = EventProxy.create("tpl", "data", function (tpl, data) {
      if (!data || !data.kind) {
        return;
      }

      if (!data.items) {
        data.items = [];
      }

      data._path = path;

      var items = [];

      for (var i in data.items) {
        var name = data.items[i].name;

        data.items[i]._id = l4iString.CryptoMd5(path + "/" + name);
        data.items[i]._abspath = path + "/" + name;

        var ext = null;
        var n = name.lastIndexOf(".");
        if (n > 0) {
          ext = name.toLowerCase().substr(n + 1);
        } else {
          continue;
        }

        if (
          ext == "jpg" ||
          ext == "jpeg" ||
          ext == "png" ||
          ext == "gif" ||
          ext == "svg"
        ) {
          data.items[i]._isimg = true;
        } else {
          data.items[i]._isimg = false;
        }
        if (!data.items[i]._isimg && hpS2.selector_opts.image_only) {
          continue;
        }
        items.push(data.items[i]);
      }

      data.items = items;

      if (tpl) {
        l4iModal.Open({
          title: "Select Images",
          tplsrc: tpl,
          width: 1000,
          height: 700,
          buttons: [
            {
              title: "Cancel",
              onclick: "l4iModal.Close()",
            },
          ],
          callback: function () {
            hpS2.ObjListSelectorRefreshRender(path, data);
          },
        });
      } else {
        hpS2.ObjListSelectorRefreshRender(path, data);
      }
    });

    ep.fail(function (err) {
      // TODO
      alert("SpecListRefresh error, Please try again later (EC:app-nodelist)");
    });

    var el = document.getElementById("hpm-s2-objls");
    if (!el || el.length < 1) {
      hpMgr.TplCmd("s2/selector", {
        callback: ep.done("tpl"),
      });
    } else {
      ep.emit("tpl", null);
    }

    hpMgr.ApiCmd("s2-obj/list?path=" + path, {
      callback: ep.done("data"),
    });
  });
};

hpS2.ObjListSelectorEntry = function (path) {
  if (!hpS2.selector_cb) {
    return;
  }
  hpS2.selector_opts.path = path;
  hpS2.selector_cb(hpS2.selector_opts);
  l4iModal.Close();
};

hpS2.ObjList = function (path) {
  if (!path) {
    path = l4iStorage.Get("hpm_s2_obj_path_active");
    if (!path) {
      path = hpS2.bucket;
    }
  }

  path = path.replace(/\/+/g, "/");
  l4iStorage.Set("hpm_s2_obj_path_active", path);
  if (path.indexOf(hpS2.bucket) != 0) {
    path = hpS2.bucket;
    l4iStorage.Set("hpm_s2_obj_path_active", path);
  }

  hpMgr.ApiCmd("s2-obj/list?path=" + path, {
    callback: function (err, data) {
      if (err || !data || !data.kind) {
        return;
      }

      if (!data.items) {
        data.items = [];
      }

      data._path = path;

      for (var i in data.items) {
        var name = data.items[i].name;

        data.items[i]._id = l4iString.CryptoMd5(path + "/" + name);

        data.items[i]._abspath = path + "/" + name;

        var ext = "bin";
        var n = name.lastIndexOf(".");
        if (n > 0) {
          ext = name.toLowerCase().substr(n + 1);
        }

        if (
          ext == "jpg" ||
          ext == "jpeg" ||
          ext == "png" ||
          ext == "gif" ||
          ext == "svg"
        ) {
          data.items[i]._isimg = true;
        } else {
          data.items[i]._isimg = false;
        }
      }

      l4iTemplate.Render({
        dstid: "hpm-s2-objls",
        tplid: "hpm-s2-objls-tpl",
        data: data,
      });

      var dirnav = [];

      //
      path = l4i.StringTrim(path, "/");
      if (path.length > 0) {
        var prs = path.split("/");
        var ppath = "";
        for (var i in prs) {
          ppath += "/" + prs[i];
          dirnav.push({
            path: ppath,
            name: prs[i],
          });
        }
        dirnav[0].name = "Bucket: deft";
      }

      l4iTemplate.Render({
        dstid: "hpm-s2-objls-dirnav",
        tplid: "hpm-s2-objls-dirnav-tpl",
        data: {
          items: dirnav,
        },
      });
    },
  });
};

// html5 file uploader
var hps2_fsUploadRequestId = "";
var hps2_fsUploadAreaId = "";
var hps2_fsUploadBind = null;

function hps2_fsUploadTraverseTree(reqid, item, path, cap) {
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

      hps2_fsUploadCommit(reqid, file, cap);
    });
  } else if (item.isDirectory) {
    // Get folder contents
    var dirReader = item.createReader();
    dirReader.readEntries(function (entries) {
      for (var i = 0; i < entries.length; i++) {
        hps2_fsUploadTraverseTree(
          reqid,
          entries[i],
          path + item.name + "/",
          cap
        );
      }
    });
  }
}

function hps2_fsUploadHanderDragEnter(evt) {
  this.setAttribute("style", "border-style:dashed;");
}

function hps2_fsUploadHanderDragLeave(evt) {
  this.setAttribute("style", "");
}

function hps2_fsUploadHanderDragOver(evt) {
  evt.stopPropagation();
  evt.preventDefault();
}

function hps2_fsUploadCommit(reqid, file, cap) {
  var reader = new FileReader();

  reader.onload = (function (file) {
    return function (e) {
      if (e.target.readyState != FileReader.DONE) {
        return;
      }

      var ppath = $("#" + reqid + " :input[name=path]").val();

      hpMgr.ApiCmd("s2-obj/put", {
        method: "POST",
        data: JSON.stringify({
          path: ppath + "/" + file.name,
          size: file.size,
          body: e.target.result,
          encode: "base64",
        }),
        callback: function (err, rsp) {
          if (err) {
            return;
          }

          if (rsp && rsp.kind && rsp.kind == "FsFile") {
            $("#" + reqid + "-alert")
              .show()
              .append(
                "<div class='hps2-fsupload-item'>" + file.name + " OK</div>"
              );
          } else {
            if (rsp.error) {
              $("#" + reqid + "-alert")
                .show()
                .append(
                  "<div class='hps2-fsupload-item'>" +
                    file.name +
                    " Failed: " +
                    rsp.error.message +
                    "</div>"
                );
            } else {
              $("#" + reqid + "-alert")
                .show()
                .append(
                  "<div class='hps2-fsupload-item'>" +
                    file.name +
                    " Failed</div>"
                );
            }
          }

          if (cap && cap > 0) {
            var items = $("#" + reqid + "-alert").find(".hps2-fsupload-item");
            if (items.length >= cap) {
              setTimeout(function () {
                hpS2.ObjList(ppath);
                l4iModal.Close();
              }, 1000);
            }
          }
        },
      });
    };
  })(file);

  reader.readAsDataURL(file);
}

function hps2_fsUploadHander(evt) {
  evt.stopPropagation();
  evt.preventDefault();

  var items = evt.dataTransfer.items;
  for (var i = 0; i < items.length; i++) {
    // webkitGetAsEntry is where the magic happens
    var item = items[i].webkitGetAsEntry();
    if (item) {
      hps2_fsUploadTraverseTree(
        hps2_fsUploadRequestId,
        item,
        null,
        items.length
      );
    }
  }
}

hpS2.fsUploadRequestId = "hpm-s2-fsupload";
hpS2.fsUploadAreaId = "hpm-s2-fsupload-area";

hpS2.ObjNew = function (type, path, file) {
  if (!path) {
    path = l4iStorage.Get("hpm_s2_obj_path_active");
    if (!path) {
      path = "/";
    }
  }
  path = "/" + l4i.StringTrim(path.replace(/\/+/g, "/"), "/") + "/";

  var formid = Math.random().toString(36).slice(2);

  var req = {
    title: type == "dir" ? "New Folder" : "New File",
    width: 850,
    height: 550,
    tplid: "hpm-s2-objnew-tpl",
    data: {
      formid: formid,
      file: file,
      path: path,
      type: type,
    },
    buttons: [
      {
        onclick: 'hpS2.ObjNewSave("' + formid + '")',
        title: "Upload",
        style: "btn-primary",
      },
      {
        onclick: "l4iModal.Close()",
        title: "Close",
      },
    ],
    callback: function (err, data) {
      hps2_fsUploadRequestId = formid;

      // console.log("ids: "+ hps2_fsUploadRequestId +", "+ areaid);

      if (hps2_fsUploadBind != null) {
        hps2_fsUploadBind.removeEventListener(
          "dragenter",
          hps2_fsUploadHanderDragEnter,
          false
        );
        hps2_fsUploadBind.removeEventListener(
          "dragover",
          hps2_fsUploadHanderDragOver,
          false
        );
        hps2_fsUploadBind.removeEventListener(
          "drop",
          hps2_fsUploadHander,
          false
        );
        hps2_fsUploadBind.removeEventListener(
          "dragleave",
          hps2_fsUploadHanderDragLeave,
          false
        );

        hps2_fsUploadBind = null;
      }

      // console.log("id:"+ areaid);

      hps2_fsUploadBind = document.getElementById(hpS2.fsUploadAreaId);

      // console.log(hps2_fsUploadBind);

      hps2_fsUploadBind.addEventListener(
        "dragenter",
        hps2_fsUploadHanderDragEnter,
        false
      );
      hps2_fsUploadBind.addEventListener(
        "dragover",
        hps2_fsUploadHanderDragOver,
        false
      );
      hps2_fsUploadBind.addEventListener("drop", hps2_fsUploadHander, false);
      hps2_fsUploadBind.addEventListener(
        "dragleave",
        hps2_fsUploadHanderDragLeave,
        false
      );
    },
  };

  l4iModal.Open(req);
};

hpS2.ObjNewSave = function (formid) {
  var elem = document.getElementById("hpm-s2-objnew-files");
  for (var i = 0; i < elem.files.length; i++) {
    hpS2._objNewUpload(formid, elem.files[i]);
  }
};

hpS2._objNewUpload = function (formid, file) {
  var reader = new FileReader();

  reader.onload = (function (file) {
    return function (e) {
      if (e.target.readyState != FileReader.DONE) {
        return;
      }

      var ppath = $("#" + formid + " :input[name=path]").val();

      hpMgr.ApiCmd("s2-obj/put", {
        method: "POST",
        data: JSON.stringify({
          path: ppath + "/" + file.name,
          size: file.size,
          body: e.target.result,
          encode: "base64",
        }),
        callback: function (err, rsp) {
          if (rsp && rsp.kind && rsp.kind == "FsFile") {
            $("#" + formid + "-alert")
              .show()
              .append("<div>" + file.name + " OK</div>");
            hpS2.ObjList(ppath);

            setTimeout(function () {
              l4iModal.Close();
            }, 1000);
          } else {
            if (rsp.error) {
              $("#" + formid + "-alert")
                .show()
                .append(
                  "<div>" +
                    file.name +
                    " Failed: " +
                    rsp.error.message +
                    "</div>"
                );
            } else {
              $("#" + formid + "-alert")
                .show()
                .append("<div>" + file.name + " Failed</div>");
            }
          }
        },
        error: function (status, message) {
          $("#" + formid + "-alert")
            .show()
            .append("<div>" + file.name + " Failed</div>");
        },
      });
    };
  })(file);

  reader.readAsDataURL(file);
};

hpS2.ObjDel = function (path) {
  //
  hpMgr.ApiCmd("s2-obj/del?path=" + path, {
    callback: function (err, data) {
      if (data.kind && data.kind == "FsFile") {
        $("#obj" + l4iString.CryptoMd5(path)).remove();
      } else if (data.error) {
        alert(data.error.message);
      }
    },
  });
};

hpS2.UtilResourceSizeFormat = function (size) {
  if (!size) {
    size = 0;
  }

  var ms = [
    [6, "EB"],
    [5, "PB"],
    [4, "TB"],
    [3, "GB"],
    [2, "MB"],
    [1, "KB"],
  ];
  for (var i in ms) {
    if (size > Math.pow(1024, ms[i][0])) {
      return (
        (size / Math.pow(1024, ms[i][0])).toFixed(0) +
        " <span>" +
        ms[i][1] +
        "</span>"
      );
    }
  }

  if (size == 0) {
    return size;
  }

  return size + " <span>B</span>";
};
