package web

import (
	"net/http"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const ADMIN = "admin"

func init() {
	Index.MergeCommands(ice.Commands{
		ADMIN: {Name: ADMIN, Help: "管理", Hand: func(m *ice.Message, arg ...string) {
			args := []string{}
			kit.If(len(arg) == 0, func() { arg = append(arg, SPACE, DOMAIN) })
			kit.For(arg[1:], func(v string) { args = append(args, ice.ARG, v) })
			m.Cmdy(SPIDE, ice.OPS, SPIDE_RAW, http.MethodGet, CHAT_CMD+arg[0]+"?debug=true", args)
		}},
	})
}
