package git

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/nfs"
)

const SPIDES = "spides"

func init() {
	Index.MergeCommands(ice.Commands{
		SPIDES: {Name: "spides repos auto", Help: "构架图", Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(REPOS)
			} else if p := _repos_path(m, arg[0]); len(arg) == 1 {
				nfs.DirDeepAll(m, p, "", nil, nfs.PATH).Options(nfs.DIR_ROOT, p+nfs.PS)
				ctx.DisplayStory(m, "")
			}
		}},
	})
}
