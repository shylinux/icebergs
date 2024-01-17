package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	GIT        = "git"
	REMOTE     = "remote"
	REMOTE_URL = "remoteURL"
	INIT       = "init"
	ADD        = "add"
)
const REPOS = nfs.REPOS

func init() {
	Index.MergeCommands(ice.Commands{
		REPOS: {Name: "repos repos branch commit file auto", Actions: ice.Actions{
			mdb.STATUS: {Help: "状态", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(web.CODE_GIT_STATUS, arg) }},
		}, Hand: func(m *ice.Message, arg ...string) { m.Cmdy(web.CODE_GIT_REPOS, arg) }},
	})
}
func GitVersion(m *ice.Message) string { return m.Cmdx(cli.SYSTEM, GIT, VERSION) }
func ReposAddFile(m *ice.Message, dir string, file string) {
	m.Cmd("web.code.git.repos", "add", kit.Dict(nfs.REPOS, path.Base(kit.Select(kit.Path(""), dir)), nfs.FILE, file))
}
