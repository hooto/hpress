<nav
  class="navbar navbar-expand-lg _bg-body-tertiary {{.topbar_class}}"
  id="hp-topbar"
>
  <div class="container">
    <a class="navbar-brand d-inline-flex" href="/">
      {{if SysConfig "frontend_header_site_logo_url"}}
      <img
        src="{{SysConfig `frontend_header_site_logo_url`}}"
        height="36"
        alt=""
      />
      {{end}}
      <span class="">{{SysConfig `frontend_header_site_name`}}</span>
    </a>

    <button
      class="navbar-toggler"
      type="button"
      data-bs-toggle="collapse"
      data-bs-target="#hp-topbar-navbar-toggler"
    >
      <span class="navbar-toggler-icon"></span>
    </button>

    <div class="collapse navbar-collapse" id="hp-topbar-navbar-toggler">
      <ul class="navbar-nav me-auto" id="hp-topbar-nav-main">
        {{range $v := .topnav.Items}}
        <li class="nav-item">
          <a class="nav-link _is-tab" href="{{FieldString $v.Fields `url`}}"
            >{{FieldStringPrint $v `title` $.LANG}}</a
          >
        </li>
        {{end}}
      </ul>
      <hr class="d-mode-none text-width-50" />
      <ul class="navbar-nav d-flex">
        <li
          class="nav-item"
          id="hp-topbar-userbar"
          onclick="hp.NavbarMenuUserToggle()"
        ></li>
      </ul>
    </div>
  </div>
</nav>

<script id="hp-topbar-user-unsigned-tpl" type="text/html">
  <a class="btn btn-outline-dark" href="{{HttpSrvBasePath `hp/auth/login`}}"
    >Sign In</a
  >
</script>

<script id="hp-topbar-user-signed-tpl" type="text/html">
  <div class="btn btn-outline-dark">
    <span class="status-name">{[=it.display_name]}</span>
    <img
      id="hp-topbar-user-signed"
      class="status-photo"
      src="{[=it.photo_url]}"
      width="28"
      height="28"
    />

    <div id="hp-topbar-user-signed-modal" style="display:none;">
      <div
        class="hp-topbar-user-signed-modal"
        onclick="hp.NavbarMenuUserClose()"
      >
        <img class="iam-photo" src="{[=it.photo_url]}" />
        <div class="iam-name">{[=it.display_name]}</div>
        {[? it.instance_owner]}
        <a
          class="btn btn-outline-dark iam-btn"
          href="{[=hp.HttpSrvBasePath('mgr')]}"
          target="_blank"
          >Content Manage</a
        >
        {[?]}
        <a
          class="btn btn-outline-dark iam-btn"
          href="{[=it.iam_url]}"
          target="_blank"
          >Account Center</a
        >
        <a
          class="btn btn-outline-dark iam-btn"
          href="{[=hp.HttpSrvBasePath('auth/sign-out')]}"
          >Sign Out</a
        >
      </div>
    </div>
  </div>
</script>

<script type="text/javascript">
  window.onload_hooks.push(function () {
    hp.NavActive("hp-topbar-nav-main", "{{.http_request_path}}");
    hp.AuthSessionRefresh();
  });
</script>
