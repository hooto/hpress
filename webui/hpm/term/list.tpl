<div class="hpm-block-gap-column">
  <div id="hpm-termls" class="hpm-table-std"></div>
  <div id="hpm-termls-pager" class=""></div>
</div>

<script id="hpm-node-term-opts" type="text/html">
    <div class="btn btn-primary"
       onclick="hpTerm.Set()" id="hpm-term-list-new-title">
        New Term
    </div>
    <div >
      <form onsubmit="hpTerm.List(); return false;" action="#" class="">
        <input id="qry_text" type="text"
          class="form-control hpm-query-input"
          placeholder="Press Enter to Search"
          value="">
      </form>
    </div>
  </div>
</script>

<script id="hpm-termls-tpl" type="text/html">
  <table class="table table-hover align-middle" style="margin: 0">
    <thead>
      <tr>
        <th width="80px">ID</th>
        <th>Title</th>
        {[? hpMgr.Equal(it.model.type,"taxonomy")]}
        <th>Weight</th>
        {[?]}
        <th>Created</th>
        <th>Updated</th>
        <th></th>
      </tr>
    </thead>
    <tbody>
      {[~it.items :v]} {[? hpMgr.Equal(v.pid,0)]}
      <tr>
        <td>{[=v.id]}</td>
        <td>{[=v.title]}</td>
        {[? hpMgr.Equal(it.model.type,"taxonomy")]}
        <td>{[=v.weight]}</td>
        {[?]}
        <td>{[=v.created]}</td>
        <td>{[=v.updated]}</td>
        <td align="right">
          <button
            class="btn btn-sm btn-outline-dark"
            onclick="hpTerm.Set('{[=it.modname]}', '{[=it.modelid]}', '{[=v.id]}')"
          >
            Edit
          </button>
        </td>
      </tr>
      {[? v._subs]} {[~v._subs :v2]}
      <tr>
        <td>{[=v2.id]}</td>
        <td>{[=hpTerm.Sprint(v2._dp)]}{[=v2.title]}</td>
        <td>{[=hpTerm.Sprint(v2._dp)]}{[=v2.weight]}</td>
        <td>{[=v2.created]}</td>
        <td>{[=v2.updated]}</td>
        <td align="right">
          <button
            class="btn btn-sm btn-outline-dark"
            onclick="hpTerm.Set('{[=it.modname]}', '{[=it.modelid]}', '{[=v2.id]}')"
          >
            Edit
          </button>
        </td>
      </tr>
      {[~]} {[?]} {[?]} {[~]}
    </tbody>
  </table>
</script>

<script id="hpm-termls-pager-tpl" type="text/html">
  {[ if (it.RangePages.length > 1) { ]}
  <nav>
    <ul class="pagination pagination-sm justify-content-center">
      {[ if (it.FirstPageNumber > 0) { ]}
      <li class="page-item">
        <a
          class="page-link"
          href="#{[=it.FirstPageNumber]}"
          onclick="hpTerm.ListPage({[=it.FirstPageNumber]})"
          >First</a
        >
      </li>
      {[ } ]} {[~it.RangePages :v]}
      <li
        class="page-item {[? hpMgr.Equal(v,it.CurrentPageNumber)]}active{[?]}"
      >
        <a class="page-link" href="#{[=v]}" onclick="hpTerm.ListPage({[=v]})"
          >{[=v]}</a
        >
      </li>
      {[~]} {[ if (it.LastPageNumber > 0) { ]}
      <li class="page-item">
        <a
          class="page-link"
          href="#{[=it.LastPageNumber]}"
          onclick="hpTerm.ListPage({[=it.LastPageNumber]})"
          >Last</a
        >
      </li>
      {[ } ]}
    </ul>
    {[ } ]}
  </nav>
</script>

<script type="text/javascript">
  $("#hpm-termls").on("click", ".term-item", function () {
    var id = $(this).attr("href").substr(1);
    hpTerm.Set($(this).attr("modname"), $(this).attr("modelid"), id);
  });
</script>
