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

var hpEditor = {
  editors: {},
  s2_bucket_default: "/deft/",
};

hpEditor.Open = function (name, format) {
  lynkui.use(
    ["~/cm/5/lib/codemirror.css", "~/cm/5/lib/codemirror.js"],
    function () {
      lynkui.use(
        [
          "~/cm/5/mode/markdown/markdown.js",
          "~/cm/5/mode/xml/xml.js",
          "~/cm/5/addon/selection/active-line.js",
        ],
        function () {
          var lineNumbers = false;
          if (format == "md" || format == "html" || format == "shtml") {
            lineNumbers = true;
          }

          if (format == "md") {
            $("#field_" + name + "_editor_mdr").show();
          } else {
            $("#field_" + name + "_editor_mdr").hide();
          }

          $("#field_" + name + "_editor_nav")
            .find("button.active")
            .removeClass("active");
          $("#field_" + name + "_editor_nav")
            .find("button.editor-nav-" + format)
            .addClass("active");

          var editor = hpEditor.editors[name];
          if (editor) {
            $("#field_" + name + "_attr_format").val(format);

            hpEditor.editors[name].setOption("lineNumbers", lineNumbers);

            if (format != "md") {
              hpEditor.PreviewClose(name);
            }

            return;
          }

          var elem_layout = $("#field_" + name + "_layout");
          if (!elem_layout) {
            return;
          }
          var bw = $(window).width();
          var height = elem_layout.height(),
            width = elem_layout.width();
          if (bw - width < 100) {
            width = width - 150;
          }
          $("#field_" + name + "_editor").css({
            width: width + "px",
          });

          // $(".CodeMirror-line").css({"width": (width - 200) +"px"});
          // $("#field_"+ name).css({"width": (width - 200) +"px"});

          // var elem_layout_editor = document.getElementById("field_"+name+"_editor");
          // if (!elem_layout_editor) {
          // 	return;
          // }
          // elem_layout_editor.className = "hpm-nodeset-auto-height";

          var elem = document.getElementById("field_" + name);
          if (!elem) {
            return;
          }
          hpEditor.editors[name] = CodeMirror.fromTextArea(elem, {
            mode: "markdown",
            lineNumbers: lineNumbers,
            theme: "default",
            lineWrapping: true,
            styleActiveLine: true,
            // viewportMargin: Infinity,
          });

          // TODO hpEditor.editors[name].setSize(width, height);
          hpEditor.editors[name].setSize("100%", height);

          hpEditor.editors[name].on("change", function (cm) {
            hpEditor.previewChanged(name);
          });

          $("#field_" + name + "_layout").addClass("hpm-editor-cm");
          $("#field_" + name + "_inner_toolbar")
            .find(".preview_open")
            .show();
        }
      );
    }
  );
};

hpEditor.sizeRefresh = function () {
  var dels = [];
  for (var name in hpEditor.editors) {
    var ok = document.getElementById("field_" + name + "_layout");
    if (!ok) {
      dels.push(name);
      continue;
    }

    if (!$("#field_" + name + "_colpreview").is(":visible")) {
      continue;
    }

    hpEditor.PreviewOpen(name);
  }

  for (var i in dels) {
    delete hpEditor.editors[dels[i]];
  }
};

hpEditor.PreviewOpen = function (name) {
  var width = $("#field_" + name + "_layout").width(),
    height = $("#field_" + name + "_layout").height();

  var width5 = (width - 20) / 2;

  $("#field_" + name + "_editor")
    .find(".CodeMirror")
    .css({
      width: width5 + "px",
      height: height + "px",
    });

  $("#field_" + name + "_preview").css({
    width: width5 + "px",
    height: height + "px",
  });

  $("#field_" + name + "_colpreview").show();

  $("#field_" + name + "_editor")
    .find(".CodeMirror")
    .hover(
      function () {
        hpEditor.editorBindScroll(name);
      },
      function () {
        hpEditor.editorUnBindScroll(name);
      }
    );

  $("#field_" + name + "_preview").hover(
    function () {
      hpEditor.previewBindScroll(name);
    },
    function () {
      hpEditor.previewUnBindScroll(name);
    }
  );

  hpEditor.previewChanged(name);

  $("#field_" + name + "_inner_toolbar")
    .find(".preview_close")
    .show();
  $("#field_" + name + "_inner_toolbar")
    .find(".preview_open")
    .hide();
};

