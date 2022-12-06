package git

import (
	"os"
	"path"
	"strings"

	"shylinux.com/x/gogit"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _repos_path(name string) string {
	if strings.Contains(name, ":\\") {
		return name
	}
	return kit.Select(path.Join(ice.USR, name)+ice.PS, nfs.PWD, name == path.Base(kit.Pwd()))
}
func _repos_cmd(m *ice.Message, name string, arg ...string) *ice.Message {
	m.Option(cli.CMD_DIR, _repos_path(name))
	return m.Copy(_git_cmd(m, arg...))
}
func _repos_init(m *ice.Message, p string) string {
	os.MkdirAll(path.Join(p, "refs/heads/"), ice.MOD_DIR)
	os.MkdirAll(path.Join(p, "objects/info/"), ice.MOD_DIR)
	m.Cmd(nfs.SAVE, path.Join(p, "HEAD"), "ref: refs/heads/master")
	return p
}
func _repos_insert(m *ice.Message, name string, path string) bool {
	if repos, e := gogit.OpenRepository(_git_dir(path)); e == nil {
		if ci, e := repos.GetCommit(); e == nil {
			mdb.HashCreate(m, REPOS, name, nfs.PATH, path, mdb.TIME, ci.Author.When.Format(ice.MOD_TIME), COMMIT, strings.TrimSpace(ci.Message),
				BRANCH, repos.GetBranch(), ORIGIN, kit.Select("", kit.Slice(kit.Split(repos.GetOrigin()), -1), 0))
		} else {
			mdb.HashCreate(m, REPOS, name, nfs.PATH, path, mdb.TIME, m.Time(),
				BRANCH, repos.GetBranch(), ORIGIN, kit.Select("", kit.Slice(kit.Split(repos.GetOrigin()), -1), 0))
		}
		return true
	}
	return false
}

const (
	ORIGIN = "origin"
	BRANCH = "branch"
	MASTER = "master"
	INIT   = "init"
)
const REPOS = "repos"

func init() {
	Index.MergeCommands(ice.Commands{
		REPOS: {Name: "repos repos path auto create", Help: "代码库", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DIR, ice.USR, "name,path", func(value ice.Maps) { _repos_insert(m, value[mdb.NAME], value[nfs.PATH]) })
				_repos_insert(m, path.Base(kit.Pwd()), kit.Pwd())
				cli.IsSystem(m, GIT)
			}},
			INIT: {Hand: func(m *ice.Message, arg ...string) {
				if dir := _repos_init(m, _git_dir(m.Option(cli.CMD_DIR))); m.Option(ORIGIN, kit.Select("", kit.Split(m.Option(ORIGIN)), -1)) != "" {
					m.Cmd(nfs.SAVE, path.Join(dir, "config"), kit.Format(_repos_config, m.Option(ORIGIN)))
					_git_cmd(m, PULL, ORIGIN, m.OptionDefault(BRANCH, MASTER))
				}
			}},
			mdb.CREATE: {Name: "create origin name path", Hand: func(m *ice.Message, arg ...string) {
				m.OptionDefault(mdb.NAME, strings.TrimSuffix(path.Base(m.Option(ORIGIN)), ".git"))
				m.OptionDefault(nfs.PATH, path.Join(ice.USR, m.Option(mdb.NAME)))
				if _repos_insert(m, m.Option(mdb.NAME), m.Option(nfs.PATH)) {
					return
				}
				m.Cmd("", INIT, kit.Dict(cli.CMD_DIR, m.Option(nfs.PATH)))
				_repos_insert(m, m.Option(mdb.NAME), m.Option(nfs.PATH))
			}},
		}, mdb.HashAction(mdb.SHORT, REPOS, mdb.FIELD, "time,repos,branch,commit,origin"), mdb.ClearHashOnExitAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m, arg...)
			} else {
				m.Cmdy(nfs.CAT, kit.Select(nfs.PWD, arg, 1), "time,line,path", kit.Dict(nfs.DIR_ROOT, _repos_path(arg[0])))
			}
		}},
	})
}
func ReposList(m *ice.Message) *ice.Message { return m.Cmd(REPOS, ice.OptionFields("repos,path")) }

var _repos_config = `
[remote "origin"]
	url = %s
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "master"]
	remote = origin
	merge = refs/heads/master
`
