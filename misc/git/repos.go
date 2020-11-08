package git

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
)

func _repos_insert(m *ice.Message, name string, dir string) {
	if s, e := os.Stat(m.Option(cli.CMD_DIR, path.Join(dir, ".git"))); e == nil && s.IsDir() {
		ls := strings.SplitN(strings.Trim(m.Cmdx(cli.SYSTEM, GIT, "log", "-n1", `--pretty=format:"%ad %s"`, "--date=iso"), "\""), " ", 4)
		m.Rich(REPOS, nil, kit.Data(
			kit.MDB_NAME, name, kit.MDB_PATH, dir,
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
	COMMIT = "commit"
)
const REPOS = "repos"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			REPOS: {Name: REPOS, Help: "仓库", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_FIELD, "time,name,branch,commit",
				"owner", "https://github.com/shylinux",
			)},
		},
		Commands: map[string]*ice.Command{
			REPOS: {Name: "repos name path auto create", Help: "代码库", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create repos branch name path", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Option(kit.MDB_NAME, kit.Select(strings.TrimSuffix(path.Base(m.Option(kit.SSH_REPOS)), ".git"), m.Option(kit.MDB_NAME)))
					m.Option(kit.MDB_PATH, kit.Select(path.Join("usr", m.Option(kit.MDB_NAME)), m.Option(kit.MDB_PATH)))
					m.Option(kit.SSH_REPOS, kit.Select(m.Conf(REPOS, "meta.owner")+"/"+m.Option(kit.MDB_NAME), m.Option(kit.SSH_REPOS)))

					if _, e := os.Stat(path.Join(m.Option(kit.MDB_PATH), ".git")); e != nil && os.IsNotExist(e) {
						// 下载仓库
						if _, e := os.Stat(m.Option(kit.MDB_PATH)); e == nil {
							m.Option(cli.CMD_DIR, m.Option(kit.MDB_PATH))
							m.Cmd(cli.SYSTEM, GIT, "init")
							m.Cmd(cli.SYSTEM, GIT, REMOTE, "add", ORIGIN, m.Option(kit.SSH_REPOS))
							m.Cmd(cli.SYSTEM, GIT, "pull", ORIGIN, MASTER)
						} else {
							m.Cmd(cli.SYSTEM, GIT, "clone", "-b", kit.Select(MASTER, m.Option(BRANCH)),
								m.Option(kit.SSH_REPOS), m.Option(kit.MDB_PATH))

						}
						_repos_insert(m, m.Option(kit.MDB_NAME), m.Option(kit.MDB_PATH))
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					if wd, _ := os.Getwd(); arg[0] != path.Base(wd) {
						m.Option(nfs.DIR_ROOT, path.Join("usr", arg[0]))
					}
					file := kit.Select("./", path.Join(arg[1:]...))
					if strings.HasSuffix(file, "/") {
						m.Cmdy(nfs.DIR, file)
					} else {
						m.Cmdy(nfs.CAT, file)
					}
					return
				}

				m.Option(mdb.FIELDS, m.Conf(REPOS, kit.META_FIELD))
				m.Cmdy(mdb.SELECT, m.Prefix(REPOS), "", mdb.HASH, kit.MDB_NAME, arg)
				m.Sort(kit.MDB_NAME)
			}},
		},
	})
}
