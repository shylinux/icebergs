package web

import (
	"net/http"
	"path"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const ADMIN = "admin"

func init() {
	Index.MergeCommands(ice.Commands{
		ADMIN: {Name: "admin index list", Help: "管理", Hand: func(m *ice.Message, arg ...string) {
			kit.If(len(arg) == 0, func() { arg = append(arg, SPACE, DOMAIN) })
			m.Cmdy(SPIDE, ice.OPS, SPIDE_RAW, http.MethodPost, CHAT_CMD+path.Join(arg...), "pwd", kit.Path(""))
		}},
	})
}
func AdminCmd(m *ice.Message, cmd string) string {
	return m.Cmdx(SPIDE, ice.OPS, SPIDE_RAW, http.MethodGet, C(cmd))
}
