<div class="">
  <div id="hpm-spec-viewls-alert"></div>

  <table class="table table-hover align-middle">
    <thead>
      <tr>
        <th>Template</th>
        <th></th>
      </tr>
    </thead>
    <tbody id="hpm-spec-viewls">
      {[~it.items :v]}
      <tr>
        <td><strong>{[=v.path]}</strong></td>
        <td align="right">
          <button
            class="btn btn-sm btn-outline-dark"
            onclick="hpSpec.RouteSetTemplateSelectOne('{[=v.path]}')"
          >
            Select
          </button>
        </td>
      </tr>
      {[~]}
    </tbody>
  </table>
</div>

<script id="hpm-spec-viewls-tpl" type="text/html"></script>
