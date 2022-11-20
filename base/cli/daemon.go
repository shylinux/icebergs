package cli

import (
	"io"
	"os/exec"
	"strings"

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
	h := mdb.HashCreate(m.Spawn(), ice.CMD, kit.Join(cmd.Args, ice.SP),
		STATUS, START, DIR, cmd.Dir, ENV, kit.Select("", cmd.Env),
		m.OptionSimple(CMD_INPUT, CMD_OUTPUT, CMD_ERRPUT, mdb.CACHE_CLEAR_ON_EXIT),
	)
	if e := cmd.Start(); m.Warn(e, ice.ErrNotStart, cmd.Args) {
		mdb.HashModify(m, h, STATUS, ERROR, ERROR, e)
		return
	}
	mdb.HashSelectUpdate(m, h, func(value ice.Map) { value[PID] = cmd.Process.Pid })
	m.Echo("%d", cmd.Process.Pid)

	m.Go(func() {
		if e := cmd.Wait(); !m.Warn(e, ice.ErrNotStart, cmd.Args) && cmd.ProcessState.ExitCode() == 0 {
			m.Cost(CODE, cmd.ProcessState.ExitCode(), ctx.ARGS, cmd.Args)
			mdb.HashModify(m, mdb.HASH, h, STATUS, STOP)
		} else {
			mdb.HashSelectUpdate(m, h, func(value ice.Map) {
				if value[STATUS] == START {
					value[STATUS], value[ERROR] = ERROR, e
				}
			})
		}
		status := mdb.HashSelectField(m, h, STATUS)
		switch m.Sleep300ms(); cb := m.OptionCB("").(type) {
		case func(string) bool:
			if !cb(status) {
				m.Cmdy(DAEMON, cmd.Path, cmd.Args)
			}
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
		DAEMON: {Name: "daemon hash auto start prunes", Help: "守护进程", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashPrunesValue(m, mdb.CACHE_CLEAR_ON_EXIT, ice.TRUE)
			}},
			START: {Name: "start cmd dir env", Hand: func(m *ice.Message, arg ...string) {
				m.Options(CMD_DIR, m.Option(DIR), CMD_ENV, kit.Split(m.Option(ENV), " ="))
				_daemon_exec(m, _system_cmd(m, kit.Split(m.Option(ice.CMD))...))
			}},
			RESTART: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy("", STOP).Sleep3s().Cmdy("", START)
			}},
			STOP: {Hand: func(m *ice.Message, arg ...string) {
				m.OptionFields(m.Config(mdb.FIELD))
				h, pid := m.Option(mdb.HASH), m.Option(PID)
				mdb.HashSelect(m, m.Option(mdb.HASH)).Tables(func(value ice.Maps) {
					if h == "" && value[PID] != pid {
						return
					}
					mdb.HashModify(m, mdb.HASH, value[mdb.HASH], STATUS, STOP)
					m.Cmd(gdb.SIGNAL, gdb.KILL, value[PID])
				})
			}},
		}, mdb.HashStatusAction(mdb.FIELD, "time,hash,status,pid,cmd,dir,env")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || !strings.Contains(arg[0], ice.PS) {
				if mdb.HashSelect(m, kit.Slice(arg, 0, 1)...).Tables(func(value ice.Maps) {
					switch value[STATUS] {
					case START:
						m.PushButton(RESTART, STOP)
					default:
						m.PushButton(mdb.REMOVE)
					}
				}); len(arg) == 0 || m.Length() > 0 {
					return
				}
			}

			if len(arg) == 1 {
				arg = kit.Split(arg[0])
			}
			if _daemon_exec(m, _system_cmd(m, arg...)); IsSuccess(m) && m.Append(CMD_ERR) == "" {
				m.SetAppend()
			}
		}},
	})
}
