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

var field = `<fieldset class="story {{.Option "name"}}" data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}" data-meta='{{.Optionv "meta"|Format}}'>
<legend>{{.Option "name"}}</legend>
<form class="option"></form>
<div class="action"></div>
<div class="output"></div>
<div class="status"></div>
</fieldset>
`
var field0 = `<fieldset class="story" data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">
<legend>{{.Option "name"}} </legend>{{$meta := .Optionv "meta"}}
<form class="option">
	{{if $meta}}
		{{range $index, $value := index $meta "inputs"}}
			{{$type := index $value "_input"}}
			<div class="item input">
				<input
					{{if eq $type "text"}}class="args"{{end}}
					type="{{index $value "_input"}}"
					name="{{index $value "name"|Format}}"
					value="{{index $value "value"|Format}}"
					data-action="{{index $value "action"|Format}}"
				>
			</div>
		{{end}}
	{{end}}
</form>
<div class="action">
</div>
<div class="output">
	{{if .Result}}
		<div class="code">{{.Result}}</div>
	{{else if .Appendv "append"}}
		<table>{{$msg := .}}
			<tr>{{range $index, $value := .Appendv "append"}}<th>{{$value}}</th>{{end}}</tr>
			{{range $index, $_ := .Appendv "_index"}}<tr>{{range $_, $key := $msg.Appendv "append"}}{{$line := $msg.Appendv $key}}
				<td>{{index $line $index}}</td>
			{{end}}</tr>{{end}}
		</table>
	{{end}}
</div>
</fieldset>
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

var prefix = `<svg class="story" vertion="1.1" xmlns="http://www.w3.org/2000/svg" dominant-baseline="middle" text-anchor="middle"
	data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}"
	width="{{.Option "width"}}" height="{{.Option "height"}}"
	font-size="{{.Option "font-size"}}" stroke="{{.Option "stroke"}}" fill="{{.Option "fill"}}"
	stroke-width="{{.Option "stroke-width"}}" font-family="{{.Option "font-family"}}"
	style="{{.Option "style"}}"
>`

var premenu = `<ul class="story" data-type="premenu"></ul>`
var endmenu = `<ul class="story" data-type="endmenu">{{$menu := .Optionv "menu"}}{{range $index, $value := Value $menu "list"}}
<li>{{Value $value "prefix"}} {{Value $value "content"}}</li>{{end}}
</ul>`
