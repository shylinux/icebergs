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
	// case nfs.SHY:
	// 	web.RenderCmd(m, "web.wiki.word", p)
	case nfs.GO:
		web.RenderCmd(m, ctx.GetFileCmd(p))
	case nfs.JS:
		ctx.DisplayBase(m, ctx.FileURI(p))
		web.RenderCmd(m, kit.Select(ice.CAN_PLUGIN, ctx.GetFileCmd(p)))
	default:
		if p = strings.TrimPrefix(p, ice.SRC+nfs.PS); nfs.Exists(m, path.Join(ice.SRC, p)) {
			if msg := m.Cmd(mdb.ENGINE, kit.Ext(p)); msg.Length() > 0 {
				m.Cmdy(mdb.ENGINE, kit.Ext(p), p, ice.SRC+nfs.PS).RenderResult()
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
		CMD: {Name: "cmd path auto upload up home", Help: "命令", Actions: ice.MergeActions(
			mdb.HashAction(mdb.SHORT, mdb.TYPE, nfs.PATH, nfs.PWD), ctx.CmdAction(), web.ApiAction(), aaa.WhiteAction(ice.RUN),
		), Hand: func(m *ice.Message, arg ...string) {
			if _cmd_file(m, arg...) {
				return
			} else if len(arg[0]) == 0 || arg[0] == "" {
				return
			} else if m.IsCliUA() {
				m.Cmdy(arg, m.Optionv(ice.ARG)).RenderResult()
				return
			}
			if arg[0] == "web.chat.portal" {
				web.RenderMain(m)
			} else if m.Cmdy(ctx.COMMAND, arg[0]); m.Length() > 0 {
				web.RenderCmd(m, m.Append(ctx.INDEX), arg[1:])
			}
		}},
	})
}
