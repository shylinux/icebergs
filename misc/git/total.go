package git

import (
	"path"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/task"
)

const TOTAL = "total"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		TOTAL: {Name: TOTAL, Help: "统计量", Value: kit.Data(
			"skip", kit.Dict(
				"wubi-dict", ice.TRUE, "word-dict", ice.TRUE,
				"websocket", ice.TRUE, "go-qrcode", ice.TRUE,
				"go-sql-mysql", ice.TRUE, "echarts", ice.TRUE,
			),
		)},
	}, Commands: ice.Commands{
		TOTAL: {Name: "total repos auto pie", Help: "统计量", Actions: ice.Actions{
			PIE: {Name: "pie", Help: "饼图", Hand: func(m *ice.Message, arg ...string) {
				defer ctx.DisplayStory(m, "pie.js")
				m.Cmd(TOTAL, func(value ice.Maps) {
					if value[REPOS] == mdb.TOTAL {
						m.StatusTimeCount(REPOS, mdb.TOTAL, mdb.VALUE, "1", mdb.TOTAL, value["rest"])
						return
					}
					m.Push(REPOS, value[REPOS])
					m.Push(mdb.VALUE, value["rest"])
				})
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 { // 提交详情
				arg[0] = kit.Replace(arg[0], ice.SRC, ice.CONTEXTS)
				m.Cmd(REPOS, ice.OptionFields("name,path"), func(value ice.Maps) {
					if value[REPOS] == arg[0] {
						m.Cmdy("_sum", value[nfs.PATH], arg[1:])
					}
				})
				return
			}

			// 提交统计
			from, days, commit, adds, dels, rest := "", 0, 0, 0, 0, 0
			m.Cmd(REPOS, ice.OptionFields("name,path")).TableGo(func(value ice.Maps, lock *task.Lock) {
				if m.Config(kit.Keys("skip", value[REPOS])) == ice.TRUE {
					return
				}
				msg := m.Cmd("_sum", value[nfs.PATH], mdb.TOTAL, "10000")
				// msg := m.Cmd("_sum", value[nfs.PATH], mdb.TOTAL, "2022-01-01")

				defer lock.Lock()()
				msg.Tables(func(value ice.Maps) {
					if kit.Int(value["days"]) > days {
						days = kit.Int(value["days"])
						from = value["from"]
					}
					commit += kit.Int(value["commit"])
					adds += kit.Int(value["adds"])
					dels += kit.Int(value["dels"])
					rest += kit.Int(value["rest"])
				})

				m.Push(REPOS, value[REPOS])
				m.Copy(msg)
			})

			m.Push(REPOS, mdb.TOTAL)
			m.Push("tags", "v3.0.0")
			m.Push("days", days)
			m.Push("commit", commit)
			m.Push("adds", adds)
			m.Push("dels", dels)
			m.Push("rest", rest)
			m.SortIntR("rest")
			m.StatusTimeCount("from", from)
		}},
		"_sum": {Name: "_sum [path] [total] [count|date] args...", Help: "统计量", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 {
				if s, e := nfs.StatFile(m, path.Join(arg[0], ".git")); e == nil && s.IsDir() {
					m.Option(cli.CMD_DIR, arg[0])
					arg = arg[1:]
				} else if s, e := nfs.StatFile(m, path.Join(arg[0], "refs")); e == nil && s.IsDir() {
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
			from, count, count_add, count_del := "", 0, 0, 0
			for i, v := range strings.Split(_git_cmds(m, args...), "commit: ") {
				l := strings.Split(v, ice.NL)
				hs := strings.Split(l[0], ice.SP)
				if len(l) < 2 {
					continue
				}

				add, del := "0", "0"
				if len(l) > 3 {
					for _, v := range kit.Split(strings.TrimSpace(l[3]), ice.FS) {
						switch {
						case strings.Contains(v, "insert"):
							add = kit.Split(v)[0]
						case strings.Contains(v, "delet"):
							del = kit.Split(v)[0]
						}
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

				if i == 0 {
					from = hs[0]
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
				m.Push("from", from)
				m.Push("tags", _git_cmds(m, "describe", "--tags"))
				m.Push("days", int(total_day.Hours())/24)
				m.Push("commit", count)
				m.Push("adds", count_add)
				m.Push("dels", count_del)
				m.Push("rest", count_add-count_del)
			}
		}},
	}})
}
