package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const TEXT = "text"

func init() {
	Index.MergeCommands(ice.Commands{
		TEXT: {Help: "文本", Actions: ice.Actions{
			ctx.CMDS: {Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(arg)
				kit.If(msg.IsErrNotFound(), func() { msg.SetResult().Cmdy(cli.SYSTEM, arg) })
				kit.If(msg.Result() == "", func() { msg.TableEcho() })
				m.Cmdy("", msg.Result())
			}},
			web.LINK: {Name: "link link name text icons", Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.ICONS) == "", func() { m.Option(mdb.ICONS, m.Cmdv(ACCESS, m.Option(ACCESS), mdb.ICONS)) })
				m.Option(mdb.ICONS, web.ShareLocal(m, m.Option(mdb.ICONS)))
				m.Cmdy("", m.OptionDefault(mdb.TEXT, ice.Info.Titles), "link.xml")
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			m.Echo(nfs.Template(m.Options(mdb.TEXT, arg[0]), kit.Select("welcome.xml", arg, 1))).RenderResult()
			m.Debug("text: %v", m.Result())
		}},
	})
}
