package portal

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

type apply struct {
	ice.Hash
	email string `data:"admin"`
	field string `data:"time,hash,mobile,email,usernick,username,userrole,status"`
	sso   string `name:"sso" help:"登录"`
	input string `name:"input" help:"注册" role:"void"`
	apply string `name:"apply" help:"申请" role:"void"`
	agree string `name:"agree userrole=tech" help:"同意" icon:"bi bi-check2-square"`
	login string `name:"login" help:"登录" role:"void"`
	list  string `name:"list hash auto sso" help:"注册"`
}

func (s apply) Sso(m *ice.Message, arg ...string) {
	m.Cmd(web.CHAT_HEADER, mdb.CREATE, mdb.TYPE, cli.QRCODE, mdb.NAME, "扫码登录", mdb.ORDER, "10")
	m.Cmd(web.CHAT_HEADER, mdb.CREATE, mdb.TYPE, mdb.PLUGIN, mdb.NAME, "邮箱登录", mdb.ORDER, "20", ctx.INDEX, m.PrefixKey(), ctx.ARGS, kit.FuncName(s.Email))
	m.Cmd(web.CHAT_HEADER, mdb.CREATE, mdb.TYPE, mdb.PLUGIN, mdb.NAME, "注册用户", mdb.ORDER, "30", ctx.INDEX, m.PrefixKey(), ctx.ARGS, kit.FuncName(s.Input))
}
func (s apply) Input(m *ice.Message, arg ...string) {
	kit.If(m.Option(mdb.HASH) == kit.FuncName(s.Input), func() { m.Options(mdb.HASH, "") })
	if k := _cookie_key(m); m.Option(k) == "" || s.List(m, m.Option(k)).Length() == 0 && m.Result() == "" {
		ctx.DisplayStoryForm(m.Message, "email*", aaa.USERNICK, s.Apply)
	}
}
func (s apply) Apply(m *ice.Message, arg ...string) {
	if m.Warn(m.Options(arg).Cmd(aaa.USER, m.Option(aaa.EMAIL)).Length() > 0, ice.ErrAlreadyExists) {
		return
	}
	m.ProcessCookie(_cookie_key(m), s.Hash.Create(m, kit.Simple(append(arg, aaa.USERNAME, m.Option(aaa.EMAIL)), mdb.STATUS, s.Apply,
		aaa.IP, m.Option(ice.MSG_USERIP), aaa.UA, m.Option(ice.MSG_USERUA), cli.DAEMON, m.Option(ice.MSG_DAEMON))...))
}
func (s apply) Agree(m *ice.Message, arg ...string) {
	msg := s.Hash.List(m.Spawn(), m.Option(mdb.HASH))
	s.Hash.Modify(m, kit.Simple(m.OptionSimple(mdb.HASH, aaa.USERROLE), mdb.STATUS, s.Agree)...)
	m.Cmd(aaa.USER, mdb.CREATE, msg.AppendSimple(aaa.USERNICK, aaa.USERNAME), m.Option(aaa.USERROLE))
	web.PushNotice(m.Spawn(kit.Dict(ice.MSG_DAEMON, msg.Append(cli.DAEMON))).Message, html.REFRESH)
}
func (s apply) Login(m *ice.Message, arg ...string) {
	kit.If(m.Option(mdb.HASH) == kit.FuncName(s.Input), func() { m.Options(mdb.HASH, "") })
	if m.Options(arg).Option(aaa.EMAIL) == "" {
		m.OptionDefault(mdb.HASH, m.Option(_cookie_key(m)))
		s.Hash.Modify(m, kit.Simple(m.OptionSimple(mdb.HASH), mdb.STATUS, s.Login)...)
		web.RenderCookie(m.Message, m.Cmdx(aaa.SESS, mdb.CREATE, s.Hash.List(m.Spawn(), m.Option(mdb.HASH)).Append(aaa.USERNAME)))
		m.ProcessLocation(nfs.PS)
	} else {
		if m.Warn(m.Cmd(aaa.USER, m.Option(aaa.EMAIL)).Length() == 0, ice.ErrNotFound) {
			return
		}
		m.Options(ice.MSG_USERNAME, m.Option(aaa.EMAIL))
		space := kit.Keys(kit.Slice(kit.Split(m.Option(ice.MSG_DAEMON), nfs.PT), 0, -1))
		share := m.Cmd(web.SHARE, mdb.CREATE, mdb.TYPE, web.FIELD, mdb.NAME, web.CHAT_GRANT, mdb.TEXT, space).Append(mdb.LINK)
		m.Cmdy(aaa.EMAIL, aaa.SEND, mdb.Config(m.Message, aaa.EMAIL), m.Option(aaa.EMAIL), "", "login contexts, please grant", html.FormatA(share))
		m.ProcessHold()
	}
}
func (s apply) Email(m *ice.Message, arg ...string) {
	ctx.DisplayStoryForm(m.Message, "email*", s.Login).Echo("please auth login in mailbox, after email sent")
}
func (s apply) List(m *ice.Message, arg ...string) *ice.Message {
	if m.Cmd(aaa.EMAIL, mdb.Config(m.Message, aaa.EMAIL)).Length() == 0 {
		m.Echo("please add admin email")
		return m
	}
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
	} else if m.Option(_cookie_key(m)) != "" || !kit.IsIn(m.ActionKey(), "", ice.LIST, mdb.SELECT) {
		switch m.Append(mdb.STATUS) {
		case kit.FuncName(s.Login):
			if m.ActionKey() == kit.FuncName(s.Input) {
				m.ProcessCookie(_cookie_key(m), "")
			} else {
				m.SetAppend().ProcessLocation(nfs.PS)
			}
		case kit.FuncName(s.Agree):
			m.SetAppend().EchoInfoButton("please login", s.Login)
		case kit.FuncName(s.Apply):
			m.SetAppend().EchoInfoButton("please wait admin agree")
		}
	}
	return m
}

func init() { ice.Cmd("aaa.apply", apply{}) }

func _cookie_key(m *ice.Message) string { return kit.Keys(m.PrefixKey(), mdb.HASH) }
