package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func _share_list(m *ice.Message, key string, fields ...string) {
	if key == "" {
		m.Grows(SHARE, nil, "", "", func(index int, value map[string]interface{}) {
			m.Push("", value, []string{kit.MDB_TIME, kit.MDB_SHARE, kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT})
			m.Push(kit.MDB_LINK, fmt.Sprintf(m.Conf(SHARE, "meta.template.link"), m.Conf(SHARE, "meta.domain"), value[kit.MDB_SHARE], value[kit.MDB_SHARE]))
		})
		return
	}

	m.Richs(SHARE, nil, key, func(key string, value map[string]interface{}) {
		m.Push("detail", value)

		m.Push(kit.MDB_KEY, kit.MDB_LINK)
		m.Push(kit.MDB_VALUE, m.Cmdx(mdb.RENDER, RENDER.A, key, URL(m, kit.Format("/share/%s", key))))
		m.Push(kit.MDB_KEY, kit.MDB_SHARE)
		m.Push(kit.MDB_VALUE, m.Cmdx(mdb.RENDER, RENDER.IMG, URL(m, kit.Format("/share/%s/share", key))))
		m.Push(kit.MDB_KEY, kit.MDB_VALUE)
		m.Push(kit.MDB_VALUE, m.Cmdx(mdb.RENDER, RENDER.IMG, URL(m, kit.Format("/share/%s/value", key))))
	})
}
func _share_show(m *ice.Message, key string, value map[string]interface{}, arg ...string) bool {
	switch kit.Select("", arg, 0) {
	case "check", "安全码":
		m.Render(ice.RENDER_QRCODE, kit.Format(kit.Dict(
			kit.MDB_TYPE, SHARE, kit.MDB_NAME, value[kit.MDB_TYPE], kit.MDB_TEXT, key,
		)))
	case kit.MDB_SHARE, "共享码":
		m.Render(ice.RENDER_QRCODE, kit.Format("%s/share/%s/?share=%s", m.Conf(SHARE, "meta.domain"), key, key))
	case kit.MDB_VALUE, "数据值":
		m.Render(ice.RENDER_QRCODE, kit.Format(value), kit.Select("256", arg, 1))
	case kit.MDB_TEXT:
		m.Render(ice.RENDER_QRCODE, kit.Format(value[kit.MDB_TEXT]))
	case "detail", "详情":
		m.Render(kit.Formats(value))
	case "download", "下载":
		if strings.HasPrefix(kit.Format(value["text"]), m.Conf(CACHE, "meta.path")) {
			m.Render(ice.RENDER_DOWNLOAD, value["text"], value["type"], value["name"])
		} else {
			m.Render("%s", value["text"])
		}
	default:
		return false
	}
	return true
}
func _share_repos(m *ice.Message, repos string, arg ...string) {
	prefix := m.Conf(SERVE, "meta.volcanos.require")
	if _, e := os.Stat(path.Join(prefix, repos)); e != nil {
		m.Cmd(cli.SYSTEM, "git", "clone", "https://"+repos, path.Join(prefix, repos))
	}
	m.Render(ice.RENDER_DOWNLOAD, path.Join(prefix, repos, path.Join(arg...)))
}
func _share_local(m *ice.Message, arg ...string) {
	p := path.Join(arg...)
	switch ls := strings.Split(p, "/"); ls[0] {
	case "etc", "var":
		// 私有文件
		m.Render(STATUS, http.StatusUnauthorized, "not auth")
		return
	default:
		if m.Warn(!m.Right(ls), ice.ErrNotAuth, m.Option(ice.MSG_USERROLE), " of ", p) {
			m.Render(STATUS, http.StatusUnauthorized, "not auth")
			return
		}
	}

	if m.Option("pod") != "" {
		// 远程文件
		pp := path.Join("var/proxy", m.Option("pod"), p)
		cache := time.Now().Add(-time.Hour * 240000)
		if s, e := os.Stat(pp); e == nil {
			cache = s.ModTime()
		}
		m.Cmdy(SPACE, m.Option("pod"), SPIDE, "dev", "raw", kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/share/proxy/"),
			"part", "pod", m.Option("pod"), "path", p, "cache", cache.Format(ice.MOD_TIME), "upload", "@"+p)

		m.Render(ice.RENDER_DOWNLOAD, path.Join("var/proxy", m.Option("pod"), p))
		return
	}

	// 本地文件
	m.Render(ice.RENDER_DOWNLOAD, p)
}
func _share_proxy(m *ice.Message, arg ...string) {
	switch m.Option(ice.MSG_METHOD) {
	case http.MethodGet:
		m.Render(ice.RENDER_DOWNLOAD, path.Join("var/proxy", path.Join(m.Option("pod"), m.Option("path"), m.Option("name"))))
	case http.MethodPost:
		m.Cmdy(CACHE, UPLOAD)
		m.Cmdy(CACHE, WATCH, m.Option("data"), path.Join("var/proxy", m.Option("pod"), m.Option("path")))
		m.Render(ice.RENDER_RESULT, m.Option("path"))
	}

}
func _share_remote(m *ice.Message, pod string, arg ...string) {
	m.Cmdy(SPACE, pod, "web./publish/", arg)
	m.Render(ice.RENDER_RESULT)
}
func _share_create(m *ice.Message, kind, name, text string, arg ...string) string {
	h := m.Rich(SHARE, nil, kit.Dict(
		kit.MDB_TIME, m.Time(m.Conf(SHARE, "meta.expire")),
		kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, text,
		kit.MDB_EXTRA, kit.Dict(
			aaa.USERROLE, m.Option(ice.MSG_USERROLE),
			aaa.USERNAME, m.Option(ice.MSG_USERNAME),
			"river", m.Option(ice.MSG_RIVER),
			"storm", m.Option(ice.MSG_STORM),
			arg),
	))

	// 创建列表
	m.Grow(SHARE, nil, kit.Dict(
		kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, text,
		kit.MDB_SHARE, h,
	))
	m.Log_CREATE(kit.MDB_SHARE, h, kit.MDB_TYPE, kind, kit.MDB_NAME, name)
	m.Echo(h)
	return h
}

