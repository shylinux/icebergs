package wiki

var title = `<{{.Option "level"}} class="story" data-type="{{.Option "type"}}" data-name="{{.Option "prefix"}}" data-text="{{.Option "text"}}">{{.Option "prefix"}} {{.Option "content"}}</{{.Option "level"}}>`
var brief = `<p class="story" data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">{{.Option "text"}}</p>`
var refer = `<ul class="story"
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">
{{range $index, $value := .Optionv "list"}}<li>{{index $value 0}}: {{index $value 1}}</li>{{end}}</ul>`
var spark = `<p class="story" data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">{{.Option "text"}}</p>`

var local = `<div class="story"
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">
{{range $index, $value := .Optionv "input"}}{{$value}}{{end}}</div>`

var shell = `<code class="story" data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "input"}}" data-dir="{{.Option "cmd_dir"}}">$ {{.Option "input"}} # {{.Option "name"}}
{{.Option "output"}}</code>
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

var stack = `<div class="story"
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">`

var prefix = `<svg class="story" vertion="1.1" xmlns="http://www.w3.org/2000/svg"
	data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}"
	width="{{.Option "width"}}" height="{{.Option "height"}}"
	font-size="{{.Option "font-size"}}" style="{{.Option "style"}}"
	dominant-baseline="middle" text-anchor="middle"
	stroke="{{.Option "stroke"}}" fill="{{.Option "fill"}}"
	stroke-width="{{.Option "stroke-width"}}"
	font-family="{{.Option "font-family"}}"
>`

var premenu = `<ul class="story" data-type="premenu"></ul>`
var endmenu = `<ul class="story" data-type="endmenu">{{$menu := .Optionv "menu"}}{{range $index, $value := Value $menu "list"}}
<li>{{Value $value "prefix"}} {{Value $value "content"}}</li>{{end}}
</ul>`
