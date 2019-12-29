package git

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/toolkits"
	"os"
	"strings"
	"time"
)

var Index = &ice.Context{Name: "git", Help: "代码管理",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"repos": {Name: "repos", Help: "仓库", Value: kit.Data(kit.MDB_SHORT, "name")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 前端代码
			m.Rich("repos", nil, kit.Data(
				"name", "volcanos", "path", "usr/volcanos", "branch", "master",
				"remote", "https://github.com/shylinux/volcanos",
			))
			m.Watch(ice.SYSTEM_INIT, "cli.git.check", "volcanos")
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"check": {Name: "check", Help: "检查", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Richs("repos", nil, arg[0], func(key string, value map[string]interface{}) {
				if _, e := os.Stat(kit.Format(kit.Value(value, "meta.path"))); e != nil && os.IsNotExist(e) {
					m.Cmd("cli.system", "git", "clone", kit.Value(value, "meta.remote"),
						"-b", kit.Value(value, "meta.branch"), kit.Value(value, "meta.path"))
				}
			})
		}},
		"sum": {Name: "sum", Help: "统计", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			total := false
			if len(arg) > 0 && arg[0] == "total" {
				total, arg = true, arg[1:]
			}

			args := []string{}
			if len(arg) > 0 {
				if s, e := os.Stat(arg[0] + "/.git"); e == nil && s.IsDir() {
					args, arg = append(args, "-C", arg[0]), arg[1:]
				}
			}

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

func init() { cli.Index.Register(Index, nil) }
