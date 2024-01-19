package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func require(arg ...string) string { return path.Join(nfs.PS, ice.REQUIRE, path.Join(arg...)) }

const HTML = "html"

func init() {
	Index.MergeCommands(ice.Commands{
		HTML: {Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				if arg[1] == "main.html" {
					m.EchoIFrame(m.MergePodCmd("", web.ADMIN))
				} else {
					m.EchoIFrame(m.MergeLink(require(arg[2], arg[1]), "_v", kit.Hashs(mdb.UNIQ)))
				}
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
