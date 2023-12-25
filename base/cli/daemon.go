package cli

import (
	"bytes"
	"io"
	"os/exec"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _daemon_exec(m *ice.Message, cmd *exec.Cmd) {
	if r, ok := m.Optionv(CMD_INPUT).(io.Reader); ok {
		cmd.Stdin = r
	}
	err := bytes.NewBuffer(make([]byte, 0, ice.MOD_BUFS))
	cmd.Stderr = err
	if w := _system_out(m, CMD_OUTPUT); w != nil {
		cmd.Stdout, cmd.Stderr = w, w
	}
	if w := _system_out(m, CMD_ERRPUT); w != nil {
		cmd.Stderr = w
	}
	h := mdb.HashCreate(m.Spawn(), STATUS, START,
		ice.CMD, kit.Join(cmd.Args, lex.SP), DIR, cmd.Dir, ENV, kit.Select("", cmd.Env),
		m.OptionSimple(CMD_INPUT, CMD_OUTPUT, CMD_ERRPUT, mdb.CACHE_CLEAR_ONEXIT),
	)
	if e := cmd.Start(); m.Warn(e, ice.ErrNotStart, cmd.Args, err.String()) {
		mdb.HashModify(m, h, STATUS, ERROR, ERROR, e)
		return
	}
	mdb.HashSelectUpdate(m, h, func(value ice.Map) { value[PID] = cmd.Process.Pid })
	m.Echo("%d", cmd.Process.Pid)
	m.Go(func() {
		if e := cmd.Wait(); !m.Warn(e, ice.ErrNotStart, cmd.Args, err.String()) && cmd.ProcessState != nil && cmd.ProcessState.Success() {
			mdb.HashModify(m, mdb.HASH, h, STATUS, STOP)
			m.Cost(CODE, "0", ctx.ARGS, cmd.Args)
		} else {
			mdb.HashSelectUpdate(m, h, func(value ice.Map) { value[STATUS], value[ERROR] = ERROR, e })
		}
		switch status := mdb.HashSelectField(m.Sleep300ms(), h, STATUS); cb := m.OptionCB("").(type) {
		case func(string) bool:
			kit.If(!cb(status), func() { m.Cmdy(DAEMON, cmd.Path, cmd.Args) })
		case func(string):
			cb(status)
		case func():
			cb()
		case nil:
		default:
			m.ErrorNotImplement(cb)
		}
		kit.For(kit.Simple(CMD_INPUT, CMD_OUTPUT, CMD_ERRPUT), func(p string) { nfs.Close(m, m.Optionv(p)) })
	})
}

const (
	DIR = "dir"
	ENV = "env"
	API = "api"
	MOD = "mod"
	PID = "pid"
	PWD = "pwd"
)
const (
	BUILD = "build"
	ORDER = "order"
	SPAWN = "spawn"
	CHECK = "check"
	BENCH = "bench"
	PPROF = "pprof"

	TIMEOUT  = "timeout"
	STATUS   = "status"
	ERROR    = "error"
	CLEAR    = "clear"
	DELAY    = "delay"
	RELOAD   = "reload"
	RESTART  = "restart"
	INTERVAL = "interval"

	BEGIN = "begin"
	END   = "end"
	START = "start"
	STOP  = "stop"
	OPEN  = "open"
	CLOSE = "close"

	MAIN = "main"
	CODE = "code"
	COST = "cost"
	FROM = "from"
	BACK = "back"
)

const DAEMON = "daemon"

func init() {
	Index.MergeCommands(ice.Commands{
		DAEMON: {Name: "daemon hash auto", Help: "守护进程", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { mdb.HashPrunesValue(m, mdb.CACHE_CLEAR_ONEXIT, ice.TRUE) }},
			START: {Name: "start cmd* dir env", Hand: func(m *ice.Message, arg ...string) {
				m.Options(CMD_DIR, m.Option(DIR), CMD_ENV, kit.Split(m.Option(ENV), " ="))
				_daemon_exec(m, _system_cmd(m, kit.Split(m.Option(ice.CMD))...))
			}},
			RESTART: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy("", STOP).Sleep3s().Cmdy("", START) }},
			STOP: {Hand: func(m *ice.Message, arg ...string) {
				h, pid := m.Option(mdb.HASH), m.Option(PID)
				mdb.HashSelects(m, h).Table(func(value ice.Maps) {
					if h == "" && value[PID] != pid {
						return
					}
					mdb.HashModify(m, mdb.HASH, kit.Select(h, value[mdb.HASH]), STATUS, STOP)
					kit.If(value[PID], func() { m.Cmd(gdb.SIGNAL, gdb.KILL, value[PID]) })
				})
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				h, pid := m.Option(mdb.HASH), m.Option(PID)
				mdb.HashSelects(m, h).Table(func(value ice.Maps) {
					if h == "" && value[PID] != pid {
						return
					}
					mdb.HashRemove(m, kit.Select(h, value[mdb.HASH]))
				})
			}},
		}, mdb.StatusHashAction(mdb.FIELD, "time,hash,status,pid,cmd,dir,env")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Table(func(value ice.Maps) {
				switch value[STATUS] {
				case START:
					m.PushButton(RESTART, STOP)
				default:
					m.PushButton(START, mdb.REMOVE)
				}
			})
			kit.If(len(arg) == 0, func() { m.Action(START, mdb.PRUNES) })
			if len(arg) > 0 && m.Length() == 0 {
				_daemon_exec(m, _system_cmd(m, kit.Simple(kit.Split(arg[0]), arg[1:])...))
				kit.If(IsSuccess(m) && m.Append(CMD_ERR) == "", func() { m.SetAppend() })
			}
		}},
	})
}

func Opens(m *ice.Message, arg ...string) {
	if !tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
		return
	}
	if len(arg) == 0 || arg[0] == "" {
		return
	}
	switch runtime.GOOS {
	case DARWIN:
		if kit.Ext(arg[0]) == "app" {
			m.Cmdy(SYSTEM, OPEN, "-a", arg[0])
		} else {
			m.Cmdy(SYSTEM, OPEN, arg[0])
		}
	case WINDOWS:
		if kit.Ext(arg[0]) == "exe" {
			m.Cmdy(SYSTEM, arg[0])
		} else {
			m.Cmdy(SYSTEM, "explorer", arg[0])
		}
	}
}
func OpenCmds(m *ice.Message, arg ...string) *ice.Message {
	if !tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
		return m
	}
	if len(arg) == 0 || arg[0] == "" {
		return m
	}
	TellApp(m, "Terminal", kit.Format(`do script %s`, strings.Join(arg, "; ")), "activate")
	return m
}
func TellApp(m *ice.Message, app string, arg ...string) {
	OSAScript(m, kit.Format(`
tell application "%s"
	%s
end tell
`, app, strings.Join(arg, lex.NL)))
}
func OSAScript(m *ice.Message, arg ...string) {
	m.Cmd(SYSTEM, "osascript", "-e", arg)
}
