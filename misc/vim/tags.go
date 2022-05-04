package vim

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const TAGS = "tags"

func init() {
	const (
		MODULE  = "module"
		PATTERN = "pattern"

		ONIMPORT = "onimport"
		ONACTION = "onaction"
		ONEXPORT = "onexport"

		defs_pattern = "4\n%s\n/\\<%s: /\n"
		func_pattern = "4\n%s\n/\\<%s: \\(shy\\|func\\)/\n"
		libs_pattern = "4\nusr/volcanos/lib/%s.js\n/\\<%s: \\(shy\\|func\\)/\n"
	)
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		TAGS: {Name: TAGS, Help: "索引", Value: kit.Data(
			mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,id,type,name,text,path,file,line",
		)},
	}, Commands: map[string]*ice.Command{
		"/tags": {Name: "/tags", Help: "跳转", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch m.Option(MODULE) {
			case ONIMPORT, ONACTION, ONEXPORT:
				m.Echo(func_pattern, m.Option(BUF), m.Option(PATTERN))
			case "msg", "res":
				m.Echo(libs_pattern, ice.MISC, m.Option(PATTERN))
			default:
				if mdb.ZoneSelectCB(m, m.Option(MODULE), func(value map[string]string) {
					if value[mdb.NAME] == m.Option(PATTERN) {
						m.Echo(kit.Select(defs_pattern, func_pattern, value[mdb.TYPE] == "function"),
							path.Join(value[nfs.PATH], value[nfs.FILE]), m.Option(PATTERN))
					}
				}); m.Length() == 0 {
					m.Echo(defs_pattern, "usr/volcanos/proto.js", m.Option(PATTERN))
				}
			}
		}},
		TAGS: {Name: "tags zone id auto", Help: "索引", Action: ice.MergeAction(map[string]*ice.Action{
			"listTags": {Name: "listTags", Help: "索引", Hand: func(m *ice.Message, arg ...string) {
				kit.Fetch(kit.UnMarshal(m.Option(mdb.TEXT)), func(index int, value map[string]interface{}) {
					m.Cmd(TAGS, mdb.INSERT, mdb.ZONE, value[mdb.ZONE], kit.Simple(value))
				})
				m.ProcessRefresh300ms()
			}},
			mdb.INSERT: {Name: "insert zone=core type name=hi text=hello path file line", Help: "添加"},
			code.INNER: {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessCommand(code.INNER, m.OptionSplit("path,file,line"), arg...)
			}},
			INPUT: {Name: "input name text", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(mdb.TEXT) == "" {
					return
				}
				mdb.ZoneSelectCB(m, kit.Slice(kit.Split(m.Option(mdb.TEXT), ice.PT), -1)[0], func(value map[string]string) {
					if !strings.Contains(value[mdb.NAME], m.Option(mdb.NAME)) {
						return
					}
					if m.Length() == 0 {
						m.Echo("func" + ice.NL)
					}
					m.Echo(value[mdb.NAME] + ice.NL)
					m.Echo("%s: %s"+ice.NL, value[mdb.NAME], strings.Split(value[mdb.TEXT], ice.NL)[0])
				})
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.ZoneSelectAll(m, arg...); len(arg) == 0 {
				m.Action("listTags", mdb.CREATE, mdb.EXPORT, mdb.IMPORT)
			} else {
				m.Action(mdb.INSERT).PushAction(code.INNER).StatusTimeCount()
			}
		}},
	}})
}
