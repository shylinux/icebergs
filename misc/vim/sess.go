package vim

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
	PRE = "pre"
	PWD = "pwd"
	BUF = "buf"
	ROW = "row"
	COL = "col"
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
			case "/input", "/sess":
				return // 登录入口
			}

			if m.Warn(m.Option(SID, strings.TrimSpace(m.Option(SID))) == "", ice.ErrNotLogin, arg[0]) {
				return
			}

			msg := m.Cmd(SESS, m.Option(SID))
			m.Option(aaa.USERNAME, msg.Append(aaa.USERNAME))
			m.Option(tcp.HOSTNAME, msg.Append(tcp.HOSTNAME))
			m.Warn(m.Option(aaa.USERNAME) == "", ice.ErrNotLogin, arg[0])
		}},
		"/sess": {Name: "/sess", Help: "会话", Actions: ice.Actions{
			aaa.LOGOUT: {Name: "logout", Help: "退出", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SESS, mdb.MODIFY, mdb.STATUS, aaa.LOGOUT, ice.Option{mdb.HASH, m.Option(SID)})
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Option(SID) == "" { // 终端登录
				m.Option(SID, m.Cmdx(SESS, mdb.CREATE, mdb.STATUS, aaa.LOGIN, m.OptionSimple(aaa.USERNAME, tcp.HOSTNAME, cli.PID, cli.PWD)))
			} else {
				m.Cmdy(SESS, mdb.MODIFY, mdb.STATUS, aaa.LOGIN, ice.Option{mdb.HASH, m.Option(SID)})
			}
			m.Echo(m.Option(SID))
		}},
		SESS: {Name: "sess hash auto prunes", Help: "会话流", Actions: ice.MergeActions(ice.Actions{
			mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.OptionFields(mdb.HashField(m))
				m.Cmdy(mdb.PRUNES, m.PrefixKey(), "", mdb.HASH, mdb.STATUS, aaa.LOGOUT)
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,status,username,hostname,pid,pwd")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...)
		}},
	})
}
