package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func init() {
	Index.MergeCommands(ice.Commands{
		VERSION: {Name: "version refresh", Help: "版本", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				ls := kit.Split(kit.Split(strings.TrimPrefix(m.Option(VERSION), "v"), "-")[0], ".")
				if len(ls) > 2 {
					m.Push(arg[0], kit.Format("v%d.%d.%d", kit.Int(ls[0]), kit.Int(ls[1]), kit.Int(ls[2])+1))
					m.Push(arg[0], kit.Format("v%d.%d.%d", kit.Int(ls[0]), kit.Int(ls[1])+1, 0))
					m.Push(arg[0], kit.Format("v%d.%d.%d", kit.Int(ls[0])+1, 0, 0))
				} else {
					m.Push(arg[0], "v0.0.1")
				}
			}},
			"sync": {Help: "同步", Hand: func(m *ice.Message, arg ...string) {
				repos := map[string]string{}
				m.Cmd("web.code.mod", mdb.RENDER, MOD, ice.GO_MOD, nfs.PWD).Table(func(value ice.Maps) { repos[value[REQUIRE]] = value[VERSION] })
				res := m.Cmdx(nfs.CAT, path.Join(nfs.USR_LOCAL_WORK, m.Option(SPACE), ice.GO_MOD), func(ls []string, text string) string {
					if len(ls) == 2 {
						if v, ok := repos[ls[0]]; ok {
							text = lex.TB + ls[0] + lex.SP + v
						}
					}
					return text
				})
				m.Cmd(nfs.SAVE, path.Join(nfs.USR_LOCAL_WORK, m.Option(SPACE), ice.GO_MOD), res)
				m.Cmd(SPACE, m.Option(SPACE), cli.SYSTEM, GO, MOD, "tidy")
			}},
			"add": {Name: "add version", Hand: func(m *ice.Message, arg ...string) {
			}},
		}), Hand: func(m *ice.Message, arg ...string) {
			repos := map[string]string{}
			list := map[string]map[string]string{}
			list[ice.Info.Pathname] = map[string]string{}
			m.Cmd("web.code.mod", mdb.RENDER, MOD, ice.GO_MOD, nfs.PWD).Table(func(value ice.Maps) {
				list[ice.Info.Pathname][value[REQUIRE]] = value[VERSION]
				if value[REPLACE] == nfs.PWD {
					list[ice.Info.Pathname][MODULE] = value[REQUIRE]
					list[ice.Info.Pathname][VERSION] = value[VERSION]
				} else {
					repos[value[REQUIRE]] = value[VERSION]
				}
			})
			web.DreamEach(m, m.Option(mdb.NAME), "", func(name string) {
				list[name] = map[string]string{}
				m.Cmd(web.SPACE, name, "web.code.mod", mdb.RENDER, MOD, ice.GO_MOD, nfs.PWD).Table(func(value ice.Maps) {
					if value[REPLACE] == nfs.PWD {
						list[name][MODULE] = value[REQUIRE]
						list[name][VERSION] = value[VERSION]
					} else {
						list[name][value[REQUIRE]] = value[VERSION]
					}
				})
				list[name]["diff"] = kit.ReplaceAll(m.Cmdx(web.SPACE, name, cli.SYSTEM, "git", "diff", "--shortstat"), " changed", "", "tions", "")
			})
			for space, v := range list {
				diff := false
				m.Push(web.SPACE, space)
				m.Push(MODULE, list[space][MODULE])
				m.Push(VERSION, list[space][VERSION])
				kit.For(repos, func(k, _v string) {
					m.Push(k, v[k])
					kit.If(v[k] != "" && v[k] != _v, func() { diff = true })
				})
				m.Push("diff", list[space]["diff"])
				if diff {
					m.Push(mdb.STATUS, html.DANGER).PushButton("sync", "add")
				} else {
					if strings.Contains(m.Option(VERSION), "-") {
						m.Push(mdb.STATUS, "").PushButton("add")
					} else {
						m.Push(mdb.STATUS, "").PushButton("")
					}
				}
			}
			m.Sort(web.SPACE, ice.STR_R)
		}},
	})

}
