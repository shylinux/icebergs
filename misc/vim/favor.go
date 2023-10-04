package vim

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

const FAVOR = "favor"

func init() {
	Index.MergeCommands(ice.Commands{
		FAVOR: {Name: "favor zone id auto insert", Help: "收藏夹", Actions: ice.MergeActions(ice.Actions{
			code.INNER: {Hand: func(m *ice.Message, arg ...string) {
				ls := nfs.SplitPath(m, path.Join(kit.Select("", m.Option(cli.PWD), !path.IsAbs(m.Option(nfs.FILE))), m.Option(nfs.FILE)))
				ctx.ProcessField(m, "", []string{ls[0], ls[1], m.Option(nfs.LINE)}, arg...)
			}},
		}, mdb.ZoneAction(mdb.FIELDS, "time,id,type,name,text,file,line,pwd")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.Action(mdb.EXPORT, mdb.IMPORT)
			} else {
				m.PushAction(code.INNER)
			}
		}},
		web.PP(FAVOR): {Actions: ice.Actions{
			mdb.INSERT: {Hand: func(m *ice.Message, arg ...string) { m.Cmd(FAVOR, mdb.INSERT) }},
			mdb.SELECT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(FAVOR, func(value ice.Maps) { m.EchoLine(value[mdb.ZONE]) })
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(FAVOR, m.Option(mdb.ZONE), func(value ice.Maps) {
				m.EchoLine(m.Option(mdb.ZONE)).EchoLine("%v:%v:%v:(%v): %v", value[nfs.FILE], value[nfs.LINE], "1", value[mdb.NAME], value[mdb.TEXT])
			})
		}},
	})
}
