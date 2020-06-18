package chat

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"strconv"
)

func _action_share_create(m *ice.Message, arg ...string) {
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
func _action_share_list(m *ice.Message, river, storm string) {
	m.Richs(ice.WEB_SHARE, nil, m.Option("share"), func(key string, value map[string]interface{}) {
		kit.Fetch(kit.Value(value, "extra.tool"), func(index int, value map[string]interface{}) {
			m.Push("river", river)
			m.Push("storm", storm)
			m.Push("action", index)

			m.Push("node", value["pod"])
			m.Push("group", value["ctx"])
			m.Push("index", value["cmd"])
			m.Push("args", value["args"])

			msg := m.Cmd(m.Space(value["pod"]), ice.CTX_COMMAND, kit.Keys(value["ctx"], value["cmd"]))
			m.Push("name", value["cmd"])
			m.Push("help", kit.Select(msg.Append("help"), kit.Format(value["help"])))
			m.Push("inputs", msg.Append("list"))
			m.Push("feature", msg.Append("meta"))
		})
	})
}
func _action_share_show(m *ice.Message, river, storm, index string, arg ...string) {
	if i, e := strconv.Atoi(index); e == nil {
		m.Richs(ice.WEB_SHARE, nil, m.Option("share"), func(key string, value map[string]interface{}) {
			kit.Fetch(kit.Value(value, kit.Keys("extra.tool", i-1)), func(value map[string]interface{}) {
				cmds := kit.Simple(kit.Keys(value["ctx"], value["cmd"]), arg)
				m.Cmdy(_action_proxy(m), cmds).Option("cmds", cmds)
			})
		})
	}
}
func _action_order_list(m *ice.Message, river, storm string, arg ...string) {
	for _, v := range arg {
		m.Push("river", river)
		m.Push("storm", storm)
		m.Push("action", v)

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

func _action_action(m *ice.Message, action string, arg ...string) bool {
	switch action {
	case "upload":
		msg := m.Cmd(ice.WEB_STORY, "upload")
		m.Option("name", msg.Append("name"))
		m.Option("data", msg.Append("data"))
	}
	return false
}
func _action_proxy(m *ice.Message) (proxy []string) {
	if m.Option("pod") != "" {
		proxy = append(proxy, ice.WEB_PROXY, m.Option("pod"))
		m.Option("pod", "")
	}
	return proxy
}
func _action_list(m *ice.Message, river, storm string) {
	prefix := kit.Keys(kit.MDB_HASH, river, "tool", kit.MDB_HASH, storm)
	m.Grows(ice.CHAT_RIVER, prefix, "", "", func(index int, value map[string]interface{}) {
		if meta, ok := kit.Value(value, "meta").(map[string]interface{}); ok {
			m.Push("river", river)
			m.Push("storm", storm)
			m.Push("action", index)

			m.Push("node", meta["pod"])
			m.Push("group", meta["ctx"])
			m.Push("index", meta["cmd"])
			m.Push("args", kit.Select("[]", kit.Format(meta["args"])))

			msg := m.Cmd(m.Space(meta["pod"]), ice.CTX_COMMAND, kit.Keys(meta["ctx"], meta["cmd"]))
			m.Push("name", meta["cmd"])
			m.Push("help", kit.Select(msg.Append("help"), kit.Format(meta["help"])))
			m.Push("feature", msg.Append("meta"))
			m.Push("inputs", msg.Append("list"))
		}
	})
}
func _action_show(m *ice.Message, river, storm, index string, arg ...string) {
	prefix := kit.Keys(kit.MDB_HASH, river, "tool", kit.MDB_HASH, storm)
	cmds := []string{}

	if i, e := strconv.Atoi(index); e == nil {
		m.Grows(ice.CHAT_RIVER, prefix, kit.MDB_ID, kit.Format(i+1), func(index int, value map[string]interface{}) {
			if meta, ok := kit.Value(value, "meta").(map[string]interface{}); ok {
				cmds = kit.Simple(m.Space(meta["pod"]), kit.Keys(meta["ctx"], meta["cmd"]), arg[3:])
			}
		})
	} else if !m.Warn(!m.Right(index), "no right of %v", index) {
		cmds = kit.Simple(index, arg)
	}
	if !m.Right(cmds) {
		m.Render("status", 403, "not auth")
		return
	}
	m.Cmdy(_action_proxy(m), cmds).Option("cmds", cmds)
}

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/action": {Name: "/action", Help: "工作台", Action: map[string]*ice.Action{
			kit.MDB_SHARE: {Name: "share arg...", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
				_action_share_create(m, arg...)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 2 {
				if m.Option("share") != "" {
					// 共享列表
					_action_share_list(m, arg[0], arg[1])
				}
				if p := m.Option("pod"); p != "" {
					m.Option("pod", "")
					// 代理列表
					m.Cmdy(ice.WEB_SPACE, p, "web.chat./action", arg)
				}
				// 命令列表
				_action_list(m, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM))
				return
			}
			switch arg[2] {
			case "index":
				// 前端列表
				_action_order_list(m, arg[0], arg[1], arg[3:]...)
				return
			}

			if arg[0] == "" && m.Option("share") != "" {
				// 共享命令
				_action_share_show(m, arg[0], arg[1], arg[2], arg[3:]...)
				return
			}
			if len(arg) > 3 && arg[3] == "action" && _action_action(m, arg[3]) {
				// 前置命令
				return
			}
			// 执行命令
			_action_show(m, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM), kit.Select(arg[2], m.Option("index")), arg[3:]...)
		}},
	}}, nil)
}
