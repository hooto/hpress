<div id="hpm-s2-objls-navbar">
  <ul id="hpm-s2-objls-dirnav" class="hpm-breadcrumb"></ul>
</div>

<table class="table table-hover">
  <thead>
    <tr>
      <th width="64px"></th>
      <th>Name</th>
      <th style="text-align: right">Size</th>
      <th></th>
      <th></th>
    </tr>
  </thead>
  <tbody id="hpm-s2-objls"></tbody>
</table>

<script id="hpm-s2-objls-dirnav-tpl" type="text/html">
  {[~it.items :v]}
  <li>
    <a href="#{[=v.path]}" onclick="hpS2.ObjList('{[=v.path]}')">{[=v.name]}</a>
  </li>
  {[~]}
</script>

<script id="hpm-s2-objls-tpl" type="text/html">
  {[~it.items :v]}
  <tr id="obj{[=v._id]}">
    <td>
      {[ if (v.isdir) { ]}
      <svg class="bi" width="16" height="16" fill="currentColor">
        <use xlink:href="/hp/lynkui/~/bi/v1/bootstrap-icons.svg#folder" />
      </svg>
      {[ } else if (v._isimg) { ]}
      <a href="{[=v.self_link]}" target="_blank"
        ><img src="{[=v.self_link]}?ipl=w64,h64,c" width="64" height="64"
      /></a>
      {[ } ]}
    </td>
    <td class="ts3-fontmono">
      {[ if (v.isdir) { ]}
      <a class="obj-item-dir" href="#objs" path="{[=v._abspath]}"
        >{[=v.name]}</a
      >
      {[ } else { ]}
      <a class="obj-item-file" href="{[=v.self_link]}" target="_blank"
        >{[=v.name]}</a
      >
      {[ } ]}
    </td>
    <td align="right">
      {[?!v.isdir]} {[=hpS2.UtilResourceSizeFormat(v.size)]} {[?]}
    </td>
    <td align="right">{[=lynkui.utilx.timeParseFormat(v.modtime, "Y-m-d H:i:s")]}</td>
    <td align="right">
      {[ if (!v.isdir) { ]}
      <button
        class="pure-button button-xsmall"
        onclick="hpS2.ObjListSelectorEntry('{[=v._abspath]}')"
      >
        Select
      </button>
      {[ } ]}
    </td>
  </tr>
  {[~]}
</script>

<script type="text/javascript">
  $("#hpm-s2-objls").on("click", ".obj-item-dir", function () {
    hpS2.ObjList($(this).attr("path"));
  });
  $("#hpS2-object-dirnav").on("click", ".obj-item-dir", function () {
    hpS2.ObjList($(this).attr("path"));
  });
</script>
