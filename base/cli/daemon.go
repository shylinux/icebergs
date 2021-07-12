package cli

import (
	"io"
	"os/exec"
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _daemon_show(m *ice.Message, cmd *exec.Cmd, out, err string) {
	if w, ok := m.Optionv(CMD_OUTPUT).(io.Writer); ok {
		cmd.Stdout = w
		cmd.Stderr = w
	} else if f, p, e := kit.Create(out); m.Assert(e) {
		m.Log_EXPORT(kit.MDB_META, DAEMON, CMD_OUTPUT, p)
		m.Optionv(CMD_OUTPUT, f)
		cmd.Stdout = f
		cmd.Stderr = f
	}
	if w, ok := m.Optionv(CMD_ERRPUT).(io.Writer); ok {
		cmd.Stderr = w
	} else if f, p, e := kit.Create(err); m.Assert(e) {
		m.Log_EXPORT(kit.MDB_META, DAEMON, CMD_ERRPUT, p)
		m.Optionv(CMD_ERRPUT, f)
		cmd.Stderr = f
	}

	// 启动进程
	if e := cmd.Start(); m.Warn(e != nil, cmd.Args, " ", e) {
		return
	}
	m.Echo("%d", cmd.Process.Pid)

	m.Go(func() {
		h := m.Cmdx(mdb.INSERT, DAEMON, "", mdb.HASH,
			kit.MDB_STATUS, START, kit.SSH_PID, cmd.Process.Pid,
			kit.SSH_CMD, strings.Join(cmd.Args, " "),
			kit.SSH_DIR, cmd.Dir, kit.SSH_ENV, kit.Select("", cmd.Env),
			mdb.CACHE_CLEAR_ON_EXIT, m.Option(mdb.CACHE_CLEAR_ON_EXIT),
			CMD_OUTPUT, out, CMD_ERRPUT, err,
		)

		if e := cmd.Wait(); m.Warn(e != nil, cmd.Args, " ", e) {
			m.Cmd(mdb.MODIFY, DAEMON, "", mdb.HASH, kit.MDB_HASH, h,
				kit.MDB_STATUS, ERROR, kit.MDB_ERROR, e)
		} else {
			m.Cost("args", cmd.Args, "code", cmd.ProcessState.ExitCode())
			m.Cmd(mdb.MODIFY, DAEMON, "", mdb.HASH, kit.MDB_HASH, h, kit.MDB_STATUS, STOP)
		}

		switch cb := m.Optionv(kit.Keycb(DAEMON)).(type) {
		case func():
			cb()
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
	PID = "pid"
	DIR = "dir"
	ENV = "env"
	CMD = "cmd"
	API = "api"
	ARG = "arg"
	RUN = "run"
	RES = "res"
	ERR = "err"
)
const (
	ERROR = "error"
	BUILD = "build"
	SPAWN = "spawn"
	CHECK = "check"
	BENCH = "bench"
	PPROF = "pprof"

	START   = "start"
	RESTART = "restart"
	RELOAD  = "reload"
	STOP    = "stop"
)

const DAEMON = "daemon"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			DAEMON: {Name: DAEMON, Help: "守护进程", Value: kit.Data(kit.MDB_PATH, path.Join(ice.USR_LOCAL, DAEMON))},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.PRUNES, DAEMON, "", mdb.HASH, mdb.CACHE_CLEAR_ON_EXIT, ice.TRUE)
			}},

			DAEMON: {Name: "daemon hash auto start prunes", Help: "守护进程", Action: map[string]*ice.Action{
				START: {Name: "start cmd env dir", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Option(CMD_DIR, m.Option(DIR))
					m.Option(CMD_ENV, kit.Split(m.Option(ENV), " ="))
					m.Cmdy(DAEMON, kit.Split(m.Option(CMD)))
				}},
				RESTART: {Name: "restart", Help: "重启", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(DAEMON, STOP)
					m.Sleep("3s")
					m.Cmdy(DAEMON, START)
				}},
				STOP: {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,hash,status,pid,cmd,dir,env")
					m.Cmd(mdb.SELECT, DAEMON, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH)).Table(func(index int, value map[string]string, head []string) {
						m.Cmd(mdb.MODIFY, DAEMON, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH), kit.MDB_STATUS, STOP)
						m.Cmdy(SYSTEM, "kill", "-9", value[kit.SSH_PID])
					})
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, DAEMON, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,hash,status,pid,cmd,dir,env")
					m.Cmdy(mdb.PRUNES, DAEMON, "", mdb.HASH, kit.MDB_STATUS, ERROR)
					m.Cmdy(mdb.PRUNES, DAEMON, "", mdb.HASH, kit.MDB_STATUS, STOP)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 { // 进程列表
					m.Fields(len(arg) == 0, "time,hash,status,pid,cmd,dir,env")
					m.Cmdy(mdb.SELECT, DAEMON, "", mdb.HASH).Table(func(index int, value map[string]string, head []string) {
						switch value[kit.MDB_STATUS] {
						case START:
							m.PushButton(RESTART, STOP)
						default:
							m.PushButton(mdb.REMOVE)
						}
					})

				} else if m.Richs(DAEMON, "", arg[0], nil) != nil { // 进程详情
					m.Option(mdb.FIELDS, mdb.DETAIL)
					m.Cmdy(mdb.SELECT, DAEMON, "", mdb.HASH, kit.MDB_HASH, arg)

				} else { // 启动进程
					m.Option(CMD_TYPE, DAEMON)
					m.Cmdy(SYSTEM, arg)
				}
			}},
		},
	})
}
