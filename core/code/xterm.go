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
	"shylinux.com/x/icebergs/base/log"
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
func (s _xterm) Writeln(data string, arg ...ice.Any) {
	s.Write(kit.Format(data, arg...) + ice.NL)
}
func (s _xterm) Write(data string) (int, error) {
	return s.File.Write([]byte(data))
}
func (s _xterm) Close() error {
	return s.Cmd.Process.Kill()
}
func _xterm_get(m *ice.Message, h string) _xterm {
	if h = kit.Select(m.Option(mdb.HASH), h); m.Assert(h != "") {
		if mdb.HashSelectField(m, m.Option(mdb.HASH, h), mdb.TYPE) == "" {
			mdb.HashCreate(m, mdb.HASH, h, m.OptionSimple("type,name,text"))
		}
		mdb.HashModify(m, web.VIEW, m.Option(ice.MSG_DAEMON))
	}
	t := mdb.HashSelectField(m, m.Option(mdb.HASH, h), mdb.TYPE)
	return mdb.HashSelectTarget(m, h, func() ice.Any {
		ls := kit.Split(kit.Select(nfs.SH, strings.Split(t, " # ")[0]))
		cmd := exec.Command(cli.SystemFind(m, ls[0]), ls[1:]...)
		cmd.Env = append(cmd.Env, os.Environ()...)
		cmd.Env = append(cmd.Env, "TERM=xterm")
		tty, err := pty.Start(cmd)
		m.Assert(err)
		m.Go(func() {
			defer tty.Close()
			defer mdb.HashRemove(m, mdb.HASH, h)
			m.Option(ice.LOG_DISABLE, ice.TRUE)
			buf := make([]byte, ice.MOD_BUFS)
			for {
				if n, e := tty.Read(buf); !m.Warn(e) && e == nil {
					_xterm_echo(m, h, string(buf[:n]))
				} else {
					_xterm_echo(m, h, "~~~end~~~")
					break
				}
			}
		})
		m.Go(func() { _xterm_cmds(m.Sleep("1s"), h, mdb.HashSelectField(m, h, mdb.TEXT)) })
		return _xterm{cmd, tty}
	}).(_xterm)
}
func _xterm_echo(m *ice.Message, h string, str string) {
	m.Options(ice.MSG_DAEMON, mdb.HashSelectField(m, h, web.VIEW))
	mdb.HashModify(m, mdb.HASH, h, mdb.TIME, m.Time())
	web.PushNoticeGrow(m, h, str)
}
func _xterm_cmds(m *ice.Message, h string, cmd string, arg ...ice.Any) {
	if cmd != "" {
		_xterm_get(m, h).Writeln(cmd, arg...)
	}
	ctx.ProcessHold(m)
}

const XTERM = "xterm"

func init() {
	Index.MergeCommands(ice.Commands{
		XTERM: {Name: "xterm hash auto", Help: "命令行", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case mdb.TYPE:
					if m.Option(nfs.LINE) != "" && m.Option(nfs.FILE) != "" {
						m.Push(arg[0], VIM+" +"+m.Option(nfs.LINE)+ice.SP+m.Option(nfs.PATH)+m.Option(nfs.FILE))
					}
					m.Push(arg[0], BASH, SH)
				case mdb.NAME:
					m.Push(arg[0], path.Base(m.Option(mdb.TYPE)), ice.Info.Hostname)
				}
			}},
			mdb.CREATE: {Name: "create type*=sh name text theme:textarea", Hand: func(m *ice.Message, arg ...string) { mdb.HashCreate(m) }},
			web.RESIZE: {Hand: func(m *ice.Message, arg ...string) {
				_xterm_get(m, "").Setsize(m.OptionDefault("rows", "24"), m.OptionDefault("cols", "80"))
			}},
			web.INPUT: {Hand: func(m *ice.Message, arg ...string) {
				if b, e := base64.StdEncoding.DecodeString(strings.Join(arg, "")); !m.Warn(e) {
					_xterm_get(m, "").Write(string(b))
					mdb.HashModify(m, mdb.TIME, m.Time())
				}
			}},
			INSTALL: {Hand: func(m *ice.Message, arg ...string) {
				_xterm_cmds(m, kit.Select("", arg, 0), m.Cmdx(PUBLISH, ice.CONTEXTS, INSTALL))
			}},
			log.DEBUG: {Hand: func(m *ice.Message, arg ...string) {
				_xterm_cmds(m, kit.Select("", arg, 0), "cd ~/contexts; tail -f var/log/bench.log")
			}},
			web.OUTPUT: {Hand: func(m *ice.Message, arg ...string) {
				web.ProcessWebsite(m, "", "", m.OptionSimple(mdb.HASH), ctx.STYLE, web.OUTPUT)
			}},
			ctx.PROCESS: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, m.PrefixKey(), func() string { return m.Cmdx("", mdb.CREATE, arg) }, arg...)
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,type,name,text,view,theme", mdb.TOOLS, FAVOR), ctx.ProcessAction()), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				m.PushAction(web.OUTPUT, mdb.REMOVE).Action(mdb.CREATE, mdb.PRUNES)
			} else {
				if m.Length() == 0 {
					arg[0] = m.Cmdx("", mdb.CREATE, mdb.TYPE, arg[0])
					mdb.HashSelect(m, arg[0]).Push(mdb.HASH, arg[0])
				}
				m.Push(mdb.HASH, arg[0])
				// m.Action(INSTALL, log.DEBUG)
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}

func _xterm_show(m *ice.Message, cmds, text string, arg ...string) {
	m.Cmdy(ctx.COMMAND, XTERM).Push(ctx.ARGS, kit.Format([]string{m.Cmdx(XTERM, mdb.CREATE, mdb.TYPE, cmds, mdb.NAME, kit.Select("", arg, 0), mdb.TEXT, text)})).ProcessField(XTERM)
}
