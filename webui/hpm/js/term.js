// Copyright 2015 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

var hpTerm = {
  taxonomy_ls_cache: null,
};

hpTerm.SpecActive = function () {
  return l4iStorage.Get("hpm_spec_active");
};

hpTerm.SpecTermModelActive = function (value) {
  if (!hpTerm.SpecActive()) {
    return null;
  }
  var k = "hpm_stm_" + hpTerm.SpecActive();
  if (value && value.length > 1) {
    l4iStorage.Set(k, value);
  }
  return l4iStorage.Get(k);
};

hpTerm.List = function (modname, modelid) {
  var alertid = "#hpm-node-alert",
    page = 0;

  if (!modname && hpTerm.SpecActive()) {
    modname = hpTerm.SpecActive();
  }
  if (!modelid && hpTerm.SpecTermModelActive()) {
    modelid = hpTerm.SpecTermModelActive();
  }
  if (l4iStorage.Get("hpm_termls_page")) {
    page = l4iStorage.Get("hpm_termls_page");
  }

  if (!modname || !modelid) {
    return;
  }

  var uri = "modname=" + modname + "&modelid=" + modelid + "&page=" + page;
  if (document.getElementById("qry_text")) {
    uri += "&qry_text=" + $("#qry_text").val();
  }

  seajs.use(["ep"], function (EventProxy) {
    var ep = EventProxy.create("tpl", "data", function (tpl, rsj) {
      if (tpl) {
        $("#work-content").html(tpl);
      }

      hpTerm.SpecTermModelActive(modelid);

      if (
        !rsj ||
        rsj.kind != "TermList" ||
        !rsj.items ||
        rsj.items.length < 1
      ) {
        $("#hpm-nodels").empty();
        $("#hpm-termls").empty();

        l4i.InnerAlert(alertid, "alert-info", "Item Not Found");
      } else {
        $(alertid).hide();
      }
      $("#hpm-term-list-new-title").text("New " + rsj.model.title);

      if (!rsj.items) {
        rsj.items = [];
      }

      for (var i in rsj.items) {
        rsj.items[i].created = l4i.UnixTimeFormat(
          rsj.items[i].created,
          "Y-m-d"
        );
        rsj.items[i].updated = l4i.UnixTimeFormat(
          rsj.items[i].updated,
          "Y-m-d H:i:s"
        );

        if (!rsj.items[i].weight) {
          rsj.items[i].weight = 0;
        }

        if (!rsj.items[i].pid) {
          rsj.items[i].pid = 0;
        }

        if (rsj.model.type == "taxonomy" && rsj.items[i].pid == 0) {
          rsj.items[i]._subs = hpTerm.ListSubRange(
            rsj.items,
            null,
            rsj.items[i].id,
            0
          );
        }
      }

      hpTerm.taxonomy_ls_cache = rsj;

      l4iTemplate.Render({
        dstid: "hpm-termls",
        tplid: "hpm-termls-tpl",
        data: {
          model: rsj.model,
          modname: modname,
          modelid: modelid,
          items: rsj.items,
        },
        success: function () {
          if (rsj.model.type != "taxonomy") {
            rsj.meta.RangeLen = 20;

            l4iTemplate.Render({
              dstid: "hpm-termls-pager",
              tplid: "hpm-termls-pager-tpl",
              data: l4i.Pager(rsj.meta),
            });
          } else {
            $("#hpm-termls-pager").empty();
          }

          hpNode.OpToolsRefresh("#hpm-node-term-opts");
        },
      });
    });

    ep.fail(function (err) {
      // TODO
      alert("SpecListRefresh error, Please try again later (EC:app-termlist)");
    });

    // template
    var el = document.getElementById("hpm-termls");
    if (!el || el.length < 1) {
      hpMgr.TplCmd("term/list", {
        callback: function (err, tpl) {
          if (err) {
            return ep.emit("error", err);
          }

          ep.emit("tpl", tpl);
        },
      });
    } else {
      ep.emit("tpl", null);
    }

    hpMgr.ApiCmd("term/list?" + uri, {
      callback: ep.done("data"),
    });
  });
};

