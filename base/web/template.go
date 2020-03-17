package web

import (
	"github.com/shylinux/toolkits"
)

var share_template = kit.Dict(
	"value", `<img src="/share/%s/value">`,
	"share", `<img src="/share/%s/share">`,
	"text", `<img src="/share/%s/text">`,

	"link", `<a href="/share/%s" target="_blank">%s</a>`,
	"simple", `<!DOCTYPE html>
<head>
<meta charset='utf-8'>
<title>{{.Option "name"}}</title>
</head>
<body>
{{.Option "text"}}
</body>
`,
)
var favor_template = kit.Dict(
	"shell", `<div class="code">{{$msg := .}}{{range $index, $_ := .Appendv "_index"}}{{$name := $msg.Appendv "name"}}{{$text := $msg.Appendv "text"}}
# {{index $name $index}}
{{index $text $index}}
{{end}}
</div>`,
	"vimrc", `<div class="code">{{$msg := .}}{{range $index, $_ := .Appendv "_index"}}{{$name := $msg.Appendv "name"}}{{$id := $msg.Appendv "id"}}
{{$res := index $id $index|$msg.Prefile ""}}
# {{index $name $index}} {{index $res "extra.buf"}}:{{index $res "extra.row"}}
{{index $res "content"}}
{{end}}
</div>`,
	"field", `{{$msg := .}}
{{range $index, $_ := .Appendv "_index"}}
{{$type := $msg.Appendv "type"}}{{$name := $msg.Appendv "name"}}{{$text := $msg.Appendv "text"}}
<fieldset class="story {{index $name $index}}" data-type="{{index $type $index}}" data-name="{{index $name $index}}" data-text="{{index $text $index}}" data-meta='{{index $text $index|$msg.Preview}}'>
<legend>{{index $name $index}}</legend>
<form class="option"></form>
<div class="action"></div>
<div class="output"></div>
<div class="status"></div>
</fieldset>
{{end}}
`,
	"spide", `<ul>{{$msg := .}}
{{range $index, $_ := .Appendv "_index"}}
	{{$name := $msg.Appendv "name"}}
	{{$text := $msg.Appendv "text"}}
	<li>{{index $name $index}}: <a href="{{index $text $index}}" target="_blank">{{index $text $index}}</a></li>
{{end}}
</ul>`,
)
