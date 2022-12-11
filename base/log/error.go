package log

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
)

const ERROR = "error"

func init() {
	Index.MergeCommands(ice.Commands{
		ERROR: {Name: "error auto", Help: "错误", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(nfs.CAT, path.Join(ice.VAR_LOG, "error.log"))
		}},
	})
}
