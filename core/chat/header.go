package chat

import (
	"os"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _header_check(m *ice.Message, arg ...string) {
	if m.Option(web.SHARE) != "" {
		switch msg := m.Cmd(web.SHARE, m.Option(web.SHARE)); msg.Append(kit.MDB_TYPE) {
		case web.LOGIN: // 扫码登录
			if m.Option(ice.MSG_USERNAME) != msg.Append(aaa.USERNAME) {
				web.RenderCookie(m, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME, msg.Append(aaa.USERNAME))))
			}
		case web.STORM:
			m.Option(ice.MSG_USERNAME, msg.Append(aaa.USERNAME))
		case web.FIELD:
			m.Option(ice.MSG_USERNAME, msg.Append(aaa.USERNAME))
		}
	}

	if m.Option(ice.MSG_USERNAME) == "" { // 单点登录
		m.Option(web.SSO, m.Conf(web.SERVE, kit.Keym(web.SSO)))
		m.Option(web.LOGIN, m.Config(web.LOGIN))
	}
}
func _header_share(m *ice.Message, arg ...string) {
	if m.Option(kit.MDB_LINK) == "" {
		m.Cmdy(web.SHARE, mdb.CREATE, kit.MDB_TYPE, web.LOGIN, arg)
	} else {
		m.Option(kit.MDB_LINK, tcp.ReplaceLocalhost(m, m.Option(kit.MDB_LINK)))
	}

	m.Option(kit.MDB_LINK, kit.MergeURL(m.Option(kit.MDB_LINK), RIVER, "", STORM, ""))

	m.Set(kit.MDB_NAME, kit.MDB_TEXT)
	m.Push(kit.MDB_NAME, m.Option(kit.MDB_LINK))
	m.PushQRCode(kit.MDB_TEXT, m.Option(kit.MDB_LINK))
}
func _header_grant(m *ice.Message, arg ...string) {
	if m.PodCmd(m.PrefixKey(), ctx.ACTION, GRANT, arg) {
		return // 下发命令
	}

	// 授权登录
	m.Cmd(aaa.ROLE, kit.Select(aaa.TECH, aaa.VOID, m.Option(ice.MSG_USERROLE) == aaa.VOID), m.Option(ice.MSG_USERNAME))
	m.Cmd(web.SPACE, m.Option(web.SPACE), ice.MSG_SESSID, aaa.SessCreate(m, m.Option(ice.MSG_USERNAME)))
}
func _header_users(m *ice.Message, key string, arg ...string) {
	m.Option(aaa.USERNAME, m.Option(ice.MSG_USERNAME))
	m.Cmdy(aaa.USER, ctx.ACTION, mdb.MODIFY, key, m.Option(key, arg[0]))
}

const (
	TOPIC = "topic"
	TITLE = "title"
	MENUS = "menus"
	TRANS = "trans"
	AGENT = "agent"
	CHECK = "check"
	SHARE = "share"
	GRANT = "grant"
	LOGIN = "login"
)
const HEADER = "header"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		HEADER: {Name: HEADER, Help: "标题栏", Value: kit.Data(
			TITLE, "shylinux.com/x/contexts", MENUS, `["header", ["setting", "black", "white", "print", "webpack", "unpack"]]`,
			LOGIN, kit.List("登录", "扫码"),
		)},
	}, Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(web.SERVE, aaa.WHITE, HEADER, RIVER, ACTION, FOOTER)
		}},
		"/header": {Name: "/header", Help: "标题栏", Action: map[string]*ice.Action{
			AGENT: {Name: "agent", Help: "应用宿主", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy("web.chat.wx.access", "config")
			}},
			CHECK: {Name: "check", Help: "登录检查", Hand: func(m *ice.Message, arg ...string) {
				_header_check(m, arg...)
			}},
			GRANT: {Name: "grant space", Help: "授权", Hand: func(m *ice.Message, arg ...string) {
				_header_grant(m, arg...)
			}},
			SHARE: {Name: "share type", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
				_header_share(m, arg...)
			}},

			aaa.LOGIN: {Name: "login", Help: "密码登录", Hand: func(m *ice.Message, arg ...string) {
				if aaa.UserLogin(m, arg[0], arg[1]) {
					web.RenderCookie(m, aaa.SessCreate(m, arg[0]))
				}
				m.Echo(m.Option(ice.MSG_USERNAME))
			}},
			aaa.LOGOUT: {Name: "logout", Help: "退出登录", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.SESS, mdb.REMOVE, ice.OptionHash(m.Option(ice.MSG_SESSID)))
			}},
			aaa.USERNICK: {Name: "usernick", Help: "用户昵称", Hand: func(m *ice.Message, arg ...string) {
				_header_users(m, aaa.USERNICK, arg...)
			}},
			aaa.BACKGROUND: {Name: "background", Help: "用户壁纸", Hand: func(m *ice.Message, arg ...string) {
				_header_users(m, aaa.BACKGROUND, arg...)
			}},
			aaa.AVATAR: {Name: "avatar", Help: "用户头像", Hand: func(m *ice.Message, arg ...string) {
				_header_users(m, aaa.AVATAR, arg...)
			}},

			code.WEBPACK: {Name: "webpack", Help: "打包页面", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.WEBPACK, mdb.CREATE, m.OptionSimple(kit.MDB_NAME))
				p := path.Join("src/release", ice.GO_MOD)
				if _, e := os.Stat(p); e == nil {
					m.Cmd(nfs.COPY, ice.GO_MOD, p)
				}
			}},
			"unpack": {Name: "unpack", Help: "开发模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.WEBPACK, "unpack")

				p := path.Join("src/debug", ice.GO_MOD)
				if _, e := os.Stat(p); e == nil {
					m.Cmd(nfs.COPY, ice.GO_MOD, p)
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(TRANS, kit.Format(kit.Value(c.Commands[cmd].Meta, "_trans")))
			m.Option(MENUS, m.Conf(HEADER, kit.Keym(MENUS)))
			msg := m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME))
			for _, k := range []string{aaa.USERNICK, aaa.BACKGROUND, aaa.AVATAR} {
				m.Option(k, msg.Append(k))
			}
			m.Echo(m.Conf(HEADER, kit.Keym(TITLE)))
		}},
		HEADER: {Name: "header", Help: "标题栏", Action: map[string]*ice.Action{
			GRANT: {Name: "grant space", Help: "授权", Hand: func(m *ice.Message, arg ...string) {
				_header_grant(m, arg...)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
	}})
}
