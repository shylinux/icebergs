package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
)

const (
	GIT = "git"
)
const REPOS = nfs.REPOS

func init() {
	const GIT_REPOS = "web.code.git.repos"
	Index.MergeCommands(ice.Commands{
		REPOS: {Name: "repos name auto", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(GIT_REPOS, arg) }},
	})
}
