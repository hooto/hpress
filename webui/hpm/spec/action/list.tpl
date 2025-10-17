<div>
  <div id="hpm-spec-actionls-alert"></div>

  <table class="table table-hover align-middle">
    <thead>
      <tr>
        <th>Name</th>
        <th>Datax</th>
        <th></th>
      </tr>
    </thead>
    <tbody id="hpm-spec-actionls"></tbody>
  </table>
</div>

<script id="hpm-spec-actionls-tpl" type="text/html">
  {[~it.actions :v]}
  <tr>
    <td>{[=v.name]}</td>
    <td>{[=v._dataxNum]}</td>
    <td align="right">
      <button
        class="btn btn-sm btn-outline-dark"
        onclick="hpSpec.ActionSet('{[=it._modname]}', '{[=v.name]}')"
      >
        Setting
      </button>
    </td>
  </tr>
  {[~]}
</script>