func _share_story(m *ice.Message, value map[string]interface{}, arg ...string) map[string]interface{} {
	msg := m.Cmd(STORY, INDEX, value["text"])
	if msg.Append("text") == "" && kit.Value(value, "extra.pod") != "" {
		msg = m.Cmd(SPACE, kit.Value(value, "extra.pod"), STORY, INDEX, value["text"])
	}
	value = kit.Dict("type", msg.Append("scene"), "name", msg.Append("story"), "text", msg.Append("text"), "file", msg.Append("file"))
	m.Log(ice.LOG_EXPORT, "%s: %v", arg, kit.Format(value))
	return value
}
func _share_action(m *ice.Message, value map[string]interface{}, arg ...string) bool {
	if len(arg) == 1 || arg[1] == "" {
		return _share_action_redirect(m, value, arg[0])
	}
	if arg[1] == "" {
		return _share_action_page(m, value)
	}
	if len(arg) == 2 {
		return _share_action_list(m, value, arg[0], arg[1])
	}

	// 默认参数
	meta := kit.Value(value, kit.Format("extra.tool.%s", arg[2])).(map[string]interface{})
	if meta["single"] == "yes" && kit.Select("", arg, 3) != "action" {
		arg = append(arg[:3], kit.Simple(kit.UnMarshal(kit.Format(meta["args"])))...)
		for i := len(arg) - 1; i >= 0; i-- {
			if arg[i] != "" {
				return true
			}
			arg = arg[:i]
		}
	}

	// 执行命令
	cmds := kit.Simple(m.Space(meta["pod"]), kit.Keys(meta["ctx"], meta["cmd"]), arg[3:])
	m.Cmdy(cmds).Option("cmds", cmds)
	m.Option("title", value["name"])
	if strings.HasPrefix(kit.Format(value["text"]), m.Conf(CACHE, "meta.path")) {
		m.Render(ice.RENDER_DOWNLOAD, value["text"], value["type"], value["name"])
	} else {
		m.Render("%s", value["text"])
	}
	return true
}
func _share_action_redirect(m *ice.Message, value map[string]interface{}, share string) bool {
	tool := kit.Value(value, "extra.tool.0").(map[string]interface{})
	m.Render("redirect", "/share", "share", share, "title", kit.Format(value["name"]),
		"river", kit.Format(kit.Value(value, "extra.river")),
		"storm", kit.Format(kit.Value(value, "extra.storm")),
		"pod", kit.Format(tool["pod"]), kit.UnMarshal(kit.Format(tool["val"])),
	)
	return true
}
func _share_action_page(m *ice.Message, value map[string]interface{}) bool {
	Render(m, ice.RENDER_DOWNLOAD, m.Conf(SERVE, "meta.page.share"))
	return true
}
func _share_action_list(m *ice.Message, value map[string]interface{}, river, storm string) bool {
	value["count"] = kit.Int(value["count"]) + 1
	kit.Fetch(kit.Value(value, "extra.tool"), func(index int, value map[string]interface{}) {
		m.Push("river", river)
		m.Push("storm", storm)
		m.Push("action", index)

		m.Push("node", value["pod"])
		m.Push("group", value["ctx"])
		m.Push("index", value["cmd"])
		m.Push("args", value["args"])
		m.Push("value", value["value"])

		msg := m.Cmd(m.Space(value["pod"]), ctx.COMMAND, value["ctx"], value["cmd"])
		ls := strings.Split(kit.Format(value["cmd"]), ".")
		m.Push("name", ls[len(ls)-1])
		m.Push("help", kit.Select(msg.Append("help"), kit.Format(value["help"])))
		m.Push("inputs", msg.Append("list"))
		m.Push("feature", msg.Append("meta"))
	})
	return true
}

