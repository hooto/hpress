<div class="hpm-block-gap-column">
  <div id="hpm-specls-alert"></div>

  <div class="hpm-table-std">
    <table class="table table-hover align-middle">
      <thead>
        <tr>
          <th>Title</th>
          <th>Name</th>
          <th>Service Name</th>
          <th>Version</th>
          <th>Nodes</th>
          <th>Actions</th>
          <th>Routes</th>
          <th>Status</th>
          <th></th>
        </tr>
      </thead>
      <tbody id="hpm-specls"></tbody>
    </table>
  </div>
</div>

<script id="hpm-specls-tpl" type="text/html">
  {[~it.items :v]}
  <tr>
    <td>{[=v.title]}</td>
    <td>{[=v.meta.name]}</td>
    <td>{[=v.srvname]}</td>
    <td>{[=v.meta.version]}</td>
    <td>
      <button
        class="btn btn-outline-dark btn-sm"
        onclick="hpSpec.NodeList('{[=v.meta.name]}')"
      >
        {[=v._nodeModelsNum]}
      </button>
    </td>
    <td>
      <button
        class="btn btn-outline-dark btn-sm"
        onclick="hpSpec.ActionList('{[=v.meta.name]}')"
      >
        {[=v._actionsNum]}
      </button>
    </td>
    <td>
      <button
        class="btn btn-outline-dark btn-sm"
        onclick="hpSpec.RouteList('{[=v.meta.name]}')"
      >
        {[=v._routesNum]}
      </button>
    </td>
    <td>
      {[if (v.status) {]}
      <span class="badge text-bg-success">Enable</span>
      {[} else {]}
      <span class="badge text-bg-secondary">Disable</span>
      {[}]}
    </td>
    <td align="right">
      <button
        class="btn btn-outline-dark btn-sm"
        onclick="hpSpecEditor.Open('{[=v.meta.name]}')"
      >
        Develop
      </button>
      <button
        class="btn btn-outline-dark btn-sm"
        onclick="hpSpec.InfoSet('{[=v.meta.name]}')"
      >
        Setting
      </button>
    </td>
  </tr>
  {[~]}
</script>
