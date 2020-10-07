package web

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

const GROUP = "group"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			GROUP: {Name: "group", Help: "分组", Value: kit.Data(kit.MDB_SHORT, "group")},
		},
		Commands: map[string]*ice.Command{
			GROUP: {Name: "group group=auto name=auto auto", Help: "分组", Meta: kit.Dict(
				"exports", []string{"grp", "group"}, "detail", []string{"标签", "添加", "退还"},
			), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 1 && arg[0] == "action" {
					switch arg[1] {
					case "label", "标签":
						if m.Option(ice.EXPORT_LABEL) != "" && m.Option(cmd) != "" {
							m.Cmdy(LABEL, m.Option(ice.EXPORT_LABEL), "add", m.Option(cmd), m.Option(kit.MDB_NAME))
							m.Option(ice.FIELD_RELOAD, "true")
						}
					case "add", "添加":
						if m.Option(cmd) != "" && m.Option(kit.MDB_NAME) != "" {
							m.Cmdy(cmd, m.Option(cmd), "add", m.Option(kit.MDB_NAME))
							m.Option(ice.FIELD_RELOAD, "true")
						}
					case "del", "退还":
						if m.Option(cmd) != "" && m.Option(kit.MDB_NAME) != "" {
							m.Cmdy(cmd, m.Option(cmd), "del", m.Option(kit.MDB_NAME))
							m.Option(ice.FIELD_RELOAD, "true")
						}
					case "prune", "清理":
						m.Richs(cmd, nil, m.Option(cmd), func(key string, value map[string]interface{}) {
							m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), kit.MDB_FOREACH, func(sub string, value map[string]interface{}) {
								if value[kit.MDB_STATUS] != "busy" {
									m.Cmdy(cmd, m.Option(cmd), "del", value[kit.MDB_NAME])
									m.Option(ice.FIELD_RELOAD, "true")
								}
							})
						})
					case "clear", "清空":
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
					case "delete", "删除":
						m.Richs(cmd, nil, m.Option(cmd), func(key string, value map[string]interface{}) {
							m.Echo(m.Conf(cmd, kit.Keys(kit.MDB_HASH, key)))
							m.Logs(ice.LOG_REMOVE, cmd, m.Option(cmd), kit.MDB_VALUE, m.Conf(cmd, kit.Keys(kit.MDB_HASH, key)))
							m.Conf(cmd, kit.Keys(kit.MDB_HASH, key), "")
							m.Option(ice.FIELD_RELOAD, "true")
						})
					}
					return
				}

				if len(arg) < 3 {
					m.Richs(cmd, nil, kit.Select(kit.MDB_FOREACH, arg, 0), func(key string, value map[string]interface{}) {
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
								m.Option(ice.FIELD_DETAIL, "标签", "添加", "退还", "清理", "清空")
								m.Push(key, value, []string{kit.MDB_TIME, kit.MDB_STATUS, kit.MDB_NAME})
								return
							}
							// 分组详情
							m.Option(ice.FIELD_DETAIL, "标签", "添加", "退还")
							m.Push("detail", value)
						})
					})
					if len(arg) < 1 {
						m.Sort(cmd)
					} else if len(arg) < 2 {
						m.Sort(kit.MDB_NAME)
					}
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
						if m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), arg[2], func(key string, value map[string]interface{}) {
							if value[kit.MDB_STATUS] == "void" {
								value[kit.MDB_STATUS] = "free"
								m.Logs(ice.LOG_MODIFY, cmd, arg[0], kit.MDB_NAME, arg[2], kit.MDB_STATUS, value[kit.MDB_STATUS])
							}
						}) == nil {
							m.Logs(ice.LOG_INSERT, cmd, arg[0], kit.MDB_NAME, arg[2])
							m.Rich(cmd, kit.Keys(kit.MDB_HASH, key), kit.Dict(
								kit.MDB_NAME, arg[2], kit.MDB_STATUS, "free",
							))
						}
						m.Echo(arg[0])
					case "del": // 删除设备
						m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), arg[2], func(sub string, value map[string]interface{}) {
							if value[kit.MDB_STATUS] == "free" {
								value[kit.MDB_STATUS] = "void"
								m.Logs(ice.LOG_MODIFY, cmd, arg[0], kit.MDB_NAME, arg[2], kit.MDB_STATUS, value[kit.MDB_STATUS])
								m.Echo(arg[2])
							}
						})
					case "get": // 分配设备
						m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), kit.Select("%", arg, 2), func(sub string, value map[string]interface{}) {
							if value[kit.MDB_STATUS] == "free" {
								value[kit.MDB_STATUS] = "busy"
								m.Logs(ice.LOG_MODIFY, cmd, arg[0], kit.MDB_NAME, arg[2], kit.MDB_STATUS, value[kit.MDB_STATUS])
								m.Echo("%s", value[kit.MDB_NAME])
							}
						})
					case "put": // 回收设备
						m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), arg[2], func(sub string, value map[string]interface{}) {
							if value[kit.MDB_STATUS] == "busy" {
								value[kit.MDB_STATUS] = "free"
								m.Logs(ice.LOG_MODIFY, cmd, arg[0], kit.MDB_NAME, arg[2], kit.MDB_STATUS, value[kit.MDB_STATUS])
								m.Echo("%s", value[kit.MDB_NAME])
							}
						})
					default:
						m.Richs(cmd, kit.Keys(kit.MDB_HASH, key), arg[1], func(key string, value map[string]interface{}) {
							// 执行命令
							m.Cmdy(PROXY, value[kit.MDB_NAME], arg[2:])
						})
					}
				})
			}},
		}}, nil)
}
