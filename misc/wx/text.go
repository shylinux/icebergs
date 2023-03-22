package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _wx_reply(m *ice.Message, tmpl string) {
	if res, err := kit.Render(mdb.Config(m, nfs.TEMPLATE), m); err == nil {
		m.SetResult().RenderResult(string(res))
	}
}

const TEXT = "text"

func init() {
	Index.MergeCommands(ice.Commands{
		TEXT: {Name: "text", Help: "文本", Actions: ice.MergeActions(ice.Actions{
			MENU: {Name: "menu name=home", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(MENU, m.Option(mdb.NAME)) }},
		}, ctx.ConfAction(nfs.TEMPLATE, text)), Hand: func(m *ice.Message, arg ...string) {
			if m.Cmdy(arg); m.IsErrNotFound() {
				m.SetResult().Cmdy(cli.SYSTEM, arg)
			}
			kit.If(m.Result() == "", func() { m.Table() })
			_wx_reply(m, m.CommandKey())
		}},
	})
}

var text = `<xml>
<FromUserName><![CDATA[{{.Option "ToUserName"}}]]></FromUserName>
<ToUserName><![CDATA[{{.Option "FromUserName"}}]]></ToUserName>
<CreateTime>{{.Option "CreateTime"}}</CreateTime>
<MsgType><![CDATA[text]]></MsgType>
<Content><![CDATA[{{.Result}}]]></Content>
</xml>`
