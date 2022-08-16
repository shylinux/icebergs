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
)

const (
	SID = "sid"
	ARG = "arg"
	SUB = "sub"
)
const SESS = "sess"

func init() {
	Index.MergeCommands(ice.Commands{
		web.WEB_LOGIN: {Hand: func(m *ice.Message, arg ...string) {
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
			m.Auth(aaa.USERROLE, m.Option(ice.MSG_USERROLE), aaa.USERNAME, m.Option(ice.MSG_USERNAME))
			if arg[0] == "/run/" {
				return
			}
			m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin, arg)
		}},
		"/qrcode": {Name: "/qrcode", Help: "二维码", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(cli.QRCODE, m.Option(mdb.TEXT), m.Option(cli.FG), m.Option(cli.BG))
		}},
		"/sess": {Name: "/sess", Help: "会话", Actions: ice.Actions{
			aaa.LOGOUT: {Name: "logout", Help: "退出", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, mdb.HASH, m.Option(SID), mdb.STATUS, aaa.LOGOUT)
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Option(SID) == "" { // 终端登录
				m.Option(SID, mdb.HashCreate(m, mdb.STATUS, aaa.LOGIN, m.OptionSimple(aaa.USERNAME, tcp.HOSTNAME, cli.PID, cli.PWD)))
			} else { // 更新状态
				mdb.HashModify(m, mdb.HASH, m.Option(SID), mdb.STATUS, aaa.LOGIN)
				m.Echo(m.Option(SID))
			}
		}},
		SESS: {Name: "sess hash auto prunes", Help: "会话流", Actions: ice.MergeActions(ice.Actions{
			mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashPrunesValue(m, mdb.STATUS, aaa.LOGOUT)
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,status,username,hostname,pid,pwd,grant"))},
	})
}
