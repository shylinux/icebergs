package log

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const DEBUG = "debug"

func init() {
	Index.MergeCommands(ice.Commands{
		DEBUG: {Name: "debug level=watch,bench,watch,error,trace offset filter auto doc", Help: "后台日志", Actions: ice.Actions{
			"doc": {Help: "文档", Hand: func(m *ice.Message, arg ...string) { m.ProcessOpen("https://pkg.go.dev/std") }},
		}, Hand: func(m *ice.Message, arg ...string) {
			offset := kit.Int(kit.Select("0", arg, 1))
			switch arg[0] {
			case "bench", ERROR, "trace":
				m.Cmd(nfs.CAT, ice.VAR_LOG+arg[0]+".log", func(line string, index int) {
					if len(arg) > 2 && !strings.Contains(line, arg[2]) || index < offset {
						return
					}
					ls := strings.SplitN(line, ice.SP, 6)
					m.Push(mdb.TIME, ls[0]+ice.SP+ls[1])
					m.Push(mdb.ID, ls[2])
					m.Push("ship", ls[3])

					i := strings.LastIndex(ls[5], ice.SP)
					if strings.HasPrefix(ls[5][i+1:], "base") || strings.HasPrefix(ls[5][i+1:], "core") || strings.HasPrefix(ls[5][i+1:], "misc") {
						m.Push(nfs.PATH, ice.USR_ICEBERGS)
						m.Push(nfs.FILE, strings.TrimSpace(strings.Split(ls[5][i:], ice.DF)[0]))
						m.Push(nfs.LINE, strings.TrimSpace(strings.Split(ls[5][i:], ice.DF)[1]))
						ls[5] = ls[5][:i]
					} else if strings.HasPrefix(ls[5][i+1:], "usr/icebergs/") {
						m.Push(nfs.PATH, ice.USR_ICEBERGS)
						m.Push(nfs.FILE, strings.TrimPrefix(strings.TrimSpace(strings.Split(ls[5][i:], ice.DF)[0]), ice.USR_ICEBERGS))
						m.Push(nfs.LINE, strings.TrimSpace(strings.Split(ls[5][i:], ice.DF)[1]))
						ls[5] = ls[5][:i]
					} else {
						m.Push(nfs.PATH, ice.USR_ICEBERGS)
						m.Push(nfs.FILE, "init.go")
						m.Push(nfs.LINE, "90")
					}
					m.Push(ctx.ACTION, ls[4])
					m.Push(mdb.TEXT, ls[5])
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
				})
			}
			m.StatusTimeCountTotal(offset + m.Length())
		}},
	})
}
