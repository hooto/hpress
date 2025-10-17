<div class="hpm-block-gap-column">
  <div id="hpm-nodels" class="hpm-table-std"></div>
  <div id="hpm-nodels-pager" class=""></div>
</div>

<script id="hpm-node-list-opts" type="text/html">
  <div
    class="btn btn-primary"
    id="hpm-node-list-refer-back"
    style="display: none"
    onclick="hpNode.ReferBack()"
  >
    Back
  </div>
  <div
    class="btn btn-primary"
    onclick="hpNode.Set()"
    id="hpm-node-list-new-title"
  >
    New Content
  </div>
  <div class="">
    <form onsubmit="hpNode.List(); return false;" action="#" class="">
      <input
        id="qry_text"
        type="text"
        class="form-control hpm-query-input"
        placeholder="Press Enter to Search"
        value=""
      />
    </form>
  </div>
  <div
    id="hpm-nodels-batch-select-todo-btn"
    class="btn btn-outline-primary"
    style="display: none"
    onclick="hpNode.ListBatchSelectTodo()"
  >
    Select Contents todo ...
  </div>
</script>

<script id="hpm-nodels-tpl" type="text/html">
  <table class="table table-hover align-middle" style="margin: 0">
    <thead>
      <tr>
        <th width="20">
          <input
            class="row-checkbox hpm-nodels-chk-all"
            type="checkbox"
            onclick="hpNode.ListBatchSelectAll()"
          />
        </th>
        <th scope="col">Title</th>
        {[if (it.model.extensions.node_sub_refer) {]}
        <th></th>
        {[}]}
        <th>Status</th>
        {[if (it.model.extensions.access_counter) { ]}
        <th>Access</th>
        {[}]}
        <th>Created</th>
        <th>Updated</th>
        <th></th>
      </tr>
    </thead>
    <tbody>
      {[~it.items :v]}
      <tr>
        <td>
          <input
            class="row-checkbox hpm-nodels-chk-item"
            type="checkbox"
            value="{[=v.id]}"
            onclick="hpNode.ListBatchSelectTodoBtnRefresh()"
          />
        </td>
        <td>
          <a
            class="node-item hpm-no-underline"
            onclick="hpNode.Set('{[=it.modname]}', '{[=it.modelid]}', '{[=v.id]}')"
            href="#{[=v.id]}"
            >{[=v.title]}</a
          >
        </td>
        {[if (it.model.extensions.node_sub_refer) {]}
        <td>
          <!--<button class="btn btn-sm" onclick="hpNode.Set('{[=it.modname]}', '{[=it.model.extensions.node_sub_refer]}', null, '{[=v.id]}')">New Sub Content</button>-->
          <button
            class="btn btn-sm btn-outline-dark"
            onclick="hpNode.List('{[=it.modname]}', '{[=it.model.extensions.node_sub_refer]}', '{[=v.id]}')"
          >
            Sub Contents
          </button>
        </td>
        {[}]}
        <td>
          {[~it._status_def :sv]} {[if (sv.type == v.status) { ]}{[=sv.name]}{[
          } ]} {[~]}
        </td>
        {[if (it.model.extensions.access_counter) {]}
        <td>{[=v.ext_access_counter]}</td>
        {[}]}
        <td>{[=v.created]}</td>
        <td>{[=v.updated]}</td>
        <td align="right">
          <button
            class="btn btn-sm btn-outline-dark"
            onclick="hpNode.Del('{[=it.modname]}', '{[=it.modelid]}', '{[=v.id]}')"
          >
            Delete
          </button>
          <button
            class="btn btn-sm btn-outline-dark"
            onclick="hpNode.Set('{[=it.modname]}', '{[=it.modelid]}', '{[=v.id]}')"
          >
            Edit
          </button>
        </td>
      </tr>
      {[~]}
    </tbody>
  </table>
</script>

<script id="hpm-nodels-pager-tpl" type="text/html">
  {[ if (it.RangePages.length > 1) { ]}
  <nav>
    <ul class="pagination pagination-sm justify-content-center">
      {[ if (it.FirstPageNumber > 0) { ]}
      <li class="page-item">
        <a
          class="page-link"
          href="#{[=it.FirstPageNumber]}"
          onclick="hpNode.ListPage({[=it.FirstPageNumber]})"
          >First</a
        >
      </li>
      {[ } ]} {[~it.RangePages :v]}
      <li
        class="page-item {[ if (v== it.CurrentPageNumber) { ]} active {[ } ]}"
      >
        <a class="page-link" href="#{[=v]}" onclick="hpNode.ListPage({[=v]})"
          >{[=v]}</a
        >
      </li>
      {[~]} {[ if (it.LastPageNumber > 0) { ]}
      <li class="page-item">
        <a
          class="page-link"
          href="#{[=it.LastPageNumber]}"
          onclick="hpNode.ListPage({[=it.LastPageNumber]})"
          >Last</a
        >
      </li>
      {[ } ]}
    </ul>
    {[ } ]}
  </nav>
</script>
