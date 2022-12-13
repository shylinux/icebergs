package vim

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/misc/bash"
	kit "shylinux.com/x/toolkits"
)

const INPUT = "input"

func init() {
	const (
		CMDS = "cmds"
		WUBI = "wubi"
	)
	Index.MergeCommands(ice.Commands{
		INPUT: {Name: "input hash auto export import", Help: "输入法", Actions: mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,hash,type,name,text")},
		web.P(INPUT): {Hand: func(m *ice.Message, arg ...string) {
			if arg[0] == ice.PT {

			} else if strings.Contains(m.Option("pre"), "ice") {
				m.EchoLine(kit.Join(bash.Complete(m, true, kit.Split(strings.Split(m.Option("pre")+m.Option("cmds"), "ice")[1])...), ice.NL))
				// mdb.HashCreate(m.Spawn(), kit.SimpleKV("", CMDS, strings.TrimSpace(arg[0]), m.Result()))
			} else if m.Cmdy(TAGS, INPUT, arg[0], m.Option("pre")); len(m.Result()) > 0 {
				// mdb.HashCreate(m, kit.SimpleKV("", TAGS, arg[0], m.Result()))
			} else if m.Cmdy("web.code.input.wubi", INPUT, arg[0]); len(m.Result()) > 0 {
				// mdb.HashCreate(m.Spawn(), kit.SimpleKV("", WUBI, arg[0], m.Result()))
			}
		}},
	})
}
