<div class="">
  <div id="hpm-spec-termls-alert"></div>

  <table class="table table-hover align-middle">
    <thead>
      <tr>
        <th>Title</th>
        <th>Name</th>
        <th>Type</th>
        <th></th>
      </tr>
    </thead>
    <tbody id="hpm-spec-termls"></tbody>
  </table>
</div>

<script id="hpm-spec-termls-tpl" type="text/html">
  {[~it.termModels :v]}
  <tr>
    <td>{[=v.title]}</td>
    <td>{[=v.meta.name]}</td>
    <td>{[=v.type]}</td>
    <td align="right">
      <button
        class="btn btn-sm btn-outline-dark"
        onclick="hpSpec.TermSet('{[=it.meta.name]}', '{[=v.meta.name]}')"
      >
        Setting
      </button>
    </td>
  </tr>
  {[~]}
</script>
