package git

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func init() {
	const TAGS = "tags"
	Index.MergeCommands(ice.Commands{
		TAGS: {Name: "tags path name auto", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Icon: "bi bi-star"},
			mdb.SHOW: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessFloat(m, web.VIMER, []string{m.Option(nfs.PATH), m.Option(nfs.FILE), m.Option(nfs.LINE)}, arg...)
			}},
			code.DOC: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessFloat(m, "web.code.doc", []string{m.Option(nfs.PATH), m.Option(mdb.NAME)}, arg...)
			}},
		}, mdb.ExportHashAction(mdb.SHORT, "path,file,line", mdb.FIELD, "time,path,file,line,type,name,text,hash")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m).PushAction(code.DOC, mdb.SHOW, mdb.REMOVE).Action()
				return
			}
			if len(arg) > 1 {
				m.Cmdy("web.code.doc", arg)
				kit.For(kit.SplitLine(m.Result()), func(text string) {
					ls := kit.Split(text, " (*)")
					if strings.HasPrefix(text, "func (") {
						m.Push(mdb.NAME, ls[3])
						m.Push(mdb.TEXT, text)
					} else if strings.HasPrefix(text, "func") {
						m.Push(mdb.NAME, ls[1])
						m.Push(mdb.TEXT, text)
					}
				})
				m.Action(html.FILTER)
				return
			}
			m.Cmdy("web.code.doc", arg)
			list := map[string]bool{}
			mdb.HashSelect(m.Spawn()).Table(func(value ice.Maps) { list[kit.Fields(value[nfs.PATH], value[nfs.FILE], value[nfs.LINE])] = true })
			kit.For(kit.SplitLine(m.Cmdx(cli.SYSTEM, cli.GOTAGS, "-f", "-", "-R", kit.Path(arg[0]))), func(text string) {
				if strings.HasPrefix(text, "!_") {
					return
				}
				ls := kit.Split(text)
				info := map[string]string{}
				kit.For(kit.Split(ls[3][1:], lex.TB, lex.TB), func(text string) {
					if ls := strings.SplitN(text, ":", 2); len(ls) > 1 {
						info[ls[0]] = ls[1]
					}
				})
				if strings.Contains(ls[1], "/internal/") || strings.HasSuffix(ls[1], "_test.go") {
					return
				}
				if len(arg) == 1 && info[aaa.ACCESS] == aaa.PRIVATE {
					return
				}
				t := kit.Split(ls[3])[0]
				if kit.IsIn(t, "p", "i") {
					return
				}
				file := kit.TrimPrefix(ls[1], kit.Path("")+nfs.PS, arg[0])
				has := list[kit.Fields(arg[0], file, ls[2])]
				if len(arg) == 1 {
					if kit.IsIn(t, "e", "w", "m") && !has {
						return
					}
				} else if info["ctype"] != arg[1] && ls[0] != arg[1] {
					return
				}
				text = info["signature"]
				if kit.IsIn(t, "f", "m") {
					p := "func "
					if t == "m" {
						if info["ntype"] == "" {
							p += "(" + strings.ToLower(info["ctype"]) + " " + info["ctype"] + ") "
						} else {
							p = ""
						}
					}
					if info[mdb.TYPE] == "" {
						text = kit.Format("%s%s%s", p, ls[0], text)
					} else if strings.Contains(info[mdb.TYPE], ",") {
						text = kit.Format("%s%s%s (%s)", p, ls[0], text, info[mdb.TYPE])
					} else {
						text = kit.Format("%s%s%s %s", p, ls[0], text, info[mdb.TYPE])
					}
				}
				m.Push(mdb.STATUS, has)
				m.Push(mdb.TYPE, t)
				m.Push(mdb.NAME, ls[0])
				m.Push(nfs.FILE, file)
				m.Push(nfs.LINE, ls[2])
				m.Push(mdb.TEXT, text)
				if has {
					m.PushButton(code.DOC, mdb.SHOW)
				} else {
					m.PushButton(code.DOC, mdb.SHOW, mdb.CREATE)
				}
			})
			m.Action(html.FILTER).StatusTimeCountStats(mdb.TYPE)
			m.Sort("status,type,name", []string{ice.TRUE, ice.FALSE}, []string{"t", "n", "f", "v", "c"}, ice.STR)
		}},
	})
}
