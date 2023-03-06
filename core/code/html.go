package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const HTML = "html"

func init() {
	Index.MergeCommands(ice.Commands{
		HTML: {Name: "html path auto", Help: "网页", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.EchoIFrame(kit.MergeURL(path.Join("/require/", arg[2], arg[1]), "_v", kit.Hashs("uniq")))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				m.EchoIFrame(kit.MergeURL(path.Join("/require/", arg[2], arg[1]), "_v", kit.Hashs("uniq")))
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(kit.Renders(_html_template, ice.Maps{ice.LIST: kit.Format(kit.List(kit.Dict(ctx.INDEX, ctx.GetFileCmd(kit.ExtChange(path.Join(arg[2], arg[1]), GO)))))})).RenderResult()
			}},
		}, PlugAction())},
	})
}

var _html_template = `<!DOCTYPE html>
<head>
	<meta charset="utf-8"><title>volcanos</title>
	<link href="/publish/can.css" rel="stylesheet">
</head>
<body>
	<script src="/publish/can.js"></script>
	<script>Volcanos({{.list}})</script>
</body>
`
