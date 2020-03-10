package web

import (
	"github.com/shylinux/toolkits"
)

var share_template = kit.Dict(
	"share", `<a href="/share/%s" target="_blank">%s</a>`,
	"qrcode", `<img src="/share/%s/qrcode">`,

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
