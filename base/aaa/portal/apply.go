package portal

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

type apply struct {
	ice.Hash
	field string `data:"time,hash,email,mobile,username,status"`
	input string `name:"input" help:"输入" role:"void"`
	apply string `name:"apply" help:"申请" role:"void"`
	agree string `name:"agree userrole=tech" help:"同意" icon:"bi bi-check2-square"`
	login string `name:"login" help:"登录" role:"void"`
	list  string `name:"list hash auto" help:"注册"`
}

func (s apply) Input(m *ice.Message, arg ...string) {
	kit.If(m.Option(mdb.HASH) == "input", func() { m.Option(mdb.HASH, "") })
	if k := _cookie_key(m); m.Option(k) == "" || s.List(m, m.Option(k)).Length() == 0 && m.Result() == "" {
		ctx.DisplayStoryForm(m.Message, "email*", s.Apply)
	}
}
func (s apply) Apply(m *ice.Message, arg ...string) {
	if m.Warn(m.Options(arg).Cmd(aaa.USER, m.Option(aaa.EMAIL)).Length() > 0, "already exists") {
		return
	}
	arg = append(arg, aaa.USERNAME, m.Option(aaa.EMAIL))
	m.ProcessCookie(_cookie_key(m), s.Hash.Create(m, kit.Simple(arg, mdb.STATUS, s.Apply, aaa.IP, m.Option(ice.MSG_USERIP), aaa.UA, m.Option(ice.MSG_USERUA), cli.DAEMON, m.Option(ice.MSG_DAEMON))...))
}
func (s apply) Agree(m *ice.Message, arg ...string) {
	msg := s.Hash.List(m.Spawn(), m.Option(mdb.HASH))
	s.Hash.Modify(m, kit.Simple(m.OptionSimple(mdb.HASH, aaa.USERROLE), mdb.STATUS, s.Agree)...)
	m.Cmd(aaa.USER, mdb.CREATE, msg.AppendSimple(aaa.USERNICK, aaa.USERNAME), m.Option(aaa.USERROLE))
	web.PushNotice(m.Spawn(kit.Dict(ice.MSG_DAEMON, msg.Append(cli.DAEMON))).Message, "refresh")
}
func (s apply) Login(m *ice.Message, arg ...string) {
	m.Options(arg)
	if m.Option(aaa.EMAIL) != "" {
		m.Cmdy(aaa.EMAIL, aaa.SEND, m.Option(aaa.EMAIL), "", "login contexts", m.Cmd(web.SHARE, mdb.CREATE, mdb.TYPE, aaa.LOGIN).Append(mdb.LINK))
		m.ProcessHold()
		return
	}
	kit.If(m.Option(mdb.HASH) == "input", func() { m.Option(mdb.HASH, "") })
	m.OptionDefault(mdb.HASH, m.Option(_cookie_key(m)))
	msg := s.Hash.List(m.Spawn(), m.Option(mdb.HASH))
	s.Hash.Modify(m, kit.Simple(m.OptionSimple(mdb.HASH), mdb.STATUS, s.Login)...)
	sessid := m.Cmdx(aaa.SESS, mdb.CREATE, msg.Append(aaa.USERNAME))
	web.RenderCookie(m.Message, sessid)
	m.ProcessLocation(nfs.PS)
}
func (s apply) Email(m *ice.Message, arg ...string) {
	ctx.DisplayStoryForm(m.Message, "email*", s.Login)
}
func (s apply) List(m *ice.Message, arg ...string) *ice.Message {
	kit.If(m.Option(_cookie_key(m)), func(p string) { arg = []string{p} })
	s.Hash.List(m, arg...).Table(func(value ice.Maps) {
		switch value[mdb.STATUS] {
		case kit.FuncName(s.Apply):
			m.PushButton(s.Agree, s.Remove)
		case kit.FuncName(s.Agree):
			m.PushButton(s.Login, s.Remove)
		default:
			m.PushButton(s.Remove)
		}
	})
	if len(arg) == 0 {
		m.EchoQRCode(m.MergePodCmd("", "", ctx.ACTION, s.Input))
	} else {
		switch m.Append(mdb.STATUS) {
		case kit.FuncName(s.Apply):
			m.SetAppend()
			m.EchoInfoButton("please wait agree")
		case kit.FuncName(s.Agree):
			m.SetAppend()
			m.EchoInfoButton("please login", s.Login)
		case kit.FuncName(s.Login):
			m.SetAppend()
			m.EchoAnchor(nfs.PS)
			m.ProcessLocation(nfs.PS)
		}
	}
	return m
}

func init() { ice.Cmd("aaa.apply", apply{}) }

func _cookie_key(m *ice.Message) string { return kit.Keys(m.PrefixKey(), mdb.HASH) }
