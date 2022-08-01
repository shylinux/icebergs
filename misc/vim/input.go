package vim

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const INPUT = "input"

func init() {
	Index.MergeCommands(ice.Commands{
		"/input": {Name: "/input", Help: "输入法", Hand: func(m *ice.Message, arg ...string) {
			if m.Cmdy(TAGS, ctx.ACTION, INPUT, arg[0], m.Option("pre")); m.Length() > 0 {
				m.Cmd(m.PrefixKey(), mdb.INSERT, kit.SimpleKV("", "tags", arg[0], m.Result()))
				return // 代码补全
			}
			if arg[0] == ice.PT {
				return
			}
			if m.Cmdy("web.code.input.wubi", ctx.ACTION, INPUT, arg[0]); m.Length() > 0 {
				m.Cmd(m.PrefixKey(), mdb.INSERT, kit.SimpleKV("", "wubi", arg[0], m.Result()))
				return // 五笔输入
			}
			if arg[0] = strings.TrimSpace(arg[0]); strings.HasPrefix(arg[0], "ice") {
				list := kit.Split(arg[0])
				if m.Cmdy(list[1:]); m.IsErrNotFound() {
					m.SetResult().Cmdy(cli.SYSTEM, list[1:])
				}
				if len(m.Resultv()) == 0 {
					m.Table()
				}
				m.Cmd(m.PrefixKey(), mdb.INSERT, kit.SimpleKV("", "cmds", strings.TrimSpace(strings.Join(list[1:], ice.SP)), m.Result()))
				return // 本地命令
			}
		}},
		INPUT: {Name: "input id auto export import", Help: "输入法", Actions: mdb.ListAction(mdb.FIELD, "time,id,type,name,text")},
	})
}
