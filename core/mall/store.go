package mall

import (
	ice "shylinux.com/x/icebergs"
)

const STORE = "store"

func init() {
	Index.MergeCommands(ice.Commands{
		STORE: {Help: "商店", Hand: func(m *ice.Message, arg ...string) {
		}},
	})
}
