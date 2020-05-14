package web

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"fmt"
	"path"
	"strings"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ice.WEB_SHARE: {Name: "share", Help: "共享链", Value: kit.Data(
				"index", "usr/volcanos/share.html",
				"template", share_template,
			)},
		},
		Commands: map[string]*ice.Command{
			ice.WEB_SHARE: {Name: "share share auto", Help: "共享链", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					// 共享列表
					m.Grows(ice.WEB_SHARE, nil, "", "", func(index int, value map[string]interface{}) {
						m.Push("", value, []string{kit.MDB_TIME, "share", kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT})
						m.Push("value", fmt.Sprintf(m.Conf(ice.WEB_SHARE, "meta.template.link"), m.Conf(ice.WEB_SHARE, "meta.domain"), value["share"], value["share"]))
					})
					return
				}
				if len(arg) == 1 {
					// 共享详情
					if m.Richs(ice.WEB_SHARE, nil, arg[0], func(key string, value map[string]interface{}) {
						m.Push("detail", value)
						m.Push("key", "link")
						m.Push("value", fmt.Sprintf(m.Conf(ice.WEB_SHARE, "meta.template.link"), m.Conf(ice.WEB_SHARE, "meta.domain"), key, key))
						m.Push("key", "share")
						m.Push("value", fmt.Sprintf(m.Conf(ice.WEB_SHARE, "meta.template.share"), m.Conf(ice.WEB_SHARE, "meta.domain"), key))
						m.Push("key", "value")
						m.Push("value", fmt.Sprintf(m.Conf(ice.WEB_SHARE, "meta.template.value"), m.Conf(ice.WEB_SHARE, "meta.domain"), key))
					}) != nil {
						return
					}
				}

				switch arg[0] {
				case "invite":
					arg = []string{arg[0], m.Cmdx(ice.WEB_SHARE, "add", "invite", kit.Select("tech", arg, 1), kit.Select("miss", arg, 2))}

					fallthrough
				case "check":
					m.Richs(ice.WEB_SHARE, nil, arg[1], func(key string, value map[string]interface{}) {
						m.Render(ice.RENDER_QRCODE, kit.Format(kit.Dict(
							kit.MDB_TYPE, "share", kit.MDB_NAME, value["type"], kit.MDB_TEXT, key,
						)))
					})

				case "auth":
					m.Richs(ice.WEB_SHARE, nil, arg[1], func(key string, value map[string]interface{}) {
						switch value["type"] {
						case "active":
							m.Cmdy(ice.WEB_SPACE, value["name"], "sessid", m.Cmdx(ice.AAA_SESS, "create", arg[2]))
						case "user":
							m.Cmdy(ice.AAA_ROLE, arg[2], value["name"])
						default:
							m.Cmdy(ice.AAA_SESS, "auth", value["text"], arg[2])
						}
					})

				case "add":
					arg = arg[1:]
					fallthrough
				default:
					if len(arg) == 2 {
						arg = append(arg, "")
					}
					extra := kit.Dict(arg[3:])

					// 创建共享
					h := m.Rich(ice.WEB_SHARE, nil, kit.Dict(
						kit.MDB_TIME, m.Time("10m"),
						kit.MDB_TYPE, arg[0], kit.MDB_NAME, arg[1], kit.MDB_TEXT, arg[2],
						"extra", extra,
					))
					// 创建列表
					m.Grow(ice.WEB_SHARE, nil, kit.Dict(
						kit.MDB_TYPE, arg[0], kit.MDB_NAME, arg[1], kit.MDB_TEXT, arg[2],
						"share", h,
					))
					m.Log(ice.LOG_CREATE, "share: %s extra: %s", h, kit.Format(extra))
					m.Echo(h)
				}
			}},
			"/share/": {Name: "/share/", Help: "共享链", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch arg[0] {
				case "local":
					m.Render(ice.RENDER_DOWNLOAD, m.Cmdx(arg[1], path.Join(arg[2:]...)))
					return
				}

				m.Richs(ice.WEB_SHARE, nil, arg[0], func(key string, value map[string]interface{}) {
					m.Log(ice.LOG_EXPORT, "%s: %v", arg, kit.Format(value))
					if m.Option(ice.MSG_USERROLE) != ice.ROLE_ROOT && kit.Time(kit.Format(value[kit.MDB_TIME])) < kit.Time(m.Time()) {
						m.Echo("invalid")
						return
					}

					switch value["type"] {
					case ice.TYPE_SPACE:
					case ice.TYPE_STORY:
						// 查询数据
						msg := m.Cmd(ice.WEB_STORY, ice.STORY_INDEX, value["text"])
						if msg.Append("text") == "" && kit.Value(value, "extra.pod") != "" {
							msg = m.Cmd(ice.WEB_SPACE, kit.Value(value, "extra.pod"), ice.WEB_STORY, ice.STORY_INDEX, value["text"])
						}
						value = kit.Dict("type", msg.Append("scene"), "name", msg.Append("story"), "text", msg.Append("text"), "file", msg.Append("file"))
						m.Log(ice.LOG_EXPORT, "%s: %v", arg, kit.Format(value))
					}

					switch kit.Select("", arg, 1) {
					case "download", "下载":
						if strings.HasPrefix(kit.Format(value["text"]), m.Conf(ice.WEB_CACHE, "meta.path")) {
							m.Render(ice.RENDER_DOWNLOAD, value["text"], value["type"], value["name"])
						} else {
							m.Render("%s", value["text"])
						}
						return
					case "detail", "详情":
						m.Render(kit.Formats(value))
						return
					case "share", "共享码":
						m.Render(ice.RENDER_QRCODE, kit.Format("%s/share/%s/", m.Conf(ice.WEB_SHARE, "meta.domain"), key))
						return
					case "check", "安全码":
						m.Render(ice.RENDER_QRCODE, kit.Format(kit.Dict(
							kit.MDB_TYPE, "share", kit.MDB_NAME, value["type"], kit.MDB_TEXT, key,
						)))
						return
					case "value", "数据值":
						m.Render(ice.RENDER_QRCODE, kit.Format(value), kit.Select("256", arg, 2))
						return
					case "text":
						m.Render(ice.RENDER_QRCODE, kit.Format(value["text"]))
						return
					}

					switch value["type"] {
					case ice.TYPE_RIVER:
						// 共享群组
						m.Render("redirect", "/", "share", key, "river", kit.Format(value["text"]))

					case ice.TYPE_STORM:
						// 共享应用
						m.Render("redirect", "/", "share", key, "storm", kit.Format(value["text"]), "river", kit.Format(kit.Value(value, "extra.river")))

					case ice.TYPE_ACTION:
						if len(arg) == 1 {
							// 跳转主页
							m.Render("redirect", "/share/"+arg[0]+"/", "title", kit.Format(value["name"]))
							break
						}

						if arg[1] == "" {
							// 返回主页
							Render(m, ice.RENDER_DOWNLOAD, m.Conf(ice.WEB_SERVE, "meta.page.share"))
							break
						}

						if len(arg) == 2 {
							// 应用列表
							value["count"] = kit.Int(value["count"]) + 1
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
							break
						}

						// 默认参数
						meta := kit.Value(value, kit.Format("extra.tool.%s", arg[2])).(map[string]interface{})
						if meta["single"] == "yes" && kit.Select("", arg, 3) != "action" {
							arg = append(arg[:3], kit.Simple(kit.UnMarshal(kit.Format(meta["args"])))...)
							for i := len(arg) - 1; i >= 0; i-- {
								if arg[i] != "" {
									break
								}
								arg = arg[:i]
							}
						}

						// 执行命令
						cmds := kit.Simple(m.Space(meta["pod"]), kit.Keys(meta["ctx"], meta["cmd"]), arg[3:])
						m.Cmdy(cmds).Option("cmds", cmds)
						m.Option("title", value["name"])

					default:
						// 查看数据
						m.Option("type", value["type"])
						m.Option("name", value["name"])
						m.Option("text", value["text"])
						m.Render(ice.RENDER_TEMPLATE, m.Conf(ice.WEB_SHARE, "meta.template.simple"))
						m.Option(ice.MSG_OUTPUT, ice.RENDER_RESULT)
					}
				})
			}},
		}}, nil)
}
