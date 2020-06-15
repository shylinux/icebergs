package chat

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
)

func _action_share(m *ice.Message, arg ...string) {
	if m.Option("_index") != "" {
		m.Cmdy(ice.WEB_SHARE, ice.TYPE_ACTION, m.Option("_name"), m.Option("_text"),
			"tool.0.pod", kit.Select(m.Option("_pod"), m.Option("_node")),
			"tool.0.ctx", m.Option("_group"),
			"tool.0.cmd", m.Option("_index"),
			"tool.0.args", m.Option("_args"),
			"tool.0.value", m.Option("_value"),
			"tool.0.single", "yes",
		)
	} else {
		m.Option(ice.MSG_RIVER, arg[5])
		m.Option(ice.MSG_STORM, arg[7])
		m.Cmd("/action", arg[5], arg[7]).Table(func(index int, value map[string]string, head []string) {
			arg = append(arg, kit.Format("tool.%d.pod", index), value["node"])
			arg = append(arg, kit.Format("tool.%d.ctx", index), value["group"])
			arg = append(arg, kit.Format("tool.%d.cmd", index), value["index"])
			arg = append(arg, kit.Format("tool.%d.args", index), value["args"])
		})
	}
}
func _action_share_list(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
	m.Richs(ice.WEB_SHARE, nil, m.Option("share"), func(key string, value map[string]interface{}) {
		kit.Fetch(kit.Value(value, "extra.tool"), func(index int, value map[string]interface{}) {
			m.Push("river", arg[0])
			m.Push("storm", arg[1])
			m.Push("action", index)

			m.Push("node", value["pod"])
			m.Push("group", value["ctx"])
			m.Push("index", value["cmd"])
			m.Push("args", value["args"])

			msg := m.Cmd(m.Space(value["pod"]), ice.CTX_COMMAND, value["ctx"], value["cmd"])
			m.Push("name", value["cmd"])
			m.Push("help", kit.Select(msg.Append("help"), kit.Format(value["help"])))
			m.Push("inputs", msg.Append("list"))
			m.Push("feature", msg.Append("meta"))
		})
	})
}
func _action_share_show(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
	if len(arg) > 3 && arg[3] == "action" && _action_action(m, arg[4], arg[5:]...) {
		return
	}

	m.Richs(ice.WEB_SHARE, nil, m.Option("share"), func(key string, value map[string]interface{}) {
		kit.Fetch(kit.Value(value, kit.Keys("extra.tool", arg[2])), func(value map[string]interface{}) {
			cmds := kit.Simple(kit.Keys(value["ctx"], value["cmd"]), arg[3:])
			m.Cmdy(_action_proxy(m), cmds).Option("cmds", cmds)
		})
	})
}
func _action_order(m *ice.Message, arg ...string) {
	if arg[2] == "index" {
		for i, v := range arg[3:] {
			m.Push("river", arg[0])
			m.Push("storm", arg[1])
			m.Push("action", i)

			m.Push("node", "")
			m.Push("group", "")
			m.Push("index", v)
			m.Push("args", "[]")

			msg := m.Cmd(m.Space(m.Option("pod")), ice.CTX_COMMAND, v)
			m.Push("name", msg.Append("name"))
			m.Push("help", msg.Append("help"))
			m.Push("feature", msg.Append("meta"))
			m.Push("inputs", msg.Append("list"))
		}
	}
	if len(arg) > 3 && arg[3] == "action" && _action_action(m, arg[4], arg[5:]...) {
		return
	}

	if m.Option("index") == "" || !m.Right(kit.Keys(m.Option("group"), m.Option("index"))) {
		m.Render("status", 403, "not auth")
		return
	}

	cmds := kit.Simple(kit.Keys(m.Option("group"), m.Option("index")), arg[3:])
	if m.Set(ice.MSG_RESULT); !m.Right(cmds) {
		m.Render("status", 403, "not auth")
		return
	}
	m.Add("option", "_option", "data", "name")
	m.Cmdy(_action_proxy(m), cmds).Option("cmds", cmds)
}

