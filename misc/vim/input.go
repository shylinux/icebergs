package vim

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const INPUT = "input"

func init() {
	const (
		CMDS = "cmds"
		WUBI = "wubi"
		ICE_ = "ice "
	)
	Index.MergeCommands(ice.Commands{
		INPUT: {Name: "input hash auto export import", Help: "输入法", Actions: mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,hash,type,name,text")},
		web.P(INPUT): {Hand: func(m *ice.Message, arg ...string) {
			if arg[0] == ice.PT {

			} else if strings.HasPrefix(arg[0], "ice") {
				args := kit.Split(arg[0])
				if list := ctx.CmdList(m.Spawn()).Appendv(ctx.INDEX); len(args) == 1 || kit.IndexOf(list, args[1]) == -1 {
					if len(args) > 1 {
						list = kit.Simple(list, func(item string) bool { return strings.HasPrefix(item, args[1]) })
					}
					m.Echo(ICE_ + kit.Join(list, ice.NL+ICE_)).Echo(ice.NL)
					return
				}
				msg := m.Cmd(args[1:])
				if msg.IsErrNotFound() {
					msg.SetResult().Cmdy(cli.SYSTEM, args[1:])
				}
				if msg.Result() == "" {
					msg.Table()
				}
				m.Echo(arg[0]).Echo(ice.NL).Search(args[1], func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
					m.Echo(arg[0] + ice.SP + kit.Join(msg.Appendv(kit.Format(kit.Value(cmd.List, kit.Keys(len(args)-2, mdb.NAME)))), ice.NL+arg[0]+ice.SP)).Echo(ice.NL)
				}).Copy(msg)
				kit.Fetch(kit.UnMarshal(msg.Option(ice.MSG_STATUS)), func(index int, value ice.Map) { m.Echo("%s: %v ", value[mdb.NAME], value[mdb.VALUE]) })
				mdb.HashCreate(m.Spawn(), kit.SimpleKV("", CMDS, strings.TrimSpace(strings.Join(args[1:], ice.SP)), m.Result()))
			} else if m.Cmdy(TAGS, INPUT, arg[0], m.Option("pre")); m.Length() > 0 {
				mdb.HashCreate(m, kit.SimpleKV("", TAGS, arg[0], m.Result()))
			} else if m.Cmdy("web.code.input.wubi", INPUT, arg[0]); len(m.Result()) > 0 {
				mdb.HashCreate(m.Spawn(), kit.SimpleKV("", WUBI, arg[0], m.Result()))
			}
		}},
	})
}
