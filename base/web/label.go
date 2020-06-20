package web

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"sync"
)

func _label_add(m *ice.Message, cmd string) {
	if m.Option(cmd) != "" && m.Option(kit.MDB_GROUP) != "" && m.Option(kit.MDB_NAME) != "" {
		m.Cmdy(cmd, m.Option(cmd), "add", m.Option(kit.MDB_GROUP), m.Option(kit.MDB_NAME))
		m.Option(ice.FIELD_RELOAD, "true")
	}
}
func _label_del(m *ice.Message, cmd string) {
	if m.Option(cmd) != "" && m.Option(kit.MDB_GROUP) != "" && m.Option(kit.MDB_NAME) != "" {
		m.Cmdy(cmd, m.Option(cmd), "del", m.Option(kit.MDB_GROUP), m.Option(kit.MDB_NAME))
		m.Option(ice.FIELD_RELOAD, "true")
	}
}
func _label_prune(m *ice.Message, cmd string) {
	m.Richs(cmd, nil, m.Option(cmd), func(key string, value map[string]interface{}) {
		m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), kit.MDB_FOREACH, func(sub string, value map[string]interface{}) {
			if value[kit.MDB_STATUS] != "busy" {
				m.Cmdy(cmd, m.Option(cmd), "del", value[kit.MDB_GROUP], value[kit.MDB_NAME])
				m.Option(ice.FIELD_RELOAD, "true")
			}
		})
	})
}
func _label_clear(m *ice.Message, cmd string) {
	m.Richs(cmd, nil, m.Option(cmd), func(key string, value map[string]interface{}) {
		m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), kit.MDB_FOREACH, func(sub string, value map[string]interface{}) {
			if value[kit.MDB_STATUS] == "void" {
				last := m.Conf(cmd, kit.Keys(kit.MDB_HASH, key, kit.MDB_HASH, sub))
				m.Logs(ice.LOG_DELETE, cmd, m.Option(cmd), kit.MDB_NAME, value[kit.MDB_NAME], kit.MDB_VALUE, last)
				m.Conf(cmd, kit.Keys(kit.MDB_HASH, key, kit.MDB_HASH, sub), "")
				m.Option(ice.FIELD_RELOAD, "true")
				m.Echo(last)
			}
		})
	})
}
func _label_delete(m *ice.Message, cmd string) {
	m.Richs(cmd, nil, m.Option(cmd), func(key string, value map[string]interface{}) {
		m.Echo(m.Conf(cmd, kit.Keys(kit.MDB_HASH, key)))
		m.Logs(ice.LOG_REMOVE, cmd, m.Option(cmd), kit.MDB_VALUE, m.Conf(cmd, kit.Keys(kit.MDB_HASH, key)))
		m.Conf(cmd, kit.Keys(kit.MDB_HASH, key), "")
		m.Option(ice.FIELD_RELOAD, "true")
	})
}

