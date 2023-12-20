package macos

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const DESKTOP = "desktop"

func init() {
	Index.MergeCommands(ice.Commands{
		ice.CTX_OPEN: {Hand: func(m *ice.Message, arg ...string) {
			if m.Cmd(DESKTOP).Length() == 0 {
				DeskAppend(m, "Books.png", web.WIKI_WORD)
				DeskAppend(m, "Photos.png", web.WIKI_FEEL)
				DeskAppend(m, "Grapher.png", web.WIKI_DRAW)
				DeskAppend(m, "Calendar.png", web.TEAM_PLAN)
			}
			if m.Cmd(DOCK).Length() == 0 {
				DockAppend(m, "Finder.png", Prefix(FINDER))
				DockAppend(m, "Safari.png", web.CHAT_IFRAME)
				DockAppend(m, "Terminal.png", web.CODE_XTERM)
				DockAppend(m, "go.png", web.CODE_COMPILE)
				DockAppend(m, "git.png", web.CODE_GIT_STATUS)
				DockAppend(m, "vimer.png", web.CODE_VIMER)
			}
			m.Travel(func(p *ice.Context, c *ice.Context, key string, cmd *ice.Command) {
				kit.If(cmd.Icon, func() {
					if !kit.HasPrefix(cmd.Icon, nfs.PS, web.HTTP) {
						nfs.Exists(m, path.Join(path.Dir(strings.TrimPrefix(ctx.GetCmdFile(m, m.PrefixKey()), kit.Path(""))), cmd.Icon), func(p string) { cmd.Icon = p })
					}
					AppInstall(m, cmd.Icon, m.PrefixKey())
				})
			})
			Notify(m, "Infomation.png", cli.RUNTIME, "系统启动成功", ctx.INDEX, cli.RUNTIME)
		}},
		DESKTOP: {Help: "应用桌面", Actions: ice.MergeActions(ice.Actions{
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.Switch(m.Option(mdb.TYPE), kit.Simple(web.WORKER, web.SERVER), func() { m.PushButton(kit.Dict(m.CommandKey(), "桌面")) })
			}},
			web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { web.DreamProcess(m, []string{}, arg...) }},
		}, aaa.RoleAction(), PodCmdAction(), CmdHashAction(), mdb.ExportHashAction())},
	})
}

func DeskAppend(m *ice.Message, icon, index string) { install(m, DESKTOP, icon, index) }
