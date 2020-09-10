package git

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
	"sync"
	"time"
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

const (
	GIT   = "git"
	REPOS = "repos"
	TOTAL = "total"
	TREND = "trend"
	SPIDE = "spide"
)

var Index = &ice.Context{Name: GIT, Help: "代码库",
	Configs: map[string]*ice.Config{
		GIT: {Name: GIT, Help: "代码库", Value: kit.Data(
			"source", "https://mirrors.edge.kernel.org/pub/software/scm/git/git-1.8.3.1.tar.gz", "config", kit.Dict(
				"alias", kit.Dict("s", "status", "b", "branch"),
				"color", kit.Dict("ui", "true"),
				"push", kit.Dict("default", "simple"),
				"credential", kit.Dict("helper", "store"),
			),
		)},
		REPOS: {Name: REPOS, Help: "仓库", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_FIELD, "time,name,branch,last",
			"owner", "https://github.com/shylinux",
		)},
		TOTAL: {Name: TOTAL, Help: "统计", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_NAME, "skip", kit.Dict(
				"wubi-dict", "true", "word-dict", "true",
			),
		)},
		"progress": {Name: "progress", Help: "进度", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 系统项目
			wd, _ := os.Getwd()
			_repos_insert(m, path.Base(wd), wd)
		}},
		"init": {Name: "init", Help: "初始化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 官方项目
			m.Cmd(nfs.DIR, "usr", "name path").Table(func(index int, value map[string]string, head []string) {
				_repos_insert(m, value["name"], value["path"])
			})
		}},

		GIT: {Name: "git port=auto path=auto auto 构建 下载", Help: "代码库", Action: map[string]*ice.Action{
			"download": {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, "download", m.Conf(GIT, kit.META_SOURCE))
			}},
			"build": {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, "build", m.Conf(GIT, kit.META_SOURCE))
			}},
			"start": {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Optionv("prepare", func(p string) []string {
					m.Option(cli.CMD_DIR, p)
					kit.Fetch(m.Confv(GIT, "meta.config"), func(conf string, value interface{}) {
						kit.Fetch(value, func(key string, value string) {
							m.Cmd(cli.SYSTEM, "bin/git", "config", "--global", conf+"."+key, value)
						})
					})
					return []string{}
				})
				m.Cmdy(code.INSTALL, "start", m.Conf(GIT, kit.META_SOURCE), "bin/git")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(code.INSTALL, path.Base(m.Conf(GIT, kit.META_SOURCE)), arg)
		}},

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

		TOTAL: {Name: "total name=auto auto", Help: "提交统计", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 {
				// 提交详情
				m.Richs(REPOS, nil, arg[0], func(key string, value map[string]interface{}) {
					m.Cmdy("_sum", kit.Value(value, "meta.path"), arg[1:])
				})
				return
			}

			// 提交统计
			days := 0
			commit, adds, dels, rest := 0, 0, 0, 0
			wg := &sync.WaitGroup{}
			m.Richs(REPOS, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
				if m.Conf(TOTAL, kit.Keys("meta.skip", kit.Value(value, "meta.name"))) == "true" {
					return
				}
				wg.Add(1)
				m.Gos(m, func(m *ice.Message) {
					msg := m.Cmd("_sum", kit.Value(value, "meta.path"), "total", "10000").Table(func(index int, value map[string]string, head []string) {
						if kit.Int(value["days"]) > days {
							days = kit.Int(value["days"])
						}
						commit += kit.Int(value["commit"])
						adds += kit.Int(value["adds"])
						dels += kit.Int(value["dels"])
						rest += kit.Int(value["rest"])
					})
					m.Push("name", kit.Value(value, "meta.name"))
					m.Copy(msg)
					wg.Done()
				})
			})
			wg.Wait()
			m.Push("name", "total")
			m.Push("days", kit.Int(days)+1)
			m.Push("commit", commit)
			m.Push("adds", adds)
			m.Push("dels", dels)
			m.Push("rest", rest)
			m.Sort("rest", "int_r")
		}},
		"_sum": {Name: "_sum [path] [total] [count|date] args...", Help: "统计", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 {
				if s, e := os.Stat(arg[0] + "/.git"); e == nil && s.IsDir() {
					m.Option(cli.CMD_DIR, arg[0])
					arg = arg[1:]
				}
			}

			total := false
			if len(arg) > 0 && arg[0] == "total" {
				total, arg = true, arg[1:]
			}

			args := []string{}
			args = append(args, "log", kit.Format("--author=%s\\|shylinux", m.Option(ice.MSG_USERNAME)), "--shortstat", "--pretty=commit: %ad %n%s", "--date=iso", "--reverse")
			if len(arg) > 0 {
				args = append(args, kit.Select("-n", "--since", strings.Contains(arg[0], "-")))
				if strings.Contains(arg[0], "-") && !strings.Contains(arg[0], ":") {
					arg[0] = arg[0] + " 00:00:00"
				}
				args = append(args, arg[0:]...)
			} else {
				args = append(args, "-n", "30")
			}

			var total_day time.Duration
			count, count_add, count_del := 0, 0, 0
			for i, v := range strings.Split(m.Cmdx(cli.SYSTEM, "git", args), "commit: ") {
				if i > 0 {
					l := strings.Split(v, "\n")
					hs := strings.Split(l[0], " ")

					add, del := "0", "0"
					if len(l) > 3 {
						fs := strings.Split(strings.TrimSpace(l[3]), ", ")
						if adds := strings.Split(fs[1], " "); len(fs) > 2 {
							dels := strings.Split(fs[2], " ")
							add = adds[0]
							del = dels[0]
							// } else if adds[1] == "insertions(+)" {
						} else if strings.Contains(adds[1], "insertion") {
							add = adds[0]
						} else {
							del = adds[0]
						}
					}

					if total {
						if count++; i == 1 {
							if t, e := time.Parse("2006-01-02", hs[0]); e == nil {
								total_day = time.Now().Sub(t)
								m.Append("from", hs[0])
							}
						}
						count_add += kit.Int(add)
						count_del += kit.Int(del)
						continue
					}

					m.Push("date", hs[0])
					m.Push("adds", add)
					m.Push("dels", del)
					m.Push("rest", kit.Int(add)-kit.Int(del))
					m.Push("note", l[1])
					m.Push("hour", strings.Split(hs[1], ":")[0])
					m.Push("time", hs[1])
				}
			}
			if total {
				m.Push("days", int(total_day.Hours())/24)
				m.Push("commit", count)
				m.Push("adds", count_add)
				m.Push("dels", count_del)
				m.Push("rest", count_add-count_del)
			}
		}},
		TREND: {Name: "trend name=auto begin_time=@date auto", Help: "趋势图", Meta: kit.Dict(
			"display", "/plugin/story/trend.js",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Option("_display", "table")
			}
			m.Cmdy(TOTAL, arg)
		}},
		SPIDE: {Name: "spide path=auto file=auto auto", Help: "结构图", Meta: kit.Dict(
			"display", "/plugin/story/spide.js",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 仓库列表
				m.Option("_display", "table")
				m.Cmdy(TOTAL, arg)
				return
			}
			if len(arg) == 1 {
				// 目录列表
				m.Option(nfs.DIR_DEEP, "true")
				m.Cmdy(nfs.DIR, mdb.RENDER, nfs.DIR, "", path.Join("usr", arg[0]))
				return
			}

			if len(arg) > 1 && arg[0] == "inner" {
				// 代码详情
				arg[1] = path.Join("usr", arg[1])
				m.Cmdy("web.code.inner", arg[1:])
				return
			}

			tags := ""
			m.Option(cli.CMD_DIR, path.Join("usr", arg[0]))
			if strings.HasSuffix(arg[1], ".go") {
				tags = m.Cmdx(cli.SYSTEM, "gotags", arg[1])
				for _, line := range strings.Split(tags, "\n") {
					if len(line) == 0 || strings.HasPrefix(line, "!_") {
						continue
					}

					ls := kit.Split(line, "\t ", "\t ", "\t ")
					name := ls[3] + ":" + ls[0]
					switch ls[3] {
					case "m":
						if strings.HasPrefix(ls[5], "ctype") {
							name = strings.TrimPrefix(ls[5], "ctype:") + ":" + ls[0]
						} else if strings.HasPrefix(ls[6], "ntype") {
							name = "-" + ls[0]
						} else {

						}
					case "w":
						t := ls[len(ls)-1]
						name = "-" + ls[0] + ":" + strings.TrimPrefix(t, "type:")
					}

					m.Push("name", name)
					m.Push("file", ls[1])
					m.Push("line", strings.TrimSuffix(ls[2], ";\""))
					m.Push("type", ls[3])
					m.Push("extra", strings.Join(ls[4:], " "))
				}
			} else {
				tags = m.Cmdx(cli.SYSTEM, "ctags", "-f", "-", arg[1])
				for _, line := range strings.Split(tags, "\n") {
					if len(line) == 0 || strings.HasPrefix(line, "!_") {
						continue
					}

					ls := kit.Split(line, "\t ", "\t ", "\t ")
					m.Push("name", ls[0])
					m.Push("file", ls[1])
					m.Push("line", "1")
				}
			}
			m.Sort("line", "int")
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
}

func init() { code.Index.Register(Index, &web.Frame{}) }
