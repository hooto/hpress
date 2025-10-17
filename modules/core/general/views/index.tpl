<!DOCTYPE html>
<html lang="en">
  {{pagelet . "core/general" "v3/html-header-base.tpl"}}
  <body>
    {{pagelet . "core/general" "v3/nav-header.tpl" "topnav"}}

    <div class="container">
      <div class="hp-ctn-header">
        <h2>Hello, world</h2>
      </div>

      <div class="jumbotron">
        <p>
          This is a simple hero unit, a simple jumbotron-style component for
          calling extra attention to featured content or information.
        </p>
        <p>
          <a class="btn btn-primary btn-lg" href="#" role="button"
            >Learn more</a
          >
        </p>
      </div>
    </div>

    {{pagelet . "core/general" "v3/footer.tpl"}} {{pagelet . "core/general"
    "v3/html-footer.tpl"}}
  </body>
</html>
