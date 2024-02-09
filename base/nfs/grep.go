package nfs

import (
	"path"
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
			m.Options(mdb.VALUE, arg[0], CMD_DIR, kit.Select("", arg, 2))
			kit.For(strings.Split(m.System(GREP, "--exclude=.[a-z]*", "--exclude-dir=.[a-z]*", "-rni", arg[0], path.Join(kit.Select(SRC, arg, 1))).Result(), ice.NL), func(s string) {
				if ls := strings.SplitN(s, DF, 3); len(ls) > 2 {
					m.Push(FILE, strings.TrimPrefix(ls[0], PWD)).Push(LINE, ls[1]).Push(mdb.TEXT, ls[2])
				}
			})
			m.StatusTimeCount(kit.Dict(PATH, m.Option(CMD_DIR)))
		}},
	})
}
