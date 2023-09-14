package macos

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const APPLICATIONS = "applications"

func init() {
	Index.MergeCommands(ice.Commands{
		APPLICATIONS: {Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				FinderAppend(m, "Applications", m.PrefixKey())
				m.Travel(func(p *ice.Context, c *ice.Context, key string, cmd *ice.Command) {
					kit.If(cmd.Icon, func() {
						m.Debug("what %v", cmd.Icon)
						AppInstall(m, cmd.Icon, m.PrefixKey())
					})
				})
				Notify(m, cli.RUNTIME, "系统启动成功", ctx.INDEX, cli.RUNTIME)
			}},
			ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { mdb.Conf(m, m.PrefixKey(), mdb.HASH, "") }},
			code.INSTALL: {Hand: func(m *ice.Message, arg ...string) { AppInstall(m, arg[0], arg[1]) }},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case web.SPACE:
					m.Cmdy(web.SPACE).CutTo(mdb.NAME, arg[0])
				case ctx.INDEX:
					m.Cmdy(web.Space(m, m.Option(web.SPACE)), ctx.COMMAND)
				case ctx.ARGS:
					m.Cmdy(web.Space(m, m.Option(web.SPACE)), ctx.COMMAND, mdb.INPUTS, m.Option(ctx.INDEX))
				case mdb.ICON:
					if m.Option(ctx.INDEX) != "" {
						m.Cmd(web.Space(m, m.Option(web.SPACE)), m.Option(ctx.INDEX), mdb.INPUTS, arg[0]).Table(func(value ice.Maps) {
							m.Push(arg[0], value[arg[0]]+"?pod="+kit.Keys(m.Option(ice.MSG_USERPOD), m.Option(web.SPACE)))
						})
					}
					m.Cmd(nfs.DIR, USR_ICONS, func(value ice.Maps) { m.Push(arg[0], value[nfs.PATH]) })
				}
			}}, mdb.CREATE: {Name: "create space index args name icon"},
		}, PodCmdAction(), CmdHashAction("space,index,args"))},
	})
}
func install(m *ice.Message, cmd, icon, index string, arg ...string) {
	name := kit.TrimExt(path.Base(icon), "png")
	m.Cmd(Prefix(cmd), mdb.CREATE, mdb.NAME, name, mdb.ICON, icon, ctx.INDEX, index, arg)
}
func AppInstall(m *ice.Message, icon, index string, arg ...string) {
	install(m, APPLICATIONS, icon, index, arg...)
}
