package web

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func init() {
	Index.MergeCommands(ice.Commands{
		"admin": {Name: "admin", Help: "管理", Hand: func(m *ice.Message, arg ...string) {
			args := []string{}
			kit.For(arg[1:], func(v string) { args = append(args, ice.ARG, v) })
			m.Cmdy(SPIDE, ice.OPS, SPIDE_RAW, "/chat/cmd/"+arg[0]+"?debug=true", SPIDE_FORM, args)
		}},
	})
}
