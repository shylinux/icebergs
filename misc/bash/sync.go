package bash

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const SYNC = "sync"

func init() {
	const (
		HISTORY = "history"
		SHELL   = "shell"
	)
	Index.MergeCommands(ice.Commands{
		SYNC: {Name: "sync id auto page export import", Help: "同步流", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(FAVOR, mdb.INPUTS, arg) }},
			cli.SYSTEM: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 && arg[0] == ice.RUN {
					if msg := mdb.ListSelect(m.Spawn(), m.Option(mdb.ID)); nfs.ExistsFile(m, msg.Append(cli.PWD)) {
						m.Option(cli.CMD_DIR, msg.Append(cli.PWD))
					}
					ctx.ProcessField(m, "", nil, arg...)
				} else {
					ctx.ProcessField(m, "", kit.Split(m.Option(mdb.TEXT)))
				}
			}},
			FAVOR: {Name: "favor zone=demo type name text pwd", Help: "收藏", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(FAVOR, mdb.INSERT, arg, m.OptionSimple(aaa.USERNAME, tcp.HOSTNAME))
			}},
		}, mdb.PageListAction(mdb.FIELD, "time,id,type,name,text,pwd,username,hostname")), Hand: func(m *ice.Message, arg ...string) {
			mdb.PageListSelect(m, arg...).PushAction(cli.SYSTEM, FAVOR)
		}},
		web.PP(SYNC): {Actions: ice.Actions{
			HISTORY: {Hand: func(m *ice.Message, arg ...string) {
				ls := strings.SplitN(strings.TrimSpace(m.Option(ARG)), ice.SP, 4)
				if text := strings.TrimSpace(strings.Join(ls[3:], ice.SP)); text != "" {
					m.Cmd(SYNC, mdb.INSERT, mdb.TIME, ls[1]+ice.SP+ls[2], mdb.TYPE, SHELL, mdb.NAME, ls[0], mdb.TEXT, text, m.OptionSimple(cli.PWD, aaa.USERNAME, tcp.HOSTNAME))
				}
			}},
		}},
	})
}
