package aaa

import (
	"net/smtp"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
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
		MAILBOX = "mailbox"
		NL      = "\r\n"
		DF      = ": "
	)
	Index.MergeCommands(ice.Commands{
		EMAIL: {Help: "邮件", Actions: ice.MergeActions(ice.Actions{
			MAILBOX: {Help: "邮箱", Hand: func(m *ice.Message, arg ...string) { m.EchoIFrame(mdb.Config(m, MAILBOX)).ProcessInner() }},
			SEND: {Name: "send to*='shy@shylinux.com' cc subject*=hi content*:textarea=hello", Help: "发送", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd("", m.OptionDefault(mdb.NAME, ADMIN))
				if m.Warn(msg.Append(SERVICE) == "", ice.ErrNotValid, SERVICE) {
					return
				}
				content := []byte(kit.JoinKV(DF, NL, kit.Simple(FROM, msg.Append(USERNAME), m.OptionSimple(TO, CC, SUBJECT), DATE, time.Now().Format(time.RFC1123Z), "Content-Type", "text/html; charset=UTF-8")...) + NL + NL + m.Option(CONTENT))
				auth := smtp.PlainAuth("", msg.Append(USERNAME), msg.Append(PASSWORD), kit.Split(msg.Append(SERVICE), ice.DF)[0])
				m.Logs(EMAIL, SEND, string(content)).Warn(smtp.SendMail(msg.Append(SERVICE), auth, msg.Append(USERNAME), kit.Split(m.Option(TO)), content))
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,service,username", ice.ACTION, SEND))},
	})
}
