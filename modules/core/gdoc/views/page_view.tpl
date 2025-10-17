<!DOCTYPE html>
<html lang="en">
  {{pagelet . "core/general" "v3/html-header.tpl"}}
  <link
    rel="stylesheet"
    href="{{HttpSrvBasePath `hp/-/static/gdoc/css/main.css`}}?v={{.sys_version_sign}}"
    type="text/css"
  />
  <script src="{{HttpSrvBasePath `hp/-/static/gdoc/js/gdoc.js`}}?v={{.sys_version_sign}}"></script>
  <body id="hp-body">
    {{pagelet . "core/general" "v3/nav-header.tpl" "topnav"
    "topbar_class=navbar-light"}}

    <div class="hp-gdoc-index-frame-dark-light">
      <div class="container hp-block-gap-column">
        <nav aria-label="breadcrumb">
          <ol class="breadcrumb hp-block-zero">
            <li class="breadcrumb-item hp-icon-inline">
              <svg class="bi" width="16" height="16" fill="currentColor">
                <use
                  xlink:href="/hp/lynkui/~/bi/v1/bootstrap-icons.svg#filetype-doc"
                />
              </svg>
              <a href="{{.baseuri}}/">
                <span>{{T .LANG `Documents`}}</span>
              </a>
            </li>
            <li class="breadcrumb-item active">
              <a href="{{.baseuri}}/view/{{.doc_entry.ExtPermalinkName}}/">
                {{FieldStringPrint .doc_entry `title` .LANG}}
              </a>
            </li>
          </ol>
        </nav>

        <div class="hp-gdoc-node-content hp-block-gap-column">
          <div class="hp-block-gap-row hp-is-mobile">
            <div class="col-auto">
              <button
                class="btn btn-outline-dark hp-icon-inline"
                onclick="hp.DisplayToggle('hp-gdoc-entry-summary')"
              >
                <svg class="bi" width="16" height="16" fill="currentColor">
                  <use
                    xlink:href="/hp/lynkui/~/bi/v1/bootstrap-icons.svg#list-ul"
                  />
                </svg>
                Menu
              </button>
            </div>
          </div>

          <div class="hp-block-gap-auto">
            <div
              id="hp-gdoc-entry-summary"
              class="col col-md-2 flex-grow-1 hp-is-desktop hp-gdoc-entry-summary hp-content hp-scroll"
            >
              {{FieldHtmlPrint .doc_entry `content` .LANG}}
            </div>
            <div class="col-md-7 flex-grow-1">
              <div
                id="hp-gdoc-page-entry-content"
                class="hp-gdoc-entry-content hp-gdoc-page-content content hp-content"
              >
                {{FieldHtmlPrint .page_entry `content` .LANG}}
              </div>
            </div>
            <div
              id="hp-gdoc-page-entry-toc"
              class="col-md-2 hp-is-desktop hp-gdoc-entry-toc hp-scroll"
            ></div>
          </div>
        </div>
      </div>
    </div>

    {{pagelet . "core/general" "v3/footer.tpl"}}

    <script type="text/javascript">
      window.onload_hooks.push(function () {
        hp.CodeRender({ theme: "monokai" });
        gdoc.PageEntryRender({
          doc_base_path: "{{.baseuri}}/view/{{.doc_entry.ExtPermalinkName}}/",
        });
      });
    </script>

    {{pagelet . "core/general" "html-footer.tpl"}}
  </body>
</html>
