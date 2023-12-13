package macos

import (
	"path"
	"strings"

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
		APPLICATIONS: {Help: "应用", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				FinderAppend(m, APPLICATIONS, m.PrefixKey())
				defer Notify(m, "Infomation.png", cli.RUNTIME, "系统启动成功", ctx.INDEX, cli.RUNTIME)
				m.Travel(func(p *ice.Context, c *ice.Context, key string, cmd *ice.Command) {
					kit.If(cmd.Icon, func() {
						if !kit.HasPrefix(cmd.Icon, nfs.PS, web.HTTP) {
							nfs.Exists(m, ice.USR_ICONS+cmd.Icon, func(p string) { cmd.Icon = p })
							nfs.Exists(m, path.Join(path.Dir(strings.TrimPrefix(ctx.GetCmdFile(m, m.PrefixKey()), kit.Path(""))), cmd.Icon), func(p string) {
								cmd.Icon = p
							})
						}
						AppInstall(m, cmd.Icon, m.PrefixKey())
					})
				})
			}},
			code.INSTALL: {Hand: func(m *ice.Message, arg ...string) { AppInstall(m, arg[0], arg[1]) }},
			mdb.CREATE:   {Name: "create space index args name icon"},
		}, PodCmdAction(), CmdHashAction("space,index,args"), mdb.ClearOnExitHashAction())},
	})
}
func install(m *ice.Message, cmd, icon, index string, arg ...string) {
	if icon == "" {
		return
	}
	nfs.Exists(m, ice.USR_ICONS+icon, func(p string) { icon = p })
	name := kit.TrimExt(path.Base(icon), nfs.PNG, nfs.JPG, nfs.JPEG)
	m.Cmd(Prefix(cmd), mdb.CREATE, mdb.NAME, name, mdb.ICON, icon, ctx.INDEX, index, arg)
}
func AppInstall(m *ice.Message, icon, index string, arg ...string) {
	install(m, APPLICATIONS, icon, index, arg...)
}
