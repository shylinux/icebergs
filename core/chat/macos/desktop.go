package macos

import (
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
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(web.BINPACK, mdb.INSERT, nfs.USR_ICONS)
		}},
		ice.CTX_OPEN: {Hand: func(m *ice.Message, arg ...string) {
			if m.Cmd(DESKTOP).Length() == 0 {
				DeskAppend(m, "Books.png", web.WIKI_WORD)
				DeskAppend(m, "Photos.png", web.WIKI_FEEL)
				DeskAppend(m, "Calendar.png", web.TEAM_PLAN)
				DeskAppend(m, "Messages.png", web.CHAT_MESSAGE)
			}
			if m.Cmd(DOCK).Length() == 0 {
				DockAppend(m, "Finder.png", Prefix(FINDER))
				DockAppend(m, "Safari.png", web.CHAT_IFRAME)
				DockAppend(m, "Terminal.png", web.CODE_XTERM)
				DockAppend(m, "git.png", web.CODE_GIT_STATUS)
				DockAppend(m, "vimer.png", web.CODE_VIMER)
			}
			m.Travel(func(p *ice.Context, c *ice.Context, key string, cmd *ice.Command) {
				kit.If(cmd.Icon, func() {
					if kit.Contains(cmd.Icon, ".ico", ".png", ".jpg") {
						cmd.Icon = AppInstall(m, cmd.Icon, m.PrefixKey())
					}
				})
			})
			Notify(m, "usr/icons/Infomation.png", cli.RUNTIME, "系统启动成功", ctx.INDEX, cli.RUNTIME)
		}},
		DESKTOP: {Help: "桌面", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				web.AddPortalProduct(m, "桌面", `
一款网页版的电脑桌面，打开网页即可随时随地的使用各种软件，
无论这些软件是运行在本机，还是远程，还是任何虚拟的空间，无论是内存还是磁盘。
`, 100.0)
			}},
			// web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { web.DreamProcessIframe(m, arg...) }},
			// web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { web.DreamProcess(m, "", arg, arg...) }},
		}, web.DreamTablesAction(), PodCmdAction(), CmdHashAction(ctx.INDEX), mdb.ExportHashAction())},
	})
}

func DeskAppend(m *ice.Message, icon, index string, arg ...string) {
	install(m, DESKTOP, icon, index, arg...)
}
