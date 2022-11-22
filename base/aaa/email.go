package aaa

import (
	"net/smtp"
	
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	SEND = "send"
)
const EMAIL = "email"

func init() {
	const (
		TO = "to"
		SUBJECT = "subject"
		CONTENT = "content"
		SERVICE = "service"
		NL = "\r\n"
		DF = ": "
	)
	Index.MergeCommands(ice.Commands{
		EMAIL: {Name: "email name auto create", Help: "邮件", Actions: ice.MergeActions(ice.Actions{
			SEND: {Name: "send to='shylinux@163.com' subject=hi content:textarea=hello", Help: "发送", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(SERVICE) == "" {
					msg := m.Cmd("", "admin")
					m.Option(SERVICE, msg.Append(SERVICE))
					m.Option(USERNAME, msg.Append(USERNAME))
					m.Option(PASSWORD, msg.Append(PASSWORD))
				}
				content := []byte(kit.JoinKV(DF, NL, "To", m.Option(TO), "From", m.Option(USERNAME), "Subject", m.Option(SUBJECT), "Content-Type", "text/html; charset=UTF-8")+NL+NL+m.Option(CONTENT))
				auth := smtp.PlainAuth("", m.Option(USERNAME), m.Option(PASSWORD), kit.Split(m.Option(SERVICE), ice.DF)[0]) 
				m.Warn(smtp.SendMail(m.Option(SERVICE), auth, m.Option(USERNAME), kit.Split(m.Option(TO)), content))
				m.Debug("email send %v %v", auth, string(content))
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,username,password,service")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).PushAction(SEND)
		}},
	})	
}
