package web

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func init() {
	const ADMIN = "admin"
	Index.MergeCommands(ice.Commands{
		ADMIN: {Name: ADMIN, Help: "管理", Hand: func(m *ice.Message, arg ...string) {
			args := []string{}
			kit.If(len(arg) == 0, func() { arg = append(arg, SPACE, DOMAIN) })
			kit.For(arg[1:], func(v string) { args = append(args, ice.ARG, v) })
			m.Cmdy(SPIDE, ice.OPS, SPIDE_RAW, CHAT_CMD+arg[0]+"?debug=true", SPIDE_FORM, args)
		}},
	})
}
