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
	Index.MergeCommands(ice.Commands{REPOS: {Name: "repos name auto", Actions: ice.Actions{
		"status": {Hand: func(m *ice.Message, arg ...string) { m.Cmdy("web.code.git.status", arg) }},
	}, Hand: func(m *ice.Message, arg ...string) { m.Cmdy("web.code.git.repos", arg) }}})
}
