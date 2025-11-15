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

var hp = {
  base: "/hp/",
  sys_version_sign: "1.0",
  debug: true,
};

hp.urlver = function (debug_off) {
  var u = "?v=" + hp.sys_version_sign;
  if (!debug_off && hp.debug) {
    u += "&_=" + Math.random();
  }
  return u;
};

hp.Boot = function () {
  if (window._basepath && window._basepath.length > 1) {
    hp.base = window._basepath;
    if (hp.base.substring(hp.base.length - 1) != "/") {
      hp.base += "/";
    }
  }
  if (window._sys_version_sign && window._sys_version_sign.length > 1) {
    hp.sys_version_sign = window._sys_version_sign;
  }

  if (!hp.base || hp.base == "") {
    hp.base = "/";
  }

  lynkui.use([], function () {
    setTimeout(function () {
      for (var i in window.onload_hooks) {
        window.onload_hooks[i]();
      }
    }, 100);
  });
};

hp.HttpSrvBasePath = function (url) {
  if (hp.base == "") {
    return url;
  }

  if (url.substr(0, 1) == "/") {
    return url;
  }

  return hp.base + url;
};

hp.CodeRender = function (options) {
  $(".hp-content").each(function (i, el) {
    var isMath = /\$\$(.*)\$\$/g.test(el.innerHTML);
    if (!isMath) {
      return;
    }
    el.innerHTML = el.innerHTML.replace(/(\$\$([^\$]*)\$\$)+/g, function (v) {
      return '<span class="language-math">' + v.replace(/\$/g, "") + "</span>";
    });
  });

  options = options || {};
  $("[class^='language-']").each(function (i, el) {
    var lang = el.className.substr("language-".length);

    if (lang == "hchart" || lang == "hooto_chart") {
      return hp.hchartRender(i, el);
    }

    if (lang == "math") {
      return hp.mathRender(i, el);
    }

    var modes = [];

    if (lang == "html") {
      lang = "htmlmixed";
    }

    switch (lang) {
      case "php":
        modes.push("~/cm/5/mode/php/php.js");
      case "htmlmixed":
        modes.push("~/cm/5/mode/xml/xml.js");
        modes.push("~/cm/5/mode/javascript/javascript.js");
        modes.push("~/cm/5/mode/css/css.js");
        modes.push("~/cm/5/mode/htmlmixed/htmlmixed.js");
        break;

      case "c":
      case "cpp":
      case "clike":
      case "java":
        lang = "clike";
        break;

      case "json":
        modes.push("~/cm/5/mode/javascript/javascript.js");
        lang = "application/ld+json";
        break;

      case "clojure":
      case "cmake":
      case "coffeescript":
      case "commonlisp":
      case "css":
      case "d":
      case "dart":
      case "diff":
      case "django":
      case "dockerfile":
      case "erlang":
      case "go":
      case "groovy":
      case "haskell":
      case "http":
      case "javascript":
      case "lua":
      case "markdown":
      case "nginx":
      case "perl":
      case "protobuf":
      case "python":
      case "r":
      case "rpm":
      case "ruby":
      case "rust":
      case "shell":
      case "sql":
      case "swift":
      case "toml":
      case "xml":
      case "yaml":
        modes.push("~/cm/5/mode/" + lang + "/" + lang + ".js");
        break;

      default:
        return;
    }

    var deps = ["~/cm/5/lib/codemirror.css", "~/cm/5/lib/codemirror.js"];
    if (options.theme && options.theme == "monokai") {
      deps.push("~/cm/5/theme/monokai.css");
    } else {
      options.theme = "default";
    }
    lynkui.use(deps, function () {
      modes.push("~/cm/5/addon/runmode/runmode.js");
      modes.push("~/cm/5/mode/clike/clike.js");
      lynkui.use(modes, function () {
        if (options.theme != "default") {
          $(el).addClass("CodeMirror");
        }
        $(el).addClass("cm-s-" + options.theme); // apply a theme class
        CodeMirror.runMode($(el).text().trim(), lang, $(el)[0]);
      });
    });
  });
};

hp.hchartRender = function (i, elem) {
  var elem_id = "hchart-id-" + i;
  elem.setAttribute("id", elem_id);
  lynkui.use(["~/hchart/hchart.js"], function () {
    hooto_chart.basepath = hp.base + "/~/hchart";
    hooto_chart.opts_width = "600px";
    hooto_chart.opts_height = "400px";
    hooto_chart.JsonRenderElement(elem, elem_id);
  });
};

hp.mathRender = function (i, elem) {
  lynkui.use(["~/katex/0.10/katex.css", "~/katex/0.10/katex.js"], function () {
    var txt = elem.innerHTML.replace(/\\‘/g, "'");
    txt = txt.replace(/\\“/g, '"');
    txt = txt.replace(/\&amp;/g, "&");
    elem.innerHTML = katex.renderToString(txt, {
      throwOnError: false,
    });
  });
};

hp.NavActive = function (tplid, nav_path) {
  if (!tplid) {
    return;
  }

  var nav = $("#" + tplid);
  if (!nav) {
    return;
  }

  if (!nav_path || nav_path.length < 1) {
    nav_path = window.location.pathname;
  }
  if (!nav_path || nav_path == "") {
    nav_path = "/";
  }

  var found = false;
  while (true) {
    nav.find("a").each(function () {
      if (found) {
        return;
      }
      var href = $(this).attr("href");
      if (href && href == nav_path) {
        nav.find("a.active").removeClass("active");
        $(this).addClass("active");
        found = true;
      }
    });

    if (found) {
      break;
    }

    if (nav_path.lastIndexOf("/") > 0) {
      nav_path = nav_path.substr(0, nav_path.lastIndexOf("/"));
    } else {
      break;
    }
  }
};

hp.Ajax = function (url, options) {
  options = options || {};

  //
  if (url.substr(0, 1) != "/" && url.substr(0, 4) != "http") {
    url = hp.HttpSrvBasePath(url);
  }

  lynkui.utilx.ajax(url, options);
};

hp.ActionLoader = function (target, url) {
  hp.Ajax(hp.HttpSrvBasePath(url), {
    callback: function (err, data) {
      $("#" + target).html(data);
    },
  });
};

hp.ApiCmd = function (url, options) {
  hp.Ajax(hp.HttpSrvBasePath(url), options);
};

hp.AuthSessionRefresh = function () {
  hp.Ajax(hp.HttpSrvBasePath("auth/session"), {
    callback: function (err, data) {
      if (err || !data || data.kind != "AuthSession") {
        return lynkui.template.render({
          dstid: "hp-topbar-userbar",
          tplid: "hp-topbar-user-unsigned-tpl",
        });
      }

      if (hp.sys_version_sign == "unreg") {
        return (window.location = "/hp/mgr");
      }

      lynkui.template.render({
        dstid: "hp-topbar-userbar",
        tplid: "hp-topbar-user-signed-tpl",
        data: data,
        success: function () {
          $("#hp-topbar-userbar").hover(
            function () {
              $("#hp-topbar-user-signed-modal").fadeIn(200);
            },
            function () {}
          );
          $("#hp-topbar-user-signed-modal").hover(
            function () {},
            function () {
              $("#hp-topbar-user-signed-modal").fadeOut(200);
            }
          );
        },
      });
    },
  });
};

hp.LangChange = function (t) {
  lynkui.cookie.set("lang", t.value, null, "/");
  window.location.reload(true);
};
