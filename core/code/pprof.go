package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
	"github.com/shylinux/toolkits/task"

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
func _pprof_show(m *ice.Message, zone string, id string) {
	favor := m.Conf(PPROF, kit.Keys(kit.MDB_META, web.FAVOR))

	m.Richs(PPROF, nil, zone, func(key string, val map[string]interface{}) {
		val = val[kit.MDB_META].(map[string]interface{})

		list := []string{}
		task.Put(val, func(task *task.Task) error {
			m.Sleep("1s")
			m.Grows(PPROF, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
				// 压测命令
				m.Log_EXPORT(kit.MDB_META, PPROF, kit.MDB_ZONE, zone, kit.MDB_VALUE, kit.Format(value))
				cmd := kit.Format(value[kit.MDB_TYPE])
				arg := kit.Format(value[kit.MDB_TEXT])
				res := m.Cmd(mdb.ENGINE, value[kit.MDB_TYPE], value[kit.MDB_NAME], value[kit.MDB_TEXT], value[kit.MDB_EXTRA]).Result()
				m.Cmd(web.FAVOR, favor, cmd, arg, res)
				list = append(list, cmd+": "+arg, res)
			})
			return nil
		})

		// 收藏程序
		msg := m.Cmd(web.CACHE, web.CATCH, kit.MIME_FILE, kit.Format(val[BINNARY]))
		bin := msg.Append(kit.MDB_TEXT)
		m.Cmd(web.FAVOR, favor, kit.MIME_FILE, bin, val[BINNARY])

		// 性能分析
		msg = m.Cmd(web.SPIDE, "self", web.CACHE, http.MethodGet, kit.Select("/code/pprof/profile", val[SERVICE]), "seconds", kit.Select("5", kit.Format(val[SECONDS])))
		m.Cmd(web.FAVOR, favor, PPROF, msg.Append(kit.MDB_TEXT), kit.Keys(zone, "pd.gz"))

		// 结果摘要
		cmd := kit.Simple(m.Confv(PPROF, "meta.pprof"), "-text", val[BINNARY], msg.Append(kit.MDB_TEXT))
		res := strings.Split(m.Cmdx(cli.SYSTEM, cmd), "\n")
		if len(res) > 20 {
			res = res[:20]
		}
		m.Cmd(web.FAVOR, favor, web.TYPE_SHELL, strings.Join(cmd, " "), strings.Join(res, "\n"))
		list = append(list, web.TYPE_SHELL+": "+strings.Join(cmd, " "), strings.Join(res, "\n"))

		// 结果展示
		u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
		p := kit.Format("%s:%s", u.Hostname(), m.Cmdx(tcp.PORT, "get"))
		m.Option(cli.CMD_STDOUT, "var/daemon/stdout")
		m.Option(cli.CMD_STDERR, "var/daemon/stderr")
		m.Cmd(cli.DAEMON, m.Confv(PPROF, "meta.pprof"), "-http="+p, val[BINNARY], msg.Append(kit.MDB_TEXT))

		url := u.Scheme + "://" + p + "/ui/top"
		m.Cmd(web.FAVOR, favor, web.SPIDE, url, msg.Append(kit.MDB_TEXT))
		m.Set(ice.MSG_RESULT).Echo(url).Echo(" \n").Echo("\n")
		m.Echo(strings.Join(list, "\n")).Echo("\n")

		m.Push("url", url)
		m.Push(PPROF, msg.Append(kit.MDB_TEXT))
		m.Push(SERVICE, strings.Replace(kit.Format(val[SERVICE]), "profile", "", -1))
		m.Push("bin", bin)
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
				web.FAVOR, "pprof", PPROF, []string{"go", "tool", "pprof"},
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
					_pprof_show(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID))
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
