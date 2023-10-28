package log

import (
	"strings"
	"time"
	"unicode"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _debug_file(k string) string { return ice.VAR_LOG + k + ".log" }

const DEBUG = "debug"

func init() {
	const (
		LEVEL = "level"
	)
	Index.MergeCommands(ice.Commands{
		DEBUG: {Name: "debug level=error,bench,debug,error,watch offset limit filter auto reset doc", Help: "后台日志", Actions: ice.Actions{
			"doc": {Help: "文档", Hand: func(m *ice.Message, arg ...string) { m.ProcessOpen("https://pkg.go.dev/std") }},
			"reset": {Help: "重置", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.CAT, _debug_file(arg[0]), func(line string, index int) { m.ProcessRewrite(mdb.OFFSET, index+2, mdb.LIMIT, 1000) })
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			offset, limit, stats := kit.Int(kit.Select("0", arg, 1)), kit.Int(kit.Select("1000", arg, 2)), map[string]int{}
			switch arg[0] {
			case BENCH, ERROR, DEBUG:
				m.Cmd(nfs.CAT, _debug_file(arg[0]), func(text string, index int) {
					if index < offset || index >= offset+limit || !strings.Contains(text, kit.Select("", arg, 3)) {
						return
					}
					ls := strings.SplitN(strings.ReplaceAll(text, "  ", " "), lex.SP, 8)
					if _, e := time.Parse(kit.Split(ice.MOD_TIMES)[0], ls[0]); e != nil || len(ls) < 8 {
						m.Push(mdb.TIME, "").Push(ice.LOG_TRACEID, "").Push(mdb.ID, "")
						m.Push(nfs.PATH, "").Push(nfs.FILE, "").Push(nfs.LINE, "")
						m.Push(ctx.SHIP, "").Push(LEVEL, "").Push(nfs.CONTENT, text)
						return
					}
					m.Push(mdb.TIME, ls[0]+lex.SP+ls[1]).Push(ice.LOG_TRACEID, ls[3]).Push(mdb.ID, ls[4])
					m.Push(nfs.PATH, ice.USR_ICEBERGS)
					if i := strings.LastIndex(ls[7], lex.SP); strings.HasPrefix(ls[7][i+1:], ice.BASE) || strings.HasPrefix(ls[7][i+1:], ice.CORE) || strings.HasPrefix(ls[7][i+1:], ice.MISC) {
						m.Push(nfs.FILE, strings.TrimSpace(strings.Split(ls[7][i:], nfs.DF)[0]))
						m.Push(nfs.LINE, strings.TrimSpace(strings.Split(ls[7][i:], nfs.DF)[1]))
						ls[7] = ls[7][:i]
					} else if strings.HasPrefix(ls[7][i+1:], ice.USR_ICEBERGS) {
						m.Push(nfs.FILE, strings.TrimPrefix(strings.TrimSpace(strings.Split(ls[7][i:], nfs.DF)[0]), ice.USR_ICEBERGS))
						m.Push(nfs.LINE, strings.TrimSpace(strings.Split(ls[7][i:], nfs.DF)[1]))
						ls[7] = ls[7][:i]
					} else {
						m.Push(nfs.FILE, "base/web/serve.go").Push(nfs.LINE, "62")
					}
					if ls[6] == ice.LOG_CMDS {
						_ls := strings.SplitN(ls[5], lex.SP, 2)
						if ls[6], ls[7] = _ls[0], _ls[1]; !unicode.IsDigit(rune(ls[7][0])) {
							_ls := strings.SplitN(ls[7], lex.SP, 2)
							ls[6], ls[7] = ls[6]+lex.SP+_ls[0], _ls[1]
						}
					}
					m.Push(ctx.SHIP, ls[5]).Push(LEVEL, ls[6]).Push(nfs.CONTENT, ls[7])
					stats[ls[6]]++
				})
			case WATCH:
				m.Cmd(nfs.CAT, ice.VAR_LOG+arg[0]+".log", func(text string, index int) {
					if len(arg) > 2 && !strings.Contains(text, arg[2]) || index < offset {
						return
					}
					ls := strings.SplitN(strings.ReplaceAll(text, "  ", " "), lex.SP, 8)
					m.Push(mdb.TIME, ls[0]+lex.SP+ls[1]).Push(ice.LOG_TRACEID, ls[3]).Push(mdb.ID, ls[4])
					i := strings.LastIndex(ls[7], lex.SP)
					m.Push(nfs.PATH, ice.USR_ICEBERGS)
					m.Push(nfs.FILE, strings.TrimSpace(strings.Split(ls[7][i:], nfs.DF)[0]))
					m.Push(nfs.LINE, strings.TrimSpace(strings.Split(ls[7][i:], nfs.DF)[1]))
					m.Push(ctx.SHIP, ls[5]).Push(LEVEL, ls[6]).Push(nfs.CONTENT, ls[7][:i])
					stats[ls[6]]++
				})
			}
			m.StatusTimeCountTotal(offset+m.Length(), stats)
		}},
	})
}
