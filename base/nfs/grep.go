package nfs

import ice "shylinux.com/x/icebergs"

const GREP = "grep"

func init() {
	Index.MergeCommands(ice.Commands{
		GREP: {Name: "grep word path auto", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
			m.Option("cmd_dir", arg[1])
			m.Split(m.Cmdx("cli.system", GREP, "--exclude=.[a-z]*", "--exclude-dir=.[a-z]*", "-rni", arg[0], ice.PT), "file:line:text", ice.DF)
		}},
	})
}
