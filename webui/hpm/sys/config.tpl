<style>
  .hp-sys-table {
    font-size: 12px;
  }
  .hp-sys-table td {
    padding: 5px !important;
  }
  .hp-sys-table tr.line {
    border-top: 1px solid #ccc;
  }
</style>

<div class="card">
  <div class="card-header">
    <div>System Config</div>
  </div>

  <div id="hpm-sys-configset" class="card-body">
    <div id="hpm-sys-configset-alert"></div>

    <table width="100%" class="table table-striped">
      <thead>
        <tr>
          <th>Key</th>
          <th>Value</th>
          <th>Comment</th>
        </tr>
      </thead>
      <tbody>
        {[~it.items :v]}
        <tr>
          <td width="20%">{[=v.key]}</td>
          <td width="40%">
            {[ if (v.type !== undefined && v.type == "text") { ]}
            <textarea
              class="form-control hpm-sys-config-item"
              name="{[=v.key]}"
              rows="3"
            >
{[=v.value]}</textarea
            >
            {[ } else { ]}
            <input
              type="text"
              class="form-control hpm-sys-config-item"
              name="{[=v.key]}"
              value="{[=v.value]}"
            />
            {[ } ]}
          </td>
          <td>{[=v.comment]}</td>
        </tr>
        {[~]}
      </tbody>
    </table>

    <button class="btn btn-primary" onclick="hpSys.ConfigSetCommit()">
      Save
    </button>
  </div>
</div>
