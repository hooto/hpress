<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <title>{{SysConfig "frontend_html_head_subtitle"}}</title>
    <script src="{{HttpSrvBasePath `hp/~/lessui/js/sea.js`}}?v={{.sys_version_sign}}"></script>
    <script src="{{HttpSrvBasePath `hp/~/hpm/js/main.js`}}?v={{.sys_version_sign}}"></script>
    <link
      rel="shortcut icon"
      type="image/x-icon"
      href="{{HttpSrvBasePath `hp/~/hp/img/ap.ico`}}?v={{.sys_version_sign}}"
    />
    <script type="text/javascript">
      window._basepath = {{HttpSrvBasePath "/hp"}};
    </script>
  </head>
  <body id="body-content">
    <div id="hpm-topbar" style="display: none">
      <div class="hpm-topbar-collapse">
        <ul class="hpm-nav" id="hpm-topbar-siteinfo">
          {{if SysConfig "frontend_header_site_logo_url"}}
          <li>
            <img class="hpm-topbar-logo" src="{{SysConfig `frontend_header_site_logo_url`}}" />
          </li>
          {{end}}
          <li class="hpm-topbar-brand">{{SysConfig "frontend_header_site_name"}}</li>
        </ul>
        <ul class="hpm-nav hpm-topbar-nav" id="hpm-topbar-nav-node-specls"></ul>
        <ul class="hpm-nav hpm-nav-right" id="hpm-topbar-userbar">
          <li><a href="{{HttpSrvBasePath `hp/auth/sign-out`}}">Sign Out</a></li>
        </ul>
        <ul class="hpm-nav hpm-nav-right">
          <li><a class="l4i-nav-item" href="#s2/index">Storage</a></li>
          <li><a class="l4i-nav-item" href="#spec/index">Modules</a></li>
          <li><a class="l4i-nav-item" href="#sys/index">System</a></li>
        </ul>
      </div>
    </div>
    <div id="com-content" class="hpm-block-gap-frame">loading</div>
    <script id="hpm-topbar-nav-node-specls-tpl" type="text/html">
      {[~it.items :v]}
      <li><a class="l4i-nav-item" href="#node/index/{[=v.meta.name]}">{[=v.title]}</a></li>
      {[~]}
    </script>
  </body>
  <script type="text/javascript">
    window._sys_version_sign = {{.sys_version_sign}};
    window.onload = hpMgr.Boot();
  </script>
</html>
