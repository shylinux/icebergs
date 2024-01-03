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
	"shylinux.com/x/icebergs/base/web/html"
	"shylinux.com/x/icebergs/core/chat"
	"shylinux.com/x/icebergs/core/chat/location"
	kit "shylinux.com/x/toolkits"
)

func _wx_sign(m *ice.Message, nonce, stamp string) string {
	return kit.Format(sha1.Sum([]byte(kit.Join(kit.Sort(kit.Simple(
		kit.Format("jsapi_ticket=%s", m.Cmdx(ACCESS, TICKET)),
		kit.Format("url=%s", m.R.Header.Get(html.Referer)),
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
				kit.If(strings.Index(m.Option(ice.MSG_USERUA), "MicroMessenger") > -1, func() { m.Option(mdb.PLUGIN, m.PrefixKey()) })
			}},
			"getLocation": {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(location.LOCATION, mdb.CREATE, arg) }},
			"scanQRCode1": {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(chat.FAVOR, mdb.CREATE, arg) }},
		}, gdb.EventsAction(chat.HEADER_AGENT), ctx.ConfAction(nfs.SCRIPT, "https://res.wx.qq.com/open/js/jweixin-1.6.0.js")), Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(ACCESS, AGENT).Options(SIGNATURE, _wx_sign(m, m.Option(NONCESTR, ice.Info.Pathname), m.Option(TIMESTAMP, kit.Format(time.Now().Unix())))).Display("")
			ctx.OptionFromConfig(m, nfs.SCRIPT)
		}},
	})
}
