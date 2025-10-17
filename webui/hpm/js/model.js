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

var hpMgrModel = {};

hpMgrModel.List = function (tplid) {
  if (!tplid) {
    tplid = "hpm-ctnls";
  }

  var qry_grpdev = $("#" + tplid + "-grpdev-qryid")
    .attr("href")
    .substr(1);
  var qry_grpdev_dp = "Groups";
  if (qry_grpdev != "") {
    qry_grpdev_dp = $("#" + tplid + "-grpdev-qrydp").text();
  }

  var uri = "qry_text=" + $("#" + tplid + "-qry-text").val();
  uri += "&qry_grpdev=" + qry_grpdev;

  var req = {
    items: [
      {
        name: "groups",
        uri: hpMgr.base + "ext/lps/group/list",
      },
      {
        name: "info",
        uri: hpMgr.base + "ext/lps/pkg-info/list?" + uri,
      },
    ],
  };

  //
  $.ajax({
    type: "POST",
    url: hpMgr.base + "v1/mixer",
    data: JSON.stringify(req),
    timeout: 3000,
    success: function (rsp) {
      var rsj = JSON.parse(rsp);
      if (rsj === undefined || rsj.kind != "Mixer" || rsj.items === undefined) {
        $("#" + tplid + "-empty-alert").show();
        return;
      }

      if (rsj.items.groups === undefined) {
        $("#" + tplid + "-empty-alert").show();
        return;
      }

      if (
        rsj.items.info === undefined ||
        rsj.items.info.kind != "PackageInfoList" ||
        rsj.items.info.items === undefined
      ) {
        rsj.items.info.items = [];
      }

      if (rsj.items.info.items.length > 0) {
        $("#" + tplid + "-empty-alert").hide();
      } else {
        $("#" + tplid + "-empty-alert").show();
      }

      for (var i in rsj.items.info.items) {
        rsj.items.info.items[i].meta.updated = l4i.TimeParseFormat(
          rsj.items.info.items[i].meta.updated,
          "Y-m-d"
        );
      }

      lessTemplate.Render({
        dstid: tplid,
        tplid: tplid + "-tpl",
        data: rsj.items.info.items,
      });

      if (
        rsj.items.groups.kind !== undefined &&
        rsj.items.groups.kind == "PackageGroupList" &&
        rsj.items.groups.dev !== undefined &&
        rsj.items.groups.dev.length > 0
      ) {
        lessTemplate.Render({
          dstid: tplid + "-grpdev",
          tplid: tplid + "-grpdev-tpl",
          data: {
            grpdev: rsj.items.groups.dev,
            qry_grpdev: qry_grpdev,
            qry_grpdev_dp: qry_grpdev_dp,
          },
        });
      }
    },
    error: function (xhr, textStatus, error) {
      //lessAlert("#azt02e", 'alert-danger', textStatus+' '+xhr.responseText);
    },
  });
};
