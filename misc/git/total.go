package git

import (
	"os"
	"path"
	"strings"
	"sync"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const TOTAL = "total"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		TOTAL: {Name: TOTAL, Help: "统计量", Value: kit.Data(
			"skip", kit.Dict(
				"wubi-dict", ice.TRUE, "word-dict", ice.TRUE,
				"websocket", ice.TRUE, "go-sql-mysql", ice.TRUE,
				"echarts", ice.TRUE, "go-qrcode", ice.TRUE,
			),
		)},
	}, Commands: map[string]*ice.Command{
		TOTAL: {Name: "total repos auto pie", Help: "统计量", Action: map[string]*ice.Action{
			PIE: {Name: "pie", Help: "饼图", Hand: func(m *ice.Message, arg ...string) {
				defer m.Display("/plugin/story/pie.js")
				m.Cmd(TOTAL).Table(func(index int, value map[string]string, head []string) {
					if value[REPOS] == "total" {
						m.StatusTimeCount(REPOS, "total", "value", "1", "total", value["rest"])
						return
					}
					m.Push(REPOS, value[REPOS])
					m.Push("value", value["rest"])
				})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 { // 提交详情
				arg[0] = kit.Replace(arg[0], "src", "contexts")
				m.Cmd(REPOS, ice.OptionFields("name,path")).Table(func(index int, value map[string]string, head []string) {
					if value[REPOS] == arg[0] {
						m.Cmdy("_sum", value[nfs.PATH], arg[1:])
					}
				})
				return
			}

			// 提交统计
			days, commit, adds, dels, rest := 0, 0, 0, 0, 0
			m.Richs(REPOS, nil, mdb.FOREACH, func(mu *sync.Mutex, key string, value map[string]interface{}) {
				value = kit.GetMeta(value)
				if m.Config(kit.Keys("skip", value[mdb.NAME])) == ice.TRUE {
					return
				}

				msg := m.Cmd("_sum", value[nfs.PATH], mdb.TOTAL, "10000")

				mu.Lock()
				defer mu.Unlock()

				msg.Table(func(index int, value map[string]string, head []string) {
					if kit.Int(value["days"]) > days {
						days = kit.Int(value["days"])
					}
					commit += kit.Int(value["commit"])
					adds += kit.Int(value["adds"])
					dels += kit.Int(value["dels"])
					rest += kit.Int(value["rest"])
				})

				m.Push(REPOS, value[mdb.NAME])
				m.Copy(msg)
			})

			m.Push(REPOS, "total")
			m.Push("tags", "v3.0.0")
			m.Push("days", days)
			m.Push("commit", commit)
			m.Push("adds", adds)
			m.Push("dels", dels)
			m.Push("rest", rest)
			m.SortIntR("rest")
			m.StatusTimeCount()
		}},
		"_sum": {Name: "_sum [path] [total] [count|date] args...", Help: "统计量", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 {
				if s, e := os.Stat(path.Join(arg[0], ".git")); e == nil && s.IsDir() {
					m.Option(cli.CMD_DIR, arg[0])
					arg = arg[1:]
				} else if s, e := os.Stat(path.Join(arg[0], "refs")); e == nil && s.IsDir() {
					m.Option(cli.CMD_DIR, arg[0])
					arg = arg[1:]
				}
			}

			total := false // 累积求和
			if len(arg) > 0 && arg[0] == mdb.TOTAL {
				total, arg = true, arg[1:]
			}

			args := []string{"log", "--shortstat", "--pretty=commit: %ad %n%s", "--date=iso", "--reverse"}
			if len(arg) > 0 {
				if strings.Contains(arg[0], "-") && !strings.Contains(arg[0], ":") {
					arg[0] = arg[0] + " 00:00:00"
				}
				args = append(args, kit.Select("-n", "--since", strings.Contains(arg[0], "-")))
				args = append(args, arg...)
			} else {
				args = append(args, "-n", "30")
			}

			var total_day time.Duration
			count, count_add, count_del := 0, 0, 0
			for i, v := range strings.Split(m.Cmdx(cli.SYSTEM, GIT, args), "commit: ") {
				l := strings.Split(v, ice.NL)
				hs := strings.Split(l[0], ice.SP)
				if len(l) < 2 {
					continue
				}

				add, del := "0", "0"
				if len(l) > 3 {
					fs := strings.Split(strings.TrimSpace(l[3]), ", ")
					if adds := strings.Split(fs[1], ice.SP); len(fs) > 2 {
						dels := strings.Split(fs[2], ice.SP)
						add = adds[0]
						del = dels[0]
					} else if strings.Contains(adds[1], "insertion") {
						add = adds[0]
					} else {
						del = adds[0]
					}
				}

				if total { // 累积求和
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

			if total { // 累积求和
				m.Push("tags", m.Cmdx(cli.SYSTEM, GIT, "describe", "--tags"))
				m.Push("days", int(total_day.Hours())/24)
				m.Push("commit", count)
				m.Push("adds", count_add)
				m.Push("dels", count_del)
				m.Push("rest", count_add-count_del)
			}
		}},
	}})
}
