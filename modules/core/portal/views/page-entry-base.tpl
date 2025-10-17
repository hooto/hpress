<!DOCTYPE html>
<html lang="en">
  {{pagelet . "core/general" "v3/html-header.tpl"}}
  <link
    rel="stylesheet"
    href="{{HttpSrvBasePath `hp/-/static/pt/css/main.css`}}?v={{.sys_version_sign}}"
    type="text/css"
  />
  <body id="hp-body">
    {{pagelet . "core/general" "v3/nav-header.tpl" "topnav"
    "topbar_class=navbar-light"}}
    <!-- -->
    {{FieldHtmlPrint .page_entry `content` .LANG}}
    <!-- -->
    {{pagelet . "core/general" "v3/footer.tpl"}}
    <!-- -->
    {{pagelet . "core/general" "html-footer.tpl"}}
  </body>
</html>
