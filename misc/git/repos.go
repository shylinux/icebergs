package git

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
)

func _repos_insert(m *ice.Message, name string, dir string) {
	if s, e := os.Stat(m.Option(cli.CMD_DIR, path.Join(dir, ".git"))); e == nil && s.IsDir() {
		ls := strings.SplitN(strings.Trim(m.Cmdx(cli.SYSTEM, "git", "log", "-n1", `--pretty=format:"%ad %s"`, "--date=iso"), "\""), " ", 4)
		m.Rich(REPOS, nil, kit.Data(
			"name", name, "path", dir,
			"last", kit.Select("", ls, 3), "time", strings.Join(ls[:2], " "),
			"branch", strings.TrimSpace(m.Cmdx(cli.SYSTEM, "git", "branch")),
			"remote", strings.TrimSpace(m.Cmdx(cli.SYSTEM, "git", "remote", "-v")),
		))
	}
}

const REPOS = "repos"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			REPOS: {Name: REPOS, Help: "仓库", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_FIELD, "time,name,branch,last",
				"owner", "https://github.com/shylinux",
			)},
			"progress": {Name: "progress", Help: "进度", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			REPOS: {Name: "repos name=auto path=auto auto 添加", Help: "代码库", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: `create remote branch name path`, Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Option("name", kit.Select(strings.TrimSuffix(path.Base(m.Option("remote")), ".git"), m.Option("name")))
					m.Option("path", kit.Select(path.Join("usr", m.Option("name")), m.Option("path")))
					m.Option("remote", kit.Select(m.Conf(REPOS, "meta.owner")+"/"+m.Option("name"), m.Option("remote")))

					if _, e := os.Stat(path.Join(m.Option("path"), ".git")); e != nil && os.IsNotExist(e) {
						// 下载仓库
						if _, e := os.Stat(m.Option("path")); e == nil {
							m.Option(cli.CMD_DIR, m.Option("path"))
							m.Cmd(cli.SYSTEM, GIT, "init")
							m.Cmd(cli.SYSTEM, GIT, "remote", "add", "origin", m.Option("remote"))
							m.Cmd(cli.SYSTEM, GIT, "pull", "origin", "master")
						} else {
							m.Cmd(cli.SYSTEM, GIT, "clone", "-b", kit.Select("master", m.Option("branch")),
								m.Option("remote"), m.Option("path"))

						}
						_repos_insert(m, m.Option("name"), m.Option("path"))
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					if wd, _ := os.Getwd(); arg[0] != path.Base(wd) {
						m.Option(nfs.DIR_ROOT, path.Join("usr", arg[0]))
					}
					m.Cmdy(nfs.DIR, kit.Select("./", path.Join(arg[1:]...)))
					return
				}

				m.Option(mdb.FIELDS, m.Conf(REPOS, kit.META_FIELD))
				m.Cmdy(mdb.SELECT, m.Prefix(REPOS), "", mdb.HASH, kit.MDB_NAME, arg)
				m.Sort(kit.MDB_NAME)
			}},
			"status": {Name: "status name=auto auto 提交 编译 下载", Help: "代码状态", Action: map[string]*ice.Action{
				"pull": {Name: "pull", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					if m.Richs("progress", "", m.Option("_progress"), func(key string, value map[string]interface{}) {
						m.Push("count", value["count"])
						m.Push("total", value["total"])
						m.Push("name", value["name"])
					}) != nil {
						return
					}

					count, total := 0, len(m.Confm(REPOS, "hash"))
					h := m.Rich("progress", "", kit.Dict("progress", 0, "count", count, "total", total))
					m.Gos(m, func(m *ice.Message) {
						m.Richs(REPOS, nil, kit.Select(kit.MDB_FOREACH, arg, 0), func(key string, value map[string]interface{}) {
							count++
							m.Conf("progress", kit.Keys("hash", h, "name"), kit.Value(value, "meta.name"))
							m.Conf("progress", kit.Keys("hash", h, "count"), count)
							m.Conf("progress", kit.Keys("hash", h, "progress"), count*100/total)
							m.Option(cli.CMD_DIR, kit.Value(value, "meta.path"))
							m.Echo(m.Cmdx(cli.SYSTEM, GIT, "pull"))
						})
					})
					m.Option("_progress", h)
					m.Push("count", count)
					m.Push("total", total)
					m.Push("name", "")
				}},
				"compile": {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, "make")
				}},

				"add": {Name: "add", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					if strings.Contains(m.Option("name"), ":\\") {
						m.Option(cli.CMD_DIR, m.Option("name"))
					} else {
						m.Option(cli.CMD_DIR, path.Join("usr", m.Option("name")))
					}
					m.Cmdy(cli.SYSTEM, "git", "add", m.Option("file"))
				}},
				"submit": {Name: "submit action=opt,add comment=some", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
					if m.Option("name") == "" {
						return
					}

					if strings.Contains(m.Option("name"), ":\\") {
						m.Option(cli.CMD_DIR, m.Option("name"))
					} else {
						m.Option(cli.CMD_DIR, path.Join("usr", m.Option("name")))
					}

					if arg[0] == "action" {
						m.Cmdy(cli.SYSTEM, "git", "commit", "-am", kit.Select("opt some", arg[1]+" "+arg[3]))
					} else {
						m.Cmdy(cli.SYSTEM, "git", "commit", "-am", kit.Select("opt some", strings.Join(arg, " ")))
					}
				}},
				"push": {Name: "push", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					if m.Option("name") == "" {
						return
					}
					if strings.Contains(m.Option("name"), ":\\") {
						m.Option(cli.CMD_DIR, m.Option("name"))
					} else {
						m.Option(cli.CMD_DIR, path.Join("usr", m.Option("name")))
					}
					m.Cmdy(cli.SYSTEM, "git", "push")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(REPOS, nil, kit.Select(kit.MDB_FOREACH, arg, 0), func(key string, value map[string]interface{}) {
					if m.Option(cli.CMD_DIR, kit.Value(value, "meta.path")); len(arg) > 0 {
						// 更改详情
						m.Echo(m.Cmdx(cli.SYSTEM, GIT, "diff"))
						return
					}

					// 更改列表
					for _, v := range strings.Split(strings.TrimSpace(m.Cmdx(cli.SYSTEM, GIT, "status", "-sb")), "\n") {
						vs := strings.SplitN(strings.TrimSpace(v), " ", 2)
						m.Push("name", kit.Value(value, "meta.name"))
						m.Push("tags", vs[0])
						m.Push("file", vs[1])
						list := []string{}
						switch vs[0] {
						case "##":
							if strings.Contains(vs[1], "ahead") {
								list = append(list, m.Cmdx(mdb.RENDER, web.RENDER.Button, "上传"))
							}
						default:
							if strings.Contains(vs[0], "??") {
								list = append(list, m.Cmdx(mdb.RENDER, web.RENDER.Button, "添加"))
							} else {
								list = append(list, m.Cmdx(mdb.RENDER, web.RENDER.Button, "提交"))
							}
						}
						m.Push("action", strings.Join(list, ""))
					}
				})
				m.Sort("name")
			}},
		},
	}, nil)
}
