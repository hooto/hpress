{{if (SysConfig `frontend_html_footer`)}}
<!-- -->
{{raw (SysConfig `frontend_html_footer`)}}
<!-- -->
{{else}}
<footer class="hp-footer">
  <div class="container">
    <div class="d-flex flex-column flex-lg-row justify-content-lg-between">
      <div>{{raw (SysConfig `frontend_footer_copyright`)}}</div>
      <div>
        <span class="hp-footer-powerby-item"
          >Published by
          <strong
            ><a href="https://github.com/hooto/hpress" target="_blank"
              >Hooto Press CMS</a
            ></strong
          >,</span
        >
        <span class="hp-footer-powerby-item"
          >Powered by
          <strong
            ><a href="https://www.sysinner.cn" target="_blank"
              >InnerStack PaaS Engine</a
            ></strong
          ></span
        >
        {{if $.frontend_langs}}
        <span class="hp-footer-powerby-item"
          >Language
          <select onchange="hp.LangChange(this)" class="hp-footer-langs">
            {{range $v := $.frontend_langs}}
            <option value="{{$v.Id}}" {{if eq $v.Id $.LANG}}selected{{end}}>
              {{$v.Name}}
            </option>
            {{end}}
          </select>
        </span>
        {{end}}
      </div>
    </div>
  </div>
</footer>
{{end}}
<!-- -->
{{raw (SysConfig `frontend_footer_analytics_scripts`)}}
