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

var lcEditor = {};

lcEditor.Config = {
  theme: "monokai",
  tabSize: 4,
  lineWrapping: true,
  smartIndent: true,
  tabs2spaces: true,
  codeFolding: false,
  fontSize: 13,
  EditMode: "win",
  LangEditMode: "Editor Mode Settings",
  // 'TmpEditorZone' : 'w0',
  TmpScrollLeft: 0,
  TmpScrollTop: 0,
  TmpCursorLine: 0,
  TmpCursorCh: 0,
  TmpLine2Str: null,
  TmpUrid: null,
};

lcEditor.isInited = false;

lcEditor.TabDefault = "lctab-default";

// lcEditor.MessageReply = function(cb, msg)
// {
//     if (cb != null && cb.length > 0) {
//         eval(cb +"(msg)");
//     }
// }
// lcEditor.MessageReplyStatus = function(cb, status, message)
// {
//     lcEditor.MessageReply(cb, {status: status, message: message});
// }

lcEditor.TabletOpen = function (urid, callback) {
  // console.log("lcEditor.TabletOpen 1: "+ urid);
  var item = l9rTab.pool[urid];
  if (l9rTab.frame[item.target].urid == urid) {
    callback(true);
    return;
  }

  // console.log("lcEditor.TabletOpen 2: "+ urid);
  // console.log(item);

  lcData.Get("files", urid, function (ret) {
    // console.log("lcData.Get.files");

    if (
      ret &&
      urid == ret.id &&
      ((ret.ctn1_sum && ret.ctn1_sum.length > 30) ||
        (ret.ctn0_sum && ret.ctn0_sum.length > 30))
    ) {
      //l9rTab.pool[urid].data = ret.ctn1_src;
      //l9rTab.pool[urid].hash = l4iString.CryptoMd5(ret.ctn1_src);
      // console.log(ret);
      lcEditor.LoadInstance(ret);
      callback(true);
      return;
    }

    //$("#src"+urid).remove(); // Force remove

    //var t = '<textarea id="src'+urid+'" class="displaynone"></textarea>';
    //$("#lctab-body"+ item.target).prepend(t);

    // var req = {
    //     "access_token" : l4iCookie.Get("access_token"),
    //     "data" : {
    //         "path" : l4iSession.Get("ProjPath") +"/"+ item.url
    //     }
    // }

    var req = {
      path: item.url,
    };

    req.error = function (status, message) {
      // console.log("error 964: "+ status +", "+ message);
      callback(false);
    };

    req.success = function (file) {
      // console.log("success 964:");
      // console.log(file);

      if (file.body === undefined || file.body == null) {
        file.body = "";
      }

      var entry = {
        id: urid,
        modname: l4iSession.Get("hp-speceditor-modname"),
        projdir: l4iSession.Get("hp-speceditor-modname"),
        filepth: item.url,
        ctn0_src: file.body,
        ctn0_sum: l4iString.CryptoMd5(file.body),
        ctn1_src: "",
        ctn1_sum: "",
        mime: file.mime,
        cabid: item.target,
      };
      if (item.icon) {
        entry.icon = item.icon;
      }

      lcData.Put("files", entry, function (ret) {
        if (ret) {
          $("#lctab-bar" + item.target).empty();
          $("#lctab-body" + item.target).empty();

          //l9rTab.pool[urid].mime = obj.data.mime;
          lcEditor.LoadInstance(entry);
          // l9r.HeaderAlert('success', "OK");
          callback(true);
        } else {
          // TODO
          l9r.HeaderAlert("error", "Can not write to IndexedDB");
          callback(false);
        }
      });

      // callback(true);
    };

    l9rPodFs.Get(req);
  });
};

