package code

import (
	"encoding/base64"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const XTERM = "xterm"

func init() {
	cache := sync.Map{}
	add := func(m *ice.Message, key string) string {
		cmd := exec.Command("/bin/sh")
		cmd.Env = append(os.Environ(), "TERM=xterm")
		tty, err := pty.Start(cmd)
		m.Assert(err)
		m.Go(func() {
			buf := make([]byte, 1024)
			for {
				if n, e := tty.Read(buf); m.Assert(e) {
					m.PushNotice("grow", base64.StdEncoding.EncodeToString(buf[:n]))
				}
			}
		})

		cache.Store(key, tty)
		return key
	}
	Index.MergeCommands(ice.Commands{
		XTERM: {Name: "xterm auto", Help: "终端", Actions: ice.MergeAction(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Watch(web.SPACE_STOP, m.PrefixKey())
			}},
			web.SPACE_STOP: {Name: "space.stop", Help: "断开连接", Hand: func(m *ice.Message, arg ...string) {

			}},
			"input": {Name: "input", Help: "输入", Hand: func(m *ice.Message, arg ...string) {
				if w, ok := cache.Load(m.Option("channel")); ok {
					if w, ok := w.(*os.File); ok {
						w.Write([]byte(arg[0]))
						return
					}
				}
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			m.DisplayLocal("", "channel", add(m, kit.Hashs(mdb.UNIQ)))
		}},
	})
}
