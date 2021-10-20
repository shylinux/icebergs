package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	kit "shylinux.com/x/toolkits"
)

const TEXT = "text"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		TEXT: {Name: TEXT, Help: "文本", Value: kit.Data(kit.MDB_TEMPLATE, text)},
	}, Commands: map[string]*ice.Command{
		TEXT: {Name: "text", Help: "文本", Action: map[string]*ice.Action{
			MENU: {Name: "menu name", Help: "菜单", Hand: func(m *ice.Message, arg ...string) {
				_wx_action(m.Cmdy(MENU, kit.Select("home", m.Option(kit.MDB_NAME))))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 执行命令
			if m.Cmdy(arg); len(m.Appendv(ice.MSG_APPEND)) == 0 && len(m.Result()) == 0 {
				m.Cmdy(cli.SYSTEM, arg)
			}
			if len(m.Result()) == 0 {
				m.Table()
			}

			// 返回结果
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
