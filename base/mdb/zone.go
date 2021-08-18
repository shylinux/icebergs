package mdb

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const ZONE_FIELD = "time,zone,count"

func ZoneAction(fields ...string) map[string]*ice.Action {
	_zone := func(m *ice.Message) string {
		return kit.Select(kit.MDB_ZONE, m.Conf(m.PrefixKey(), kit.Keym(kit.MDB_SHORT)))
	}
	return selectAction(map[string]*ice.Action{
		CREATE: {Name: "create zone", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INSERT, m.PrefixKey(), "", HASH, arg)
		}},
		INSERT: {Name: "insert zone= type=go name=hi text=hello", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INSERT, m.PrefixKey(), "", HASH, _zone(m), arg[1])
			m.Cmdy(INSERT, m.PrefixKey(), "", ZONE, m.Option(_zone(m)), arg[2:])
		}},
		MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(MODIFY, m.PrefixKey(), "", ZONE, m.Option(_zone(m)), m.Option(kit.MDB_ID), arg)
		}},
		REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(DELETE, m.PrefixKey(), "", HASH, m.OptionSimple(_zone(m)))
		}},
		EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
			m.OptionFields(_zone(m), m.Conf(m.PrefixKey(), kit.META_FIELD))
			m.Cmdy(EXPORT, m.PrefixKey(), "", ZONE)
			m.Conf(m.PrefixKey(), kit.MDB_HASH, "")
		}},
		IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
			m.OptionFields(_zone(m))
			m.Cmdy(IMPORT, m.PrefixKey(), "", ZONE)
		}},
		INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
			switch arg[0] {
			case _zone(m):
				m.Cmdy(INPUTS, m.PrefixKey(), "", HASH, arg)
			default:
				m.Cmdy(INPUTS, m.PrefixKey(), "", ZONE, m.Option(_zone(m)), arg)
			}
		}},
	}, fields...)
}
