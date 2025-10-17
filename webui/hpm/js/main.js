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

var hpMgr = {
  frtbase: "/hp/",
  base: "/hp/mgr/",
  api: "/hp/v1/",
  basetpl: "/hp/~/hpm/",
  sys_version_sign: "1.0",
  debug: true,
  hotkey_ctrl_s: null,
};

hpMgr.urlver = function (debug_off) {
  var u = "?_v=" + hpMgr.sys_version_sign;
  if (!debug_off && hpMgr.debug) {
    u += "&_=" + Math.random();
  }
  return u;
};

hpMgr.Boot = function () {
  if (window._basepath && window._basepath.length > 1) {
    hpMgr.frtbase = window._basepath;
    if (hpMgr.frtbase.substring(hpMgr.frtbase.length - 1) != "/") {
      hpMgr.frtbase += "/";
    }
    if (!hpMgr.frtbase || hpMgr.frtbase == "") {
      hpMgr.frtbase = "/";
    }
    hpMgr.base = hpMgr.frtbase + "/mgr/";
    hpMgr.api = hpMgr.frtbase + "v1/";
    hpMgr.basetpl = hpMgr.frtbase + "~/hpm/";
  }

  if (window._sys_version_sign && window._sys_version_sign.length > 1) {
    hpMgr.sys_version_sign = window._sys_version_sign;
  }

  document.addEventListener(
    "keydown",
    function (e) {
      if (
        e.keyCode == 83 &&
        (navigator.platform.match("Mac") ? e.metaKey : e.ctrlKey)
      ) {
        e.preventDefault();
        if (hpMgr.hotkey_ctrl_s) {
          hpMgr.hotkey_ctrl_s();
        }
      }
    },
    false
  );

  seajs.config({
    base: hpMgr.frtbase,
    alias: {
      ep: "~/lessui/js/eventproxy.js" + hpMgr.urlver(),
    },
  });

  seajs.use(
    [
      "~/lessui/js/browser-detect.js" + hpMgr.urlver(),
      "~/hp/js/jquery.js" + hpMgr.urlver(),
      "~/lessui/js/eventproxy.js" + hpMgr.urlver(),
    ],
    function () {
      var browser = BrowserDetect.browser;
      var version = BrowserDetect.version;
      var OS = BrowserDetect.OS;

      if (
        !(
          (browser == "Chrome" && version >= 22) ||
          (browser == "Firefox" && version >= 31.0)
        )
      ) {
        $("body").load(window._basepath + "/error/browser?" + hpMgr.urlver());
        return;
      }

      seajs.use(
        [
          "lynkui/~/bs/v5/css/bootstrap.css" + hpMgr.urlver(),
          "~/lessui/js/lessui.js" + hpMgr.urlver(),
          "~/lessui/css/lessui.css" + hpMgr.urlver(),
          "~/hp/css/main.v2.css" + hpMgr.urlver(),
          "~/hp/js/marked.js" + hpMgr.urlver(),
          "~/hpm/css/main.css" + hpMgr.urlver(),
          "~/hpm/css/defx.css" + hpMgr.urlver(),
          "~/hpm/js/spec.js" + hpMgr.urlver(),
          "~/hpm/js/spec-editor.js" + hpMgr.urlver(),
          "~/hpm/js/tablet.js" + hpMgr.urlver(),
          "~/hpm/js/lc-editor.js" + hpMgr.urlver(),
          "~/hpm/js/model.js" + hpMgr.urlver(),
          "~/hpm/js/term.js" + hpMgr.urlver(),
          "~/hpm/js/node.js" + hpMgr.urlver(),
          "~/hpm/js/sys.js" + hpMgr.urlver(),
          "~/hpm/js/s2.js" + hpMgr.urlver(),
          "~/hpm/js/editor.js" + hpMgr.urlver(),
        ],
        function () {
          setTimeout(hpMgr.BootInit, 300);
        }
      );
    }
  );
};

hpMgr.BootInit = function () {
  l4i.debug = hpMgr.debug;

  $("#hpm-topbar").css({
    display: "block",
  });

  hpSys.Init();

  marked.setOptions({
    renderer: new marked.Renderer(),
    gfm: true,
    tables: true,
    breaks: false,
    pedantic: false,
    sanitize: true,
    smartLists: true,
    smartypants: true,
  });

  $(window).resize(function () {
    hpEditor.sizeRefresh();
  });

  hpSys.Init();
  hpSpec.Init();
  hpS2.Init();

  hpNode.Init(function () {
    var navlast = l4iStorage.Get("hpm_nav_last_active");
    if (!navlast) {
      navlast = "sys/index";
    }
    l4i.UrlEventHandler(navlast);
  });
};

hpMgr.HttpSrvBasePath = function (url) {
  if (hpMgr.base == "") {
    return url;
  }

  if (url.substr(0, 1) == "/") {
    return url;
  }

  return hpMgr.base + url;
};

hpMgr.Ajax = function (url, options) {
  options = options || {};

  //
  if (url.substr(0, 1) != "/" && url.substr(0, 4) != "http") {
    url = hpMgr.HttpSrvBasePath(url);
  }

  l4i.Ajax(url, options);
};

hpMgr.AlertUserLogin = function () {
  l4iAlert.Open(
    "warn",
    "You are not logged in, or your login session has expired. Please sign in again",
    {
      close: false,
      buttons: [
        {
          title: "SIGN IN",
          href: hpMgr.frtbase + "auth/login",
        },
      ],
    }
  );
};

hpMgr.ApiCmd = function (url, options) {
  if (options.nocache === undefined) {
    options.nocache = true;
  }

  var appcb = null;
  if (options.callback) {
    appcb = options.callback;
  }
  options.callback = function (err, data) {
    if (err == "Unauthorized") {
      return hpMgr.AlertUserLogin();
    }
    if (appcb) {
      appcb(err, data);
    }
  };

  hpMgr.Ajax(hpMgr.api + url, options);
};

hpMgr.TplCmd = function (url, options) {
  hpMgr.Ajax(hpMgr.basetpl + url + ".tpl" + hpMgr.urlver(true), options);
};

hpMgr.Loader = function (target, uri) {
  hpMgr.Ajax(hpMgr.basetpl + uri + ".tpl" + hpMgr.urlver(true), {
    callback: function (err, data) {
      $(target).html(data);
    },
  });
};

hpMgr.BodyLoader = function (uri) {
  hpMgr.Loader("#body-content", uri);
};

hpMgr.ComLoader = function (uri) {
  hpMgr.Loader("#com-content", uri);
};

hpMgr.WorkLoader = function (uri) {
  hpMgr.Loader("#work-content", uri);
};

hpMgr.Equal = function (a, b) {
  return a == b;
};

hpMgr.NotEqual = function (a, b) {
  return a != b;
};

hpMgr.GreaterThan = function (a, b) {
  return a > b;
};
