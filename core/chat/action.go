package chat

import (
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"strconv"
)

func _action_share_create(m *ice.Message, name, text string, arg ...string) {
	list := []string{}
	for i := 0; i < len(arg)-3; i += 5 {
		p := kit.Format("tool.%d.", i/5)
		list = append(list, p+POD, arg[i])
		list = append(list, p+CTX, arg[i+1])
		list = append(list, p+CMD, arg[i+2])
		list = append(list, p+ARG, arg[i+3])
		list = append(list, p+VAL, arg[i+4])
	}
	m.Cmdy(web.SHARE, ACTION, name, text, list)
}
func _action_share_list(m *ice.Message, key string) {
	m.Richs(web.SHARE, nil, key, func(key string, value map[string]interface{}) {
		m.Option(ice.MSG_USERNAME, kit.Value(value, "extra.username"))
		m.Option(ice.MSG_USERROLE, kit.Value(value, "extra.userrole"))
		m.Log_AUTH(aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERROLE, m.Option(ice.MSG_USERROLE))
		kit.Fetch(kit.Value(value, "extra.tool"), func(index int, value map[string]interface{}) {
			m.Push(RIVER, "")
			m.Push(STORM, "")
			m.Push(ACTION, index)

			m.Push("node", value[POD])
			m.Push("group", value[CTX])
			m.Push("index", value[CMD])
			m.Push("args", value[ARG])

			msg := m.Cmd(m.Space(value[POD]), ctx.COMMAND, kit.Keys(value[CTX], value[CMD]))
			ls := strings.Split(kit.Format(value[CMD]), ".")
			m.Push("name", ls[len(ls)-1])
			m.Push("help", kit.Select(msg.Append("help"), kit.Format(value["help"])))
			m.Push("inputs", msg.Append("list"))
			m.Push("feature", msg.Append("meta"))
		})
	})
}
func _action_order_list(m *ice.Message, river, storm string, arg ...string) {
	for _, v := range arg {
		m.Push(RIVER, river)
		m.Push(STORM, storm)
		m.Push(ACTION, v)

		m.Push("node", "")
		m.Push("group", "")
		m.Push("index", v)
		m.Push("args", "[]")

		msg := m.Cmd(m.Space(m.Option(POD)), ctx.COMMAND, v)
		m.Push("name", msg.Append("name"))
		m.Push("help", msg.Append("help"))
		m.Push("feature", msg.Append("meta"))
		m.Push("inputs", msg.Append("list"))
	}
}
func _action_list(m *ice.Message, river, storm string) {
	if p := m.Option(POD); p != "" {
		m.Option(POD, "")
		// 代理列表
		m.Cmdy(web.SPACE, p, "web.chat./action")
	}
	if m.Option("share") != "" {
		// 共享列表
		_action_share_list(m, m.Option("share"))
	}

	prefix := kit.Keys(kit.MDB_HASH, river, TOOL, kit.MDB_HASH, storm)
	m.Grows(RIVER, prefix, "", "", func(index int, value map[string]interface{}) {
		if meta, ok := kit.Value(value, kit.MDB_META).(map[string]interface{}); ok {
			m.Push(RIVER, river)
			m.Push(STORM, storm)
			m.Push(ACTION, index)

			m.Push("node", meta[POD])
			m.Push("group", meta[CTX])
			m.Push("index", meta[CMD])
			m.Push("args", kit.Select("[]", kit.Format(meta["args"])))

			msg := m.Cmd(m.Space(meta[POD]), ctx.COMMAND, kit.Keys(meta[CTX], meta[CMD]))
			ls := strings.Split(kit.Format(meta["cmd"]), ".")
			m.Push("name", ls[len(ls)-1])
			m.Push("help", kit.Select(msg.Append("help"), kit.Format(meta["help"])))
			m.Push("feature", msg.Append("meta"))
			m.Push("inputs", msg.Append("list"))
		}
	})
}
func _action_show(m *ice.Message, river, storm, index string, arg ...string) {
	prefix := kit.Keys(kit.MDB_HASH, river, TOOL, kit.MDB_HASH, storm)
	cmds := []string{}

	if i, e := strconv.Atoi(index); e == nil {
		m.Richs(web.SHARE, nil, m.Option("share"), func(key string, value map[string]interface{}) {
			m.Option(ice.MSG_USERNAME, kit.Value(value, "extra.username"))
			m.Option(ice.MSG_USERROLE, kit.Value(value, "extra.userrole"))
			m.Log_AUTH(aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERROLE, m.Option(ice.MSG_USERROLE))
			kit.Fetch(kit.Value(value, kit.Keys("extra.tool", i)), func(value map[string]interface{}) {
				// 共享命令
				if value[POD] != "" {
					m.Option(POD, value[POD])
				}
				cmds = kit.Simple(kit.Keys(value[CTX], value[CMD]), arg)
			})
		})
		m.Grows(RIVER, prefix, kit.MDB_ID, kit.Format(i+1), func(index int, value map[string]interface{}) {
			if value, ok := kit.Value(value, kit.MDB_META).(map[string]interface{}); ok {
				// 群组命令
				if value[POD] != "" {
					m.Option(POD, value[POD])
				}
				cmds = kit.Simple(kit.Keys(value[CTX], value[CMD]), arg)
			}
		})
	} else if m.Right(index) {
		// 定制命令
		cmds = kit.Simple(index, arg)
	} else {
		return
	}
	if len(cmds) == 0 {
		if p := m.Option(POD); p != "" {
			m.Option(POD, "")
			// 代理列表
			m.Cmdy(web.SPACE, p, "web.chat./action", index, arg)
			return
		}
		m.Render("status", 404, "not found")
		return
	}
	if !m.Right(cmds) {
		m.Render("status", 403, "not auth")
		return
	}
	m.Cmdy(_action_proxy(m), cmds)
}
func _action_proxy(m *ice.Message) (proxy []string) {
	if m.Option(POD) != "" {
		proxy = append(proxy, web.SPACE, m.Option(POD))
		m.Option(POD, "")
	}
	return proxy
}

const (
	ORDER = "order"
)
const ACTION = "action"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/" + ACTION: {Name: "/action", Help: "工作台", Action: map[string]*ice.Action{
			web.SHARE: {Name: "share name text [pod ctx cmd arg]...", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
				_action_share_create(m, arg[0], arg[1], arg[2:]...)
			}},
			web.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(web.STORY, web.UPLOAD)
				m.Option(kit.MDB_NAME, msg.Append(kit.MDB_NAME))
				m.Option(web.DATA, msg.Append(web.DATA))
				_action_show(m, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM), m.Option(ice.MSG_ACTION),
					append([]string{ACTION, web.UPLOAD}, arg...)...)
			}},
			ORDER: {Name: "order cmd...", Help: "定制", Hand: func(m *ice.Message, arg ...string) {
				_action_order_list(m, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM), arg...)
			}},

			"command": {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 1 {
					m.Cmdy(ctx.COMMAND, arg[0])
					return
				}
				m.Cmdy(arg[0], arg[1:])
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 命令列表
				_action_list(m, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM))
				return
			}
			// 执行命令
			_action_show(m, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM), arg[0], arg[1:]...)
		}},
	}}, nil)
}
