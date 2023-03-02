package log

import (
	"strings"
	"unicode"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const DEBUG = "debug"

func init() {
	Index.MergeCommands(ice.Commands{
		DEBUG: {Name: "debug level=watch,bench,debug,error,watch offset filter auto doc", Help: "后台日志", Actions: ice.Actions{
			"doc": {Help: "文档", Hand: func(m *ice.Message, arg ...string) { m.ProcessOpen("https://pkg.go.dev/std") }},
		}, Hand: func(m *ice.Message, arg ...string) {
			offset := kit.Int(kit.Select("0", arg, 1))
			stats := map[string]int{}
			switch arg[0] {
			case BENCH, ERROR, DEBUG:
				m.Cmd(nfs.CAT, ice.VAR_LOG+arg[0]+".log", func(line string, index int) {
					if len(arg) > 2 && !strings.Contains(line, arg[2]) || index < offset {
						return
					}
					ls := strings.SplitN(line, ice.SP, 6)
					m.Push(mdb.TIME, ls[0]+ice.SP+ls[1])
					m.Push(mdb.ID, ls[2])
					i := strings.LastIndex(ls[5], ice.SP)
					if strings.HasPrefix(ls[5][i+1:], ice.BASE) || strings.HasPrefix(ls[5][i+1:], ice.CORE) || strings.HasPrefix(ls[5][i+1:], ice.MISC) {
						m.Push(nfs.PATH, ice.USR_ICEBERGS)
						m.Push(nfs.FILE, strings.TrimSpace(strings.Split(ls[5][i:], ice.DF)[0]))
						m.Push(nfs.LINE, strings.TrimSpace(strings.Split(ls[5][i:], ice.DF)[1]))
						ls[5] = ls[5][:i]
					} else if strings.HasPrefix(ls[5][i+1:], ice.USR_ICEBERGS) {
						m.Push(nfs.PATH, ice.USR_ICEBERGS)
						m.Push(nfs.FILE, strings.TrimPrefix(strings.TrimSpace(strings.Split(ls[5][i:], ice.DF)[0]), ice.USR_ICEBERGS))
						m.Push(nfs.LINE, strings.TrimSpace(strings.Split(ls[5][i:], ice.DF)[1]))
						ls[5] = ls[5][:i]
					} else {
						m.Push(nfs.PATH, ice.USR_ICEBERGS)
						m.Push(nfs.FILE, "init.go")
						m.Push(nfs.LINE, "90")
					}
					if ls[4] == "cmds" {
						_ls := strings.SplitN(ls[5], ice.SP, 2)
						ls[4] = _ls[0]
						ls[5] = _ls[1]
						if !unicode.IsDigit(rune(ls[5][0])) {
							_ls := strings.SplitN(ls[5], ice.SP, 2)
							ls[4] += ice.SP + _ls[0]
							ls[5] = _ls[1]
						}
					}
					m.Push("ship", ls[3])
					m.Push(ctx.ACTION, ls[4])
					m.Push(mdb.TEXT, ls[5])
					stats[ls[4]]++
				})
			case WATCH:
				m.Cmd(nfs.CAT, ice.VAR_LOG+arg[0]+".log", func(line string, index int) {
					if len(arg) > 2 && !strings.Contains(line, arg[2]) || index < offset {
						return
					}
					ls := strings.SplitN(line, ice.SP, 6)
					m.Push(mdb.TIME, ls[0]+ice.SP+ls[1])
					m.Push(mdb.ID, ls[2])
					m.Push("ship", ls[3])

					i := strings.LastIndex(ls[5], ice.SP)
					m.Push(nfs.PATH, ice.USR_ICEBERGS)
					m.Push(nfs.FILE, strings.TrimSpace(strings.Split(ls[5][i:], ice.DF)[0]))
					m.Push(nfs.LINE, strings.TrimSpace(strings.Split(ls[5][i:], ice.DF)[1]))

					m.Push(ctx.ACTION, ls[4])
					m.Push(mdb.TEXT, ls[5][:i])
					stats[ls[4]]++
				})
			}
			m.StatusTimeCountTotal(offset+m.Length(), stats)
		}},
	})
}
