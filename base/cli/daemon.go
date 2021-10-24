package cli

import (
	"io"
	"os/exec"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _daemon_exec(m *ice.Message, cmd *exec.Cmd) {
	if w := _system_out(m, CMD_OUTPUT); w != nil {
		cmd.Stdout = w
		cmd.Stderr = w
	}
	if w := _system_out(m, CMD_ERRPUT); w != nil {
		cmd.Stderr = w
	}

	// 启动进程
	if e := cmd.Start(); m.Warn(e != nil, cmd.Args, ice.SP, e) {
		return // 启动失败
	}
	m.Echo("%d", cmd.Process.Pid)

	m.Go(func() {
		h := m.Cmdx(mdb.INSERT, DAEMON, "", mdb.HASH,
			kit.MDB_STATUS, START, ice.CMD, kit.Join(cmd.Args, ice.SP),
			PID, cmd.Process.Pid, DIR, cmd.Dir, ENV, kit.Select("", cmd.Env),
			m.OptionSimple(CMD_OUTPUT, CMD_ERRPUT, mdb.CACHE_CLEAR_ON_EXIT),
		)

		if e := cmd.Wait(); m.Warn(e != nil, cmd.Args, ice.SP, e) {
			if m.Conf(DAEMON, kit.Keys(kit.MDB_HASH, h, kit.Keym(kit.MDB_STATUS))) == START {
				m.Cmd(mdb.MODIFY, DAEMON, "", mdb.HASH, kit.MDB_HASH, h, kit.MDB_STATUS, ERROR, ERROR, e)
			}
		} else {
			m.Cost(kit.MDB_CODE, cmd.ProcessState.ExitCode(), kit.MDB_ARGS, cmd.Args)
			m.Cmd(mdb.MODIFY, DAEMON, "", mdb.HASH, kit.MDB_HASH, h, kit.MDB_STATUS, STOP)
		}

		switch cb := m.Optionv(kit.Keycb(DAEMON)).(type) {
		case func(string):
			m.Sleep("1s")
			cb(m.Conf(DAEMON, kit.Keys(kit.MDB_HASH, h, kit.Keym(kit.MDB_STATUS))))
		case func():
			m.Sleep("1s")
			cb()
		}

		if w, ok := m.Optionv(CMD_INPUT).(io.Closer); ok {
			w.Close()
		}
		if w, ok := m.Optionv(CMD_OUTPUT).(io.Closer); ok {
			w.Close()
		}
		if w, ok := m.Optionv(CMD_ERRPUT).(io.Closer); ok {
			w.Close()
		}
	})
}

func Inputs(m *ice.Message, field string) bool {
	switch strings.TrimPrefix(field, "extra.") {
	case ice.POD:
		m.Cmdy("route")
	case ice.CTX:
		m.Cmdy(ctx.CONTEXT)
	case ice.CMD:
		m.Cmdy(ctx.CONTEXT, kit.Select(m.Option(ice.CTX), m.Option(kit.Keys(kit.MDB_EXTRA, ice.CTX))), ctx.COMMAND)
	case ice.ARG:

	default:
		return false
	}
	return true
}

const (
	DIR = "dir"
	ENV = "env"
	API = "api"
	PID = "pid"
	PWD = "pwd"
)
const (
	ERROR = "error"
	BUILD = "build"
	ORDER = "order"
	SPAWN = "spawn"
	CHECK = "check"
	BENCH = "bench"
	PPROF = "pprof"

	START   = "start"
	RESTART = "restart"
	RELOAD  = "reload"
	STOP    = "stop"

	OPEN  = "open"
	CLOSE = "close"
	BEGIN = "begin"
	END   = "end"
)

const DAEMON = "daemon"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		DAEMON: {Name: DAEMON, Help: "守护进程", Value: kit.Data(
			kit.MDB_PATH, ice.USR_LOCAL_DAEMON, kit.MDB_FIELD, "time,hash,status,pid,cmd,dir,env",
		)},
	}, Commands: map[string]*ice.Command{
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(mdb.PRUNES, DAEMON, "", mdb.HASH, mdb.CACHE_CLEAR_ON_EXIT, ice.TRUE)
		}},

		DAEMON: {Name: "daemon hash auto start prunes", Help: "守护进程", Action: ice.MergeAction(map[string]*ice.Action{
			START: {Name: "start cmd env dir", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Option(CMD_DIR, m.Option(DIR))
				m.Option(CMD_ENV, kit.Split(m.Option(ENV), " ="))
				m.Cmdy(DAEMON, kit.Split(m.Option(ice.CMD)))
			}},
			RESTART: {Name: "restart", Help: "重启", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(DAEMON, STOP)
				m.Sleep("3s")
				m.Cmdy(DAEMON, START)
			}},
			STOP: {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
				m.OptionFields(m.Config(kit.MDB_FIELD))
				m.Cmd(mdb.SELECT, DAEMON, "", mdb.HASH, m.OptionSimple(kit.MDB_HASH)).Table(func(index int, value map[string]string, head []string) {
					m.Cmd(mdb.MODIFY, DAEMON, "", mdb.HASH, m.OptionSimple(kit.MDB_HASH), kit.MDB_STATUS, STOP)
					m.Cmdy(SYSTEM, "kill", "-9", value[PID])
				})
			}},
			mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.OptionFields(m.Config(kit.MDB_FIELD))
				m.Cmdy(mdb.PRUNES, DAEMON, "", mdb.HASH, kit.MDB_STATUS, STOP)
				m.Cmdy(mdb.PRUNES, DAEMON, "", mdb.HASH, kit.MDB_STATUS, ERROR)
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...).Table(func(index int, value map[string]string, head []string) {
				switch value[kit.MDB_STATUS] {
				case START:
					m.PushButton(RESTART, STOP)
				default:
					m.PushButton(mdb.REMOVE)
				}
			})

			if len(arg) == 0 || m.Length() > 0 {
				return
			}
			_daemon_exec(m, _system_cmd(m, arg...))
		}},
	}})
}
