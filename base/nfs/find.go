package nfs

import (
	"strings"

	ice "shylinux.com/x/icebergs"
)

const FIND = "find"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		FIND: {Name: "find path word auto", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
			for _, file := range strings.Split(m.Cmdx("cli.system", FIND, PWD, "-name", arg[1]), ice.NL) {
				m.Push(FILE, strings.TrimPrefix(file, PWD))
			}
		}},
	}})
}
