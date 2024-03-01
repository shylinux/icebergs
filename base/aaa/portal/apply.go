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
	email    string `data:"admin"`
	checkbox string `data:"true"`
	online   string `data:"true"`
	field    string `data:"time,hash,status,email,usernick,username,userrole,icons,agent,system,ip,ua"`
	apply    string `name:"apply" help:"申请" role:"void"`
	agree    string `name:"agree userrole=tech,void" help:"同意" icon:"bi bi-check2-square"`
	login    string `name:"login" help:"登录" role:"void"`
	list     string `name:"list hash auto sso" help:"注册"`
}

func (s apply) Sso(m *ice.Message, arg ...string) {
	m.AddHeaderLogin(cli.QRCODE, cli.QRCODE, "扫码登录", "10")
	m.AddHeaderLogin(mdb.PLUGIN, aaa.EMAIL, "邮箱登录", "20", ctx.INDEX, m.ShortKey(), ctx.ARGS, kit.FuncName(s.Login))
	m.AddHeaderLogin(mdb.PLUGIN, aaa.APPLY, "注册用户", "30", ctx.INDEX, m.ShortKey(), ctx.ARGS, kit.FuncName(s.Apply))
}
func (s apply) Apply(m *ice.Message, arg ...string) {
	if m.IsGetMethod() {
		if k := _cookie_key(m); m.Option(k) == "" || s.List(m, m.Option(k)).Length() == 0 && m.Result() == "" {
			m.DisplayForm(m, "email*", aaa.USERNICK, s.Apply)
		}
	} else if !m.WarnAlreadyExists(m.Options(arg).Cmd(aaa.USER, m.Option(aaa.EMAIL)).Length() > 0, m.Option(aaa.EMAIL)) {
		m.Option(ice.MSG_USERNAME, m.Option(aaa.EMAIL), ice.MSG_USERNICK, kit.Split(m.Option(aaa.EMAIL), "@")[0])
		m.ProcessCookie(_cookie_key(m), s.Hash.Create(m, kit.Simple(arg, mdb.STATUS, kit.FuncName(s.Apply), web.ParseUA(m.Message))...))
		m.StreamPushRefreshConfirm()
	}
}
func (s apply) Agree(m *ice.Message, arg ...string) {
	if m.WarnNotValid(m.Option(mdb.HASH) == "", mdb.HASH) {
		return
	}
	msg := s.Hash.List(m.Spawn(), m.Option(mdb.HASH))
	if m.WarnNotFound(msg.Length() == 0, m.Option(mdb.HASH)) {
		return
	}
	s.Hash.Modify(m, kit.Simple(m.OptionSimple(mdb.HASH, aaa.USERROLE), mdb.STATUS, s.Agree)...)
	m.Cmd(aaa.USER, mdb.CREATE, msg.AppendSimple(aaa.USERNICK, aaa.USERNAME), m.OptionSimple(aaa.USERROLE))
	m.PushRefresh(msg.Append(cli.DAEMON))
}
func (s apply) Login(m *ice.Message, arg ...string) {
	kit.If(m.Option(mdb.HASH) == kit.FuncName(s.Apply), func() { m.Options(mdb.HASH, "") })
	if m.IsGetMethod() {
		m.DisplayForm("email*", s.Login)
	} else if m.Options(arg).Option(aaa.EMAIL) == "" {
		if m.WarnNotValid(m.OptionDefault(mdb.HASH, m.Option(_cookie_key(m))) == "", mdb.HASH) {
			m.ProcessCookie(_cookie_key(m), "")
			return
		}
		msg := s.Hash.List(m.Spawn(), m.Option(mdb.HASH))
		if m.WarnNotFound(msg.Length() == 0, m.Option(mdb.HASH)) {
			m.ProcessCookie(_cookie_key(m), "")
			return
		}
		s.Hash.Modify(m, kit.Simple(m.OptionSimple(mdb.HASH), mdb.STATUS, s.Login)...)
		web.RenderCookie(m.Message, m.Cmdx(aaa.SESS, mdb.CREATE, msg.Append(aaa.USERNAME)))
		m.ProcessLocation(nfs.PS)
	} else {
		if m.WarnNotFound(m.Cmd(aaa.USER, m.Option(aaa.EMAIL)).Length() == 0, m.Option(aaa.EMAIL)) {
			return
		}
		m.Options(ice.MSG_USERNAME, m.Option(aaa.EMAIL))
		space := kit.Keys(kit.Slice(kit.Split(m.Option(ice.MSG_DAEMON), nfs.PT), 0, -1))
		share := m.Cmd(web.SHARE, mdb.CREATE, mdb.TYPE, web.FIELD, mdb.NAME, web.CHAT_GRANT, mdb.TEXT, space).Append(mdb.LINK)
		m.Options(web.LINK, share).SendEmail("", "", "")
		m.ProcessHold(m.Trans("please auth login in mailbox", "请注意查收邮件"))
	}
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
		m.EchoQRCode(m.MergePodCmd("", "", ctx.ACTION, s.Apply))
	} else if m.Option(_cookie_key(m)) != "" || m.ActionKey() != "" {
		switch m.Append(mdb.STATUS) {
		case kit.FuncName(s.Login):
			if m.ActionKey() == kit.FuncName(s.Apply) {
				m.ProcessCookie(_cookie_key(m), "")
			} else {
				m.SetAppend().ProcessLocation(nfs.PS)
			}
		case kit.FuncName(s.Agree):
			m.SetAppend().EchoInfoButton(m.Trans("please login", "请登录"), s.Login)
		case kit.FuncName(s.Apply):
			m.SetAppend().EchoInfoButton(m.Trans("please wait admin agree", "请等待管理员同意"), nil)
		}
	}
	return m
}

func init() { ice.Cmd("aaa.apply", apply{}) }

func _cookie_key(m *ice.Message) string { return kit.Keys(m.PrefixKey(), mdb.HASH) }
