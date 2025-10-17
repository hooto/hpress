{{if .categories}}
<div class="">
  <div class="list-group term-taxonomy-group">
    <div class="list-group-item">
      <a class="term-taxonomy-item" href="{{$.baseuri}}/list">All</a>
    </div>
    {{range $v := .categories.Items}} {{if ne $v.PID 0}}{{continue}}{{end}}
    <div class="list-group-item">
      <a
        class='term-taxonomy-item {{if eq $.term_categories "$v.ID"}} active{{end}}'
        href="{{$.baseuri}}/list?term_{{$.categories.Model.Meta.Name}}={{$v.ID}}"
        >{{$v.Title}}</a
      >
      {{range $v2 := $.categories.Items}} {{if ne $v2.ID
      $v2.PID}}{{continue}}{{end}}
      <a
        class='term-taxonomy-subitem {{if eq $.term_categories "$v2.ID"}} active{{end}}'
        href="{{$.baseuri}}/list?term_{{$.categories.Model.Meta.Name}}={{$v2.ID}}"
        >{{$v2.Title}}</a
      >
      {{end}}
    </div>
    {{end}}
  </div>
</div>
{{end}}
