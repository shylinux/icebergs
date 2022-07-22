package code

import (
	"io"
	"os"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
)

const XTERM = "xterm"

func init() {
	cache := map[string]io.Writer{}
	cache1 := map[string]io.Writer{}

	Index.MergeCommands(ice.Commands{
		XTERM: {Name: "xterm auto", Help: "终端", Actions: ice.MergeAction(ice.Actions{
			"input": {Name: "input", Help: "输入", Hand: func(m *ice.Message, arg ...string) {
				if w, ok := cache[m.Option("channel")]; ok {
					m.Debug("write %v", []byte(arg[0]))
					w.Write([]byte(arg[0]))
				}
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			r0, w0, e0 := os.Pipe()
			m.Assert(e0)
			r1, w1, e1 := os.Pipe()
			m.Assert(e1)
			m.Option(cli.CMD_INPUT, r0)
			m.Option(cli.CMD_OUTPUT, w1)
			m.Cmd(cli.DAEMON, "bash")
			m.Go(func() {
				buf := make([]byte, 1024)
				for {
					n, e := r1.Read(buf)
					m.Assert(e)
					m.Debug("what %v", string(buf[:n]))
					m.PushNotice("grow", string(buf[:n]))
				}
			})

			cache["1"] = w0
			cache1["1"] = w1
			m.DisplayLocal("", "channel", "1")
		}},
	})
}
