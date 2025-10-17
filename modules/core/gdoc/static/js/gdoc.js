var gdoc = {};

gdoc.PageEntryRender = function (opts) {
  opts = opts || {};

  if (window.location.pathname) {
    $("#hp-gdoc-entry-summary a").each(function (i, el) {
      var basepath = $(this).attr("href");
      if (basepath && basepath == window.location.pathname) {
        $(this).addClass("active");
      }
    });
  }

  var idre = /^H\d{1}$/;
  var el = document.getElementById("hp-gdoc-page-entry-content");
  if (!el) {
    return;
  }
  var elo = document.getElementById("hp-gdoc-page-entry-toc");
  if (!elo) {
    return;
  }

  var items = el.childNodes;
  var nav = "";
  var last = 1;
  var num = 0;
  for (var i in items) {
    var item = items[i];
    if (!item.previousSibling) {
      continue;
    }
    if (!idre.test(item.previousSibling.nodeName)) {
      continue;
    }
    var level = parseInt(item.previousSibling.nodeName.substring(1));
    if (level > 4) {
      continue;
    } else if (level < 2) {
      level = 2;
    }
    if (level > last) {
      nav += "<ul>\n";
    } else if (level < last) {
      nav += "</li>\n</ul>\n";
    } else {
      nav += "</li>\n";
    }
    var tocid = "hp-gdoc-toc-" + i;
    var toctitle = item.previousSibling.outerText.trim();
    if (toctitle.length < 1) {
      continue;
    } else if (toctitle.length > 40) {
      toctitle = toctitle.substr(0, 30) + "...";
    }
    nav += '<li><a href="#' + tocid + '">' + toctitle + "</a>";
    last = level;
    item.previousSibling.id = tocid;
    num += 1;
  }

  if (num < 1) {
    // elo.style.display = "none";
    // el.classList.replace("is-8", "is-10");
    return;
  }

  while (last > 1) {
    nav += "</li></ul>\n";
    last -= 1;
  }

  elo.innerHTML = '<nav class="hp-gdoc-page-toc-menu">\n<h1>Page Nav</h1>' + nav + "</nav>";
};

gdoc.textWidth = function (txt, opts) {
  opts = opts || {};

  var el = document.createElement("div");
  if (opts.fontSize) {
    el.style.fontSize = opts.fontSize;
  }
  el.style.position = "absolute";
  el.style.whiteSpace = "nowrap";
  el.style.left = -1000;
  el.style.top = -1000;
  el.innerHTML = txt;

  document.body.appendChild(el);
  var width = el.clientWidth;
  document.body.removeChild(el);

  return width;
};
