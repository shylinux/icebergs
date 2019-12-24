package web

import (
	"github.com/shylinux/toolkits"
)

var share_template = kit.Dict(
	"download", `<a href="/code/zsh?cmd=download&arg=%s" target="_blank">%s</a>`,
	"share", `<a href="/share/%s" target="_blank">%s</a>`,
	"shy/story", `{{.}}`,
	"shy/chain", `<!DOCTYPE html>
<head>
<meta charset="utf-8">
<link rel="stylesheet" text="text/css" href="/style.css">
</head>
<body>
<fieldset>
	{{.Append "type"}}
	{{.Append "text"}}
</fieldset>
</body>
`,
)