hpEditor.PreviewClose = function (name) {
  hpEditor.editorUnBindScroll(name);
  hpEditor.previewUnBindScroll(name);

  $("#field_" + name + "_preview").empty();

  $("#field_" + name + "_colpreview").hide();

  $("#field_" + name + "_editor")
    .find(".CodeMirror")
    .css({
      width: "100%",
    });

  $("#field_" + name + "_inner_toolbar")
    .find(".preview_close")
    .hide();
  $("#field_" + name + "_inner_toolbar")
    .find(".preview_open")
    .show();
};

hpEditor.previewChanged = function (name) {
  if (!$("#field_" + name + "_colpreview").is(":visible")) {
    return;
  }

  var editor = hpEditor.editors[name];
  if (!editor) {
    return;
  }

  var text = editor.getValue();

  // frountend markdown render
  $("#field_" + name + "_preview").html(marked(text));

  // // backend markdown render
  // hpMgr.ApiCmd("/text/markdown-render", {
  //     method : "POST",
  //     data   : text,
  //     callback : function(err, data) {
  //         $("#field_"+ name +"_preview").html(data);
  //     }
  // });
};

hpEditor.editorBindScroll = function (name) {
  hpEditor.previewUnBindScroll(name);

  $("#field_" + name + "_editor")
    .find(".CodeMirror-scroll")
    .on("scroll", function () {
      var height = $(this).outerHeight();
      var scrollTop = $(this).scrollTop();
      var percent = scrollTop / $(this)[0].scrollHeight;
      var preview = $("#field_" + name + "_preview");

      if (scrollTop === 0) {
        preview.scrollTop(0);
      } else if (scrollTop + height >= $(this)[0].scrollHeight) {
        preview.scrollTop(preview[0].scrollHeight);
      } else {
        preview.scrollTop(preview[0].scrollHeight * percent);
      }
    });
};

hpEditor.editorUnBindScroll = function (name) {
  $("#field_" + name + "_editor")
    .find(".CodeMirror-scroll")
    .unbind("scroll");
};

hpEditor.previewBindScroll = function (name) {
  hpEditor.editorUnBindScroll(name);

  $("#field_" + name + "_preview").on("scroll", function () {
    var height = $(this).outerHeight();
    var scrollTop = $(this).scrollTop();
    var percent = scrollTop / $(this)[0].scrollHeight;
    var editorView = $("#field_" + name + "_editor").find(".CodeMirror-scroll");

    if (scrollTop === 0) {
      editorView.scrollTop(0);
    } else if (scrollTop + height >= $(this)[0].scrollHeight) {
      editorView.scrollTop(editorView[0].scrollHeight);
    } else {
      editorView.scrollTop(editorView[0].scrollHeight * percent);
    }
  });
};

hpEditor.previewUnBindScroll = function (name) {
  $("#field_" + name + "_preview").unbind("scroll");
};

hpEditor.StorageImageSelector = function (name) {
  var editor = hpEditor.editors[name];
  if (!editor) {
    return;
  }
  hpS2.ObjListSelector(hpEditor.StorageImageSelectorCallback, {
    name: name,
    image_only: true,
  });
};

hpEditor.StorageImageSelectorCallback = function (options) {
  if (!options || !options.name || !options.path) {
    return;
  }
  var editor = hpEditor.editors[options.name];
  if (!editor) {
    return;
  }
  var cs = editor.getCursor();
  options.path = options.path.replace(hpEditor.s2_bucket_default, "");
  var line =
    "\n![FIG]({{hp_storage_service_endpoint}}/" +
    options.path +
    "?ipl=w960,h960)\n\n";
  line = line.replace(/\/+/g, "/");
  editor.replaceRange(
    line,
    {
      line: cs.line,
      ch: cs.ch,
    },
    {
      line: cs.line,
      ch: cs.ch,
    }
  );
};

hpEditor.Content = function (name) {
  var editor = hpEditor.editors[name];
  if (editor) {
    return editor.getValue();
  }

  return null;
};

hpEditor.ContentSet = function (name, value) {
  var editor = hpEditor.editors[name];
  if (editor) {
    return editor.setValue(value);
  }
  return null;
};

hpEditor.Close = function (name) {
  var edr = hpEditor.editors[name];
  if (edr) {
    hpEditor.editors[name] = null;
    delete hpEditor.editors[name];
  }
};

hpEditor.Clean = function () {
  for (var i in hpEditor.editors) {
    hpEditor.editors[i] = null;
    delete hpEditor.editors[i];
  }
};
