package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _wx_reply(m *ice.Message, tmpl string) {
	if res, err := kit.Render(m.Config(nfs.TEMPLATE), m); err == nil {
		m.SetResult().RenderResult(string(res))
	}
}

const TEXT = "text"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		TEXT: {Name: TEXT, Help: "文本", Value: kit.Data(nfs.TEMPLATE, text)},
	}, Commands: ice.Commands{
		TEXT: {Name: "text", Help: "文本", Actions: ice.Actions{
			MENU: {Name: "menu name", Help: "菜单", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(MENU, kit.Select("home", m.Option(mdb.NAME)))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Cmdy(arg); m.Length() == 0 && (m.Result() == "" || m.IsErrNotFound()) {
				m.SetResult().Cmdy(cli.SYSTEM, arg) // 系统命令
			}
			if m.Result() == "" {
				m.Table()
			}
			_wx_reply(m, m.CommandKey())
		}},
	}})
}

var text = `<xml>
<FromUserName><![CDATA[{{.Option "ToUserName"}}]]></FromUserName>
<ToUserName><![CDATA[{{.Option "FromUserName"}}]]></ToUserName>
<CreateTime>{{.Option "CreateTime"}}</CreateTime>
<MsgType><![CDATA[text]]></MsgType>
<Content><![CDATA[{{.Result}}]]></Content>
</xml>`
