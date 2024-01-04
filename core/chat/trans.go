package chat

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const TRANS = "trans"

func init() {
	const (
		SEND = "send"
		FROM = "from"
		TO   = "to"
	)
	Index.MergeCommands(ice.Commands{
		TRANS: {Name: "trans from@key to@key auto", Help: "文件传输", Icon: "Migration.png", Actions: ice.Actions{
			SEND: {Hand: func(m *ice.Message, arg ...string) {
				defer web.ToastProcess(m)()
				p := web.ProxyUpload(m, m.Option(FROM), m.Option(nfs.PATH))
				h := m.Cmdx(web.SHARE, mdb.CREATE, mdb.TYPE, web.DOWNLOAD, mdb.TEXT, p)
				defer m.Cmd(web.SHARE, mdb.REMOVE, mdb.HASH, h)
				m.Cmdy(web.SPACE, m.Option(TO), web.SPIDE, ice.DEV, web.SPIDE_SAVE, path.Join(m.Option("to_path"), path.Base(m.Option(nfs.PATH))), m.MergeLink(web.PP(web.SHARE, h)))
			}},
			ctx.RUN: {Hand: func(m *ice.Message, arg ...string) {
				m.Options(ice.MSG_USERPOD, m.Option(FROM), ice.POD, m.Option(FROM))
				kit.If(!ctx.PodCmd(m, arg) && aaa.Right(m, arg), func() { m.Cmdy(arg) })
				kit.If(arg[0] == nfs.DIR && len(arg) < 3, func() { m.PushAction(SEND, mdb.SHOW, nfs.TRASH) })
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			defer m.Options(ice.MSG_ACTION, "")
			if len(arg) == 0 {
				m.Cmdy(web.SPACE).RenameAppend(mdb.NAME, FROM).Toast("请选择空间1")
			} else if len(arg) == 1 {
				m.Cmdy(web.SPACE).RenameAppend(mdb.NAME, TO).Toast("请选择空间2")
			} else {
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}
