package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func require(arg ...string) string { return path.Join(nfs.PS, ice.REQUIRE, path.Join(arg...)) }

const HTML = "html"

func init() {
	Index.MergeCommands(ice.Commands{
		HTML: {Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.EchoIFrame(m.MergeLink(require(arg[2], arg[1]), "_v", kit.Hashs(mdb.UNIQ)))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				m.EchoIFrame(m.MergeLink(require(arg[2], arg[1]), "_v", kit.Hashs(mdb.UNIQ)))
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(nfs.Template(m, DEMO_HTML))
			}},
		}, PlugAction())},
	})
}
