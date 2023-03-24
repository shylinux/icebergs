package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

func _wx_action(m *ice.Message) (count int) {
	m.SetResult().RenderResult()
	m.Echo(`<xml>
<FromUserName><![CDATA[%s]]></FromUserName>
<ToUserName><![CDATA[%s]]></ToUserName>
<CreateTime>%s</CreateTime>
<MsgType><![CDATA[%s]]></MsgType>
`, m.Option("ToUserName"), m.Option("FromUserName"), m.Option("CreateTime"), "news")
	m.Table(func(value ice.Maps) { count++ })
	m.Echo(`<ArticleCount>%d</ArticleCount>`, count).Echo(`<Articles>`)
	share := m.Cmdx(web.SHARE, mdb.CREATE, mdb.TYPE, web.LOGIN)
	m.Table(func(value ice.Maps) {
		m.Echo(`<item>
<Title><![CDATA[%s]]></Title>
<Description><![CDATA[%s]]></Description>
<PicUrl><![CDATA[%s]]></PicUrl>
<Url><![CDATA[%s]]></Url>
</item>
`, value[wiki.TITLE], value[wiki.SPARK], value[wiki.IMAGE], kit.MergeURL2(kit.Format(value[wiki.REFER]), "/share/"+share))
	}).Echo(`</Articles>`).Echo(`</xml>`)
	m.Debug("echo: %v", m.Result())
	return
}

const MENU = "menu"

func init() {
	Index.MergeCommands(ice.Commands{
		MENU: {Name: "menu zone id auto insert", Help: "菜单", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert zone=home title=hi refer=hello image"},
		}, mdb.ZoneAction(mdb.FIELD, "time,id,title,refer,image")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) > 0 {
				_wx_action(m)
			}
		}},
	})
}
