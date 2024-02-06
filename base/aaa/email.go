package aaa

import (
	"net/smtp"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	ADMIN   = "admin"
	SEND    = "send"
	DATE    = "date"
	FROM    = "from"
	TO      = "to"
	CC      = "cc"
	SUBJECT = "subject"
	CONTENT = "content"
)
const EMAIL = "email"

func init() {
	const (
		ADMIN   = "admin"
		SERVICE = "service"
		NL      = "\r\n"
		DF      = ": "
	)
	Index.MergeCommands(ice.Commands{
		EMAIL: {Help: "邮件", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create name*=admin service*='mail.shylinux.com:25' username*='shy@shylinux.com' password*"},
			SEND: {Name: "send from=admin to*='shy@shylinux.com' cc subject*=hi content*:textarea=hello", Help: "发送", Icon: "bi bi-send-plus", Hand: func(m *ice.Message, arg ...string) {
				msg := mdb.HashSelects(m.Spawn(), m.OptionDefault(FROM, ADMIN))
				if m.WarnNotFound(msg.Append(SERVICE) == "", m.Option(FROM)) {
					return
				}
				m.Toast(ice.PROCESS, "", "-1")
				content := []byte(kit.JoinKV(DF, NL, kit.Simple(FROM, msg.Append(USERNAME), m.OptionSimple(TO, CC, SUBJECT), DATE, time.Now().Format(time.RFC1123Z), "Content-Type", "text/html; charset=UTF-8")...) + NL + NL + m.Option(CONTENT))
				auth := smtp.PlainAuth("", msg.Append(USERNAME), msg.Append(PASSWORD), kit.Split(msg.Append(SERVICE), ice.DF)[0])
				m.Logs(EMAIL, SEND, string(content))
				if !m.Warn(smtp.SendMail(msg.Append(SERVICE), auth, msg.Append(USERNAME), kit.Split(m.Option(TO)), content)) {
					m.Toast(ice.SUCCESS)
				}
			}},
		}, mdb.DevDataAction("name,service,username,password"), mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,service,username", ice.ACTION, SEND)), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 && m.Length() == 0 {
				m.EchoInfoButton(m.Trans("please add admin email", "请配置管理员邮箱"), mdb.CREATE, mdb.DEV_REQUEST)
			} else if len(arg) == 0 {
				m.Action(mdb.CREATE, mdb.DEV_REQUEST)
			}
		}},
	})
}
func SendEmail(m *ice.Message, from, to, cc string, arg ...string) {
	m.Option(ice.MSG_USERHOST, strings.Split(m.Option(ice.MSG_USERHOST), "://")[1])
	m.Cmdy(EMAIL, SEND, kit.Select(mdb.Config(m, EMAIL), from), kit.Select(m.Option(EMAIL), to), cc,
		strings.TrimSpace(kit.Select(ice.Render(m, ice.RENDER_TEMPLATE, SUBJECT_HTML), arg, 0)),
		kit.Select(ice.Render(m, ice.RENDER_TEMPLATE, CONTENT_HTML), arg, 1),
	)
}
