<!DOCTYPE html>
<html lang="en">
  {{pagelet . "core/general" "v3/html-header.tpl"}}
  <link
    rel="stylesheet"
    href="{{HttpSrvBasePath `hp/-/static/gdoc/css/main.css`}}?v={{.sys_version_sign}}"
    type="text/css"
  />
  <body id="hp-body">
    {{pagelet . "core/general" "v3/nav-header.tpl" "topnav"
    "topbar_class=navbar-light"}}

    <div
      class="hp-gdoc-index-frame-dark hp-gdoc-node-content hp-gdoc-bgimg-hexagons"
    >
      <div class="container text-center">
        <div class="hp-gdoc-index-frame-title">Explore Documents</div>
      </div>
    </div>

    <div class="container">
      <div class="hp-block-gap-column">
        <!-- items/ -->
        <div
          style="grid-template-columns: repeat(auto-fit, minmax(360px, 1fr))"
          class="hp-gdoc-nodels hp-block-gap-grid"
        >
          {{range $v := .doc_list.Items}}
          <div
            class="hp-gdoc-nodels-item"
            style="min-width: 360px; max-width: 900px"
          >
            <div class="card">
              <div class="row">
                <div class="col-auto" style="width: 70px; padding: 20px 20px">
                  <div class="media-left hp-icon-inline">
                    <svg class="bi" width="48" height="48" fill="currentColor">
                      <use
                        xlink:href="/hp/lynkui/~/bi/v1/bootstrap-icons.svg#book"
                      />
                    </svg>
                  </div>
                </div>
                <div class="col">
                  <div class="card-body">
                    <p
                      class="card-title"
                      style="font-size: 1.2rem; font-weight: bold"
                    >
                      <a href="{{$.baseuri}}/view/{{$v.ExtPermalinkName}}/"
                        >{{FieldStringPrint $v `title` $.LANG}}</a
                      >
                    </p>
                    <p class="subtitle">
                      <span class="hp-icon-inline">
                        <svg
                          class="bi"
                          width="16"
                          height="16"
                          fill="currentColor"
                        >
                          <use
                            xlink:href="/hp/lynkui/~/bi/v1/bootstrap-icons.svg#stopwatch"
                          />
                        </svg>
                        {{UnixtimeFormat $v.Updated "2006-01-02"}}
                      </span>
                      <span>
                        {{range $term := $v.Terms}} {{if eq $term.Name "tags"}}
                        {{if $term.Items}}
                        <span class="hp-icon-inline">
                          <svg
                            class="bi"
                            width="16"
                            height="16"
                            fill="currentColor"
                          >
                            <use
                              xlink:href="/hp/lynkui/~/bi/v1/bootstrap-icons.svg#tags"
                            />
                          </svg>
                          {{range $term_item := $term.Items}}
                          <a
                            href="{{$.baseuri}}/list?term_tags={{$term_item.Title}}"
                            class="tag-item"
                            >{{$term_item.Title}}</a
                          >
                          {{end}}
                        </span>
                        {{end}} {{end}} {{end}}
                      </span>
                    </p>
                  </div>
                  <div class="card-footer text-right bg-transparent border-0">
                    <a
                      class="btn btn-dark"
                      href="{{$.baseuri}}/view/{{$v.ExtPermalinkName}}/"
                      >Read</a
                    >
                  </div>
                </div>
              </div>
            </div>
          </div>
          {{end}}
        </div>
        <!-- /items -->

        <!-- pager/ -->
        {{if .list_pager}}
        <nav class="">
          <ul class="pagination justify-content-center d-flex d-lg-none">
            {{if .list_pager.FirstPageNumber}}
            <li class="page-item">
              <a
                class="page-link"
                href="{{$.baseuri}}/list?page={{.list_pager.FirstPageNumber}}"
                >First</a
              >
            </li>
            {{end}}
            <!-- -->
            {{range $index, $page := .list_pager.RangePages}}
            <li class="page-item">
              <a
                class="page-link {{if eq $page $.list_pager.CurrentPageNumber}} active{{end}}"
                href="{{$.baseuri}}/list?{{FilterUri $ `page` $page}}"
                >{{$page}}</a
              >
            </li>
            {{end}}
            <!-- -->
            {{if .list_pager.LastPageNumber}}
            <li class="page-item">
              <a
                class="page-link"
                href="{{$.baseuri}}/list?page={{.list_pager.LastPageNumber}}"
                >Last</a
              >
            </li>
            {{end}}
          </ul>
        </nav>
        {{end}}
        <!-- /pager -->
      </div>
    </div>

    {{pagelet . "core/general" "v3/footer.tpl"}}
    <!-- -->
    {{pagelet . "core/general" "html-footer.tpl"}}
  </body>
</html>
