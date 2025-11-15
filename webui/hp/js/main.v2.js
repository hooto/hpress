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
        modes.push("~/cm/5/mode/php/php.js" + hp.urlver(true));
      case "htmlmixed":
        modes.push("~/cm/5/mode/xml/xml.js" + hp.urlver(true));
        modes.push("~/cm/5/mode/javascript/javascript.js" + hp.urlver(true));
        modes.push("~/cm/5/mode/css/css.js" + hp.urlver(true));
        modes.push("~/cm/5/mode/htmlmixed/htmlmixed.js" + hp.urlver(true));
        break;

      case "c":
      case "cpp":
      case "clike":
      case "java":
        lang = "clike";
        break;

      case "json":
        modes.push("~/cm/5/mode/javascript/javascript.js" + hp.urlver(true));
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
        modes.push(
          "~/cm/5/mode/" + lang + "/" + lang + ".js" + hp.urlver(true)
        );
        break;

      default:
        return;
    }

    var deps = [
      "~/cm/5/lib/codemirror.css" + hp.urlver(true),
      "~/cm/5/lib/codemirror.js" + hp.urlver(true),
    ];
    if (options.theme && options.theme == "monokai") {
      deps.push("~/cm/5/theme/monokai.css" + hp.urlver(true));
    } else {
      options.theme = "default";
    }
    lynkui.use(deps, function () {
      modes.push("~/cm/5/addon/runmode/runmode.js" + hp.urlver(true));
      modes.push("~/cm/5/mode/clike/clike.js" + hp.urlver(true));
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
  lynkui.use(["~/hchart/hchart.js" + hp.urlver(true)], function () {
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
        //
        nav.find("a.is-active").removeClass("is-active");
        $(this).addClass("is-active");
        //
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
      });
    },
  });
};

hp.LangChange = function (t) {
  lynkui.cookie.set("lang", t.value, null, "/");
  window.location.reload(true);
};

hp.DisplayToggle = function (tplid, v) {
  if (!v) {
    v = "block";
  }
  var el = document.getElementById(tplid);
  if (el) {
    var computedStyle = window.getComputedStyle(el);
    var display = computedStyle.display;

    if (display === "none") {
      el.style.display = v;
    } else {
      el.style.display = "none";
    }
  }
};

hp.NavbarMenuToggle = function (tplid) {
  var el = document.getElementById(tplid);
  if (el) {
    el.classList.toggle("d-block");
    el.classList.toggle("d-none");
  }
};

hp.NavbarMenuUserToggle = function () {
  lynkui.modal.open({
    position: "cursor",
    title: "User Overview",
    tplid: "hp-topbar-user-signed-modal",
    width: 300,
    width_min: 200,
    height: "auto",
    buttons: [
      {
        title: "Close",
        onclick: "lynkui.modal.close()",
      },
    ],
  });
};

hp.NavbarMenuUserClose = function () {
  lynkui.modal.close();
};
