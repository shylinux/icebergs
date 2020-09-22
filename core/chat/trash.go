package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			"/ocean": {Name: "/ocean", Help: "大海洋", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					// 用户列表
					m.Richs(aaa.USER, nil, "*", func(key string, value map[string]interface{}) {
						m.Push(key, value, []string{"username", "usernode"})
					})
					return
				}

				switch arg[0] {
				case "spawn":
					// 创建群组
					river := m.Rich(RIVER, nil, kit.Dict(
						kit.MDB_META, kit.Dict(kit.MDB_NAME, arg[1]),
						"user", kit.Data(kit.MDB_SHORT, "username"),
						"tool", kit.Data(),
					))
					m.Log(ice.LOG_CREATE, "river: %v name: %v", river, arg[1])
					// 添加用户
					m.Cmd("/river", river, "add", m.Option(ice.MSG_USERNAME), arg[2:])
					m.Echo(river)
				}
			}},
			"/steam": {Name: "/steam", Help: "大气层", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Warn(m.Option(ice.MSG_RIVER) == "", "not join") {
					m.Render("status", 402, "not join")
					return
				}

				if len(arg) < 2 {
					if list := []string{}; m.Option("pod") != "" {
						// 远程空间
						m.Cmdy(web.SPACE, m.Option("pod"), "web.chat./steam").Table(func(index int, value map[string]string, head []string) {
							list = append(list, kit.Keys(m.Option("pod"), value["name"]))
						})
						m.Append("name", list)
					} else {
						// 本地空间
						m.Richs(web.SPACE, nil, "*", func(key string, value map[string]interface{}) {
							switch value[kit.MDB_TYPE] {
							case web.SERVER, web.WORKER:
								m.Push(key, value, []string{"type", "name", "user"})
							}
						})
					}
					return
				}

				if m.Warn(!m.Right(cmd, arg[1])) {
					m.Render("status", 403, "not auth")
					return
				}

				switch arg[1] {
				case "spawn":
					// 创建应用
					storm := m.Rich(RIVER, kit.Keys(kit.MDB_HASH, arg[0], "tool"), kit.Dict(
						kit.MDB_META, kit.Dict(kit.MDB_NAME, arg[2]),
					))
					m.Log(ice.LOG_CREATE, "storm: %s name: %v", storm, arg[2])
					// 添加命令
					m.Cmd("/storm", arg[0], storm, "add", arg[3:])
					m.Echo(storm)

				case "append":
					// 追加命令
					m.Cmd("/storm", arg[0], arg[2], "add", arg[3:])

				default:
					// 命令列表
					m.Cmdy(web.SPACE, arg[2], ctx.COMMAND)
				}
			}},
		},
	}, nil)
}
