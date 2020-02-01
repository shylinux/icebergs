package wx

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/core/chat"
	"github.com/shylinux/toolkits"
	"regexp"
)

var Index = &ice.Context{Name: "wx", Help: "wx",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"login": {Name: "login", Help: "认证", Value: kit.Data("wechat", "https://login.weixin.qq.com")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.WEB_SPIDE, "add", "wechat", m.Conf("login", "meta.wechat"))
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		"login": {Name: "login", Help: "认证", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			what := m.Cmdx(ice.WEB_SPIDE, "wechat", "raw", "GET", "/jslogin", "appid", "wx782c26e4c19acffb", "fun", "new")
			// what := `window.QRLogin.code = 200; window.QRLogin.uuid = "gZUohppbCw==";`
			reg, _ := regexp.Compile(`window.QRLogin.code = (\d+); window.QRLogin.uuid = "(\S+?)";`)
			if list := reg.FindStringSubmatch(what); list[1] == "200" {
				m.Richs(ice.WEB_SPIDE, nil, "wechat", func(key string, value map[string]interface{}) {
					if qrcode := kit.Format("%s/l/%s", kit.Value(value, "client.url"), list[2]); m.R == nil {
						m.Cmdy("cli.python", "qrcode", qrcode)
					} else {
						m.Push("_output", "qrcode").Echo(qrcode)
					}

					m.Gos(m, func(m *ice.Message) {
						reg, _ := regexp.Compile(`window.code=(\d+)`)
						for i := 0; i < 1000; i++ {
							what := m.Cmdx(ice.WEB_SPIDE, "wechat", "raw", "GET", "/cgi-bin/mmwebwx-bin/login", "loginicon", "true", "uuid", list[2], "tip", "1", "r", kit.Int(m.Time("stamp"))/1579, "_", m.Time("stamp"))
							// window.code=200; window.redirect_uri="https://wx2.qq.com/cgi-bin/mmwebwx-bin/webwxnewloginpage?ticket=A7_l6ng7wSjNbs7-qD3ArIRJ@qrticket_0&uuid=Ia1-kbZ0wA==&lang=zh_CN&scan=1579005657";
							if list := reg.FindStringSubmatch(what); list[1] == "200" {
								reg, _ := regexp.Compile(`window.redirect_uri="(\S+)";`)
								if list := reg.FindStringSubmatch(what); len(list) > 1 {
									what := m.Cmdx(ice.WEB_SPIDE, "wechat", "raw", "GET", list[1])
									m.Info("what %s", what)
									break
								}
							}
							m.Info("wait scan %v", list)
							m.Sleep("1s")
						}
					})
				})
			}
		}},
	},
}

func init() { chat.Index.Register(Index, nil) }
