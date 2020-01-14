package wiki

var title = `<{{.Option "level"}} class="story" data-type="{{.Option "type"}}" data-name="{{.Option "prefix"}}" data-text="{{.Option "text"}}">{{.Option "prefix"}}{{.Option "content"}}</{{.Option "level"}}>`
var brief = `<p class="story" data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">{{.Option "text"}}</p>`
var refer = `<ul class="story"
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">
{{range $index, $value := .Optionv "list"}}<li>{{index $value 0}}: {{index $value 1}}</li>{{end}}</ul>`
var spark = `<p>{{.}}</p>`

var shell = `<div class="story code" data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "input"}}">$ {{.Option "input"}}
{{.Option "output"}}</div>
`

var order = `<ul class="story"
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">
{{range $index, $value := .Optionv "list"}}<li>{{$value}}</li>{{end}}</ul>`

var table = `<table class="story"
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">
<tr>{{range $i, $v := .Optionv "head"}}<th>{{$v}}</th>{{end}}</tr>
{{range $index, $value := .Optionv "list"}}
<tr>{{range $i, $v := $value}}<td>{{$v}}</td>{{end}}</tr>
{{end}}
</table>`

var prefix = `<svg class="story" vertion="1.1" xmlns="http://www.w3.org/2000/svg"
	data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}"
	width="{{.Option "width"}}" height="{{.Option "height"}}" style="{{.Option "style"}}"
>`

var premenu = `<ul class="story premenu"></ul>`
var endmenu = `<ul class="story endmenu">
{{$menu := .Optionv "menu"}}
{{range $index, $value := Value $menu "list"}}
<li>{{Value $value "prefix"}} {{Value $value "content"}}</li>
{{end}}
</ul>`
