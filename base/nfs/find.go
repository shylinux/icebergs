package nfs

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const FIND = "find"

func init() {
	const CMD_DIR = "cmd_dir"
	Index.MergeCommands(ice.Commands{
		FIND: {Name: "find word file auto", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
			kit.If(len(arg) == 0, func() { arg = append(arg, "main.go") })
			m.Options(mdb.VALUE, arg[0], CMD_DIR, kit.Select("", arg, 2))
			msg := m.System(FIND, kit.Select(SRC, arg, 1), "-name", arg[0])
			m.Echo(msg.FormatsMeta(nil))
			kit.For(strings.Split(msg.Result(), ice.NL), func(s string) { m.Push(FILE, s) })
			m.StatusTimeCount(kit.Dict(PATH, m.Option(CMD_DIR)))
		}},
	})
}