lcEditor.LoadInstance = function (entry) {
  var item = l9rTab.pool[entry.id];
  if (!item) {
    return;
  }

  if (item.modname != l4iSession.Get("hp-speceditor-modname")) {
    return;
  }

  var modes = [];
  var ext = item.url.split(".").pop();
  switch (ext) {
    case "c":
    case "h":
    case "cc":
    case "cpp":
    case "hpp":
    case "java":
      mode = "clike";
      break;

    case "php":
    case "css":
    case "xml":
    case "go":
    case "lua":
    case "sql":
      // case "less":
      mode = ext;
      break;
    // case "sql":
    //     mode = "plsql";
    //     break;
    case "js":
    case "json":
      mode = "javascript";
      break;
    case "sh":
      mode = "shell";
      break;
    case "py":
      mode = "python";
      break;
    case "rb":
      mode = "ruby";
      break;
    case "perl":
    case "prl":
    case "pl":
    case "pm":
      mode = "perl";
      break;
    case "md":
      mode = "markdown";
      break;
    case "yml":
    case "yaml":
      mode = "yaml";
      break;
    default:
      mode = "htmlmixed";
  }

  switch (entry.mime) {
    case "text/x-php":
      mode = "php";
      break;
    case "text/x-shellscript":
      mode = "shell";
      break;
  }

  //l9rTab.frame[item.target].urid = entry.id;

  if (l9rTab.frame[item.target].editor != null) {
    $("#lctab-body" + item.target).empty();
    $("#lctab-bar" + item.target).empty();
  }

  // styling
  $(".CodeMirror-lines").css({
    "font-size": lcEditor.Config.fontSize + "px",
  });

  var src = entry.ctn1_sum.length > 30 ? entry.ctn1_src : entry.ctn0_src;
  //console.log(entry);

  lcEditor.Config.TmpLine2Str = null;
  if (item.editor_strto && item.editor_strto.length > 1) {
    lcEditor.Config.TmpLine2Str = item.editor_strto;
    l9rTab.pool[entry.id].editor_strto = null;
  }

  lcEditor.Config.TmpScrollLeft = isNaN(entry.scrlef)
    ? 0
    : parseInt(entry.scrlef);
  lcEditor.Config.TmpScrollTop = isNaN(entry.scrtop)
    ? 0
    : parseInt(entry.scrtop);
  lcEditor.Config.TmpCursorLine = isNaN(entry.curlin)
    ? 0
    : parseInt(entry.curlin);
  lcEditor.Config.TmpCursorCH = isNaN(entry.curch) ? 0 : parseInt(entry.curch);
  lcEditor.Config.TmpUrid = entry.id;

  if (!lcEditor.isInited) {
    CodeMirror.defineInitHook(function (cminst) {
      l9rLayout.Resize();

      if (lcEditor.Config.TmpLine2Str != null) {
        //console.log("line to"+ lcEditor.Config.TmpLine2Str);
        var crs = cminst.getSearchCursor(
          lcEditor.Config.TmpLine2Str,
          cminst.getCursor(),
          null
        );

        if (crs.findNext()) {
          var lineto = crs.from().line + 3;
          if (lineto > cminst.lineCount()) {
            lineto = cminst.lineCount() - 1;
          }

          cminst.scrollIntoView({
            line: lineto,
            ch: 0,
          });
        }
      }

      if (
        lcEditor.Config.TmpScrollLeft > 0 ||
        lcEditor.Config.TmpScrollTop > 0
      ) {
        cminst.scrollTo(
          lcEditor.Config.TmpScrollLeft,
          lcEditor.Config.TmpScrollTop
        );
      }

      if (
        lcEditor.Config.TmpCursorLine > 0 ||
        lcEditor.Config.TmpCursorCH > 0
      ) {
        cminst.focus();
        cminst.setCursor(
          lcEditor.Config.TmpCursorLine,
          lcEditor.Config.TmpCursorCH
        );
      }
    });

    lcEditor.isInited = true;
  }

  $("#lctab-body" + item.target).empty();

  l9rTab.frame[item.target].editor = CodeMirror(
    document.getElementById("lctab-body" + item.target),
    {
      value: src,
      lineNumbers: true,
      matchBrackets: true,
      undoDepth: 1000,
      mode: mode,
      indentUnit: lcEditor.Config.tabSize,
      tabSize: lcEditor.Config.tabSize,
      theme: lcEditor.Config.theme,
      smartIndent: lcEditor.Config.smartIndent,
      lineWrapping: lcEditor.Config.lineWrapping,
      foldGutter: lcEditor.Config.codeFolding,
      gutters: ["CodeMirror-linenumbers", "CodeMirror-foldgutter"],
      rulers: [
        {
          color: "#777",
          column: 80,
          lineStyle: "dashed",
        },
      ],
      autoCloseTags: true,
      autoCloseBrackets: true,
      showCursorWhenSelecting: true,
      styleActiveLine: true,
      extraKeys: {
        Tab: function (cm) {
          if (lcEditor.Config.tabs2spaces) {
            var spaces = Array(cm.getOption("indentUnit") + 1).join(" ");
            cm.replaceSelection(spaces, "end", "+input");
          }
        },
        "Shift-Space": "autocomplete",
        // Mac
        "Cmd-S": function () {
          lcEditor.EntrySave({
            urid: entry.id,
          });
        },
        // PC
        "Ctrl-S": function () {
          lcEditor.EntrySave({
            urid: entry.id,
          });
        },
      },
    }
  );

  // CodeMirror.modeURL = l9r.basecm +"mode/%N/%N.js";
  // CodeMirror.autoLoadMode(l9rTab.frame[item.target].editor, mode);

  l9rTab.frame[item.target].editor.on("change", function (cm) {
    lcEditor.Changed(entry.id);
  });

  l9rTab.frame[item.target].editor.on("cursorActivity", function (cm) {
    lcEditor.CursorActivity(entry.id);
  });

  CodeMirror.commands.find = function (cm) {
    lcEditor.Search();
  };

  CodeMirror.commands.autocomplete = function (cm) {
    CodeMirror.showHint(cm, CodeMirror.hint.javascript);
  };

  setTimeout(l9rLayout.Resize, 200);

  // });
};

