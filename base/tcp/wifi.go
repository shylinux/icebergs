package tcp

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
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
		CONNECT      = "connect"
	)
	Index.MergeCommands(ice.Commands{
		WIFI: {Help: "无线", Icon: "AirPort Utility.png", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case SSID:
					kit.For(kit.Slice(kit.SplitLine(m.System(NETWORKSETUP, "-listpreferredwirelessnetworks", "en0").Result()), 1), func(name string) {
						m.Push(arg[0], strings.TrimSpace(name))
					})
				}
			}},
			CONNECT: {Help: "连接", Hand: func(m *ice.Message, arg ...string) {
				m.ToastProcess()
				msg := mdb.HashSelect(m.Spawn(), m.Option(SSID, strings.TrimSpace(m.Option(SSID))))
				if res := m.System(NETWORKSETUP, "-setairportnetwork", "en0", kit.Select(m.Option(SSID), msg.Append(SSID)), msg.Append(aaa.PASSWORD)); res.Result() != "" {
					m.Echo(res.Result()).ToastFailure(res.Result())
				} else {
					m.ProcessHold()
				}
			}},
		}, mdb.HashAction(mdb.SHORT, SSID, mdb.FIELD, "time,ssid,password")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...).PushAction(CONNECT, mdb.REMOVE); len(arg) > 0 {
				m.EchoQRCode(kit.Format("WIFI:T:WPA;S:%s;P:%s;H:false;;", m.Append(SSID), m.Append(aaa.PASSWORD)))
			}
		}},
	})
}