func _share_auth(m *ice.Message, share string, role string) {
	m.Richs(SHARE, nil, share, func(key string, value map[string]interface{}) {
		switch value["type"] {
		case "active":
			m.Cmdy(SPACE, value["name"], "sessid", m.Cmdx(aaa.SESS, "create", role))
		case "user":
			m.Cmdy(aaa.ROLE, role, value["name"])
		default:
			m.Cmdy(aaa.SESS, "auth", value["text"], role)
		}
	})
}
func _share_check(m *ice.Message, share string) {
	m.Richs(SHARE, nil, share, func(key string, value map[string]interface{}) {
		m.Render(ice.RENDER_QRCODE, kit.Format(kit.Dict(
			kit.MDB_TYPE, "share", kit.MDB_NAME, value["type"], kit.MDB_TEXT, key,
		)))
	})
}
func _trash(m *ice.Message, arg ...string) {
	switch arg[0] {
	case "invite":
		arg = []string{arg[0], m.Cmdx(SHARE, "invite", kit.Select("tech", arg, 1), kit.Select("miss", arg, 2))}
		fallthrough
	case "check":
		_share_check(m, arg[1])
	case "auth":
		_share_auth(m, arg[1], arg[2])
	case "add":
		_share_create(m, arg[1], arg[2], arg[3], arg[4:]...)
	default:
		if len(arg) == 1 {
			_share_list(m, arg[0])
			break
		}
	}
}

func ShareCreate(m *ice.Message, kind, name, text string, arg ...string) string {
	return _share_create(m, kind, name, text, arg...)
}

const SHARE = "share"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SHARE: {Name: "share", Help: "共享链", Value: kit.Data(
				"template", share_template, "expire", "72h",
				"limit", 10,
			)},
		},
		Commands: map[string]*ice.Command{
			SHARE: {Name: "share share=auto auto", Help: "共享链", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create type name text arg...", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_share_create(m, arg[0], arg[1], arg[2], arg[3:]...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) < 2 {
					_share_list(m, kit.Select("", arg, 0))
					return
				}
				_share_create(m, arg[0], arg[1], arg[2], arg[3:]...)
			}},
			"/share/local/": {Name: "/share/local/", Help: "共享链", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_local(m, arg...)
			}},
			"/share/proxy/": {Name: "/share/proxy/", Help: "缓存池", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_proxy(m, arg...)
			}},
			"/share/": {Name: "/share/", Help: "共享链", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(SHARE, nil, kit.Select(m.Option(kit.MDB_SHARE), arg, 0), func(key string, value map[string]interface{}) {
					m.Log_SELECT(kit.MDB_META, SHARE, "arg", arg, "value", kit.Format(value))
					if m.Warn(m.Option(ice.MSG_USERROLE) != aaa.ROOT && kit.Time(kit.Format(value[kit.MDB_TIME])) < kit.Time(m.Time()), "expired") {
						m.Echo("expired")
						return
					}

					switch value[kit.MDB_TYPE] {
					case STORY:
						value = _share_story(m, value, arg...)
					}

					if _share_show(m, key, value, kit.Select("", arg, 1), kit.Select("", arg, 2)) {
						return
					}

					switch value[kit.MDB_TYPE] {
					case TYPE_RIVER:
						// 共享群组
						m.Render("redirect", "/", "share", key, "river", kit.Format(value["text"]))

					case TYPE_STORM:
						// 共享应用
						m.Render("redirect", "/", "share", key, "storm", kit.Format(value["text"]), "river", kit.Format(kit.Value(value, "extra.river")))

					case TYPE_ACTION:
						_share_action(m, value, arg...)

					default:
						// 查看数据
						m.Option(kit.MDB_VALUE, value)
						m.Option(kit.MDB_TYPE, value[kit.MDB_TYPE])
						m.Option(kit.MDB_NAME, value[kit.MDB_NAME])
						m.Option(kit.MDB_TEXT, value[kit.MDB_TEXT])
						m.Render(ice.RENDER_TEMPLATE, m.Conf(SHARE, "meta.template.simple"))
						m.Option(ice.MSG_OUTPUT, ice.RENDER_RESULT)
					}
				})
			}},
			"/plugin/github.com/": {Name: "/space/", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_repos(m, path.Join(strings.Split(cmd, "/")[2:5]...), arg[6:]...)
			}},
			"/publish/": {Name: "/publish/", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if arg[0] == "order.js" && len(ice.BinPack) > 0 {
					m.Render(ice.RENDER_RESULT, "{}")
					return
				}
				if p := m.Option("pod"); p != "" {
					m.Option("pod", "")
					_share_remote(m, p, arg...)
					return
				}

				p := path.Join(kit.Simple(m.Conf(SERVE, "meta.publish"), arg)...)
				if m.W == nil {
					m.Cmdy("nfs.cat", p)
					return
				}
				_share_local(m, p)
			}},
		}}, nil)
}
