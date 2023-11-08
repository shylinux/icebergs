package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/chat"
)

const BMAP = "bmap"

func init() {
	chat.Index.MergeCommands(ice.Commands{
		BMAP: {Help: "百度地图", Hand: func(m *ice.Message, arg ...string) {
			m.Display("", nfs.SCRIPT, mdb.Config(m, nfs.SCRIPT))
		}},
	})
}
