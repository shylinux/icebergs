package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const CASE = "case"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		CASE: {Name: CASE, Help: "用例", Value: kit.Data(
			mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,id,name,cmd,api,arg,res",
		)},
	}, Commands: map[string]*ice.Command{
		CASE: {Name: "case dev zone id auto", Help: "用例", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.CREATE: {Name: "create name address", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.SPIDE, mdb.CREATE, arg)
			}},
			mdb.INSERT: {Name: "insert zone name=hi cmd=POST,GET api arg:textarea res:textarea", Help: "添加"},

			cli.CHECK: {Name: "check", Help: "检查", Hand: func(m *ice.Message, arg ...string) {
				if m.ProcessInner(); len(arg) > 0 {
					success := 0
					m.Cmd(m.PrefixKey(), arg[0]).Table(func(index int, value map[string]string, head []string) {
						m.Push(mdb.TIME, m.Time())
						m.Push(mdb.ID, value[mdb.ID])
						if err := m.Cmdx(m.PrefixKey(), cli.CHECK, value); err == ice.OK {
							m.Push(ice.ERR, cli.Color(m, cli.GREEN, err))
							success++
						} else {
							m.Push(ice.ERR, cli.Color(m, cli.RED, err))
						}
						m.Push(cli.API, value[cli.API])
						m.Push(ice.ARG, value[ice.ARG])
						m.Push(ice.RES, value[ice.RES])
					})
					m.StatusTimeCount(ice.SUCCESS, success)
					return
				}

				res := kit.UnMarshal(m.Cmdx(m.PrefixKey(), ice.RUN))
				if m.Option(ice.RES) != "" {
					for k, v := range kit.KeyValue(nil, "", kit.UnMarshal(m.Option(ice.RES))) {
						if v != kit.Value(res, k) {
							m.Echo(kit.Formats(res))
							return
						}
					}
				}
				m.Echo(ice.OK)
			}},
			ice.RUN: {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SPIDE_HEADER, web.ContentType, web.ContentJSON)
				m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(web.SPIDE, m.Option(ice.DEV), web.SPIDE_RAW,
					m.Option(ice.CMD), m.Option(cli.API), web.SPIDE_DATA, m.Option(ice.ARG)))))
				m.Info(`curl "` + m.Option(cli.API) + `" -H "Content-Type: application/json"` + ` -d '` + m.Option(ice.ARG) + `'`)
				m.ProcessDisplay("/plugin/story/json.js")
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(web.SPIDE)
				m.Action(mdb.CREATE)
				m.RenameAppend("client.name", "dev")
				m.RenameAppend("client.url", "address")
				return
			}

			if mdb.ZoneSelect(m, arg[1:]...); len(arg) == 1 {
				m.Action(mdb.INSERT, mdb.EXPORT, mdb.IMPORT)
				m.PushAction(mdb.INSERT, cli.CHECK, mdb.REMOVE)
			} else {
				m.Action(mdb.INSERT, cli.CHECK)
				m.PushAction(ice.RUN, cli.CHECK)
			}
			m.StatusTimeCount()
		}},
		"test": {Name: "test path func auto run case", Help: "测试用例", Action: map[string]*ice.Action{
			"run": {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
				// cli.Follow(m, "run", func() {
				m.Option(cli.CMD_DIR, kit.Select(path.Dir(arg[0]), arg[0], strings.HasSuffix(arg[0], "/")))
				m.Cmdy(cli.SYSTEM, "go", "test", ice.PWD, "-v", "-run="+arg[1])
				// })
			}},
			"case": {Name: "case", Help: "用例", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Spawn()
				if strings.HasSuffix(arg[0], "/") {
					msg.Option(cli.CMD_DIR, arg[0])
					msg.Split(msg.Cmdx(cli.SYSTEM, "grep", "-r", "func Test.*(", ice.PWD), "file:line", ":", "\n")
					msg.Table(func(index int, value map[string]string, head []string) {
						if strings.HasPrefix(strings.TrimSpace(value["line"]), "//") {
							return
						}
						ls := kit.Split(value["line"], " (", " (", " (")
						m.Push("file", value["file"])
						m.Push("func", strings.TrimPrefix(ls[1], "Test"))
					})
				} else {
					for _, line := range kit.Split(m.Cmdx(cli.SYSTEM, "grep", "^func Test.*(", arg[0]), "\n", "\n", "\n") {
						ls := kit.Split(line, " (", " (", " (")
						m.Push("func", strings.TrimPrefix(ls[1], "Test"))
					}
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 || arg[0] == "" {
				m.Cmdy(nfs.DIR, ice.PWD)
				return
			}
			if len(arg) == 1 {
				if strings.HasSuffix(arg[0], "/") {
					m.Cmdy(nfs.DIR, arg[0])
				} else {
					for _, line := range kit.Split(m.Cmdx(cli.SYSTEM, "grep", "^func Test.*(", arg[0]), "\n", "\n", "\n") {
						ls := kit.Split(line, " (", " (", " (")
						m.Push("func", strings.TrimPrefix(ls[1], "Test"))
					}
				}
				return
			}

			m.Option(cli.CMD_DIR, kit.Select(path.Dir(arg[0]), arg[0], strings.HasSuffix(arg[0], "/")))
			m.Cmdy(cli.SYSTEM, "go", "test", ice.PWD, "-v", "-run="+arg[1])
		}},
	}})
}
