package chat

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	SEND = "send"
	FROM = "from"
	TO   = "to"
)

func init() {
	Index.MergeCommands(ice.Commands{
		TRANS: {Name: "trans from to auto", Help: "传输", Actions: ice.MergeAction(ice.Actions{
			SEND: {Name: "send", Help: "发送", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.SPACE, m.Option(TO), web.SPIDE, ice.DEV, web.SPIDE_SAVE, kit.Select(nfs.PWD, m.Option("to_path")),
					m.MergeURL2(path.Join(web.SHARE_LOCAL, m.Option("from_path")), ice.POD, m.Option(FROM),
						web.SHARE, m.Cmdx(web.SHARE, mdb.CREATE, mdb.TYPE, web.LOGIN),
					),
				)
				m.Toast(ice.SUCCESS, SEND)
				m.ProcessHold()
			}},
			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				m.Option(ice.POD, m.Option("_pod"))
				m.Option(ice.MSG_USERPOD, m.Option("_pod"))
				if m.Right(arg) && !m.PodCmd(arg) {
					m.Cmdy(arg)
				}
				if arg[0] == nfs.DIR && m.Length() > 0 {
					m.PushAction(SEND, nfs.TRASH)
				}
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(web.ROUTE).RenameAppend(web.ROUTE, FROM)
				return
			}
			if len(arg) == 1 {
				m.Cmdy(web.ROUTE).RenameAppend(web.ROUTE, TO)
				return
			}
			m.DisplayLocal("")
		}},
	})
}