func _label_select(m *ice.Message, cmd string, arg ...string) {
	m.Richs(cmd, nil, kit.Select("*", arg, 0), func(key string, value map[string]interface{}) {
		if len(arg) < 1 {
			// 一级列表
			m.Option(ice.FIELD_DETAIL, "清理", "清空", "删除")
			value = value[kit.MDB_META].(map[string]interface{})
			m.Push(key, value, []string{kit.MDB_TIME})
			status := map[string]int{}
			m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
				status[kit.Format(value[kit.MDB_STATUS])]++
			})
			m.Push("count", kit.Format("%d/%d/%d", status["busy"], status["free"], status["void"]))
			m.Push(key, value, []string{cmd})
			return
		}
		m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), kit.Select("*", arg, 1), func(key string, value map[string]interface{}) {
			if len(arg) < 2 {
				// 二级列表
				m.Option(ice.FIELD_DETAIL, "添加", "退还", "清理", "清空")
				m.Push(key, value, []string{kit.MDB_TIME, kit.MDB_GROUP, kit.MDB_STATUS, kit.MDB_NAME})
				return
			}
			// 分组详情
			m.Option(ice.FIELD_DETAIL, "添加", "退还")
			m.Push("detail", value)
		})
	})
	if len(arg) < 1 {
		m.Sort(cmd)
	} else if len(arg) < 2 {
		m.Sort(kit.MDB_NAME)
	}
}
func _label_create(m *ice.Message, cmd string, key string, arg ...string) {
	if pod := m.Cmdx(GROUP, arg[2], "get", arg[3:]); pod != "" {
		if m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), pod, func(key string, value map[string]interface{}) {
			if value[kit.MDB_STATUS] == "void" {
				value[kit.MDB_STATUS] = "free"
				m.Logs(ice.LOG_MODIFY, cmd, arg[0], kit.MDB_NAME, pod, kit.MDB_STATUS, value[kit.MDB_STATUS])
			}
		}) == nil {
			m.Logs(ice.LOG_INSERT, cmd, arg[0], kit.MDB_NAME, pod)
			m.Rich(cmd, kit.Keys(kit.MDB_HASH, key), kit.Dict(
				kit.MDB_NAME, pod, kit.MDB_GROUP, arg[2], kit.MDB_STATUS, "free",
			))
		}
		m.Echo(arg[0])
	}
}
func _label_remove(m *ice.Message, cmd string, key string, arg ...string) {
	m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), arg[3], func(sub string, value map[string]interface{}) {
		if value[kit.MDB_STATUS] == "free" {
			value[kit.MDB_STATUS] = "void"
			m.Logs(ice.LOG_MODIFY, cmd, arg[0], kit.MDB_NAME, arg[3], kit.MDB_STATUS, "void")
			m.Cmdx(GROUP, value[kit.MDB_GROUP], "put", arg[3])
			m.Echo(arg[3])
		}
	})
}
func _label_remote(m *ice.Message, cmd string, key string, arg ...string) {
	wg := &sync.WaitGroup{}
	m.Option("_async", "true")
	m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), arg[1], func(key string, value map[string]interface{}) {
		wg.Add(1)
		m.Option(ice.MSG_USERPOD, value[kit.MDB_NAME])
		m.Cmd(SPACE, value[kit.MDB_NAME], arg[2:]).Call(false, func(res *ice.Message) *ice.Message {
			if wg.Done(); res != nil && m != nil {
				m.Copy(res)
			}
			return nil
		})
	})
	wg.Wait()
}

const LABEL = "label"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			LABEL: {Name: "label", Help: "标签", Value: kit.Data(kit.MDB_SHORT, "label")},
		},
		Commands: map[string]*ice.Command{
			LABEL: {Name: "label label=auto name=auto auto", Help: "标签", Meta: kit.Dict(
				"exports", []string{"lab", "label"}, "detail", []string{"归还"},
			), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 1 && arg[0] == "action" {
					switch arg[1] {
					case "add", "添加":
						_label_add(m, cmd)
					case "del", "退还":
						_label_del(m, cmd)
					case "prune", "清理":
						_label_prune(m, cmd)
					case "clear", "清空":
						_label_clear(m, cmd)
					case "delete", "删除":
						_label_delete(m, cmd)
					}
					return
				}

				if len(arg) < 3 {
					// 查询分组
					_label_select(m, cmd, arg...)
					return
				}

				if m.Richs(cmd, nil, arg[0], nil) == nil {
					// 添加分组
					m.Logs(ice.LOG_CREATE, cmd, m.Rich(cmd, nil, kit.Data(
						kit.MDB_SHORT, kit.MDB_NAME, cmd, arg[0],
					)))
				}

				m.Richs(cmd, nil, arg[0], func(key string, value map[string]interface{}) {
					switch arg[1] {
					case "add": // 添加设备
						_label_create(m, cmd, key, arg...)
					case "del": // 删除设备
						_label_remove(m, cmd, key, arg...)
					default: // 远程命令
						if arg[0] == "route" {
							m.Cmd(ROUTE).Table(func(index int, value map[string]string, field []string) {
								m.Rich(cmd, kit.Keys(kit.MDB_HASH, key), kit.Dict(
									kit.MDB_NAME, value["name"], kit.MDB_GROUP, arg[0], kit.MDB_STATUS, "free",
								))
							})
						}
						_label_remote(m, cmd, key, arg...)
					}
				})
			}},
		}}, nil)
}
