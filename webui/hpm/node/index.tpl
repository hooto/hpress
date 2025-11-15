<div class="hpm-block-gap-column">
  <div
    id="hpm-node-navbar"
    class="d-flex flex-row justify-content-between hpm-block-gap-row"
  >
    <div
      id="hpm-node-optools"
      class="d-flex flex-row align-self-center hpm-block-gap-row-sm"
    ></div>
    <div class="d-flex flex-row hpm-block-gap-row-sm">
      <div id="hpm-node-nmodels" class="hpm-block-gap-row-sm"></div>
      <div id="hpm-node-tmodels" class="hpm-block-gap-row-sm"></div>
    </div>
  </div>

  <div id="hpm-node-workspace" class="hpm-block-gap-column">
    <div id="hpm-node-alert" class=""></div>
    <div id="work-content" class=""></div>
  </div>
</div>

<script id="hpm-node-nmodels-tpl" type="text/html">
  {[~it.items :v]} {[if (!v.extensions.node_refer) {]}
  <div
    class="node-item btn btn-outline-dark {[? hpMgr.Equal(it.active,v.meta.name)]} active{[?]}"
    tgname="{[=v.meta.name]}"
    href="#{[=v.meta.name]}"
  >
    {[=v.title]}
  </div>
  {[}]} {[~]}
</script>

<script id="hpm-node-tmodels-tpl" type="text/html">
  {[~it.items :v]}
  <div
    class="term-item btn btn-outline-dark {[? hpMgr.Equal(it.active,v.meta.name)]} active{[?]}"
    tgname="{[=v.meta.name]}"
    href="#{[=v.meta.name]}"
  >
    {[=v.title]}
  </div>
  {[~]}
</script>

<script type="text/javascript">
  $("#hpm-node-nmodels").on("click", ".node-item", function () {
    $("#hpm-node-nmodels").find(".active").removeClass("active");
    $("#hpm-node-tmodels").find(".active").removeClass("active");
    $(this).addClass("active");

    hpNode.SpecNodeModelActive($(this).attr("tgname"));
    lynkui.storage.del("hpm_nodels_page");
    lynkui.storage.del("hpm_termls_page");

    hpNode.List(null, $(this).attr("tgname"));
  });

  $("#hpm-node-tmodels").on("click", ".term-item", function () {
    $("#hpm-node-nmodels").find(".active").removeClass("active");
    $("#hpm-node-tmodels").find(".active").removeClass("active");
    $(this).addClass("active");

    hpTerm.SpecTermModelActive($(this).attr("tgname"));
    lynkui.storage.del("hpm_nodels_page");
    lynkui.storage.del("hpm_termls_page");

    hpTerm.List(null, $(this).attr("tgname"));
  });
</script>
