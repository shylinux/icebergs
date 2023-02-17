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
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
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
					m.Option(SUB, string(b))
				}
			}
			switch m.RenderResult(); arg[0] {
			case web.P(cli.QRCODE), web.PP(SESS):
				return
			}
			if m.Option(SID, strings.TrimSpace(m.Option(SID))) == "" && m.Option(ice.MSG_USERNAME) != "" {
				return
			} else if m.Warn(m.Option(SID) == "", ice.ErrNotLogin, arg) {
				return
			} else if msg := m.Cmd(SESS, m.Option(SID)); msg.Append(GRANT) == "" {
				aaa.SessAuth(m, ice.Maps{aaa.USERNAME: msg.Append(aaa.USERNAME), aaa.USERNICK: msg.Append(aaa.USERNAME), aaa.USERROLE: aaa.VOID}).Options(msg.AppendSimple(aaa.USERNAME, tcp.HOSTNAME, cli.RELEASE))
			} else {
				aaa.SessAuth(m, kit.Dict(aaa.USERNAME, msg.Append(GRANT), aaa.USERNICK, aaa.UserNick(m, msg.Append(GRANT)), aaa.USERROLE, aaa.UserRole(m, msg.Append(GRANT)))).Options(msg.AppendSimple(aaa.USERNAME, tcp.HOSTNAME, cli.RELEASE))
			}
		}},
		web.P(cli.QRCODE): {Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(cli.QRCODE, tcp.PublishLocalhost(m, m.Option(mdb.TEXT)), m.Option(cli.FG), m.Option(cli.BG))
		}},
		web.PP(SESS): {Actions: ice.Actions{
			aaa.LOGOUT: {Hand: func(m *ice.Message, arg ...string) {
				if !m.Warn(m.Option(SID) == "", ice.ErrNotValid, SID) {
					mdb.HashModify(m, mdb.HASH, m.Option(SID), mdb.STATUS, aaa.LOGOUT)
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Option(SID) == "" {
				m.Option(SID, mdb.HashCreate(m, mdb.STATUS, aaa.LOGIN, m.OptionSimple(aaa.USERNAME, tcp.HOSTNAME, cli.PID, cli.PWD, cli.RELEASE)))
			} else {
				mdb.HashModify(m, mdb.HASH, m.Option(SID), mdb.STATUS, aaa.LOGIN)
				m.Echo(m.Option(SID))
			}
		}},
		SESS: {Name: "sess hash auto invite prunes", Help: "会话流", Actions: ice.MergeActions(ice.Actions{
			aaa.INVITE: {Hand: func(m *ice.Message, arg ...string) {
				code.PublishScript(m, `export ctx_dev={{.Option "domain"}}; ctx_temp=$(mktemp); curl -o $ctx_temp -fsSL $ctx_dev; source $ctx_temp`)
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,status,username,hostname,release,pid,pwd,grant"))},
	})
}
