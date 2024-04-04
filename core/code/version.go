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
			SYNC: {Help: "同步", Hand: func(m *ice.Message, arg ...string) {
				repos := map[string]string{}
				m.Cmds(web.CODE_GIT_STATUS).Table(func(value ice.Maps) { repos[strings.Split(value[web.ORIGIN], "://")[1]] = value[nfs.VERSION] })
				m.Cmd(web.CODE_MOD, mdb.RENDER, MOD, ice.GO_MOD, nfs.PWD).Table(func(value ice.Maps) { repos[value[REQUIRE]] = value[VERSION] })
				res := m.Cmdx(nfs.CAT, path.Join(nfs.USR_LOCAL_WORK, m.Option(SPACE), ice.GO_MOD), func(ls []string, text string) string {
					if len(ls) > 1 {
						if v, ok := repos[ls[0]]; ok && !strings.Contains(v, "-") {
							m.Debug("what %v %v => %v", ls[0], ls[1], v)
							text = lex.TB + ls[0] + lex.SP + v
						}
					}
					return text
				})
				m.Cmd(nfs.SAVE, path.Join(nfs.USR_LOCAL_WORK, m.Option(SPACE), ice.GO_MOD), res)
				m.Cmd(SPACE, m.Option(SPACE), cli.SYSTEM, GO, MOD, "tidy")
			}},
			TAG: {Name: "tag version", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPACE, m.Option(SPACE), cli.SYSTEM, GIT, TAG, m.Option(VERSION))
				m.Cmd(SPACE, m.Option(SPACE), VIMER, COMPILE)
				m.Sleep3s()
			}},
			XTERM: {Hand: func(m *ice.Message, arg ...string) {
				web.ProcessPodCmd(m, m.Option(web.SPACE), m.ActionKey(), cli.SH, arg...)
			}},
			STATUS: {Hand: func(m *ice.Message, arg ...string) {
				web.ProcessPodCmd(m, m.Option(web.SPACE), m.ActionKey(), nil, arg...)
			}},
		}), Hand: func(m *ice.Message, arg ...string) {
			repos := map[string]string{}
			list := map[string]map[string]string{}
			list[ice.Info.Pathname] = map[string]string{}
			m.Cmds(web.CODE_GIT_REPOS).Table(func(value ice.Maps) { repos[strings.Split(value[web.ORIGIN], "://")[1]] = value[nfs.VERSION] })
			m.Cmd(web.CODE_MOD, mdb.RENDER, MOD, ice.GO_MOD, nfs.PWD).Table(func(value ice.Maps) {
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
				m.Cmd(web.SPACE, name, web.CODE_MOD, mdb.RENDER, MOD, ice.GO_MOD, nfs.PWD).Table(func(value ice.Maps) {
					if value[REPLACE] == nfs.PWD {
						list[name][MODULE] = value[REQUIRE]
						list[name][VERSION] = value[VERSION]
					} else {
						list[name][value[REQUIRE]] = value[VERSION]
					}
				})
				list[name][DIFF] = kit.ReplaceAll(m.Cmdx(web.SPACE, name, cli.SYSTEM, GIT, DIFF, "--shortstat"), " changed", "", "tions", "")
			})
			for space, v := range list {
				button, status := []ice.Any{}, ""
				m.Push(web.SPACE, space).Push(MODULE, v[MODULE]).Push(VERSION, v[VERSION])
				kit.For(repos, func(k, _v string) {
					if kit.IsIn(k, "shylinux.com/x/websocket", "shylinux.com/x/go-qrcode") {
						return
					}
					if k == v[MODULE] || v[k] == "" {
						m.Push(k, "")
					} else if v[k] == _v {
						m.Push(k, v[k])
					} else if strings.Contains(_v, "-") {
						m.Push(k, v[k]+" != "+_v)
						status = html.NOTICE
					} else {
						m.Push(k, v[k]+" => "+_v)
						status = html.DANGER
						button = append(button, SYNC)
					}
				})
				kit.If(list[space][DIFF] != "", func() { button = append(button, STATUS) })
				kit.If(strings.Contains(list[space][VERSION], "-"), func() { button = append(button, TAG) })
				kit.If(len(button) > 0, func() { button = append(button, XTERM) })
				m.Push(DIFF, list[space][DIFF]).Push(mdb.STATUS, status).PushButton(button...)
			}
			fields := []string{}
			kit.For(m.Appendv(ice.MSG_APPEND), func(k string) { kit.If(len(kit.TrimArg(m.Appendv(k)...)) > 0, func() { fields = append(fields, k) }) })
			m.Cut(fields...).Sort(web.SPACE, ice.STR_R)
		}},
	})

}
