package nfs

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const GREP = "grep"

func init() {
	Index.MergeCommands(ice.Commands{
		GREP: {Name: "grep word path auto", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
			m.Option("cmd_dir", kit.Select(m.Option(PATH), arg, 1))
			for _, line := range strings.Split(m.Cmdx("cli.system", GREP, "--exclude=.[a-z]*", "--exclude-dir=.[a-z]*", "-rni", arg[0], ice.PT), ice.NL) {
				if ls := strings.SplitN(line, ice.DF, 3); len(ls) > 2 {
					m.Push(FILE, strings.TrimPrefix(ls[0], PWD)).Push(LINE, ls[1]).Push(mdb.TEXT, ls[2])
				}
			}
			m.StatusTimeCount(PATH, m.Option("cmd_dir"))
		}},
	})
}
