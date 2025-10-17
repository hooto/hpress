<!doctype html>
<html lang="en">
  {{pagelet . "core/general" "v3/html-header.tpl"}}
  <body id="hp-body">
    {{pagelet . "core/general" "v3/nav-header.tpl" "topnav"}}

    <div class="container hp-block-gap-column">
      <!-- sub-nav/ -->
      <div class="hp-block-gap">
        <div class="col">
          <div class="hp-ctn-title">Content Entry</div>
        </div>
      </div>
      <!-- /sub-nav -->

      <div class="hp-block-gap-column">
        <div class="hp-node-view">
          <div class="hp-header">{{FieldStringPrint .entry "title" .LANG}}</div>

          <div class="hp-info">
            <span class="info-item"
              >Published: {{UnixtimeFormat .entry.Created "Y-m-d"}}
            </span>

            {{range $term := .entry.Terms}} {{if eq $term.Name "categories"}}
            {{if $term.Items}}
            <span class="info-item">
              Categories: {{range $term_item := $term.Items}}
              <a
                href='{{$.baseuri}}/list?term_categories={{printf "%d" $term_item.ID}}'
                >{{$term_item.Title}}</a
              >
              {{end}}
            </span>
            {{end}} {{end}} {{end}} {{range $term := .entry.Terms}} {{if eq
            $term.Name "tags"}} {{if $term.Items}}
            <span class="info-item">
              Tags: {{range $term_item := $term.Items}}
              <a
                href="{{$.baseuri}}/list?term_tags={{$term_item.Title}}"
                class="tag-item"
                >{{$term_item.Title}}</a
              >
              {{end}}
            </span>
            {{end}} {{end}} {{end}}
          </div>

          <div class="content hp-content">
            {{FieldHtmlPrint .entry "content" .LANG}}
          </div>
        </div>

        {{if .entry.ExtCommentEnable}}
        <div id="core-blog-comments" class="">comments loading</div>
        {{end}}
      </div>
    </div>

    {{pagelet . "core/general" "v3/footer.tpl"}}

    <script type="text/javascript">
      {{if .entry.ExtCommentEnable}}
      window.onload_hooks.push(function() {
          seajs.use([
              "+/comment/~/index.js",
              "+/comment/~/index.css",
          ],
          function() {
              hpComment.EmbedLoader("core-blog-comments", "{{.modname}}", "{{.__datax_table__}}", "{{.entry.ID}}");
          });
      });
      {{end}}

      window.onload_hooks.push(function() {
          hp.CodeRender({"theme": "monokai"});
      });
    </script>

    {{pagelet . "core/general" "html-footer.tpl"}}
  </body>
</html>
