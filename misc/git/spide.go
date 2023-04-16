package git

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const SPIDE = "spide"

func init() {
	Index.MergeCommands(ice.Commands{
		SPIDE: {Name: "spide repos auto", Help: "构架图", Actions: ctx.CmdAction(), Hand: func(m *ice.Message, arg ...string) {
			if len(kit.Slice(arg, 0, 1)) == 0 {
				m.Cmdy(REPOS)
			} else if len(arg) == 1 {
				nfs.DirDeepAll(m, _repos_path(m, arg[0]), "", func(value ice.Maps) { m.Push("", value, []string{nfs.PATH}) }, nfs.PATH)
				m.Options(nfs.DIR_ROOT, _repos_path(m, arg[0])).StatusTimeCount()
				ctx.DisplayStory(m, "", mdb.FIELD, nfs.PATH, aaa.ROOT, arg[0])
			}
		}},
	})
}
