package code

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"net/http"
	_ "net/http/pprof"
	"strings"
)

const (
	PPROF   = "pprof"
	BINNARY = "binnary"
	SERVICE = "service"
	SECONDS = "seconds"
)

func _pprof_list(m *ice.Message, zone string, id string, field ...interface{}) {
	m.Richs(PPROF, nil, kit.Select(kit.MDB_FOREACH, zone), func(key string, val map[string]interface{}) {
		val = val[kit.MDB_META].(map[string]interface{})
		if zone = kit.Format(kit.Value(val, kit.MDB_ZONE)); id == "" {
			m.Grows(PPROF, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
				// 查看信息
				m.Push(kit.MDB_ZONE, zone)
				m.Push(BINNARY, val[BINNARY])
				m.Push(SERVICE, val[SERVICE])
				m.Push(SECONDS, val[SECONDS])
				m.Push("操作", `<input type="button" value="运行">`)
				m.Push(zone, value, []string{kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT})
			})
			return
		}
		m.Grows(PPROF, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			// 查看信息
			m.Push("detail", value)
		})
	})
}
func _pprof_show(m *ice.Message, zone string, seconds string) {
	m.Richs(PPROF, nil, zone, func(key string, val map[string]interface{}) {
		val = val[kit.MDB_META].(map[string]interface{})

		m.Gos(m.Spawn(), func(msg *ice.Message) {
			m.Sleep("1s").Grows(PPROF, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
				// 压测命令
				m.Cmd(ice.WEB_FAVOR, "pprof", "shell", value[kit.MDB_TEXT], m.Cmdx(kit.Split(kit.Format(value[kit.MDB_TEXT]))))
			})
		})

		// 启动监控
		name := zone + ".pd.gz"
		msg := m.Cmd(ice.WEB_SPIDE, "self", "cache", "GET", kit.Select("/code/pprof/profile", val[SERVICE]), "seconds", kit.Select("5", seconds))
		m.Cmd(ice.WEB_FAVOR, "pprof", "shell", "text", m.Cmdx(ice.CLI_SYSTEM, "go", "tool", "pprof", "-text", msg.Append("text")))
		m.Cmd(ice.WEB_FAVOR, "pprof", "pprof", name, msg.Append("data"))

		// 展示结果
		p := kit.Format("%s:%s", m.Conf(ice.WEB_SHARE, "meta.host"), m.Cmdx("tcp.getport"))
		m.Cmd(ice.CLI_DAEMON, "go", "tool", "pprof", "-http="+p, val[BINNARY], msg.Append("text"))
		m.Cmd(ice.WEB_FAVOR, "pprof", "bin", val[BINNARY], m.Cmd(ice.WEB_CACHE, "catch", "bin", val[BINNARY]).Append("data"))
		m.Cmd(ice.WEB_FAVOR, "pprof", "spide", msg.Append("text"), "http://"+p)
		m.Echo(p)
	})

}
func _pprof_modify(m *ice.Message, zone, id, pro, set, old string) {
	m.Richs(PPROF, nil, kit.Select(kit.MDB_FOREACH, zone), func(key string, val map[string]interface{}) {
		switch pro {
		case kit.MDB_ZONE, kit.MDB_ID, kit.MDB_TIME:
			m.Warn(true, "deny modify %v", pro)
			return
		case BINNARY, SERVICE, SECONDS:
			// 修改信息
			m.Log_MODIFY(kit.MDB_ZONE, zone, kit.MDB_KEY, pro, kit.MDB_VALUE, set, "old", old)
			val = val[kit.MDB_META].(map[string]interface{})
			kit.Value(val, pro, set)
			return
		}

		m.Grows(PPROF, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			// 修改信息
			m.Log_MODIFY(kit.MDB_ZONE, zone, kit.MDB_ID, id, kit.MDB_KEY, pro, kit.MDB_VALUE, set, "old", old)
			kit.Value(value, pro, set)
		})
	})
}
func _pprof_insert(m *ice.Message, zone, kind, name, text string, arg ...string) {
	m.Richs(PPROF, nil, zone, func(key string, val map[string]interface{}) {
		id := m.Grow(PPROF, kit.Keys(kit.MDB_HASH, key), kit.Dict(
			kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, text,
			// 添加信息
			kit.MDB_EXTRA, kit.Dict(arg),
		))
		m.Log_INSERT(kit.MDB_ZONE, zone, kit.MDB_ID, id, kit.MDB_TYPE, kind, kit.MDB_NAME, name)
		m.Echo("%d", id)
	})
}
func _pprof_create(m *ice.Message, zone string, binnary, service string, seconds string, arg ...string) {
	if m.Richs(PPROF, nil, zone, nil) == nil {
		m.Rich(PPROF, nil, kit.Data(kit.MDB_ZONE, zone,
			// 添加信息
			BINNARY, binnary, SERVICE, service, SECONDS, seconds, arg))
		m.Log_CREATE(kit.MDB_ZONE, zone)
	}
}

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PPROF: {Name: "pprof", Help: "性能分析", Value: kit.Data(kit.MDB_SHORT, kit.MDB_ZONE)},
		},
		Commands: map[string]*ice.Command{
			PPROF: {Name: "pprof zone=auto id=auto auto", Help: "性能分析", Action: map[string]*ice.Action{
				kit.MDB_CREATE: {Name: "create zone [binnary [service [seconds]]]", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_pprof_create(m, arg[0], kit.Select("bin/ice.bin", arg, 1),
						kit.Select("http://localhost:9020/code/pprof/profile", arg, 2), kit.Select("3", arg, 3))
				}},
				kit.MDB_INSERT: {Name: "insert zone type name text", Help: "插入", Hand: func(m *ice.Message, arg ...string) {
					_pprof_insert(m, arg[0], arg[1], arg[2], kit.Select("http://localhost:9020/code/bench?cmd="+arg[2], arg, 3))
				}},
				kit.MDB_MODIFY: {Name: "modify key value old", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					_pprof_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg[0], arg[1], kit.Select("", arg, 2))
				}},
				kit.MDB_SHOW: {Name: "show", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					_pprof_show(m, m.Option(kit.MDB_ZONE), m.Option(SECONDS))
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
