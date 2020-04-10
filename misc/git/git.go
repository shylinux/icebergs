package git

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
	"sync"
	"time"
)

func add(m *ice.Message, n string, p string) {
	if s, e := os.Stat(m.Option("cmd_dir", path.Join(p, ".git"))); e == nil && s.IsDir() {
		ls := strings.SplitN(strings.Trim(m.Cmdx(ice.CLI_SYSTEM, "git", "log", "-n1", `--pretty=format:"%ad %s"`, "--date=iso"), "\""), " ", 4)
		m.Rich("repos", nil, kit.Data(
			"name", n, "path", p,
			"last", kit.Select("", ls, 3), "time", strings.Join(ls[:2], " "),
			"branch", strings.TrimSpace(m.Cmdx(ice.CLI_SYSTEM, "git", "branch")),
			"remote", strings.TrimSpace(m.Cmdx(ice.CLI_SYSTEM, "git", "remote", "-v")),
		))
	}
}

var Index = &ice.Context{Name: "git", Help: "代码库",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"git": {Name: "git", Help: "代码库", Value: kit.Data(
			"source", "https://github.com/git/git.git",
		)},
		"repos": {Name: "repos", Help: "仓库", Value: kit.Data(kit.MDB_SHORT, "name", "owner", "https://github.com/shylinux")},
		"total": {Name: "total", Help: "统计", Value: kit.Data(kit.MDB_SHORT, "name", "skip", kit.Dict("wubi-dict", "true", "word-dict", "true"))},
	},
	Commands: map[string]*ice.Command{
		ice.CODE_INSTALL: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option("cmd_dir", "usr")
			msg := m.Cmd(ice.CLI_SYSTEM, "git", "clone", m.Conf("git", "meta.git"))

			m.Option("cmd_dir", "usr/git")
			m.Cmd(ice.CLI_SYSTEM, "make", "configure")
			m.Cmd(ice.CLI_SYSTEM, "./configure", "--prefix="+kit.Path("usr/local"))

			m.Cmd(ice.CLI_SYSTEM, "make", "-j4")
			m.Cmd(ice.CLI_SYSTEM, "make", "install")
		}},
		ice.CODE_PREPARE: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

		}},
		"init": {Name: "init", Help: "初始化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 系统项目
			wd, _ := os.Getwd()
			add(m, path.Base(wd), wd)

			// 官方项目
			m.Cmd("nfs.dir", "usr", "name path").Table(func(index int, value map[string]string, head []string) {
				add(m, value["name"], value["path"])
			})

			// 应用项目
			m.Cmd("nfs.dir", m.Conf(ice.WEB_DREAM, "meta.path"), "name path").Table(func(index int, value map[string]string, head []string) {
				add(m, value["name"], value["path"])
			})
		}},

		"repos": {Name: "repos [name [path [remote [branch]]]]", Help: "仓库", List: kit.List(
			kit.MDB_INPUT, "text", "name", "name", "action", "auto",
			kit.MDB_INPUT, "button", "name", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 {
				if _, e := os.Stat(path.Join(arg[1], ".git")); e != nil && os.IsNotExist(e) {
					// 下载仓库
					m.Cmd(ice.CLI_SYSTEM, "git", "clone", "-b", kit.Select("master", arg, 3),
						kit.Select(m.Conf("repos", "meta.owner")+"/"+arg[0], arg, 2), arg[1])
					add(m, arg[0], arg[1])
				}
			}

			if len(arg) > 0 {
				// 仓库详情
				m.Richs("repos", nil, arg[0], func(key string, value map[string]interface{}) {
					m.Push("detail", value["meta"])
				})
				return
			}

			// 仓库列表
			m.Richs("repos", nil, "*", func(key string, value map[string]interface{}) {
				m.Push(key, value["meta"], []string{"time", "name", "branch", "last"})
			})
			m.Sort("name")
		}},
		"total": {Name: "total", Help: "统计", List: kit.List(
			kit.MDB_INPUT, "text", "name", "name", "action", "auto",
			kit.MDB_INPUT, "button", "name", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 {
				// 提交详情
				m.Richs("repos", nil, arg[0], func(key string, value map[string]interface{}) {
					m.Cmdy("_sum", kit.Value(value, "meta.path"), arg[1:])
				})
				return
			}

			// 提交统计
			days := 0
			commit, adds, dels, rest := 0, 0, 0, 0
			wg := &sync.WaitGroup{}
			m.Richs("repos", nil, "*", func(key string, value map[string]interface{}) {
				if m.Conf("total", kit.Keys("meta.skip", kit.Value(value, "meta.name"))) == "true" {
					return
				}
				wg.Add(1)
				m.Push("name", kit.Value(value, "meta.name"))
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
					m.Copy(msg)
					wg.Done()
				})
			})
			wg.Wait()
			m.Push("name", "total")
			m.Push("days", days)
			m.Push("commit", commit)
			m.Push("adds", adds)
			m.Push("dels", dels)
			m.Push("rest", rest)
			m.Sort("adds", "int_r")
		}},
		"status": {Name: "status repos", Help: "状态", Meta: kit.Dict(
			"detail", []interface{}{"add", "reset", "remove", kit.Dict("name", "commit", "args", kit.List(
				kit.MDB_INPUT, "select", "name", "type", "values", []string{"add", "opt"},
				kit.MDB_INPUT, "text", "name", "name", "value", "some",
			))},
		), List: kit.List(
			kit.MDB_INPUT, "text", "name", "name", "action", "auto",
			kit.MDB_INPUT, "button", "name", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{ice.CLI_SYSTEM, "git"}

			if len(arg) > 1 && arg[0] == "action" {
				m.Richs("repos", nil, m.Option("name"), func(key string, value map[string]interface{}) {
					m.Option("cmd_dir", kit.Value(value, "meta.path"))
					switch arg[1] {
					case "add":
						m.Cmdy(prefix, arg[1], m.Option("file"))
					case "reset":
						m.Cmdy(prefix, arg[1], m.Option("file"))
					case "checkout":
						m.Cmdy(prefix, arg[1], m.Option("file"))
					case "commit":
						m.Cmdy(prefix, arg[1], "-m", m.Option("comment"))
					}
				})
				return
			}

			m.Richs("repos", nil, kit.Select("*", arg, 0), func(key string, value map[string]interface{}) {
				if m.Option("cmd_dir", kit.Value(value, "meta.path")); len(arg) > 0 {
					// 更改详情
					m.Echo(m.Cmdx(prefix, "diff"))
					return
				}

				// 更改列表
				for _, v := range strings.Split(strings.TrimSpace(m.Cmdx(prefix, "status", "-sb")), "\n") {
					vs := strings.SplitN(strings.TrimSpace(v), " ", 2)
					m.Push("name", kit.Value(value, "meta.name"))
					m.Push("tags", vs[0])
					m.Push("file", vs[1])
				}
			})
		}},

		"_sum": {Name: "_sum [path] [total] [count|date] args...", Help: "统计", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 {
				if s, e := os.Stat(arg[0] + "/.git"); e == nil && s.IsDir() {
					m.Option("cmd_dir", arg[0])
					arg = arg[1:]
				}
			}

			total := false
			if len(arg) > 0 && arg[0] == "total" {
				total, arg = true, arg[1:]
			}

			args := []string{}
			args = append(args, "log", "--shortstat", "--pretty=commit: %ad %n%s", "--date=iso", "--reverse")
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
			for i, v := range strings.Split(m.Cmdx(ice.CLI_SYSTEM, "git", args), "commit: ") {
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
						} else if adds[1] == "insertions(+)" {
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

		"trend": {Name: "check name [path [repos]]", Help: "检查", Meta: kit.Dict("display", "/plugin/story/trend"), List: kit.List(
			kit.MDB_INPUT, "text", "name", "repos", "action", "auto",
			kit.MDB_INPUT, "text", "name", "begin_time", "figure", "date",
			kit.MDB_INPUT, "button", "name", "执行", "action", "auto",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Option("_display", "table")
			}
			m.Cmdy("total", arg)
		}},
	},
}

func init() { code.Index.Register(Index, nil) }
