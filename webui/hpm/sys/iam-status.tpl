<style>
  .hp-sys-table {
    font-size: 14px;
  }
  .hp-sys-table td {
    padding: 5px !important;
  }
  .hp-sys-table tr.line {
    border-top: 1px solid #ccc;
  }
</style>

<div class="card">
  <div class="card-header">IAM Service Status</div>
  <div class="card-body">
    <table
      width="100%"
      class="table table-hover"
      style="margin-bottom: 20px; padding: 2px 0"
    >
      <tr>
        <td width="20%">IAM Service Url</td>
        <td>{[=it.service_url]}</td>
      </tr>
      {[if (it.service_url_frontend && it.service_url_frontend.length > 1) {]}
      <tr>
        <td width="20%">IAM Service Url Frontend</td>
        <td>{[=it.service_url_frontend]}</td>
      </tr>
      {[}]}
    </table>

    <div id="hp-mgr-sys-iam-alert"></div>

    <form id="hp-mgr-sys-iam" action="#">
      <table width="100%" class="table table-hover hp-sys-table">
        <thead>
          <tr>
            <th width="20%"></th>
            <th width="40%"><strong>Local App Info</strong></th>
            <th><strong>Registered in IAM Service</strong></th>
          </tr>
        </thead>

        <tr>
          <td>App Version</td>
          <td>{[=it.instance_self.version]}</td>
          <td>{[=it.instance_registered.version]}</td>
        </tr>

        <tr>
          <td>App ID</td>
          <td>{[=it.instance_self.app_id]}</td>
          <td>{[=it.instance_registered.app_id]}</td>
        </tr>

        <tr>
          <td>App Name</td>
          <td>
            <input
              type="text"
              name="app_title"
              class="form-control input-sm"
              placeholder="Enter the App Name"
              value="{[=it.instance_self.app_title]}"
            />
          </td>
          <td>{[=it.instance_registered.app_title]}</td>
        </tr>

        <tr>
          <td>Entry URL</td>
          <td>
            <input
              type="text"
              name="instance_url"
              class="form-control input-sm"
              placeholder="Enter the Entry URL of App Instance"
              value="{[=it.instance_self.url]}"
            />
          </td>
          <td>{[=it.instance_registered.url]}</td>
        </tr>

        <tr>
          <td>Privileges</td>
          <td>
            <table class="table">
              <thead>
                <tr>
                  <th>Privilege</th>
                  <th>Roles</th>
                </tr>
              </thead>
              <tbody>
                {[~it.instance_self.privileges :v]}
                <tr>
                  <td>
                    <p><strong>{[=v.desc]}</strong></p>
                    <p>{[=v.privilege]}</p>
                  </td>
                  <td>
                    {[ if (v.roles.length > 0) { ]}
                    <!-- -->
                    {[~v.roles :rv]}
                    <!-- -->
                    {[~it._roles.items :drv]}
                    <!-- -->
                    {[? hpMgr.Equal(rv,drv.idxid)]}
                    <p>{[=drv.meta.name]}</p>
                    {[?]}
                    <!-- -->
                    {[~]}
                    <!-- -->
                    {[~]}
                    <!-- -->
                    {[ } else {]}
                    <p>Owner</p>
                    {[ } ]}
                  </td>
                </tr>
                {[~]}
              </tbody>
            </table>
          </td>
          <td>
            <table class="table">
              <thead>
                <tr>
                  <th>Privilege</th>
                  <th>Roles</th>
                </tr>
              </thead>
              <tbody>
                {[~it.instance_registered.privileges :v]}
                <tr>
                  <td>
                    <p><strong>{[=v.desc]}</strong></p>
                    <p>{[=v.privilege]}</p>
                  </td>
                  <td>
                    {[ if (v.roles.length > 0) { ]}
                    <!-- -->
                    {[~v.roles :rv]}
                    <!-- -->
                    {[~it._roles.items :drv]}
                    <!-- -->
                    {[? hpMgr.Equal(rv,drv.idxid)]}
                    <p>{[=drv.meta.name]}</p>
                    {[?]}
                    <!-- -->
                    {[~]}
                    <!-- -->
                    {[~]}
                    <!-- -->
                    {[ } else {]}
                    <p>Owner</p>
                    {[ } ]}
                  </td>
                </tr>
                {[~]}
              </tbody>
            </table>
          </td>
        </tr>
      </table>
    </form>

    <div class="text-center">
      <button type="submit" class="btn btn-primary" onclick="hpSys.IamSync()">
        Sync to IAM Service
      </button>
    </div>
  </div>
</div>
