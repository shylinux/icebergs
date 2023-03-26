package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	GIT = "git"
)
const REPOS = nfs.REPOS

func init() {
	Index.MergeCommands(ice.Commands{REPOS: {Name: "repos name auto", Actions: ice.Actions{
		web.DREAM_CREATE: {Hand: func(m *ice.Message, arg ...string) {
			kit.If(m.Option(nfs.REPOS), func(p string) { m.Cmd(cli.SYSTEM, GIT, "clone", p, m.Option(cli.CMD_DIR), ice.Maps{cli.CMD_DIR: ""}) })
		}},
		"status": {Hand: func(m *ice.Message, arg ...string) { m.Cmdy("web.code.git.status", arg) }},
	}, Hand: func(m *ice.Message, arg ...string) { m.Cmdy("web.code.git.repos", arg) }}})
}
