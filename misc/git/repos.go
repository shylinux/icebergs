package git

import (
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _repos_cmd(m *ice.Message, name string, arg ...string) *ice.Message {
	m.Cmdy(cli.SYSTEM, GIT, arg, ice.Option{cli.CMD_DIR, _repos_path(name)})
	return m
}
func _repos_path(name string) string {
	if strings.Contains(name, ":\\") {
		return name
	}
	return kit.Select(path.Join(ice.USR, name)+ice.PS, nfs.PWD, name == path.Base(kit.Pwd()))
}
func _repos_insert(m *ice.Message, name string, dir string) {
	if s, e := os.Stat(m.Option(cli.CMD_DIR, path.Join(dir, ".git"))); e == nil && s.IsDir() {
		ls := strings.SplitN(strings.Trim(m.Cmdx(cli.SYSTEM, GIT, "log", "-n1", `--pretty=format:"%ad %s"`, "--date=iso"), `"`), ice.SP, 4)
		m.Rich(REPOS, nil, kit.Data(mdb.NAME, name, nfs.PATH, dir,
			COMMIT, kit.Select("", ls, 3), mdb.TIME, strings.Join(ls[:2], ice.SP),
			BRANCH, strings.TrimSpace(m.Cmdx(cli.SYSTEM, GIT, BRANCH)),
			REMOTE, strings.TrimSpace(m.Cmdx(cli.SYSTEM, GIT, REMOTE, "-v")),
		))
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
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		REPOS: {Name: REPOS, Help: "代码库", Value: kit.Data(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,branch,commit,remote",
			REPOS, "https://shylinux.com/x", nfs.PATH, ice.USR_LOCAL,
		)},
	}, Commands: map[string]*ice.Command{
		REPOS: {Name: "repos repos path auto create", Help: "代码库", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Conf(REPOS, mdb.HASH, "")
				_repos_insert(m, path.Base(kit.Pwd()), kit.Pwd())
				m.Cmd(nfs.DIR, ice.USR, "name,path").Table(func(index int, value map[string]string, head []string) {
					_repos_insert(m, value[mdb.NAME], value[nfs.PATH])
				})
				cli.IsAlpine(m, "git", "apk add git")
			}},
			mdb.CREATE: {Name: "create repos branch name path", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.NAME, kit.Select(strings.TrimSuffix(path.Base(m.Option(REPOS)), ".git"), m.Option(mdb.NAME)))
				m.Option(nfs.PATH, kit.Select(path.Join(ice.USR, m.Option(mdb.NAME)), m.Option(nfs.PATH)))
				m.Option(REPOS, kit.Select(m.Config(REPOS)+ice.PS+m.Option(mdb.NAME), m.Option(REPOS)))
				if strings.Contains(m.Option(REPOS), "@") {
					ls := strings.Split(m.Option(REPOS), "@")
					m.Option(REPOS, ls[0])
					m.Option(BRANCH, ls[1])
				}

				_repos_insert(m, m.Option(mdb.NAME), m.Option(nfs.PATH))
				if s, e := os.Stat(path.Join(m.Option(nfs.PATH), ".git")); e == nil && s.IsDir() {
					return
				}

				// 下载仓库
				if s, e := os.Stat(m.Option(nfs.PATH)); e == nil && s.IsDir() {
					m.Option(cli.CMD_DIR, m.Option(nfs.PATH))
					m.Cmd(cli.SYSTEM, GIT, INIT)
					m.Cmd(cli.SYSTEM, GIT, REMOTE, ADD, ORIGIN, m.Option(REPOS))
					m.Cmd(cli.SYSTEM, GIT, PULL, ORIGIN, MASTER)
				} else {
					m.Option(cli.CMD_DIR, "")
					m.Cmd(cli.SYSTEM, GIT, CLONE, "-b", kit.Select(MASTER, m.Option(BRANCH)),
						m.Option(REPOS), m.Option(nfs.PATH))
				}
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 { // 仓库列表
				mdb.HashSelect(m, arg...).Sort(mdb.NAME).RenameAppend(mdb.NAME, REPOS)
				return
			}
			m.Cmdy(nfs.DIR, kit.Select("", arg, 1), "time,line,path", kit.Dict(nfs.DIR_ROOT, _repos_path(arg[0])))
		}},
	}})
}