hpTerm.Sprint = function (num) {
  var s = "";
  for (i = 0; i < num; i++) {
    s += "&nbsp;&nbsp;&nbsp;&nbsp;";
  }

  return s;
};

hpTerm.ListSubRange = function (ls, rs, pid, dpnum) {
  if (!rs) {
    rs = [];
  }

  dpnum++;

  for (var i in rs) {
    if (rs[i].id == pid) {
      // return rs;
    }
  }

  for (var i in ls) {
    if (ls[i].pid == pid) {
      ls[i]._dp = dpnum;
      rs.push(ls[i]);
      rs = hpTerm.ListSubRange(ls, rs, ls[i].id, dpnum);
    }
  }

  return rs;
};

hpTerm.ListPage = function (page) {
  l4iStorage.Set("hpm_termls_page", parseInt(page));
  hpTerm.List();
};

hpTerm.Set = function (modname, modelid, termid) {
  var alertid = "#hpm-node-alert";

  if (!modname && hpTerm.SpecActive()) {
    modname = hpTerm.SpecActive();
  }
  if (!modelid && hpTerm.SpecTermModelActive()) {
    modelid = hpTerm.SpecTermModelActive();
  }

  if (!modname || !modelid) {
    return;
  }

  var uri = "modname=" + modname + "&modelid=" + modelid;

  seajs.use(["ep"], function (EventProxy) {
    var ep = EventProxy.create("tpl", "data", function (tpl, data) {
      if (!tpl) {
        return; // TODO
      }

      $("#work-content").html(tpl);

      if (!data || data.kind != "Term") {
        return l4i.InnerAlert(alertid, "alert-info", "Item Not Found");
      }

      if (!data.status) {
        data.status = 1;
      }
      if (!data.weight) {
        data.weight = 0;
      }
      if (!data.pid) {
        data.pid = 0;
      }

      data._taxonomy_ls = hpTerm.taxonomy_ls_cache;

      $(alertid).hide();
      hpNode.OpToolsRefresh("clean");

      l4iTemplate.Render({
        dstid: "hpm-termset",
        tplid: "hpm-termset-tpl",
        data: data,
        success: function () {},
      });
    });

    ep.fail(function (err) {
      // TODO
      alert("SpecListRefresh error, Please try again later (EC:app-nodelist)");
    });

    hpMgr.TplCmd("term/set", {
      callback: function (err, tpl) {
        if (err) {
          return ep.emit("error", err);
        }
        ep.emit("tpl", tpl);
      },
    });

    if (termid) {
      hpMgr.ApiCmd("term/entry?" + uri + "&id=" + termid, {
        callback: ep.done("data"),
      });
    } else {
      hpMgr.ApiCmd("term-model/entry?" + uri, {
        callback: function (err, data) {
          ep.emit("data", {
            kind: "Term",
            model: data,
            id: "0",
            pid: "0",
            title: "",
            status: "1",
            weight: "0",
          });
        },
      });
    }
  });
};

hpTerm.SetCommit = function () {
  var form = $("#hpm-termset"),
    alertid = "#hpm-node-alert";

  var req = {
    kind: "Term",
    id: parseInt(form.find("input[name=id]").val()),
    title: form.find("input[name=title]").val(),
    status: parseInt(form.find("input[name=status]").val()),
  };

  var model_type = form.find("input[name=model_type]").val();
  if ((model_type = "taxonomy")) {
    req.weight = parseInt(form.find("input[name=weight]").val());
    req.pid = parseInt(form.find("select[name=pid]").val());
  } else if ((model_type = "tag")) {
    //
  }

  //
  var uri =
    "modname=" +
    hpTerm.SpecActive() +
    "&modelid=" +
    hpTerm.SpecTermModelActive();

  hpMgr.ApiCmd("term/set?" + uri, {
    method: "POST",
    data: JSON.stringify(req),
    callback: function (err, data) {
      if (!data || data.kind != "Term") {
        return l4i.InnerAlert(alertid, "alert-danger", data.error.message);
      }

      form.find("input[name=id]").val(data.id);

      l4i.InnerAlert(alertid, "alert-success", "Successful operation");
      setTimeout(hpTerm.List, 500);
    },
  });
};
