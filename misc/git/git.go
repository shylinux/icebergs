package git

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
	"time"
)

var Index = &ice.Context{Name: "git", Help: "代码管理",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"repos": {Name: "repos", Help: "仓库", Value: kit.Data(kit.MDB_SHORT, "name", "owner", "https://github.com/shylinux")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 系统项目
			wd, _ := os.Getwd()
			if s, e := os.Stat(".git"); e == nil && s.IsDir() {
				m.Rich("repos", nil, kit.Data(
					"name", path.Base(wd), "path", wd, "branch", "master",
					"remote", strings.TrimSpace(m.Cmdx(ice.CLI_SYSTEM, "git", "remote", "get-url", "origin")),
				))
			}

			// 官方项目
			m.Cmd("nfs.dir", "usr", "name path").Table(func(index int, value map[string]string, head []string) {
				if s, e := os.Stat(m.Option("cmd_dir", path.Join(value["path"], ".git"))); e == nil && s.IsDir() {
					m.Rich("repos", nil, kit.Data(
						"name", value["name"], "path", value["path"], "branch", "master",
						"remote", strings.TrimSpace(m.Cmdx(ice.CLI_SYSTEM, "git", "remote", "get-url", "origin")),
					))
				}
			})

			// 应用项目
			m.Cmd("nfs.dir", m.Conf(ice.WEB_DREAM, "meta.path"), "name path").Table(func(index int, value map[string]string, head []string) {
				if s, e := os.Stat(m.Option("cmd_dir", path.Join(value["path"], ".git"))); e == nil && s.IsDir() {
					m.Rich("repos", nil, kit.Data(
						"name", value["name"], "path", value["path"], "branch", "master",
						"remote", strings.TrimSpace(m.Cmdx(ice.CLI_SYSTEM, "git", "remote", "get-url", "origin")),
					))
				}
			})
			m.Watch(ice.SYSTEM_INIT, "web.code.git.check", "volcanos")
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		"repos": {Name: "repos [name [path]]", Help: "仓库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 {
				m.Rich("repos", nil, kit.Data(
					"name", arg[0], "path", kit.Select(path.Join("usr", arg[0]), arg, 1), "branch", "master",
					"remote", kit.Select(m.Conf("repos", "meta.owner")+"/"+arg[0], arg, 2),
				))
			}
			m.Richs("repos", nil, "*", func(key string, value map[string]interface{}) {
				m.Push(key, value["meta"], []string{"time", "name", "branch", "path", "remote"})
			})
			m.Sort("name")
		}},
		"branch": {Name: "branch", Help: "分支", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{ice.CLI_SYSTEM, "git", "branch"}
			m.Richs("repos", nil, kit.Select("*", arg, 0), func(key string, value map[string]interface{}) {
				m.Option("cmd_dir", kit.Value(value, "meta.path"))
				for _, v := range strings.Split(m.Cmdx(prefix, "-v"), "\n") {
					if len(v) > 0 {
						m.Push("repos", kit.Value(value, "meta.name"))
						m.Push("tags", v[:2])
						vs := strings.SplitN(strings.TrimSpace(v[2:]), " ", 2)
						m.Push("branch", vs[0])
						m.Push("last", m.Cmdx(ice.CLI_SYSTEM, "git", "log", "-n", "1", "--pretty=%ad", "--date=short"))
						vs = strings.SplitN(strings.TrimSpace(vs[1]), " ", 2)
						m.Push("hash", vs[0])
						m.Push("note", strings.TrimSpace(vs[1]))
					}
				}
			})
			m.Sort("repos")
		}},
		"status": {Name: "status repos", Help: "状态", Meta: kit.Dict(
			"detail", []interface{}{"add", "reset", "remove", kit.Dict("name", "commit", "args", kit.List(
				kit.MDB_INPUT, "select", "name", "type", "values", []string{"add", "opt"},
				kit.MDB_INPUT, "text", "name", "name", "value", "some",
			))},
		), List: kit.List(
			kit.MDB_INPUT, "text", "name", "repos", "action", "auto",
			kit.MDB_INPUT, "button", "name", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last", "action", "auto",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{ice.CLI_SYSTEM, "git"}

			if len(arg) > 0 && arg[0] == "action" {
				m.Richs("repos", nil, m.Option("repos"), func(key string, value map[string]interface{}) {
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
					m.Echo(m.Cmdx(prefix, "diff"))
					return
				}

				for _, v := range strings.Split(strings.TrimSpace(m.Cmdx(prefix, "status", "-sb")), "\n") {
					vs := strings.SplitN(strings.TrimSpace(v), " ", 2)
					m.Push("repos", kit.Value(value, "meta.name"))
					m.Push("tags", vs[0])
					m.Push("file", vs[1])
				}
			})
		}},
		"total": {Name: "total", Help: "统计", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			days := 0
			commit, adds, dels, rest := 0, 0, 0, 0
			m.Richs("repos", nil, kit.Select("*", arg, 0), func(key string, value map[string]interface{}) {
				m.Push("repos", kit.Value(value, "meta.name"))
				m.Copy(m.Cmd("sum", kit.Value(value, "meta.path"), "total", "10000").Table(func(index int, value map[string]string, head []string) {
					if kit.Int(value["days"]) > days {
						days = kit.Int(value["days"])
					}
					commit += kit.Int(value["commit"])
					adds += kit.Int(value["adds"])
					dels += kit.Int(value["dels"])
					rest += kit.Int(value["rest"])
				}))
			})
			m.Push("repos", "total")
			m.Push("days", days)
			m.Push("commit", commit)
			m.Push("adds", adds)
			m.Push("dels", dels)
			m.Push("rest", rest)
			m.Sort("adds", "int_r")
		}},
		"check": {Name: "check name [path [repos]]", Help: "检查", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 {
				m.Cmd("repos", arg)
			}

			m.Richs("repos", nil, arg[0], func(key string, value map[string]interface{}) {
				if _, e := os.Stat(kit.Format(kit.Value(value, "meta.path"))); e != nil && os.IsNotExist(e) {
					m.Cmd(ice.CLI_SYSTEM, "git", "clone", kit.Value(value, "meta.remote"),
						"-b", kit.Value(value, "meta.branch"), kit.Value(value, "meta.path"))
				}
			})
		}},
		"sum": {Name: "sum [path] [total] [count|date] args...", Help: "统计", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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
					arg[1] = arg[1] + " 00:00:00"
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
	},
}

func init() { code.Index.Register(Index, nil) }
