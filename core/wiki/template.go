package wiki

var premenu = `<ul class="story" data-type="premenu"></ul>`
var endmenu = `<ul class="story" data-type="endmenu">{{$menu := .Optionv "menu"}}{{range $index, $value := Value $menu "list"}}
<li>{{Value $value "prefix"}} {{Value $value "content"}}</li>{{end}}
</ul>`

var title = `<{{.Option "level"}} class="story"
{{range $k, $v := .Optionv "extra"}}data-{{$k}}='{{$v}}'{{end}}
data-type="{{.Option "type"}}" data-name="{{.Option "prefix"}}" data-text="{{.Option "text"}}"
>{{.Option "prefix"}} {{.Option "content"}}</{{.Option "level"}}>`

var brief = `<p class="story"
{{range $k, $v := .Optionv "extra"}}data-{{$k}}='{{$v}}'{{end}}
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}"
>{{.Option "text"}}</p>`

var refer = `<ul class="story"
{{range $k, $v := .Optionv "extra"}}data-{{$k}}='{{$v}}'{{end}} 
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">
{{range $index, $value := .Optionv "list"}}<li>{{index $value 0}}: <a href="{{index $value 1}}" target="_blank">{{index $value 1}}</a></li>{{end}}</ul>`
var spark = `<p class="story" {{range $k, $v := .Optionv "extra"}}data-{{$k}}='{{$v}}'{{end}} data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">{{.Option "text"}}</p>`

var chart = `<svg class="story" vertion="1.1" xmlns="http://www.w3.org/2000/svg" dominant-baseline="middle" text-anchor="middle"
	data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}"
	width="{{.Option "width"}}" height="{{.Option "height"}}"
	font-size="{{.Option "font-size"}}" stroke="{{.Option "stroke"}}" fill="{{.Option "fill"}}"
	stroke-width="{{.Option "stroke-width"}}" font-family="{{.Option "font-family"}}"
	style="{{.Option "style"}}"
>`
var field = `<fieldset class="story {{.Option "name"}}" data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}" data-meta='{{.Optionv "meta"|Format}}'>
<legend>{{.Option "name"}}</legend>
<form class="option"></form>
<div class="action"></div>
<div class="output"></div>
<div class="status"></div>
</fieldset>
`
var shell = `<code class="story"
{{range $k, $v := .Optionv "extra"}}data-{{$k}}='{{$v}}'{{end}}
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "input"}}"
>$ {{.Option "input"}} # {{.Option "name"}}
{{.Option "output"}}</code>
`
var local = `<code class="story"
{{range $k, $v := .Optionv "extra"}}data-{{$k}}='{{$v}}'{{end}}
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}"
>{{range $index, $value := .Optionv "input"}}{{$value}}{{end}}</code>`

var order = `<ul class="story"
{{range $k, $v := .Optionv "extra"}}data-{{$k}}='{{$v}}'{{end}}
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">
{{range $index, $value := .Optionv "list"}}<li>{{$value}}</li>{{end}}</ul>`

var table = `<table class="story"
{{range $k, $v := .Optionv "extra"}}data-{{$k}}='{{$v}}'{{end}}
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">
<tr>{{range $i, $v := .Optionv "head"}}<th>{{$v}}</th>{{end}}</tr>
{{range $index, $value := .Optionv "list"}}
<tr>{{range $i, $v := $value}}<td>{{$v}}</td>{{end}}</tr>
{{end}}
</table>`

var image = `<img class="story"
{{range $k, $v := .Optionv "extra"}}data-{{$k}}='{{$v}}'{{end}}
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}"
title="{{.Option "text"}}" src="{{.Option "text"}}">`

var video = `<video class="story"
{{range $k, $v := .Optionv "extra"}}data-{{$k}}='{{$v}}'{{end}}
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}"
title="{{.Option "text"}}" src="{{.Option "text"}}" controls></video>`
