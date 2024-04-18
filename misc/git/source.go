package git

import (
	git "shylinux.com/x/go-git/v5"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
)

func init() {
	Index.MergeCommands(ice.Commands{
		nfs.SOURCE: {Name: "source repos path file auto", Actions: ice.Actions{
			CLONE: {Hand: func(m *ice.Message, arg ...string) {
				if _, err := git.PlainClone(m.Option(nfs.PATH), false, &git.CloneOptions{URL: m.Option(REPOS)}); !m.WarnNotValid(err) {
					return
				}
				m.ProcessRefresh()
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 3 || arg[0] == ctx.ACTION {
				m.Cmdy(web.CODE_INNER, arg)
			} else if !nfs.Exists(m, arg[1]) {
				m.EchoInfoButton("please clone repos", CLONE)
			} else {
				m.Options(nfs.PATH, arg[1], nfs.FILE, arg[2])
				m.Cmdy(web.CODE_INNER, arg[1:])
			}
		}},
	})
}
