package cli

import (
	"io"
	"os/exec"
	"runtime"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _daemon_exec(m *ice.Message, cmd *exec.Cmd) {
	if r, ok := m.Optionv(CMD_INPUT).(io.Reader); ok {
		cmd.Stdin = r
	}
	if w := _system_out(m, CMD_OUTPUT); w != nil {
		cmd.Stdout, cmd.Stderr = w, w
	}
	if w := _system_out(m, CMD_ERRPUT); w != nil {
		cmd.Stderr = w
	}
	h := mdb.HashCreate(m.Spawn(), STATUS, START,
		ice.CMD, kit.Join(cmd.Args, ice.SP), DIR, cmd.Dir, ENV, kit.Select("", cmd.Env),
		m.OptionSimple(CMD_INPUT, CMD_OUTPUT, CMD_ERRPUT, mdb.CACHE_CLEAR_ON_EXIT),
	)
	if e := cmd.Start(); m.Warn(e, ice.ErrNotStart, cmd.Args) {
		mdb.HashModify(m, h, STATUS, ERROR, ERROR, e)
		return
	}
	mdb.HashSelectUpdate(m, h, func(value ice.Map) { value[PID] = cmd.Process.Pid })
	m.Echo("%d", cmd.Process.Pid)
	m.Go(func() {
		if e := cmd.Wait(); !m.Warn(e, ice.ErrNotStart, cmd.Args) && cmd.ProcessState != nil && cmd.ProcessState.Success() {
			mdb.HashModify(m, mdb.HASH, h, STATUS, STOP)
			m.Cost(CODE, "0", ctx.ARGS, cmd.Args)
		} else {
			mdb.HashSelectUpdate(m, h, func(value ice.Map) { kit.If(value[STATUS] == START, func() { value[STATUS], value[ERROR] = ERROR, e }) })
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
		for _, p := range kit.Simple(CMD_INPUT, CMD_OUTPUT, CMD_ERRPUT) {
			nfs.CloseFile(m, m.Optionv(p))
		}
	})
}

const (
	DIR = "dir"
	ENV = "env"
	API = "api"
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

	TIMEOUT = "timeout"
	STATUS  = "status"
	ERROR   = "error"
	CLEAR   = "clear"
	RELOAD  = "reload"
	RESTART = "restart"

	DELAY = "delay"
	BEGIN = "begin"
	START = "start"
	OPEN  = "open"
	CLOSE = "close"
	STOP  = "stop"
	END   = "end"

	MAIN = "main"
	CODE = "code"
	COST = "cost"
	BACK = "back"
	FROM = "from"
)

const DAEMON = "daemon"

func init() {
	Index.MergeCommands(ice.Commands{
		DAEMON: {Name: "daemon hash auto", Help: "守护进程", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { mdb.HashPrunesValue(m, mdb.CACHE_CLEAR_ON_EXIT, ice.TRUE) }},
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
					mdb.HashModify(m, mdb.HASH, value[mdb.HASH], STATUS, STOP)
					m.Cmd(gdb.SIGNAL, gdb.KILL, value[PID])
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
	switch runtime.GOOS {
	case DARWIN:
		if kit.Ext(arg[0]) == "app" {
			m.Cmd(SYSTEM, OPEN, "-a", arg[0])
		} else {
			m.Cmd(SYSTEM, OPEN, arg[0])
		}
	case WINDOWS:
		if kit.Ext(arg[0]) == "exe" {
			m.Cmd(SYSTEM, arg[0])
		} else {
			m.Cmd(SYSTEM, "explorer", arg[0])
		}
	}
}
