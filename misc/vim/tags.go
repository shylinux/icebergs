package vim

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const TAGS = "tags"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		TAGS: {Name: TAGS, Help: "索引", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "time,id,type,name,text,file,line",
		)},
	}, Commands: map[string]*ice.Command{
		"/tags": {Name: "/tags", Help: "跳转", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch m.Option("module") {
			case "onimport", "onaction", "onexport":
				m.Echo("4\n%s\n/\\<%s: \\(shy\\|func\\)/\n", m.Option(BUF), m.Option("pattern"))
			case "msg":
				m.Echo("4\nusr/volcanos/lib/%s.js\n/\\<%s: \\(shy\\|func\\)/\n", "misc", m.Option("pattern"))
			default:
				if mdb.ZoneSelect(m, m.Option("module")); m.Length() > 0 {
					switch m.Append(kit.MDB_TYPE) {
					case "function":
						m.Echo("4\nusr/volcanos%s\n/\\<%s: \\(shy\\|func\\)/\n", m.Append(kit.MDB_FILE), m.Option("pattern"))
					default:
						m.Echo("4\nusr/volcanos%s\n/\\<%s: /\n", m.Append(kit.MDB_FILE), m.Option("pattern"))
					}
					return
				}
				m.Echo("4\n%s\n/\\<%s: /\n", "usr/volcanos/proto.js", m.Option("pattern"))
			}
		}},
		TAGS: {Name: "tags zone id auto", Help: "索引", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INSERT: {Name: "insert zone=core type name=hi text=hello file line", Help: "添加"},
			code.INNER: {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessCommand(code.INNER, []string{
					kit.Select(ice.PWD, path.Dir(m.Option(kit.MDB_FILE))),
					path.Base(m.Option(kit.MDB_FILE)),
					m.Option(kit.MDB_LINE),
				}, arg...)
			}},
			"listTags": {Name: "listTags", Help: "索引", Hand: func(m *ice.Message, arg ...string) {
				kit.Fetch(kit.UnMarshal(m.Option("content")), func(index int, value map[string]interface{}) {
					m.Cmd(TAGS, mdb.INSERT, kit.MDB_ZONE, value[kit.MDB_ZONE], kit.Simple(value))
				})
				m.ProcessRefresh30ms()
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(ice.CACHE_LIMIT, "-1")
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.Action("listTags", mdb.CREATE, mdb.EXPORT, mdb.IMPORT)
				m.PushAction(mdb.REMOVE)
			} else {
				if m.IsCliUA() {
					if m.Length() == 0 {
						return
					}
					m.Sort(kit.MDB_NAME)
					m.Echo("func\n").Table(func(index int, value map[string]string, head []string) {
						m.Echo(arg[0] + ice.PT + value[kit.MDB_NAME] + ice.NL)
						m.Echo("%s: %s: %s // %s\n", value[kit.MDB_TYPE], value[kit.MDB_NAME], strings.Split(value[kit.MDB_TEXT], ice.NL)[0], value[kit.MDB_FILE])
					})
					return
				}
				m.Action(mdb.INSERT)
				m.PushAction(code.INNER)
				m.StatusTimeCount()
			}
		}},
	}})
}
