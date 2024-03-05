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
				if nfs.Exists(m, nfs.USR_LOCAL_IMAGE) {
					DeskAppend(m, "Photos.png", web.WIKI_FEEL, ctx.ARGS, nfs.USR_LOCAL_IMAGE)
				} else {
					DeskAppend(m, "Photos.png", web.WIKI_FEEL, ctx.ARGS, nfs.USR_ICONS)
				}
				DeskAppend(m, "Grapher.png", web.WIKI_DRAW)
				DeskAppend(m, "Calendar.png", web.TEAM_PLAN)
				DeskAppend(m, "Messages.png", web.CHAT_MESSAGE)
			}
			if m.Cmd(DOCK).Length() == 0 {
				DockAppend(m, "Finder.png", Prefix(FINDER))
				DockAppend(m, "Safari.png", web.CHAT_IFRAME)
				DockAppend(m, "Terminal.png", web.CODE_XTERM)
				DockAppend(m, "go.png", web.CODE_COMPILE)
				DockAppend(m, "git.png", web.CODE_GIT_STATUS)
				DockAppend(m, "vimer.png", web.CODE_VIMER)
			}
			AppInstall(m, "App Store.png", web.STORE)
			m.Travel(func(p *ice.Context, c *ice.Context, key string, cmd *ice.Command) {
				kit.If(cmd.Icon, func() {
					if !kit.HasPrefix(cmd.Icon, nfs.PS, web.HTTP) {
						nfs.Exists(m, path.Join(path.Dir(strings.TrimPrefix(ctx.GetCmdFile(m, m.PrefixKey()), kit.Path(""))), cmd.Icon), func(p string) { cmd.Icon = p })
					}
					AppInstall(m, cmd.Icon, m.PrefixKey())
				})
			})
			Notify(m, "usr/icons/Infomation.png", cli.RUNTIME, "系统启动成功", ctx.INDEX, cli.RUNTIME)
		}},
		DESKTOP: {Help: "桌面", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			// web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { web.DreamProcessIframe(m, arg...) }},
			// web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { web.DreamProcess(m, "", arg, arg...) }},
		}, web.DreamTablesAction(), PodCmdAction(), CmdHashAction(), mdb.ExportHashAction())},
	})
}

func DeskAppend(m *ice.Message, icon, index string, arg ...string) {
	install(m, DESKTOP, icon, index, arg...)
}