lcEditor.CursorActivity = function (urid) {
  var item = l9rTab.pool[urid];
  if (!item) {
    return;
  }

  if (urid != l9rTab.frame[item.target].urid) {
    return;
  }

  lcData.Get("files", urid, function (entry) {
    if (!entry || entry.id != urid) {
      return;
    }

    var editor = l9rTab.frame[item.target].editor;

    var prevEditorScrollInfo = editor.getScrollInfo();
    var prevEditorCursorInfo = editor.getCursor();

    entry.scrlef = prevEditorScrollInfo.left;
    entry.scrtop = prevEditorScrollInfo.top;

    entry.curlin = prevEditorCursorInfo.line;
    entry.curch = prevEditorCursorInfo.ch;

    lcData.Put("files", entry, function (ret) {
      // TODO
    });
  });
};

lcEditor.Changed = function (urid) {
  //console.log("lcEditor.Changed:"+ urid);

  if (!l9rTab.pool[urid]) {
    return;
  }
  var item = l9rTab.pool[urid];

  if (urid != l9rTab.frame[item.target].urid) {
    return;
  }

  lcData.Get("files", urid, function (entry) {
    if (!entry || entry.id != urid) {
      return;
    }

    var editor = l9rTab.frame[item.target].editor;

    entry.ctn1_src = editor.getValue();
    entry.ctn1_sum = l4iString.CryptoMd5(entry.ctn1_src);

    var prevEditorScrollInfo = editor.getScrollInfo();
    var prevEditorCursorInfo = editor.getCursor();

    entry.scrlef = prevEditorScrollInfo.left;
    entry.scrtop = prevEditorScrollInfo.top;

    entry.curlin = prevEditorCursorInfo.line;
    entry.curch = prevEditorCursorInfo.ch;

    lcData.Put("files", entry, function (ret) {
      // TODO
      // console.log(entry);
    });
  });

  $("#pgtab" + urid + " .chg").show();
  $("#pgtab" + urid + " .pgtabtitle").addClass("chglight");
};

lcEditor.SaveCurrent = function () {
  lcEditor.EntrySave({
    urid: l9rTab.frame[lcEditor.TabDefault].urid,
  });
};

