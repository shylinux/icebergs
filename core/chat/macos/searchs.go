package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const SEARCHS = "searchs"

func init() {
	Index.MergeCommands(ice.Commands{
		SEARCHS: {Name: "searchs keyword list", Help: "搜索", Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(mdb.SEARCH, mdb.FOREACH, kit.Select("", arg, 0), "ctx,cmd,type,name,text")
		}},
	})
}
