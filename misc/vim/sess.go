package vim

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
)

const (
	SID = "sid"
	ARG = "arg"
	SUB = "sub"
	PRE = "pre"
	PWD = "pwd"
	BUF = "buf"
	ROW = "row"
	COL = "col"
)
const SESS = "sess"

func init() {
	Index.MergeCommands(ice.Commands{web.WEB_LOGIN: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy("web.code.bash._login", arg) }}})
}
