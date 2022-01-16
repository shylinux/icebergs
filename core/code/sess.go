package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"sess": {Name: "sess hash auto save load", Help: "会话", Action: mdb.HashAction(
			mdb.FIELD, "time,hash,tabs,tool",
			"action", "load",
		)},
	}})
}
