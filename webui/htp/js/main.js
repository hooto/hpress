var htp = {
    base : "/htp/",
}

htp.Boot = function()
{
    if (window._basepath && window._basepath.length > 1) {
        htp.base = window._basepath;
        if (htp.base.substring(htp.base.length - 1) != "/") {
            htp.base += "/";
        }
    }

    if (!htp.base || htp.base == "") {
        htp.base = "/";
    }

    seajs.config({
        base: htp.base,
    });

    seajs.use([
        "~/htp/js/jquery.js",
    ],
    function() {

        seajs.use([
            "~/lessui/js/lessui.js",
        ],
        function() {

            setTimeout(function() {
                for (var i in window.onload_hooks) {
                    window.onload_hooks[i]();
                }
            }, 100);
        });
    });
}

htp.HttpSrvBasePath = function(url)
{
    if (htp.base == "") {
        return url;
    }

    if (url.substr(0, 1) == "/") {
        return url;
    }

    return htp.base + url;
}

htp.CodeRender = function()
{
    $("code[class^='language-']").each(function(i, el) {

        var lang = el.className.substr("language-".length);
        if (lang == "hooto_chart") {
            return htp.chartRender(i, el);
        }

        var modes = [];

        if (lang == "html") {
            lang = "htmlmixed";
        }

        switch (lang) {

        case "php":
            modes.push("~/cm/5/mode/php/php.js");
        case "htmlmixed":
            modes.push("~/cm/5/mode/xml/xml.js");
            modes.push("~/cm/5/mode/javascript/javascript.js");
            modes.push("~/cm/5/mode/css/css.js");
            modes.push("~/cm/5/mode/htmlmixed/htmlmixed.js");
            break;

        case "c":
        case "cpp":
        case "clike":
        case "java":
            lang = "clike";
            break;

        case "json":
            modes.push("~/cm/5/mode/javascript/javascript.js");
            lang = "application/ld+json";
            break;

        case "go":
        case "javascript":
        case "css":
        case "xml":
        case "yaml":
        case "lua":
        case "markdown":
        case "r":
        case "shell":
        case "sql":
        case "swift":
        case "erlang":
        case "nginx":
            modes.push("~/cm/5/mode/"+ lang +"/"+ lang +".js");
            break;

        default:
            return;
        }

        seajs.use([
            "~/cm/5/lib/codemirror.css",
            "~/cm/5/lib/codemirror.js",
        ],
        function() {

            modes.push("~/cm/5/addon/runmode/runmode.js");
            modes.push("~/cm/5/mode/clike/clike.js");

            seajs.use(modes, function() {

                $(el).addClass('cm-s-default'); // apply a theme class
                CodeMirror.runMode($(el).text(), lang, $(el)[0]);
            });
        });
    });
}

htp.chartRender = function(i, elem)
{
    var elem_id = "hooto_chart-id-" + i;
    elem.setAttribute("id", elem_id);
    seajs.use([
        "~/chart/chart.js",
    ],
    function() {
        hooto_chart.basepath = htp.base + "/~/chart";
		hooto_chart.opts_width = "600px";
		hooto_chart.opts_height = "400px";
        hooto_chart.JsonRenderElement(elem, elem_id);
    });
}

htp.NavActive = function(tplid, path)
{
    if (!tplid || !path) {
        return;
    }

    var nav = $("#"+ tplid);
    nav.find("a").each(function() {

        var href = $(this).attr("href");

        if (href) {

            if (href.match(path)) {
                nav.find("a.active").removeClass("active");
                $(this).addClass("active");
            }
        }
    });
}

htp.Ajax = function(url, options)
{
    options = options || {};

    //
    if (url.substr(0, 1) != "/" && url.substr(0, 4) != "http") {
        url = htp.HttpSrvBasePath(url);
    }

    l4i.Ajax(url, options)
}

htp.ActionLoader = function(target, url)
{
    htp.Ajax(htp.HttpSrvBasePath(url), {
        callback: function(err, data) {
            $("#"+ target).html(data);
        }
    });
}

htp.ApiCmd = function(url, options)
{
    htp.Ajax(htp.HttpSrvBasePath(url), options);
}

htp.AuthSessionRefresh = function()
{
    htp.Ajax(htp.HttpSrvBasePath("auth/session"), {
        callback: function(err, data) {

            if (err || !data || data.kind != "AuthSession") {

                return l4iTemplate.Render({
                    dstid:   "htp-topbar-userbar",
                    tplid:   "htp-topbar-user-unsigned-tpl",
                });
            }

            l4iTemplate.Render({
                dstid:   "htp-topbar-userbar",
                tplid:   "htp-topbar-user-signed-tpl",
                data:    data,
                success: function() {

                    $("#htp-topbar-userbar").hover(
                        function() {
                            $("#htp-topbar-user-signed-modal").fadeIn(200);
                        },
                        function() {
                        }
                    );
                    $("#htp-topbar-user-signed-modal").hover(
                        function() {
                        },
                        function() {
                            $("#htp-topbar-user-signed-modal").fadeOut(200);
                        }
                    );
                },
            });
        },
    });
}

