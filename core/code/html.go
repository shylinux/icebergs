package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const HTML = "html"

func init() {
	Index.MergeCommands(ice.Commands{
		HTML: {Name: "html path auto", Help: "网页", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.EchoIFrame(kit.MergeURL(path.Join(ice.PS, ice.REQUIRE, arg[2], arg[1]), "_v", kit.Hashs(mdb.UNIQ)))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				m.EchoIFrame(kit.MergeURL(path.Join(ice.PS, ice.REQUIRE, arg[2], arg[1]), "_v", kit.Hashs(mdb.UNIQ)))
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(kit.Renders(nfs.TemplateText(m, "demo.html"), ice.Maps{ice.LIST: kit.Format(kit.List(kit.Dict(ctx.INDEX, ctx.GetFileCmd(kit.ExtChange(path.Join(arg[2], arg[1]), GO)))))})).RenderResult()
			}},
		}, PlugAction())},
	})
}
