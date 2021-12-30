package cli

import (
	"io"
	"os/exec"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
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
	if e := cmd.Start(); m.Warn(e, ice.ErrNotStart, cmd.Args) {
		return // 启动失败
	}
	m.Echo("%d", cmd.Process.Pid)

	m.Go(func() {
		h := m.Cmdx(mdb.INSERT, DAEMON, "", mdb.HASH,
			STATUS, START, ice.CMD, kit.Join(cmd.Args, ice.SP),
			PID, cmd.Process.Pid, DIR, cmd.Dir, ENV, kit.Select("", cmd.Env),
			m.OptionSimple(CMD_OUTPUT, CMD_ERRPUT, mdb.CACHE_CLEAR_ON_EXIT),
		)

		if e := cmd.Wait(); m.Warn(e, ice.ErrNotStart, cmd.Args) {
			if m.Conf(DAEMON, kit.Keys(mdb.HASH, h, kit.Keym(STATUS))) == START {
				m.Cmd(mdb.MODIFY, DAEMON, "", mdb.HASH, mdb.HASH, h, STATUS, ERROR, ERROR, e)
			}
		} else {
			m.Cost(kit.MDB_CODE, cmd.ProcessState.ExitCode(), kit.MDB_ARGS, cmd.Args)
			m.Cmd(mdb.MODIFY, DAEMON, "", mdb.HASH, mdb.HASH, h, STATUS, STOP)
		}

		switch cb := m.Optionv(kit.Keycb(DAEMON)).(type) {
		case func(string):
			m.Sleep300ms()
			cb(m.Conf(DAEMON, kit.Keys(mdb.HASH, h, kit.Keym(STATUS))))
		case func():
			m.Sleep300ms()
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

	STATUS  = "status"
	ERROR   = "error"
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
			nfs.PATH, ice.USR_LOCAL_DAEMON, mdb.FIELD, "time,hash,status,pid,cmd,dir,env",
		)},
	}, Commands: map[string]*ice.Command{
		DAEMON: {Name: "daemon hash auto start prunes", Help: "守护进程", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.PRUNES, DAEMON, "", mdb.HASH, mdb.CACHE_CLEAR_ON_EXIT, ice.TRUE)
			}},
			mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.OptionFields(m.Config(mdb.FIELD))
				m.Cmdy(mdb.PRUNES, DAEMON, "", mdb.HASH, STATUS, STOP)
				m.Cmdy(mdb.PRUNES, DAEMON, "", mdb.HASH, STATUS, ERROR)
			}},
			START: {Name: "start cmd env dir", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Option(CMD_DIR, m.Option(DIR))
				m.Option(CMD_ENV, kit.Split(m.Option(ENV), " ="))
				m.Cmdy(DAEMON, kit.Split(m.Option(ice.CMD)))
			}},
			RESTART: {Name: "restart", Help: "重启", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(DAEMON, STOP)
				m.Sleep3s()
				m.Cmdy(DAEMON, START)
			}},
			STOP: {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
				m.OptionFields(m.Config(mdb.FIELD))
				m.Cmd(mdb.SELECT, DAEMON, "", mdb.HASH, m.OptionSimple(mdb.HASH)).Table(func(index int, value map[string]string, head []string) {
					m.Cmd(mdb.MODIFY, DAEMON, "", mdb.HASH, m.OptionSimple(mdb.HASH), STATUS, STOP)
					m.Cmdy(SYSTEM, "kill", "-9", value[PID])
				})
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...).Table(func(index int, value map[string]string, head []string) {
				switch value[STATUS] {
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
