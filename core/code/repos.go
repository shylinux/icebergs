package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
)

const (
	GIT = "git"
)
const REPOS = nfs.REPOS

func init() {
	Index.MergeCommands(ice.Commands{
		REPOS: {Name: "repos repos branch commit file auto", Actions: ice.Actions{
			mdb.STATUS: {Help: "状态", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(web.CODE_GIT_STATUS, arg) }},
		}, Hand: func(m *ice.Message, arg ...string) { m.Cmdy(web.CODE_GIT_REPOS, arg) }},
	})
}
