package bash

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

const SYNC = "sync"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SYNC: {Name: SYNC, Help: "同步流", Value: kit.Data(
			mdb.FIELD, "time,id,type,name,text,pwd,username,hostname",
		)},
	}, Commands: map[string]*ice.Command{
		"/sync": {Name: "/sync", Help: "同步", Action: map[string]*ice.Action{
			"history": {Name: "history", Help: "历史", Hand: func(m *ice.Message, arg ...string) {
				ls := strings.SplitN(strings.TrimSpace(m.Option(ARG)), ice.SP, 4)
				if text := strings.TrimSpace(strings.Join(ls[3:], ice.SP)); text != "" {
					m.Cmd(SYNC, mdb.INSERT, mdb.TIME, ls[1]+ice.SP+ls[2],
						mdb.TYPE, "shell", mdb.NAME, ls[0], mdb.TEXT, text,
						m.OptionSimple(cli.PWD, aaa.USERNAME, tcp.HOSTNAME, tcp.HOSTNAME))
				}
			}},
		}},
		SYNC: {Name: "sync id auto page export import", Help: "同步流", Action: ice.MergeAction(map[string]*ice.Action{
			cli.SYSTEM: {Name: "system", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_DIR, m.Option(cli.PWD))
				m.ProcessCommand(cli.SYSTEM, kit.Split(m.Option(mdb.TEXT)), arg...)
				m.ProcessCommandOpt(arg, cli.PWD)
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(FAVOR, mdb.INPUTS, arg)
			}},
			FAVOR: {Name: "favor zone=some@key type name text pwd", Help: "收藏", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(FAVOR, mdb.INSERT)
			}},
		}, mdb.ListAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.OptionPage(kit.Slice(arg, 1)...)
			mdb.ListSelect(m, kit.Slice(arg, 0, 1)...)
			m.PushAction(cli.SYSTEM, FAVOR)
			m.StatusTimeCountTotal(m.Config(mdb.COUNT))
		}},
	}})
}
