package wx

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	"shylinux.com/x/icebergs/core/chat/location"
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
		EventKey     string
		Content      string
		Location_X   float64
		Location_Y   float64
		Scale        string
		Label        string
		Title        string
		Description  string
		MediaId      int64
		PicUrl       string
		Url          string
		ScanCodeInfo struct {
			ScanType   string
			ScanResult string
		}
	}{}
	defer m.R.Body.Close()
	buf, _ := ioutil.ReadAll(m.R.Body)
	xml.NewDecoder(bytes.NewBuffer(buf)).Decode(&data)
	m.Option("debug", "true")
	m.Debug("buf: %+v", string(buf))
	m.Debug("data: %+v", kit.Formats(data))
	m.Option(ACCESS, data.ToUserName)
	m.Option("CreateTime", data.CreateTime)
	m.Option(aaa.USERNAME, data.FromUserName)
	m.Option(mdb.TYPE, data.MsgType)
	m.Option(mdb.ID, data.MsgId)
	m.Option("Event", data.Event)
	m.Option("EventKey", data.EventKey)
	m.Option(mdb.TEXT, data.Content)
	m.Option(web.LINK, kit.Select(data.Url, data.PicUrl))
	m.Option(location.LATITUDE, kit.Format("%0.6f", data.Location_X))
	m.Option(location.LONGITUDE, kit.Format("%0.6f", data.Location_Y))
	m.Option(location.SCALE, data.Scale)
	m.Option("Label", data.Label)
	m.Option("Title", data.Title)
	m.Option("MediaId", data.MediaId)
	m.Option("ScanResult", data.ScanCodeInfo.ScanResult)
}

const LOGIN = "login"

func init() {
	const (
		AUTH_CODE = "/sns/jscode2session?grant_type=authorization_code"
	)
	if false {
		web.Index.MergeCommands(ice.Commands{
			"/MP_verify_0xp0zkW3fIzIq2Bo.txt": {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) { m.RenderResult("0xp0zkW3fIzIq2Bo") }},
		})
	}
	Index.MergeCommands(ice.Commands{
		web.PP(LOGIN): {Actions: ice.MergeActions(ice.Actions{
			aaa.SESS: {Name: "sess code", Help: "会话", Hand: func(m *ice.Message, arg ...string) {
				m.Option(ice.MSG_USERZONE, WX)
				if mdb.Conf(m, "header", "meta.demo") == ice.TRUE {
					m.Echo(aaa.SessCreate(m.Spawn(), ice.Info.Username))
					return
				}
				appid := kit.Select(m.Option(APPID), kit.Split(kit.ParseURL(m.Option(ice.MSG_REFERER)).Path, nfs.PS), 0)
				m.Cmd(ACCESS).Table(func(value ice.Maps) {
					kit.If(value[APPID] == appid, func() {
						msg := m.Cmd(web.SPIDE, WX, http.MethodGet, AUTH_CODE, "js_code", m.Option(cli.CODE), APPID, value[APPID], SECRET, value[SECRET])
						m.Warn(msg.Append(OPENID) == "", msg.Append("errmsg"))
						m.Echo(aaa.SessCreate(msg, msg.Append(OPENID)))
					})
				})
			}},
			aaa.USER: {Help: "用户", Hand: func(m *ice.Message, arg ...string) {
				if m.WarnNotLogin(m.Option(ice.MSG_USERNAME) == "") {
					return
				}
				if m.Cmd(aaa.USER, m.Option(aaa.USERNAME, m.Option(ice.MSG_USERNAME))).Length() == 0 {
					m.Cmd(aaa.USER, mdb.CREATE, aaa.USERROLE, aaa.VOID, m.OptionSimple(aaa.USERNAME))
				}
				m.Cmd(aaa.USER, mdb.MODIFY, m.OptionSimple(aaa.USERNAME),
					aaa.USERNICK, m.Option(aaa.USERNICK), aaa.AVATAR, m.Option(aaa.AVATAR),
					aaa.GENDER, kit.Select(kit.Select("", "女", m.Option(aaa.GENDER) == "2"), "男", m.Option(aaa.GENDER) == "1"),
					m.OptionSimple(aaa.LANGUAGE, aaa.CITY, aaa.COUNTRY, aaa.PROVINCE), aaa.USERZONE, WX,
				)
			}},
			SCENE: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(IDE, m.Option(SCENE)) }},
			SCAN:  {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(web.CHAT_FAVOR, mdb.CREATE, mdb.TYPE, "", arg) }},
		}, aaa.WhiteAction("", aaa.SESS, aaa.USER, SCENE)), Hand: func(m *ice.Message, arg ...string) {
			if m.Cmdx(ACCESS, aaa.CHECK) == "" {
				return
			} else if m.Option("echostr") != "" {
				m.RenderResult(m.Option("echostr"))
				return
			}
			_wx_parse(m)
			aaa.SessAuth(m.Options(ice.MSG_USERZONE, WX), kit.Dict(aaa.USERNAME, m.Option(aaa.USERNAME), aaa.USERROLE, aaa.UserRole(m, m.Option(aaa.USERNAME))))
			switch m.Option(mdb.TYPE) {
			case gdb.EVENT:
				m.Cmdy(EVENTS, strings.ToLower(m.Option("Event")), kit.Split(m.Option("EventKey")))
			case location.LOCATION:
				m.Cmdy(location.LOCATION, mdb.CREATE, mdb.TEXT, m.Option("Label"), m.OptionSimple(location.LONGITUDE, location.LATITUDE, location.SCALE))
			case TEXT:
				if cmds := kit.Split(m.Option(mdb.TEXT)); aaa.Right(m, cmds) {
					m.Cmdy(TEXT, ctx.CMDS, cmds)
					break
				}
				fallthrough
			default:
				m.Cmdy(chat.FAVOR, mdb.CREATE, mdb.TYPE, m.Option(mdb.TYPE), mdb.NAME, m.Option("Title"), mdb.TEXT, kit.Select(m.Option(mdb.TEXT), m.Option(web.LINK)))
			}
		}},
		LOGIN: {Name: "login list", Help: "登录", Actions: ice.Actions{
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) { m.Cmd(ACCESS, mdb.CREATE, arg) }},
		}, Hand: func(m *ice.Message, arg ...string) { m.Cmdy(ACCESS) }},
	})
}
