<!doctype html>
<html lang="en">
  {{pagelet . "core/general" "v3/html-header.tpl"}}
  <body id="hp-body">
    {{pagelet . "core/general" "v3/nav-header.tpl" "topnav"}}

    <div class="container hp-block-gap-column">
      <!-- sub-nav/ -->
      <div class="hp-block-gap">
        <div class="col-8 d-none d-lg-block">
          <div class="hp-ctn-title">Content Explore</div>
        </div>
        <div class="d-lg-none">
          <button
            type="button"
            class="btn btn-outline-dark"
            onclick="hp.NavbarMenuToggle('hp-node-list-term-categories')"
          >
            <svg class="bi" width="16" height="16" fill="currentColor">
              <use
                xlink:href="/hp/lynkui/~/bi/v1/bootstrap-icons.svg#list-ul"
              />
            </svg>
          </button>
        </div>
        <div class="col-3">
          <form class="" action="{{.baseuri}}/list">
            <div class="input-group">
              <input
                type="text"
                class="form-control"
                placeholder=""
                name="qry_text"
                value="{{.qry_text}}"
              />
              <button class="btn btn-dark" type="button">Search</button>
            </div>
          </form>
        </div>
      </div>
      <!-- /sub-nav -->

      <div class="hp-block-gap-column">
        <div id="hp-node-list-term-categories" class="row d-none">
          <div class="col col-12">
            {{pagelet . .modname "term/categories2.tpl"}}
          </div>
        </div>

        <div class="row hp-block-gap">
          <div class="col col-11 col-lg-8 hp-block-gap-column">
            <div class="hp-node-list d-flex flex-column hp-block-gap-column">
              {{range $v := .list.Items}}
              <div class="hp-node-list-item clearfix">
                <h4 class="hp-node-list-heading">
                  <a href="{{$.baseuri}}/view/{{$v.ID}}.html"
                    >{{FieldStringPrint $v "title" $.LANG}}</a
                  >
                </h4>
                <div class="hp-node-list-info">
                  <span class="info-item">
                    Published : {{UnixtimeFormat $v.Created "Y-m-d"}}
                  </span>

                  {{range $term := $v.Terms}} {{if eq $term.Name "categories"}}
                  {{if $term.Items}}
                  <span class="info-item">
                    Categories : {{range $term_item := $term.Items}}
                    <a
                      href='{{$.baseuri}}/list?term_categories={{printf "%d" $term_item.ID}}'
                      >{{$term_item.Title}}</a
                    >
                    {{end}}
                  </span>
                  {{end}} {{end}} {{end}} {{range $term := $v.Terms}} {{if eq
                  $term.Name "tags"}} {{if $term.Items}}
                  <span class="info-item">
                    Tags : {{range $term_item := $term.Items}}
                    <a
                      href="{{$.baseuri}}/list?term_tags={{$term_item.Title}}"
                      class="info-tag-item"
                      >{{$term_item.Title}}</a
                    >
                    {{end}}
                  </span>
                  {{end}} {{end}} {{end}}
                </div>

                <div class="hp-node-list-text">
                  {{FieldHtmlSubPrint $v "content" 100 $.LANG}}
                </div>
              </div>
              {{end}}
            </div>

            {{if .list_pager}}
            <nav class="">
              <ul class="pagination justify-content-center d-flex d-lg-none">
                {{if ge $.list_pager.PrevPageNumber 1}}
                <li
                  class="page-item {{if eq $.list_pager.PrevPageNumber $.list_pager.CurrentPageNumber}}disable{{end}}"
                >
                  <a
                    class="page-link"
                    href="{{$.baseuri}}/list?page={{.list_pager.PrevPageNumber}}"
                    >Prev</a
                  >
                </li>
                {{end}}
                <!-- -->
                {{if ge $.list_pager.CurrentPageNumber 1}}
                <li class="page-item">
                  <a class="page-link" href="#"
                    >{{$.list_pager.CurrentPageNumber}}</a
                  >
                </li>
                {{end}}
                <!-- -->
                {{if .list_pager.NextPageNumber}}
                <li
                  class="page-item {{if gt $.list_pager.NextPageNumber $.list_pager.PageCount}}disable{{end}}"
                >
                  <a
                    class="page-link"
                    href="{{$.baseuri}}/list?page={{.list_pager.NextPageNumber}}"
                    >Next</a
                  >
                </li>
                {{end}}
              </ul>
              <ul class="pagination justify-content-center d-none d-lg-flex">
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
                <li
                  class="page-item {{if eq $page $.list_pager.CurrentPageNumber}}active{{end}}"
                >
                  <a
                    class="page-link"
                    href='{{$.baseuri}}/list?{{FilterUri $ "page" $page}}'
                    >{{$page}}</a
                  >
                </li>
                {{end}}
                <!-- -->
                {{if gt $.list_pager.LastPageNumber 1}}
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
          </div>

          <div class="col d-none d-lg-block col-3 col-lg-3">
            {{pagelet . .modname "term/categories.tpl"}}
          </div>
        </div>
      </div>
    </div>

    {{pagelet . "core/general" "v3/footer.tpl"}}
    <!-- -->
    {{pagelet . "core/general" "html-footer.tpl"}}
  </body>
</html>
