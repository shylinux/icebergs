package nfs

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	OPENS = "opens"
)

const GREP = "grep"

func init() {
	const CMD_DIR = "cmd_dir"
	Index.MergeCommands(ice.Commands{
		GREP: {Name: "grep word file auto", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
			kit.If(len(arg) == 0, func() { arg = append(arg, ice.MAIN) })
			kit.If(len(arg) == 1, func() { arg = append(arg, ice.SRC) })
			m.Options(mdb.VALUE, arg[0])
			kit.For(kit.SplitLine(m.System(GREP, "--exclude=.[a-z]*", "--exclude-dir=.[a-z]*", "-rni", arg[0], kit.AddUniq([]string{}, arg[1:]...)).Result()), func(s string) {
				if ls := strings.SplitN(s, DF, 3); len(ls) > 2 {
					_ls := SplitPath(m, ls[0])
					m.Push(PATH, _ls[0]).Push(FILE, _ls[1]).Push(LINE, ls[1]).Push(mdb.TEXT, ls[2])
				}
			})
			m.Sort("path,file,line")
		}},
	})
}
