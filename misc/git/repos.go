package git

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _repos_path(name string) string {
	if strings.Contains(name, ":\\") {
		return name
	}
	return kit.Select(path.Join(ice.USR, name)+ice.PS, nfs.PWD, name == path.Base(kit.Pwd()))
}
func _repos_cmd(m *ice.Message, name string, arg ...string) *ice.Message {
	return m.Cmdy(cli.SYSTEM, GIT, arg, ice.Option{cli.CMD_DIR, _repos_path(name)})
}
func _repos_insert(m *ice.Message, name string, dir string) {
	if s, e := nfs.StatFile(m, m.Option(cli.CMD_DIR, path.Join(dir, ".git"))); e == nil && s.IsDir() {
		ls := strings.SplitN(strings.Trim(_git_cmds(m, "log", "-n1", `--pretty=format:"%ad %s"`, "--date=iso"), `"`), ice.SP, 4)
		mdb.HashCreate(m, mdb.NAME, name, nfs.PATH, dir,
			COMMIT, kit.Select("", ls, 3), mdb.TIME, strings.Join(ls[:2], ice.SP),
			REMOTE, strings.TrimSpace(_git_cmds(m, REMOTE, "-v")),
			BRANCH, strings.TrimSpace(_git_cmds(m, BRANCH)),
		)
	}
}

const (
	REMOTE = "remote"
	ORIGIN = "origin"
	BRANCH = "branch"
	MASTER = "master"

	CLONE = "clone"
	INIT  = "init"
)
const REPOS = "repos"

func init() {
	Index.MergeCommands(ice.Commands{
		REPOS: {Name: "repos repos path auto create", Help: "代码库", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Conf(REPOS, mdb.HASH, "")
				_repos_insert(m, path.Base(kit.Pwd()), kit.Pwd())
				m.Cmd(nfs.DIR, ice.USR, "name,path", func(value ice.Maps) { _repos_insert(m, value[mdb.NAME], value[nfs.PATH]) })
				cli.IsAlpine(m, GIT)
				cli.IsCentos(m, GIT)
				cli.IsUbuntu(m, GIT)
				m.Config(REPOS, "https://shylinux.com/x")
			}},
			mdb.CREATE: {Name: "create repos branch name path", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.NAME, kit.Select(strings.TrimSuffix(path.Base(m.Option(REPOS)), ".git"), m.Option(mdb.NAME)))
				m.Option(nfs.PATH, kit.Select(path.Join(ice.USR, m.Option(mdb.NAME)), m.Option(nfs.PATH)))
				m.Option(REPOS, kit.Select(m.Config(REPOS)+ice.PS+m.Option(mdb.NAME), m.Option(REPOS)))

				_repos_insert(m, m.Option(mdb.NAME), m.Option(nfs.PATH))
				if s, e := nfs.StatFile(m, path.Join(m.Option(nfs.PATH), ".git")); e == nil && s.IsDir() {
					return
				}

				// 下载仓库
				if s, e := nfs.StatFile(m, m.Option(nfs.PATH)); e == nil && s.IsDir() {
					m.Option(cli.CMD_DIR, m.Option(nfs.PATH))
					_git_cmd(m, INIT)
					_git_cmd(m, REMOTE, ADD, ORIGIN, m.Option(REPOS))
					_git_cmd(m, PULL, ORIGIN, kit.Select(MASTER, m.Option(BRANCH)))
				} else {
					m.Option(cli.CMD_DIR, "")
					_git_cmd(m, CLONE, "-b", kit.Select(MASTER, m.Option(BRANCH)), m.Option(REPOS), m.Option(nfs.PATH))
				}
			}},
			web.DREAM_OPEN: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("web.code.git.repos", mdb.CREATE, m.OptionSimple(nfs.REPOS), nfs.PATH, m.Option(nfs.PATH))
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,branch,commit,remote")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 { // 仓库列表
				mdb.HashSelect(m, arg...).Sort(mdb.NAME).RenameAppend(mdb.NAME, REPOS)
			} else { // 文件列表
				m.Cmdy(nfs.DIR, kit.Select("", arg, 1), "time,line,path", kit.Dict(nfs.DIR_ROOT, _repos_path(arg[0])))
			}
		}},
	})
}