lcEditor.EntrySave = function (options) {
  options = options || {};

  if (typeof options.success !== "function") {
    options.success = function () {};
  }

  if (typeof options.error !== "function") {
    options.error = function () {};
  }

  if (options.urid === undefined) {
    return;
  }

  lcData.Get("files", options.urid, function (ret) {
    if (ret.id == undefined || options.urid != ret.id) {
      options.error(options);
      return;
    }

    var req = {
      urid: options.urid,
      path: ret.filepth,
    };

    var item = l9rTab.pool[options.urid];

    if (options.urid == l9rTab.frame[item.target].urid) {
      var ctn = l9rTab.frame[item.target].editor.getValue();
      if (ctn == ret.ctn0_src) {
        $("#pgtab" + options.urid + " .chg").hide();
        $("#pgtab" + options.urid + " .pgtabtitle").removeClass("chglight");

        options.success(options);
        return; // 200
      }

      req.data = ctn;
      req.sumcheck = l4iString.CryptoMd5(ctn);
    } else if (ret.ctn1_sum.length < 30) {
      options.success(options);
      return; // 200
    } else if (ret.ctn1_src != ret.ctn0_src) {
      req.data = ret.ctn1_src;
      req.sumcheck = ret.ctn1_sum;
    } else if (ret.ctn1_src == ret.ctn0_src) {
      $("#pgtab" + options.urid + " .chg").hide();
      $("#pgtab" + options.urid + " .pgtabtitle").removeClass("chglight");

      options.success(options);
      return;
    }

    req.success = function (rsp) {
      lcData.Get("files", options.urid, function (entry) {
        if (!entry || entry.id != options.urid) {
          options.error(options);
          return;
        }

        entry.ctn0_src = entry.ctn1_src;
        entry.ctn0_sum = entry.ctn1_sum;

        entry.ctn1_src = "";
        entry.ctn1_sum = "";

        lcData.Put("files", entry, function (ret) {
          if (!ret) {
            l9r.HeaderAlert("error", "Failed on write Local.IndexedDB");
            options.error(options);
            return;
          }

          $("#pgtab" + options.urid + " .chg").hide();
          $("#pgtab" + options.urid + " .pgtabtitle").removeClass("chglight");

          options.success(options);
        });
      });
    };

    req.error = function (status, message) {
      l4iAlert.Error(message);
      options.error(options);
    };

    l9rPodFs.Post(req);
  });
};

lcEditor.DialogChanges2SaveSkip = function (urid) {
  l9rTab.Close(urid, 1);
  l4iModal.Close();
};

lcEditor.DialogChanges2SaveDone = function (urid) {
  //console.log(lcEditor.MessageReply(0, "ok"));
  lcEditor.EntrySave({
    urid: urid,
    success: function () {
      l9rTab.Close(urid, 1);
      l4iModal.Close();
    },
    error: function () {
      l4i.InnerAlert(
        "#xi1b3h",
        "alert-error",
        "<span></span>Internal Server Error<span></span>"
      );
    },
  });
};

lcEditor.IsSaved = function (urid, cb) {
  lcData.Get("files", urid, function (ret) {
    if (ret == undefined) {
      cb(true);
      return;
    }

    if (
      ret.id == urid &&
      ret.ctn1_sum.length > 30 &&
      ret.ctn0_sum != ret.ctn1_sum
    ) {
      cb(false);
    } else {
      cb(true);
    }
  });
};

lcEditor.HookOnBeforeUnload = function () {
  if (
    l9rTab.frame[lcEditor.TabDefault].editor != null &&
    l9rTab.frame[lcEditor.TabDefault].urid == lcEditor.Config.TmpUrid
  ) {
    lcData.Get(
      "files",
      l9rTab.frame[lcEditor.TabDefault].urid,
      function (prevEntry) {
        if (!prevEntry) {
          return;
        }

        var prevEditorScrollInfo =
          l9rTab.frame[lcEditor.TabDefault].editor.getScrollInfo();
        var prevEditorCursorInfo =
          l9rTab.frame[lcEditor.TabDefault].editor.getCursor();

        prevEntry.scrlef = prevEditorScrollInfo.left;
        prevEntry.scrtop = prevEditorScrollInfo.top;
        prevEntry.curlin = prevEditorCursorInfo.line;
        prevEntry.curch = prevEditorCursorInfo.ch;

        lcData.Put("files", prevEntry, function () {
          // TODO
        });
      }
    );
  }
};
