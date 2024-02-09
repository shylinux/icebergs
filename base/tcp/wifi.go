package tcp

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	SSID = "ssid"
)
const WIFI = "wifi"

func init() {
	const (
		NETWORKSETUP = "networksetup"
		DISCOVER     = "discover"
		CONNECT      = "connect"
	)
	Index.MergeCommands(ice.Commands{
		WIFI: {Help: "无线", Actions: ice.MergeActions(ice.Actions{
			DISCOVER: {Help: "查找", Hand: func(m *ice.Message, arg ...string) {
				m.Push(SSID, strings.Split(m.System(NETWORKSETUP, "-listpreferredwirelessnetworks", "en0").Result(), lex.NL)[1:])
				m.PushAction(CONNECT)
			}},
			CONNECT: {Help: "连接", Hand: func(m *ice.Message, arg ...string) {
				msg := mdb.HashSelect(m.Spawn(), m.Option(SSID, strings.TrimSpace(m.Option(SSID))))
				m.System(NETWORKSETUP, "-setairportnetwork", "en0", kit.Select(m.Option(SSID), msg.Append(SSID)), msg.Append(aaa.PASSWORD))
				m.ProcessHold()
			}},
		}, mdb.HashAction(mdb.SHORT, SSID, mdb.FIELD, "time,ssid,password")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...).PushAction(CONNECT, mdb.REMOVE).Action(mdb.CREATE, DISCOVER); len(arg) > 0 {
				m.EchoQRCode(kit.Format("WIFI:T:WPA;S:%s;P:%s;H:false;;", m.Append(SSID), m.Append(aaa.PASSWORD)))
			}
		}},
	})
}
