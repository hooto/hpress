<!doctype html>
<html lang="en">
  {{pagelet . "core/general" "v3/html-header.tpl"}}
  <body id="hp-body">
    {{pagelet . "core/general" "v3/nav-header.tpl" "topnav"}}

    <div class="container">
      <div class="hp-block-gap-column">
        <div class="">
          <div class="hp-ctn-title">
            {{FieldStringPrint .page "title" .LANG}}
          </div>
        </div>

        <div class="">
          <div class="hp-node-view">
            <div class="content hp-content">
              {{FieldHtmlPrint .page "content" .LANG}}
            </div>
          </div>
        </div>
      </div>
    </div>

    {{pagelet . "core/general" "v3/footer.tpl"}}
    <!-- -->
    {{pagelet . "core/general" "html-footer.tpl"}}
  </body>
</html>
