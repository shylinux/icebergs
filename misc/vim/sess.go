package vim

import (
	ice "shylinux.com/x/icebergs"
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
	Index.MergeCommands(ice.Commands{"_login": {Hand: func(m *ice.Message, arg ...string) { m.Cmdy("web.code.bash._login", arg) }}})
}
