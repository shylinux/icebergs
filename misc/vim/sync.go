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

const SYNC = "sync"

func init() {
	Index.MergeCommands(ice.Commands{
		SYNC: {Name: "sync id auto page export import", Help: "同步流", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(FAVOR, mdb.INPUTS, arg) }},
			code.INNER: {Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 && arg[0] == ice.RUN {
					ctx.ProcessField(m, "", nil, arg...)
				} else {
					p := path.Join(m.Option(cli.PWD), m.Option(BUF))
					ctx.ProcessField(m, "", []string{path.Dir(p) + ice.PS, path.Base(p), m.Option(ROW)})
				}
			}},
			FAVOR: {Name: "favor zone*=some type name text pwd buf row", Help: "收藏", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(FAVOR, mdb.INSERT, m.OptionSimple(mdb.ZONE, "type,name,text,pwd"), nfs.FILE, m.Option(BUF), nfs.LINE, m.Option(ROW))
			}},
		}, mdb.PageListAction(mdb.FIELD, "time,id,type,name,text,pwd,buf,row,col")), Hand: func(m *ice.Message, arg ...string) {
			mdb.PageListSelect(m, arg...).PushAction(code.INNER, FAVOR)
		}},
		web.P(SYNC): {Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(SYNC, mdb.INSERT, mdb.TYPE, VIMRC, mdb.NAME, arg[0], mdb.TEXT, kit.Select(m.Option(ARG), m.Option(SUB)), m.OptionSimple(cli.PWD, BUF, ROW, COL))
		}},
	})
}
