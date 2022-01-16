package mdb

import (
	"encoding/csv"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _zone_fields(m *ice.Message) []string {
	return kit.Split(kit.Select("zone,id,time,type,name,text", m.OptionFields()))
}
func _zone_select(m *ice.Message, prefix, chain, zone string, id string) {
	if zone == RANDOM {
		zone = RANDOMS
	}

	fields := _zone_fields(m)
	cb := m.OptionCB(SELECT)
	m.Richs(prefix, chain, kit.Select(FOREACH, zone), func(key string, val map[string]interface{}) {
		if val = kit.GetMeta(val); zone == "" {
			if m.OptionFields() == DETAIL {
				m.Push(DETAIL, val)
			} else {
				m.Push(key, val, fields)
			}
			return
		}

		m.Grows(prefix, kit.Keys(chain, HASH, key), ID, id, func(index int, value map[string]interface{}) {
			switch value = kit.GetMeta(value); cb := cb.(type) {
			case func(string, []string, map[string]interface{}, map[string]interface{}):
				cb(key, fields, value, val)
			case func(string, map[string]interface{}, map[string]interface{}):
				cb(key, value, val)
			case func(string, map[string]interface{}):
				cb(key, value)
			case func(map[string]interface{}):
				cb(value)
			case func(map[string]string):
				res := map[string]string{}
				for k, v := range value {
					res[k] = kit.Format(v)
				}
				cb(res)
			default:
				if m.FieldsIsDetail() {
					m.Push(DETAIL, value)
				} else {
					m.Push(key, value, fields, val)
				}
			}
		})
	})
}
func _zone_export(m *ice.Message, prefix, chain, file string) {
	f, p, e := kit.Create(kit.Keys(file, CSV))
	m.Assert(e)
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	fields := _zone_fields(m)
	fields = append(fields, EXTRA)
	w.Write(fields)

	count := 0
	m.Richs(prefix, chain, FOREACH, func(key string, val map[string]interface{}) {
		val = kit.GetMeta(val)

		m.Grows(prefix, kit.Keys(chain, HASH, key), "", "", func(index int, value map[string]interface{}) {
			value = kit.GetMeta(value)

			list := []string{}
			for _, k := range fields {
				list = append(list, kit.Select(kit.Format(kit.Value(val, k)), kit.Format(kit.Value(value, k))))
			}
			w.Write(list)
			count++
		})
	})

	m.Log_EXPORT(KEY, path.Join(prefix, chain), FILE, p, COUNT, count)
	m.Conf(prefix, kit.Keys(chain, HASH), "")
	m.Echo(p)
}
func _zone_import(m *ice.Message, prefix, chain, file string) {
	f, e := os.Open(kit.Keys(file, CSV))
	m.Assert(e)
	defer f.Close()

	r := csv.NewReader(f)
	head, _ := r.Read()
	count := 0

	list := map[string]string{}
	zkey := kit.Select(head[0], m.OptionFields())

	for {
		line, e := r.Read()
		if e != nil {
			break
		}

		zone := ""
		data := kit.Dict()
		for i, k := range head {
			switch k {
			case zkey:
				zone = line[i]
			case ID:
				continue
			case EXTRA:
				kit.Value(data, k, kit.UnMarshal(line[i]))
			default:
				kit.Value(data, k, line[i])
			}
		}
		if list[zone] == "" {
			list[zone] = m.Rich(prefix, chain, kit.Data(zkey, zone))
		}

		m.Grow(prefix, kit.Keys(chain, HASH, list[zone]), data)
		count++
	}

	m.Log_IMPORT(KEY, path.Join(prefix, chain), COUNT, count)
	m.Echo("%d", count)
}

const ZONE = "zone"

func ZoneAction(args ...interface{}) map[string]*ice.Action {
	_zone := func(m *ice.Message) string { return kit.Select(ZONE, m.Config(SHORT)) }

	return ice.SelectAction(map[string]*ice.Action{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
			if cs := m.Target().Configs; cs[m.CommandKey()] == nil {
				cs[m.CommandKey()] = &ice.Config{Value: kit.Data(args...)}
			}
		}},
		INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
			arg[0] = strings.TrimPrefix(arg[0], "extra.")
			arg[0] = kit.Select(arg[0], m.Config(kit.Keys(ALIAS, arg[0])))
			switch arg[0] {
			case ice.POD:
				m.Cmdy("route")
			case ice.CTX:
				m.Cmdy("context")
			case ice.CMD:
				m.Cmdy("context", kit.Select(m.Option(ice.CTX), m.Option(kit.Keys(EXTRA, ice.CTX))), "command")
			case ice.ARG:

			case "index":
				m.OptionFields(arg[0])
				m.Cmdy("command", SEARCH, "command", kit.Select("", arg, 1))

			case _zone(m):
				m.Cmdy(INPUTS, m.PrefixKey(), "", HASH, arg)
			default:
				m.Cmdy(INPUTS, m.PrefixKey(), "", ZONE, m.Option(_zone(m)), arg)
			}
		}},
		CREATE: {Name: "create zone", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INSERT, m.PrefixKey(), "", HASH, arg)
		}},
		REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(DELETE, m.PrefixKey(), "", HASH, m.OptionSimple(_zone(m)), arg)
		}},
		INSERT: {Name: "insert zone type=go name=hi text=hello", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				arg = m.OptionSimple(_zone(m), m.Config(FIELD))
			}
			m.Cmdy(INSERT, m.PrefixKey(), "", HASH, _zone(m), arg[1])
			m.Cmdy(INSERT, m.PrefixKey(), "", ZONE, m.Option(_zone(m)), arg[2:])
		}},
		MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(MODIFY, m.PrefixKey(), "", ZONE, m.Option(_zone(m)), m.Option(ID), arg)
		}},
		PLUGIN: {Name: "plugin extra.pod extra.ctx extra.cmd extra.arg", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(MODIFY, m.PrefixKey(), "", ZONE, m.Option(_zone(m)), m.Option(ID), arg)
		}},
		EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
			m.Option(ice.CACHE_LIMIT, "-1")
			m.OptionFields(_zone(m), m.Config(FIELD))
			m.Cmdy(EXPORT, m.PrefixKey(), "", ZONE)
		}},
		IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
			m.OptionFields(_zone(m))
			m.Cmdy(IMPORT, m.PrefixKey(), "", ZONE)
		}},
		PREV: {Name: "prev", Help: "上一页", Hand: func(m *ice.Message, arg ...string) {
			PrevPage(m, arg[0], arg[1:]...)
		}},
		NEXT: {Name: "next", Help: "下一页", Hand: func(m *ice.Message, arg ...string) {
			NextPageLimit(m, arg[0], arg[1:]...)
		}},
		SELECT: {Name: "select", Help: "列表", Hand: func(m *ice.Message, arg ...string) {
			ZoneSelect(m, arg...)
		}},
	})
}
func ZoneSelect(m *ice.Message, arg ...string) *ice.Message {
	m.Fields(len(arg), kit.Fields(TIME, m.Config(SHORT), COUNT), m.Config(FIELD))
	if m.Cmdy(SELECT, m.PrefixKey(), "", ZONE, arg); kit.Select("", arg, 0) == "" {
		m.Sort(m.Config(SHORT))
		m.PushAction(REMOVE)
	}
	m.StatusTimeCount()
	return m
}
func ZoneSelectAll(m *ice.Message, arg ...string) *ice.Message {
	m.Option(ice.CACHE_LIMIT, "-1")
	return ZoneSelect(m, arg...)
}
func ZoneSelectCB(m *ice.Message, zone string, cb interface{}) *ice.Message {
	m.OptionCB(SELECT, cb)
	m.Option(ice.CACHE_LIMIT, "-1")
	return ZoneSelect(m, zone)
}
