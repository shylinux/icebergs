package wx

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

func _wx_parse(m *ice.Message) {
	data := struct {
		FromUserName string
		ToUserName   string
		CreateTime   int64
		MsgType      string
		MsgId        int64
		Event        string
		Content      string
		Title        string
		Description  string
		Url          string
		PicUrl       string
		Location_X   float64
		Location_Y   float64
		Scale        string
		Label        string
	}{}
	defer m.R.Body.Close()
	buf, _ := ioutil.ReadAll(m.R.Body)
	m.Debug("buf: %+v", string(buf))
	xml.NewDecoder(bytes.NewBuffer(buf)).Decode(&data)
	m.Debug("data: %+v", data)
	m.Option("FromUserName", data.FromUserName)
	m.Option("ToUserName", data.ToUserName)
	m.Option("CreateTime", data.CreateTime)
	m.Option("MsgId", data.MsgId)
	m.Option("Event", data.Event)
	m.Option("MsgType", data.MsgType)
	m.Option("Content", data.Content)
	m.Option("Title", data.Title)
	m.Option("Description", data.Description)
	m.Option("URL", data.Url)
	m.Option("URL", data.PicUrl)
	m.Option("LocationX", kit.Int(data.Location_X*100000))
	m.Option("LocationY", kit.Int(data.Location_Y*100000))
	m.Option("Scale", data.Scale)
	m.Option("Label", data.Label)
}

const LOGIN = "login"

func init() {
	Index.MergeCommands(ice.Commands{
		web.PP(LOGIN): {Actions: aaa.WhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			if m.Cmdx(ACCESS, CHECK) == "" {
				return
			}
			_wx_parse(m)
			m.Option(ice.MSG_USERZONE, WX)
			aaa.SessAuth(m, kit.Dict(aaa.USERNAME, m.Option("FromUserName"), aaa.USERROLE, aaa.UserRole(m, m.Option("FromUserName"))))
			switch m.Option("MsgType") {
			case EVENT:
				m.Cmdy(EVENT, m.Option("Event"))
			case TEXT:
				if cmds := kit.Split(m.Option("Content")); aaa.Right(m, cmds) {
					m.Cmdy(TEXT, cmds)
				} else {
					m.Cmdy(MENU, "home")
				}
			case mdb.LINK:
				m.Cmdy(chat.FAVOR, mdb.CREATE, mdb.TYPE, mdb.LINK, mdb.NAME, m.Option("Title"), mdb.TEXT, m.Option("URL"))
			case wiki.IMAGE:
				m.Cmdy(chat.FAVOR, mdb.CREATE, mdb.TYPE, wiki.IMAGE, mdb.NAME, m.Option("Title"), mdb.TEXT, m.Option("URL"))
			case chat.LOCATION:
				m.Cmdy(chat.LOCATION, mdb.CREATE, mdb.TYPE, "", mdb.NAME, m.Option("Label"), mdb.TEXT, m.Option("Label"),
					"latitude", m.Option("LocationX"), "longitude", m.Option("LocationY"), "scale", m.Option("Scale"),
				)
			}
		}},
		LOGIN: {Name: "login", Help: "登录", Actions: ice.Actions{
			mdb.CREATE: {Name: "create appid appmm token", Hand: func(m *ice.Message, arg ...string) { m.Cmd(ACCESS, LOGIN, arg) }},
		}},
	})
}
