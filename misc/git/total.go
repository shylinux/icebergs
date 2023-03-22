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
	const (
		FROM   = "from"
		DAYS   = "days"
		ADDS   = "adds"
		DELS   = "dels"
		REST   = "rest"
		COMMIT = "commit"
	)
	Index.MergeCommands(ice.Commands{
		TOTAL: {Name: "total repos auto pie", Help: "统计量", Actions: ice.MergeActions(ice.Actions{
			PIE: {Help: "饼图", Hand: func(m *ice.Message, arg ...string) {
				defer ctx.DisplayStory(m, "pie.js")
				m.Cmd("", func(value ice.Maps) {
					if value[REPOS] != mdb.TOTAL {
						m.Push(REPOS, value[REPOS]).Push(mdb.VALUE, value[REST]).Push("", value, []string{FROM, DAYS, ADDS, DELS, COMMIT})
					}
				})
			}},
		}, ctx.ConfAction("skip", kit.DictList("wubi-dict", "word-dict", "websocket", "go-qrcode", "go-sql-mysql", "echarts"))), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 {
				ReposList(m).Tables(func(value ice.Maps) {
					kit.If(value[REPOS] == arg[0], func() { m.Cmdy("_sum", value[nfs.PATH], arg[1:]) })
				})
				m.StatusTimeCount(m.AppendSimple(FROM))
				return
			}
			from, days, adds, dels, rest, commit := "", 0, 0, 0, 0, 0
			ReposList(m).TableGo(func(value ice.Maps, lock *task.Lock) {
				if mdb.Config(m, kit.Keys("skip", value[REPOS])) == ice.TRUE {
					return
				}
				msg := m.Cmd("_sum", value[nfs.PATH], mdb.TOTAL, "10000")
				defer lock.Lock()()
				msg.Tables(func(value ice.Maps) {
					if kit.Int(value[DAYS]) > days {
						from, days = value[FROM], kit.Int(value[DAYS])
					}
					adds += kit.Int(value[ADDS])
					dels += kit.Int(value[DELS])
					rest += kit.Int(value[REST])
					commit += kit.Int(value[COMMIT])
				})
				m.Push(REPOS, value[REPOS]).Copy(msg)
			})
			m.Push(REPOS, mdb.TOTAL).Push(TAGS, "v3.0.0").Push(FROM, from).Push(DAYS, days).Push(ADDS, adds).Push(DELS, dels).Push(REST, rest).Push(COMMIT, commit)
			m.StatusTimeCount().SortIntR(REST)
		}},
		"_sum": {Name: "_sum [path] [total] [count|date] args...", Help: "统计量", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 {
				if nfs.ExistsFile(m, _git_dir(arg[0])) || nfs.ExistsFile(m, path.Join(arg[0], REFS_HEADS)) {
					m.Option(cli.CMD_DIR, arg[0])
					arg = arg[1:]
				}
			}
			total := false
			if len(arg) > 0 && arg[0] == mdb.TOTAL {
				total, arg = true, arg[1:]
			}
			args := []string{"log", "--shortstat", "--pretty=commit: %ad %n%s", "--date=iso", "--reverse"}
			if len(arg) > 0 {
				arg[0] += kit.Select("", " 00:00:00", strings.Contains(arg[0], "-") && !strings.Contains(arg[0], ice.DF))
				args = append(args, kit.Select("-n", "--since", strings.Contains(arg[0], "-")))
				args = append(args, arg...)
			} else {
				args = append(args, "-n", "30")
			}
			from, days, adds, dels, commit := "", 0, 0, 0, 0
			kit.SplitKV(ice.NL, "commit:", _git_cmds(m, args...), func(text string, ls []string) {
				add, del := "0", "0"
				for _, v := range kit.Split(strings.TrimSpace(kit.Select("", ls, -1)), ice.FS) {
					switch {
					case strings.Contains(v, "inser"):
						add = kit.Split(v)[0]
					case strings.Contains(v, "delet"):
						del = kit.Split(v)[0]
					}
				}
				if total {
					if commit++; from == "" {
						hs := strings.Split(ls[0], ice.SP)
						if t, e := time.Parse("2006-01-02", hs[0]); e == nil {
							from, days = hs[0], int(time.Now().Sub(t).Hours())/24
						}
					}
					adds += kit.Int(add)
					dels += kit.Int(del)
					return
				}
				m.Push(FROM, ls[0])
				m.Push(ADDS, add)
				m.Push(DELS, del)
				m.Push(REST, kit.Int(add)-kit.Int(del))
				m.Push(COMMIT, ls[1])
			})
			if total {
				m.Push(TAGS, _git_cmds(m, "describe", "--tags"))
				m.Push(FROM, from)
				m.Push(DAYS, days)
				m.Push(ADDS, adds)
				m.Push(DELS, dels)
				m.Push(REST, adds-dels)
				m.Push(COMMIT, commit)
			}
		}},
	})
}
