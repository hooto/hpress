<div class="">
  <div id="hpm-spec-nodels-alert"></div>

  <table class="table table-hover align-middle">
    <thead>
      <tr>
        <th>Name</th>
        <th>Title</th>
        <th>Fields</th>
        <th>Terms</th>
        <th></th>
      </tr>
    </thead>
    <tbody id="hpm-spec-nodels"></tbody>
  </table>
</div>

<script id="hpm-spec-nodels-tpl" type="text/html">
  {[~it.nodeModels :v]}
  <tr>
    <td>{[=v.meta.name]}</td>
    <td>{[=v.title]}</td>
    <td>{[=v._fieldsNum]}</td>
    <td>{[=v._termsNum]}</td>
    <td align="right">
      <button
        class="btn btn-outline-dark btn-sm"
        onclick="hpSpec.NodeSet('{[=it.meta.name]}', '{[=v.meta.name]}')"
      >
        Setting
      </button>
    </td>
  </tr>
  {[~]}
</script>
