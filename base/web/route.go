package web

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const ROUTE = "route"

func init() {
	Index.MergeCommands(ice.Commands{
		ROUTE: {Name: "route space auto travel spide cmds compile", Help: "路由表", Actions: ice.MergeActions(ice.Actions{
			ice.MAIN: {Help: "首页", Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, CHAT_IFRAME, m.MergePod(kit.Select(m.Option(SPACE), arg, 0)), arg...)
			}},
			"compile": {Hand: func(m *ice.Message, arg ...string) {
				args := []string{CODE_VIMER, "compile"}
				GoToast(m, "", func(toast func(string, int, int)) (list []string) {
					msg := m.Cmd("")
					count, total := 0, msg.Length()
					msg.Table(func(value ice.Maps) {
						if toast(value[SPACE], count, total); value[SPACE] == "" {

						} else if msg := m.Cmd(SPACE, value[SPACE], args, ice.Maps{ice.MSG_DAEMON: ""}); !cli.IsSuccess(msg) {
							list = append(list, value[SPACE])
						}
						count++
					})
					toast(ice.Info.NodeName, count, total)
					if msg := m.Cmd(args, ice.Maps{ice.MSG_DAEMON: ""}); cli.IsSuccess(msg) {
						list = append(list, ice.Info.NodeName)
					}
					return
				})
			}},
			"cmds": {Name: "cmds index* args", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				args := []string{m.Option(ctx.INDEX)}
				kit.If(m.Option(ctx.ARGS), func() { args = append(args, kit.Split(m.Option(ctx.ARGS))...) })
				GoToast(m, "", func(toast func(string, int, int)) (list []string) {
					push := func(space string, msg *ice.Message) {
						if msg.IsErr() {
							list = append(list, space)
						} else {
							msg.Table(func(index int, val ice.Maps, head []string) {
								val[SPACE], head = space, append(head, SPACE)
								m.Push("", val, head)
							})
						}
					}
					msg := m.Cmd("")
					count, total := 0, msg.Length()
					msg.Table(func(value ice.Maps) {
						if toast(value[SPACE], count, total); value[SPACE] != "" {
							push(value[SPACE], m.Cmd(SPACE, value[SPACE], args, ice.Maps{ice.MSG_DAEMON: ""}))
						}
						count++
					})
					toast(ice.Info.NodeName, count, total)
					push("", m.Cmd(args))
					m.StatusTimeCount(ice.SUCCESS, kit.Format("%d/%d", total-len(list), total))
					return
				})
			}},
			"spide": {Help: "导图", Hand: func(m *ice.Message, arg ...string) {
				ctx.DisplayStorySpide(m.Cmdy(""), nfs.DIR_ROOT, ice.Info.NodeName, mdb.FIELD, SPACE, lex.SPLIT, nfs.PT, ctx.ACTION, ice.MAIN)
			}},
			"travel": {Help: "遍历", Hand: func(m *ice.Message, arg ...string) {
				m.Push(mdb.TIME, ice.Info.Make.Time)
				m.Push("md5", ice.Info.Hash)
				m.Push(nfs.SIZE, ice.Info.Size)
				m.Push(nfs.MODULE, ice.Info.Make.Module)
				m.Push(nfs.VERSION, ice.Info.Make.Versions())
				PushPodCmd(m, "", m.ActionKey())
				m.Table(func(value ice.Maps) { kit.If(value[SPACE], func() { mdb.HashCreate(m.Spawn(), kit.Simple(value)) }) })
				m.StatusTimeCount()
			}},
		}, ctx.CmdAction(), mdb.HashAction(mdb.SHORT, SPACE, mdb.FIELD, "time,space,module,version,md5,size", mdb.ACTION, ice.MAIN)), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...).Sort(SPACE); len(arg) > 0 {
				m.EchoIFrame(m.MergePod(arg[0]))
			}
		}},
	})
}
