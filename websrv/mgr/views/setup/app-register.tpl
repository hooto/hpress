<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <title>Setup : Registration to IAM Service</title>
    <link rel="shortcut icon" type="image/x-icon" href='{{HttpSrvBasePath "hp/~/hp/img/ap.ico"}}' />
    <script src='{{HttpSrvBasePath "hp/lynkui/~/lynkui/main.js"}}'></script>
    <script type="text/javascript">
      lynkui.basepath = '{{HttpSrvBasePath "hp/lynkui"}}';
      lynkui.uipath = "~";
      window.onload = lynkui.onload();
    </script>
  </head>

  <body>
    <div class="container" style="width: 600px; margin: 20px auto">
      <div class="hp-mgr-setup-logo" style="text-align: center; padding: 30px">
        <img src='{{HttpSrvBasePath "hp/~/hp/img/alpha2.png"}}' />
      </div>

      <div class="card">
        <div class="card-body">
          <h3 class="card-title">Setup</h3>
          <p class="card-text">
            Register Application to IAM (Identity &amp; Access Management) Service
          </p>
        </div>

        <div class="card-body">
          <form id="hp-app-reg" action="#">
            <div class="mb-3">
              <label class="form-label">IAM Service URL</label>
              <input
                type="text"
                name="iam_url"
                class="form-control"
                placeholder="Enter the IAM Service URL"
                value="{{.iam_url}}"
                readonly
              />
            </div>

            <!--
            <div class="mb-3">
              <label>Instance ID</label>
              <input
                type="text"
                name="instance_id"
                class="form-control"
                value="{{.instance_id}}"
                readonly
              />
            </div>
			-->

            <div class="mb-3">
              <label class="form-label">Instance Frontend URL</label>
              <input
                type="text"
                name="instance_url"
                class="form-control"
                value="{{.instance_url}}"
              />
            </div>

            <div class="mb-3">
              <label class="form-label">Application ID</label>
              <input type="text" name="app_id" class="form-control" value="{{.app_id}}" readonly />
            </div>

            <div class="mb-3">
              <label class="form-label">Application Name</label>
              <input
                type="text"
                name="app_title"
                class="form-control"
                placeholder="Enter the name of application"
                value="{{.app_title}}"
              />
            </div>

            <div class="mb-3">
              <label class="form-label">Application Version</label>
              <input
                type="text"
                name="version"
                class="form-control"
                value="{{.version}}"
                readonly
              />
            </div>

            <div id="hp-app-reg-alert" class="alert alert-danger d-none mb-3">...</div>

            <div class="form-group2">
              <button
                id="app_register_btn"
                type="button"
                class="btn btn-success btn-block"
                onclick="_appRegisterCommit()"
              >
                Commit
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </body>
</html>

<script type="text/javascript">
  function _appRegisterCommit() {
    const btn = document.getElementById("app_register_btn");
    btn.disabled = true;
    window.setTimeout(function () {
      btn.disabled = false;
    }, 1000);

    let params = $("#hp-app-reg").serialize();
    let alertid = "#hp-app-reg-alert";

    $.ajax({
      type: "POST",
      url: '{{HttpSrvBasePath "hp/mgr/setup/app-register-sync"}}',
      data: params,
      timeout: 3000,
      success: function (data) {
        if (!data || data.kind != "AppInstanceRegister") {
          if (data.error) {
            lynkui.alert.innerShow(alertid, "error", data.error.message);
          } else {
            lynkui.alert.innerShow(alertid, "error", "Network Exception");
          }
        } else {
          lynkui.alert.innerShow(alertid, "ok", "Successfully registered ...");

          window.setTimeout(function () {
            window.location = '{{HttpSrvBasePath "hp/mgr"}}';
          }, 1500);
        }
      },
      error: function (xhr, textStatus, error) {
        lynkui.alert.innerShow(alertid, "error", textStatus + " " + xhr.responseText);
      },
    });
  }
</script>
