package bash

import (
	"io/ioutil"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	SID = "sid"
	ARG = "arg"
	SUB = "sub"
)
const SESS = "sess"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SESS: {Name: SESS, Help: "会话流", Value: kit.Data(
			kit.MDB_FIELD, "time,hash,status,username,hostname,pid,pwd,grant",
		)},
	}, Commands: map[string]*ice.Command{
		web.WEB_LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, _, e := m.R.FormFile(SUB); e == nil {
				defer f.Close()
				if b, e := ioutil.ReadAll(f); e == nil {
					m.Option(SUB, string(b)) // 文件参数
				}
			}

			switch m.RenderResult(); arg[0] {
			case "/qrcode", "/sess":
				return // 登录入口
			}

			if m.Warn(m.Option(SID, strings.TrimSpace(m.Option(SID))) == "", ice.ErrNotLogin, arg) {
				return
			}

			msg := m.Cmd(SESS, m.Option(SID))
			m.Option(ice.MSG_USERNAME, msg.Append(GRANT))
			m.Option(ice.MSG_USERROLE, aaa.UserRole(m, msg.Append(GRANT)))
			m.Option(tcp.HOSTNAME, msg.Append(tcp.HOSTNAME))
			if arg[0] == "/run/" {
				return
			}
			m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin, arg)
		}},
		"/qrcode": {Name: "/qrcode", Help: "二维码", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(cli.QRCODE, m.Option(kit.MDB_TEXT), m.Option(cli.FG), m.Option(cli.BG))
		}},
		"/sess": {Name: "/sess", Help: "会话", Action: map[string]*ice.Action{
			aaa.LOGOUT: {Name: "logout", Help: "退出", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SESS, mdb.MODIFY, kit.MDB_STATUS, aaa.LOGOUT, ice.Option{kit.MDB_HASH, m.Option(SID)})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Option(SID) == "" { // 终端登录
				m.Option(SID, m.Cmdx(SESS, mdb.CREATE, kit.MDB_STATUS, aaa.LOGIN, m.OptionSimple(aaa.USERNAME, tcp.HOSTNAME, cli.PID, cli.PWD)))
			} else { // 更新状态
				m.Cmdy(SESS, mdb.MODIFY, kit.MDB_STATUS, aaa.LOGIN, ice.Option{kit.MDB_HASH, m.Option(SID)})
			}
			m.Echo(m.Option(SID))
		}},
		SESS: {Name: "sess hash auto prunes", Help: "会话流", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.OptionFields(m.Config(kit.MDB_FIELD))
				m.Cmdy(mdb.PRUNES, m.PrefixKey(), "", mdb.HASH, kit.MDB_STATUS, aaa.LOGOUT)
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
		}},
	}})
}
