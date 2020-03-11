package web

import (
	"github.com/shylinux/toolkits"
)

var share_template = kit.Dict(
	"value", `<img src="/share/%s/value">`,
	"share", `<img src="/share/%s/share">`,

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
