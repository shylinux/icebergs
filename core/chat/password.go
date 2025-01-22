package chat

import ice "shylinux.com/x/icebergs"

func init() {
	Index.MergeCommands(ice.Commands{
		"password": {Name: "password", Hand: func(m *ice.Message, arg ...string) {
			m.Echo("hello world")
		}},
	})
}
