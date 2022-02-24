package nfs

import ice "shylinux.com/x/icebergs"

const GREP = "grep"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		GREP: {Name: "grep path word auto", Help: "搜索", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option("cmd_dir", arg[0])
			m.Split(m.Cmdx("cli.system", GREP, "--exclude=.[a-z]*", "--exclude-dir=.[a-z]*", "-rni", arg[1]), "file:line:text", ":")
		}},
	}})
}