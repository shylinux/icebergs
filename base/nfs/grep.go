package nfs

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const GREP = "grep"

func init() {
	const CMD_DIR = "cmd_dir"
	Index.MergeCommands(ice.Commands{
		GREP: {Name: "grep word file path auto", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
			m.Options(mdb.VALUE, arg[0], CMD_DIR, kit.Select("", arg, 2))
			for _, line := range strings.Split(m.Cmdx("cli.system", GREP, "--exclude=.[a-z]*", "--exclude-dir=.[a-z]*", "--exclude-dir=node_modules", "-rni", arg[0], kit.Select(ice.PT, arg, 1)), ice.NL) {
				if ls := strings.SplitN(line, ice.DF, 3); len(ls) > 2 {
					m.Push(FILE, strings.TrimPrefix(ls[0], PWD)).Push(LINE, ls[1]).Push(mdb.TEXT, ls[2])
				}
			}
			m.StatusTimeCount(PATH, m.Option(CMD_DIR))
		}},
	})
}
