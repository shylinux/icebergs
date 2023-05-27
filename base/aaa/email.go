package aaa

import (
	"net/smtp"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	SEND    = "send"
	SUBJECT = "subject"
	CONTENT = "content"
)
const EMAIL = "email"

func init() {
	const (
		TO      = "to"
		ADMIN   = "admin"
		SERVICE = "service"
		NL      = "\r\n"
		DF      = ": "
	)
	Index.MergeCommands(ice.Commands{
		EMAIL: {Name: "email name auto create", Help: "邮件", Actions: ice.MergeActions(ice.Actions{
			SEND: {Name: "send to*='shylinux@163.com' subject*=hi content*:textarea=hello", Help: "发送", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd("", m.OptionDefault(mdb.NAME, ADMIN))
				if m.Warn(msg.Append(SERVICE) == "", ice.ErrNotValid, SERVICE) {
					return
				}
				content := []byte(kit.JoinKV(DF, NL, "From", msg.Append(USERNAME), "To", m.Option(TO), "Subject", m.Option(SUBJECT), "Content-Type", "text/html; charset=UTF-8") + NL + NL + m.Option(CONTENT))
				auth := smtp.PlainAuth("", msg.Append(USERNAME), msg.Append(PASSWORD), kit.Split(msg.Append(SERVICE), ice.DF)[0])
				m.Logs(EMAIL, SEND, string(content)).Warn(smtp.SendMail(msg.Append(SERVICE), auth, msg.Append(USERNAME), kit.Split(m.Option(TO)), content))
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,service,username", ice.ACTION, SEND))},
	})
}
