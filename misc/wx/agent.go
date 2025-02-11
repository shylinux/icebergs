package wx

import (
	"crypto/sha1"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	// "shylinux.com/x/icebergs/base/web/html"
	"shylinux.com/x/icebergs/core/chat"
	"shylinux.com/x/icebergs/core/chat/location"
	kit "shylinux.com/x/toolkits"
)

func _wx_sign(m *ice.Message, nonce, stamp string) string {
	return kit.Format(sha1.Sum([]byte(kit.Join(kit.Sort(kit.Simple(
		kit.Format("jsapi_ticket=%s", m.Cmdx(ACCESS, TICKET)),
		// kit.Format("url=%s", m.R.Header.Get(html.Referer)),
		kit.Format("url=%s", m.Option(ice.MSG_REFERER)),
		kit.Format("timestamp=%s", stamp),
		kit.Format("noncestr=%s", nonce),
	)), "&"))))
}

const (
	SIGNATURE = "signature"
	TIMESTAMP = "timestamp"
	NONCESTR  = "noncestr"
	NONCE     = "nonce"
)
const AGENT = "agent"

func init() {
	Index.MergeCommands(ice.Commands{
		AGENT: {Name: "agent auto", Help: "代理", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			chat.HEADER_AGENT: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ice.MSG_USERNAME) == "" && mdb.Config(m, "oauth") != "" {
					m.ProcessOpen(mdb.Config(m, "oauth"))
				}
				kit.If(strings.Index(m.Option(ice.MSG_USERUA), "MicroMessenger") > -1, func() {
					m.Optionv(mdb.PLUGIN, m.PrefixKey(), mdb.Config(m, web.SPACE))
				})
			}},
			"getLocation": {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(location.LOCATION, mdb.CREATE, arg) }},
			"scanQRCode1": {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(chat.FAVOR, mdb.CREATE, arg) }},
			"oauth":       {Hand: func(m *ice.Message, arg ...string) { mdb.Config(m, "oauth", arg[0]) }},
		}, gdb.EventsAction(chat.HEADER_AGENT), ctx.ConfAction(
			"space", "", "oauth", "", nfs.SCRIPT, "https://res.wx.qq.com/open/js/jweixin-1.6.0.js",
		)), Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(ACCESS, AGENT).Options(SIGNATURE, _wx_sign(m, m.Option(NONCESTR, ice.Info.Pathname), m.Option(TIMESTAMP, kit.Format(time.Now().Unix())))).Display("")
			ctx.OptionFromConfig(m, nfs.SCRIPT, "oauth")
			m.Option("oauth", strings.ReplaceAll(m.Option("oauth"),
				"https%3A%2F%2Fyunxuanlinghang.com", strings.ReplaceAll(m.Option(ice.MSG_USERHOST), "://", "%3A%2F%2F"),
			))
		}},
	})
}
