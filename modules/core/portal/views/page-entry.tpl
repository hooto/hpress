<!DOCTYPE html>
<html lang="en">
  {{pagelet . "core/portal" "v3/html-header-base.tpl"}}
  <body>
    {{pagelet . "core/general" "nav-header.tpl" "topnav"}}

    <div class="container">
      <div class="hp-ctn-header">
        <h2>{{FieldStringPrint .page_entry "title" .LANG}}</h2>
      </div>

      <div class="row">
        <div class="col-md-12">
          <div class="hp-nodev">
            <div class="content">{{FieldHtmlPrint .page_entry "content" .LANG}}</div>
          </div>
        </div>
      </div>
    </div>

    {{pagelet . "core/general" "footer.tpl"}} {{pagelet . "core/general" "html-footer.tpl"}}
  </body>
</html>
