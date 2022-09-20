package code

import (
	"encoding/base64"
	"os"
	"os/exec"
	"path"
	"strings"

	pty "shylinux.com/x/creackpty"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

type _xterm struct {
	*exec.Cmd
	*os.File
}

func (s _xterm) Setsize(rows, cols string) error {
	return pty.Setsize(s.File, &pty.Winsize{Rows: uint16(kit.Int(rows)), Cols: uint16(kit.Int(cols))})
}
func (s _xterm) Write(data string) (int, error) {
	return s.File.Write([]byte(data))
}
func (s _xterm) Close() error {
	return s.Cmd.Process.Kill()
}

func _xterm_get(m *ice.Message, h string) _xterm {
	h = kit.Select(m.Option(mdb.HASH), h)
	m.Assert(h != "")

	t := mdb.HashSelectField(m, m.Option(mdb.HASH, h), mdb.TYPE)
	mdb.HashModify(m, ice.VIEW, m.Option(ice.MSG_DAEMON))
	return mdb.HashTarget(m, h, func() ice.Any {
		ls := kit.Split(kit.Select(nfs.SH, t))
		cmd := exec.Command(cli.SystemFind(m, ls[0]), ls[1:]...)

		tty, err := pty.Start(cmd)
		m.Assert(err)

		m.Go(func() {
			defer mdb.HashSelectUpdate(m, h, func(value ice.Map) { delete(value, mdb.TARGET) })
			defer tty.Close()

			m.Option("log.disable", ice.TRUE)
			buf := make([]byte, ice.MOD_BUFS)
			for {
				if n, e := tty.Read(buf); !m.Warn(e) && e == nil {
					m.Option(ice.MSG_DAEMON, mdb.HashSelectField(m, h, ice.VIEW))
					m.Option(mdb.TEXT, string(buf[:n]))
					web.PushNoticeGrow(m)
				} else {
					break
				}
			}
		})
		return _xterm{cmd, tty}
	}).(_xterm)
}

const XTERM = "xterm"

func init() {
	Index.MergeCommands(ice.Commands{
		XTERM: {Name: "xterm hash auto", Help: "命令行", Actions: ice.MergeActions(ice.Actions{
			ctx.PROCESS: {Name: "process", Help: "响应", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 || arg[0] != ice.RUN {
					arg = []string{m.Cmdx("", mdb.CREATE, arg)}
				}
				ctx.ProcessField(m, "", arg, arg...)
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case mdb.TYPE:
					m.Cmdy(FAVOR, "_xterm").Cut(mdb.TYPE).Push(arg[0], "bash", "sh")
				case mdb.NAME:
					m.Push(arg[0], ice.Info.HostName, path.Base(m.Option(mdb.TYPE)))
				}
			}},
			mdb.CREATE: {Name: "create type=sh name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, mdb.NAME, kit.Split(m.Option(mdb.TYPE))[0], m.OptionSimple(mdb.TYPE, mdb.NAME, mdb.TEXT))
				ctx.ProcessRefresh(m)
			}},
			"resize": {Name: "resize", Help: "大小", Hand: func(m *ice.Message, arg ...string) {
				_xterm_get(m, "").Setsize(m.Option("rows"), m.Option("cols"))
			}},
			"input": {Name: "input", Help: "输入", Hand: func(m *ice.Message, arg ...string) {
				if b, e := base64.StdEncoding.DecodeString(strings.Join(arg, "")); !m.Warn(e) {
					_xterm_get(m, "").Write(string(b))
				}
			}},
			INSTALL: {Name: "install", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				_xterm_get(m, kit.Select("", arg, 0)).Write(m.Cmdx(PUBLISH, ice.CONTEXTS, INSTALL) + ice.NL)
				ctx.ProcessHold(m)
			}},
			web.WEBSITE: {Name: "website", Help: "打开", Hand: func(m *ice.Message, arg ...string) {
				web.ProcessWebsite(m, "", "", m.OptionSimple(mdb.HASH))
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,type,name,text,view", mdb.TOOLS, FAVOR), ctx.ProcessAction()), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				m.PushAction(web.WEBSITE, mdb.REMOVE)
				m.Action(mdb.CREATE, mdb.PRUNES)
			} else {
				m.Action("full", INSTALL)
				ctx.DisplayLocal(m, "")
				ctx.Toolkit(m)
			}
		}},
	})
}