func _action_proxy(m *ice.Message) (proxy []string) {
	if m.Option("pod") != "" {
		proxy = append(proxy, ice.WEB_PROXY, m.Option("pod"))
		m.Option("pod", "")
	}
	return proxy
}
func _action_action(m *ice.Message, action string, arg ...string) bool {
	switch action {
	case "upload":
		msg := m.Cmd(ice.WEB_STORY, "upload")
		m.Option("name", msg.Append("name"))
		m.Option("data", msg.Append("data"))
	}
	return false
}
func _action_select(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
	prefix := kit.Keys(kit.MDB_HASH, arg[0], "tool", kit.MDB_HASH, arg[1])
	m.Grows(ice.CHAT_RIVER, prefix, "", "", func(index int, value map[string]interface{}) {
		if meta, ok := kit.Value(value, "meta").(map[string]interface{}); ok {
			m.Push("river", arg[0])
			m.Push("storm", arg[1])
			m.Push("action", index)

			m.Push("node", meta["pod"])
			m.Push("group", meta["ctx"])
			m.Push("index", meta["cmd"])
			m.Push("args", kit.Select("[]", kit.Format(meta["args"])))

			msg := m.Cmd(m.Space(meta["pod"]), ice.CTX_COMMAND, meta["ctx"], meta["cmd"])
			m.Push("name", meta["cmd"])
			m.Push("help", kit.Select(msg.Append("help"), kit.Format(meta["help"])))
			m.Push("feature", msg.Append("meta"))
			m.Push("inputs", msg.Append("list"))
		}
	})
}
func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/action": {Name: "/action", Help: "工作台", Action: map[string]*ice.Action{
			kit.MDB_SHARE: {Name: "share arg...", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
				_action_share(m, arg...)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 || arg[0] == "" {
				if m.Option("share") != "" {
					if len(arg) < 3 {
						_action_share_list(m, c, cmd, arg...)
						return
					}
					_action_share_show(m, c, cmd, arg...)
					return
				}
			}

			if m.Warn(m.Option(ice.MSG_RIVER) == "" || m.Option(ice.MSG_STORM) == "", "not join") {
				_action_order(m, arg...)
				return
			}

			prefix := kit.Keys(kit.MDB_HASH, arg[0], "tool", kit.MDB_HASH, arg[1])
			if len(arg) == 2 {
				if p := m.Option("pod"); p != "" {
					m.Option("pod", "")
					if m.Cmdy(ice.WEB_SPACE, p, "web.chat./action", arg); len(m.Appendv("river")) > 0 {
						// 远程查询
						return
					}
				}

				// 命令列表
				_action_select(m, c, cmd, arg...)
				return
			}

			switch arg[2] {
			case "save":
				if p := m.Option("pod"); p != "" {
					// 远程保存
					m.Option("pod", "")
					m.Cmd(ice.WEB_SPACE, p, "web.chat./action", arg)
					return
				}

				// 保存应用
				m.Conf(ice.CHAT_RIVER, kit.Keys(prefix, "list"), "")
				for i := 3; i < len(arg)-4; i += 5 {
					id := m.Grow(ice.CHAT_RIVER, kit.Keys(prefix), kit.Data(
						"pod", arg[i], "ctx", arg[i+1], "cmd", arg[i+2],
						"help", arg[i+3], "args", arg[i+4],
					))
					m.Log(ice.LOG_INSERT, "storm: %s %d: %v", arg[1], id, arg[i:i+5])
				}
			}

			// 查询命令
			cmds := []string{}
			m.Grows(ice.CHAT_RIVER, prefix, kit.MDB_ID, kit.Format(kit.Int(arg[2])+1), func(index int, value map[string]interface{}) {
				if meta, ok := kit.Value(value, "meta").(map[string]interface{}); ok {
					if len(arg) > 3 && arg[3] == "action" {
						// 命令补全
						switch arg[4] {
						case "input":
							switch arg[5] {
							case "location":
								// 查询位置
								m.Copy(m.Cmd("aaa.location"), "append", "name")
								return
							}

						case "favor":
							m.Cmdy(ice.WEB_FAVOR, arg[5:])
							return

						case "device":
							// 记录位置
							m.Cmd(ice.WEB_FAVOR, kit.Select("device", m.Option("hot")), arg[5], arg[6],
								kit.Select("", arg, 7), kit.KeyValue(map[string]interface{}{}, "", kit.UnMarshal(kit.Select("{}", arg, 8))))
							return

						case "upload":
							m.Cmdy(ice.WEB_STORY, "upload")
							return

						case "share":
							list := []string{}
							for k, v := range meta {
								list = append(list, k, kit.Format(v))
							}
							// 共享命令
							m.Cmdy(ice.WEB_SHARE, "action", arg[5], arg[6], list)
							return
						}
					}

					// 组装命令
					cmds = kit.Simple(m.Space(meta["pod"]), kit.Keys(meta["ctx"], meta["cmd"]), arg[3:])
				}
			})

			if len(cmds) == 0 {
				return
			}

			if !m.Right(cmds) {
				m.Render("status", 403, "not auth")
				return
			}

			// 执行命令
			m.Cmdy(_action_proxy(m), cmds).Option("cmds", cmds)
		}},
	}}, nil)
}
