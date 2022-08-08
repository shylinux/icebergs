package chat

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _cmd_file(m *ice.Message, arg ...string) bool {
	switch p := path.Join(arg...); kit.Ext(p) {
	case nfs.JS:
		m.Display(ctx.FileURI(p))
		web.RenderCmd(m, kit.Select(ctx.CAN_PLUGIN, ctx.GetFileCmd(p)))

	case nfs.GO:
		web.RenderCmd(m, ctx.GetFileCmd(p))

	case nfs.SHY:
		web.RenderCmd(m, "web.wiki.word", p)

	case nfs.IML:
		if m.Option(ice.MSG_USERPOD) == "" {
			m.RenderRedirect(path.Join(CHAT_WEBSITE, strings.TrimPrefix(p, SRC_WEBSITE)))
			m.Option(ice.MSG_ARGS, m.Option(ice.MSG_ARGS))
		} else {
			m.RenderRedirect(path.Join("/chat/pod", m.Option(ice.MSG_USERPOD), "website", strings.TrimPrefix(p, SRC_WEBSITE)))
			m.Option(ice.MSG_ARGS, m.Option(ice.MSG_ARGS))
		}

	case nfs.ZML:
		web.RenderCmd(m, "can.parse", m.Cmdx(nfs.CAT, p))

	default:
		if p = strings.TrimPrefix(p, ice.SRC+ice.PS); kit.FileExists(path.Join(ice.SRC, p)) {
			if msg := m.Cmd(mdb.RENDER, kit.Ext(p)); msg.Length() > 0 {
				m.Cmdy(mdb.RENDER, kit.Ext(p), p, ice.SRC+ice.PS).RenderResult()
				break
			}
		}
		return false
	}
	return true
}

const CMD = "cmd"

func init() {
	Index.MergeCommands(ice.Commands{
		CMD: {Name: "cmd path auto upload up home", Help: "命令", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, CMD)
				m.Cmdy(CMD, mdb.CREATE, mdb.TYPE, nfs.SHY, mdb.NAME, "web.wiki.word")
				m.Cmdy(CMD, mdb.CREATE, mdb.TYPE, nfs.SVG, mdb.NAME, "web.wiki.draw")
				m.Cmdy(CMD, mdb.CREATE, mdb.TYPE, nfs.CSV, mdb.NAME, "web.wiki.data")
				m.Cmdy(CMD, mdb.CREATE, mdb.TYPE, nfs.JSON, mdb.NAME, "web.wiki.json")
				for _, k := range []string{"mod", "sum"} {
					m.Cmdy(CMD, mdb.CREATE, mdb.TYPE, k, mdb.NAME, "web.code.inner")
				}
			}},
		}, mdb.HashAction(mdb.SHORT, "type", nfs.PATH, nfs.PWD), ctx.CmdAction(), web.ApiAction()), Hand: func(m *ice.Message, arg ...string) {
			if _cmd_file(m, arg...) {
				return
			}
			if ctx.PodCmd(m, ctx.COMMAND, arg[0]) && !m.IsErr() {
				web.RenderCmd(m, arg[0], arg[1:]) // 远程命令
			} else if m.Cmdy(ctx.COMMAND, arg[0]); m.Length() > 0 {
				web.RenderCmd(m, arg[0], arg[1:]) // 本地命令
			}
		}},
	})
}
