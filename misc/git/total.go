package git

import (
	"os"
	"strings"
	"sync"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	kit "shylinux.com/x/toolkits"
)

const TOTAL = "total"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		TOTAL: {Name: TOTAL, Help: "统计量", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_NAME, "skip", kit.Dict("wubi-dict", ice.TRUE, "word-dict", ice.TRUE),
		)},
	}, Commands: map[string]*ice.Command{
		TOTAL: {Name: "total name auto", Help: "统计量", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 { // 提交详情
				m.Richs(REPOS, nil, arg[0], func(key string, value map[string]interface{}) {
					m.Cmdy("_sum", kit.Value(value, kit.META_PATH), arg[1:])
				})
				return
			}

			// 提交统计
			days, commit, adds, dels, rest := 0, 0, 0, 0, 0
			m.Richs(REPOS, nil, kit.MDB_FOREACH, func(mu *sync.Mutex, key string, value map[string]interface{}) {
				value = kit.GetMeta(value)
				if m.Conf(TOTAL, kit.Keym("skip", value[kit.MDB_NAME])) == ice.TRUE {
					return
				}

				msg := m.Cmd("_sum", value[kit.MDB_PATH], "total", "10000")

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

				m.Push("name", value[kit.MDB_NAME])
				m.Copy(msg)
			})

			m.Push("name", "total")
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
				if s, e := os.Stat(arg[0] + "/.git"); e == nil && s.IsDir() {
					m.Option(cli.CMD_DIR, arg[0])
					arg = arg[1:]
				} else if s, e := os.Stat(arg[0] + "/refs"); e == nil && s.IsDir() {
					m.Option(cli.CMD_DIR, arg[0])
					arg = arg[1:]
				}
			}

			total := false
			if len(arg) > 0 && arg[0] == "total" {
				total, arg = true, arg[1:]
			}

			args := []string{}
			args = append(args, "log",
				// kit.Format("--author=%s\\|shylinux", m.Option(ice.MSG_USERNAME)),
				"--shortstat", "--pretty=commit: %ad %n%s", "--date=iso", "--reverse")
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
				l := strings.Split(v, "\n")
				hs := strings.Split(l[0], " ")
				if len(l) < 2 {
					continue
				}

				add, del := "0", "0"
				if len(l) > 3 {
					fs := strings.Split(strings.TrimSpace(l[3]), ", ")
					if adds := strings.Split(fs[1], " "); len(fs) > 2 {
						dels := strings.Split(fs[2], " ")
						add = adds[0]
						del = dels[0]
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

			if total {
				m.Push("tags", m.Cmdx(cli.SYSTEM, GIT, "describe", "--tags"))
				m.Push("days", int(total_day.Hours())/24)
				m.Push("commit", count)
				m.Push("adds", count_add)
				m.Push("dels", count_del)
				m.Push("rest", count_add-count_del)
			}
		}},
	},
	})
}
