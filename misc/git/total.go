package git

import (
	"os"
	"strings"
	"sync"
	"time"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	kit "github.com/shylinux/toolkits"
)

const TOTAL = "total"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TOTAL: {Name: TOTAL, Help: "统计", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, "skip", kit.Dict(
					"wubi-dict", "true", "word-dict", "true",
				),
			)},
		},
		Commands: map[string]*ice.Command{
			TOTAL: {Name: "total name auto", Help: "提交统计", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					// 提交详情
					m.Richs(REPOS, nil, arg[0], func(key string, value map[string]interface{}) {
						m.Cmdy("_sum", kit.Value(value, "meta.path"), arg[1:])
					})
					return
				}

				// 提交统计
				wg := &sync.WaitGroup{}
				days, commit, adds, dels, rest := 0, 0, 0, 0, 0
				m.Richs(REPOS, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
					if m.Conf(TOTAL, kit.Keys("meta.skip", kit.Value(value, "meta.name"))) == "true" {
						return
					}

					wg.Add(1)
					m.Go(func() {
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
				m.SortIntR("rest")
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
		},
	})
}
