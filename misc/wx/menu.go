package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

func _wx_action(m *ice.Message) {
	m.SetResult().RenderResult()

	m.Echo(`<xml>
<FromUserName><![CDATA[%s]]></FromUserName>
<ToUserName><![CDATA[%s]]></ToUserName>
<CreateTime>%s</CreateTime>
<MsgType><![CDATA[%s]]></MsgType>
`, m.Option("ToUserName"), m.Option("FromUserName"), m.Option("CreateTime"), "news")

	count := 0
	m.Tables(func(value ice.Maps) { count++ })
	m.Echo(`<ArticleCount>%d</ArticleCount>`, count)

	share := m.Cmdx(web.SHARE, mdb.CREATE, mdb.TYPE, web.LOGIN)

	m.Echo(`<Articles>`)
	m.Tables(func(value ice.Maps) {
		m.Echo(`<item>
<Title><![CDATA[%s]]></Title>
<Description><![CDATA[%s]]></Description>
<PicUrl><![CDATA[%s]]></PicUrl>
<Url><![CDATA[%s]]></Url>
</item>
`, value[wiki.TITLE], value[wiki.SPARK], value[wiki.IMAGE],
			kit.MergeURL(kit.Format(value[wiki.REFER]), web.SHARE, share))
	})
	m.Echo(`</Articles>`)
	m.Echo(`</xml>`)

	m.Debug("echo: %v", m.Result())
}

const MENU = "menu"

func init() {
	Index.MergeCommands(ice.Commands{
		MENU: {Name: "menu zone id auto insert", Help: "菜单", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert zone=home title=hi refer=hello image", Help: "添加"},
		}, mdb.ZoneAction(mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,id,title,refer,image")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) > 0 {
				_wx_action(m)
			}
		}},
	})
}
