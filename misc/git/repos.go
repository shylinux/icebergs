package git

import (
	"fmt"
	"os"
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"
)

func _repos_path(name string) string {
	if strings.Contains(name, ":\\") {
		return name
	}
	return kit.Select(path.Join(kit.SSH_USR, name), "./", name == path.Base(kit.Pwd()))
}
func _repos_insert(m *ice.Message, name string, dir string) {
	if s, e := os.Stat(m.Option(cli.CMD_DIR, path.Join(dir, ".git"))); e == nil && s.IsDir() {
		ls := strings.SplitN(strings.Trim(m.Cmdx(cli.SYSTEM, GIT, "log", "-n1", `--pretty=format:"%ad %s"`, "--date=iso"), `"`), " ", 4)
		m.Rich(REPOS, nil, kit.Data(kit.MDB_NAME, name, kit.MDB_PATH, dir,
			COMMIT, kit.Select("", ls, 3), kit.MDB_TIME, strings.Join(ls[:2], " "),
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

	INIT = "init"
)
const REPOS = "repos"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			REPOS: {Name: REPOS, Help: "代码库", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, kit.SSH_REPOS, "https://github.com/shylinux",
			)},
		},
		Commands: map[string]*ice.Command{
			REPOS: {Name: "repos name path auto create proxy", Help: "代码库", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create repos branch name path", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Option(kit.MDB_NAME, kit.Select(strings.TrimSuffix(path.Base(m.Option(kit.SSH_REPOS)), ".git"), m.Option(kit.MDB_NAME)))
					m.Option(kit.MDB_PATH, kit.Select(path.Join(kit.SSH_USR, m.Option(kit.MDB_NAME)), m.Option(kit.MDB_PATH)))
					m.Option(kit.SSH_REPOS, kit.Select(m.Conf(REPOS, kit.Keym(kit.SSH_REPOS))+"/"+m.Option(kit.MDB_NAME), m.Option(kit.SSH_REPOS)))

					if s, e := os.Stat(path.Join(m.Option(kit.MDB_PATH), ".git")); e == nil && s.IsDir() {
						return
					}

					// 下载仓库
					if s, e := os.Stat(m.Option(kit.MDB_PATH)); e == nil && s.IsDir() {
						m.Option(cli.CMD_DIR, m.Option(kit.MDB_PATH))
						m.Cmd(cli.SYSTEM, GIT, INIT)
						m.Cmd(cli.SYSTEM, GIT, REMOTE, ADD, ORIGIN, m.Option(kit.SSH_REPOS))
						m.Cmd(cli.SYSTEM, GIT, PULL, ORIGIN, MASTER)
					} else {
						m.Cmd(cli.SYSTEM, GIT, CLONE, "-b", kit.Select(MASTER, m.Option(BRANCH)),
							m.Option(kit.SSH_REPOS), m.Option(kit.MDB_PATH))
					}

					_repos_insert(m, m.Option(kit.MDB_NAME), m.Option(kit.MDB_PATH))
				}},
				"proxy": {Name: "proxy from to", Help: "代理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(cli.SYSTEM, GIT, "config", "--global", fmt.Sprintf(`url.%s.insteadOf`, m.Option("to")), m.Option("from"))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 { // 仓库列表
					m.Fields(len(arg) == 0, "time,name,branch,commit,remote")
					m.Cmdy(mdb.SELECT, m.Prefix(REPOS), "", mdb.HASH)
					m.Sort(kit.MDB_NAME)
					return
				}

				m.Option(nfs.DIR_ROOT, _repos_path(arg[0]))
				m.Cmdy(nfs.CAT, kit.Select("./", arg, 1), "time,line,path")

				m.Option(cli.CMD_DIR, _repos_path(arg[0]))
				p := strings.TrimPrefix(strings.TrimPrefix(m.Cmdx(cli.SYSTEM, GIT, REMOTE, "get-url", ORIGIN), "http://"), "https://")
				pp := kit.MergeURL2(m.Option(ice.MSG_USERWEB), fmt.Sprintf("/code/git/%s", strings.TrimSpace(p)))
				m.EchoScript(fmt.Sprintf("git clone %s", pp))
			}},
		},
	})
}
