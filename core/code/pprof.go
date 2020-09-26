package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"net/http"
	_ "net/http/pprof"
	"strings"
)

func _pprof_list(m *ice.Message, zone string, id string, field ...interface{}) {
	m.Richs(PPROF, nil, kit.Select(kit.MDB_FOREACH, zone), func(key string, val map[string]interface{}) {
		val = val[kit.MDB_META].(map[string]interface{})
		if zone = kit.Format(kit.Value(val, kit.MDB_ZONE)); id == "" {
			m.Grows(PPROF, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
				// 列表信息
				m.Push("操作", m.Cmdx(mdb.RENDER, web.RENDER.Button, "运行"))
				m.Push(zone, value, []string{
					kit.MDB_ZONE, kit.MDB_ID, kit.MDB_TYPE,
					kit.MDB_NAME, kit.MDB_TEXT, SECONDS, BINNARY, SERVICE,
				}, val)
			})
		} else {
			m.Grows(PPROF, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
				// 详细信息
				m.Push("detail", value)
				m.Push(kit.MDB_KEY, "操作")
				m.Push(kit.MDB_VALUE, m.Cmdx(mdb.RENDER, web.RENDER.Button, "运行"))
			})
		}
	})
}
func _pprof_modify(m *ice.Message, zone, id, k, v, old string) {
	k = kit.Select(k, m.Option(kit.MDB_KEY))
	m.Richs(PPROF, nil, kit.Select(kit.MDB_FOREACH, zone), func(key string, val map[string]interface{}) {
		switch k {
		case kit.MDB_ZONE, kit.MDB_ID, kit.MDB_TIME:
			// m.Warn(true, mdb.ErrDenyModify, k)
			return
		case BINNARY, SERVICE, SECONDS:
			// 修改信息
			m.Log_MODIFY(kit.MDB_ZONE, zone, kit.MDB_KEY, k, kit.MDB_VALUE, v, "old", old)
			val = val[kit.MDB_META].(map[string]interface{})
			kit.Value(val, k, v)
			return
		}

		m.Grows(PPROF, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			// 修改信息
			m.Log_MODIFY(kit.MDB_ZONE, zone, kit.MDB_ID, id, kit.MDB_KEY, k, kit.MDB_VALUE, v, "old", old)
			kit.Value(value, k, v)
		})
	})
}
func _pprof_insert(m *ice.Message, zone, kind, name, text string, arg ...string) {
	m.Richs(PPROF, nil, zone, func(key string, val map[string]interface{}) {
		id := m.Grow(PPROF, kit.Keys(kit.MDB_HASH, key), kit.Dict(
			kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, text,
			kit.MDB_EXTRA, kit.Dict(arg),
		))
		m.Log_INSERT(kit.MDB_META, PPROF, kit.MDB_ZONE, zone, kit.MDB_ID, id, kit.MDB_TYPE, kind, kit.MDB_NAME, name)
		m.Echo("%d", id)
	})
}
func _pprof_create(m *ice.Message, zone string, binnary, service string, seconds string, arg ...string) {
	m.Rich(PPROF, nil, kit.Data(kit.MDB_ZONE, zone, BINNARY, binnary, SERVICE, service, SECONDS, seconds, arg))
	m.Log_CREATE(kit.MDB_ZONE, zone)
}

const PPROF = "pprof"
const (
	BINNARY = "binnary"
	SERVICE = "service"
	SECONDS = "seconds"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PPROF: {Name: "pprof", Help: "性能分析", Value: kit.Data(kit.MDB_SHORT, kit.MDB_ZONE,
				PPROF, []string{"go", "tool", "pprof"},
			)},
		},
		Commands: map[string]*ice.Command{
			PPROF: {Name: "pprof zone=auto id=auto auto", Help: "性能分析", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create zone [binnary [service [seconds]]]", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_pprof_create(m, arg[0], kit.Select("bin/ice.bin", arg, 1),
						kit.Select("http://localhost:9020/code/pprof/profile", arg, 2), kit.Select("3", arg, 3))
				}},
				mdb.INSERT: {Name: "insert zone type name [text]", Help: "插入", Hand: func(m *ice.Message, arg ...string) {
					if len(arg) == 2 {
						arg = append(arg, "")
					}
					if len(arg) == 3 {
						arg = append(arg, "")
					}
					_pprof_insert(m, arg[0], arg[1], arg[2], kit.Select("http://localhost:9020/code/bench?cmd="+arg[2], arg, 3), arg[4:]...)
				}},
				mdb.MODIFY: {Name: "modify key value old", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					_pprof_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg[0], arg[1], kit.Select("", arg, 2))
				}},
				kit.MDB_SHOW: {Name: "show type name text arg...", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					// _pprof_show(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_pprof_list(m, kit.Select(kit.MDB_FOREACH, arg, 0), kit.Select("", arg, 1))
			}},
			"/pprof/": {Name: "/pprof/", Help: "性能分析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.R.URL.Path = strings.Replace("/code"+m.R.URL.Path, "code", "debug", 1)
				http.DefaultServeMux.ServeHTTP(m.W, m.R)
				m.Render(ice.RENDER_VOID)
			}},
		},
	}, nil)
}
