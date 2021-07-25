package bash

import (
	"io/ioutil"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const (
	SID = "sid"
	ARG = "arg"
	SUB = "sub"
)
const SESS = "sess"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SESS: {Name: SESS, Help: "会话流", Value: kit.Data(
				kit.MDB_FIELD, "time,hash,status,username,hostname,pid,pwd",
			)},
		},
		Commands: map[string]*ice.Command{
			web.WEB_LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if f, _, e := m.R.FormFile(SUB); e == nil {
					defer f.Close()
					if b, e := ioutil.ReadAll(f); e == nil {
						m.Option(SUB, string(b)) // 文件参数
					}
				}

				switch m.Render(ice.RENDER_RESULT); m.R.URL.String() {
				case "/qrcode", "/sess": // 登录入口
					return
				}

				if m.Warn(m.Option(SID, strings.TrimSpace(m.Option(SID))) == "", ice.ErrNotLogin) {
					return
				}

				msg := m.Cmd(mdb.SELECT, m.Prefix(SESS), "", mdb.HASH, kit.MDB_HASH, m.Option(SID),
					ice.OptionFields(m.Conf(SESS, kit.META_FIELD)))
				m.Option(aaa.USERNAME, msg.Append(aaa.USERNAME))
				m.Option(tcp.HOSTNAME, msg.Append(tcp.HOSTNAME))
				m.Warn(m.Option(aaa.USERNAME) == "", ice.ErrNotLogin)
			}},
			"/qrcode": {Name: "/qrcode", Help: "二维码", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(cli.QRCODE, m.Option(kit.MDB_TEXT), m.Option(cli.FG), m.Option(cli.BG))
			}},
			"/sess": {Name: "/sess", Help: "会话", Action: map[string]*ice.Action{
				aaa.LOGOUT: {Name: "logout", Help: "退出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(SESS), "", mdb.HASH, kit.MDB_HASH, m.Option(SID), kit.MDB_STATUS, aaa.LOGOUT)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Option(SID) == "" {
					m.Option(SID, m.Cmdx(mdb.INSERT, m.Prefix(SESS), "", mdb.HASH, kit.MDB_STATUS, aaa.LOGIN,
						m.OptionSimple(aaa.USERNAME, tcp.HOSTNAME, cli.PID, cli.PWD)))
				} else {
					m.Cmdy(mdb.MODIFY, m.Prefix(SESS), "", mdb.HASH, kit.MDB_HASH, m.Option(SID), kit.MDB_STATUS, aaa.LOGIN)
				}
				m.Echo(m.Option(SID))
			}},
			SESS: {Name: "sess hash auto prunes", Help: "会话流", Action: map[string]*ice.Action{
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.OptionFields(m.Conf(SESS, kit.META_FIELD))
					m.Cmdy(mdb.DELETE, m.Prefix(SESS), "", mdb.HASH, m.OptionSimple(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.OptionFields(m.Conf(SESS, kit.META_FIELD))
					m.Cmdy(mdb.PRUNES, m.Prefix(SESS), "", mdb.HASH, kit.MDB_STATUS, aaa.LOGOUT)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), m.Conf(SESS, kit.META_FIELD))
				m.Cmdy(mdb.SELECT, m.Prefix(SESS), "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction(mdb.REMOVE)
			}},
		},
	})
}
