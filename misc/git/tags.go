package git

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func init() {
	const TAGS = "tags"
	Index.MergeCommands(ice.Commands{
		TAGS: {Name: "tags path name auto", Meta: kit.Dict(
			ice.CTX_ICONS, kit.Dict(mdb.CREATE, "bi bi-star"),
		), Actions: ice.MergeActions(ice.Actions{
			"show": {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessFloat(m, web.VIMER, []string{m.Option(nfs.PATH), m.Option(nfs.FILE), m.Option(nfs.LINE)}, arg...)
			}},
		}, mdb.HashAction(mdb.SHORT, "path,file,line", mdb.FIELD, "time,hash,type,name,text,path,file,line")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m)
				return
			}
			defer m.Action(html.FILTER)
			list := map[string]bool{}
			mdb.HashSelect(m.Spawn()).Table(func(value ice.Maps) { list[kit.Fields(value[nfs.PATH], value[nfs.FILE], value[nfs.LINE])] = true })
			kit.For(kit.SplitLine(m.Cmdx(cli.SYSTEM, "gotags", "-f", "-", "-R", kit.Path(arg[0]))), func(text string) {
				if strings.HasPrefix(text, "!_") {
					return
				}
				ls := kit.Split(text)
				t := kit.Split(ls[3])[0]
				info := map[string]string{}
				kit.For(kit.Split(ls[3][1:], "\t", "\t"), func(text string) {
					if ls := strings.SplitN(text, ":", 2); len(ls) > 1 {
						info[ls[0]] = ls[1]
					}
				})
				if strings.Contains(ls[1], "/internal/") || strings.HasSuffix(ls[1], "_test.go") {
					return
				}
				if len(arg) == 1 && info["access"] == "private" {
					return
				}
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
				m.Push(mdb.STATUS, has)
				m.Push(mdb.TYPE, t)
				m.Push(mdb.NAME, ls[0])
				m.Push(nfs.FILE, file)
				m.Push(nfs.LINE, ls[2])
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
						m.Push(mdb.TEXT, kit.Format("%s%s%s", p, ls[0], info["signature"]))
					} else if strings.Contains(info[mdb.TYPE], ",") {
						m.Push(mdb.TEXT, kit.Format("%s%s%s (%s)", p, ls[0], info["signature"], info[mdb.TYPE]))
					} else {
						m.Push(mdb.TEXT, kit.Format("%s%s%s %s", p, ls[0], info["signature"], info[mdb.TYPE]))
					}
				} else {
					m.Push(mdb.TEXT, info["signature"])
				}
				m.Push("what", ls[3])
				if has {
					m.PushButton("show")
				} else {
					m.PushButton("show", mdb.CREATE)
				}
			})
			m.StatusTimeCountStats(mdb.TYPE)
			m.Sort("status,type,name", []string{"true", "false"}, []string{"t", "n", "f", "v", "c"}, ice.STR)
		}},
	})
}
