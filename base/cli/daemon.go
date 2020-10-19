package cli

import (
	"io"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"os/exec"
	"strings"
)

func _daemon_show(m *ice.Message, cmd *exec.Cmd, out, err string) {
	if w, ok := m.Optionv(CMD_STDOUT).(io.Writer); ok {
		cmd.Stdout = w
		cmd.Stderr = w
	} else if f, p, e := kit.Create(out); m.Assert(e) {
		m.Log_EXPORT(kit.MDB_META, DAEMON, CMD_STDOUT, p)
		m.Optionv(CMD_STDOUT, f)
		cmd.Stdout = f
		cmd.Stderr = f
	}
	if w, ok := m.Optionv(CMD_STDERR).(io.Writer); ok {
		cmd.Stderr = w
	} else if f, p, e := kit.Create(err); m.Assert(e) {
		m.Log_EXPORT(kit.MDB_META, DAEMON, CMD_STDERR, p)
		m.Optionv(CMD_STDERR, f)
		cmd.Stderr = f
	}

	if e := cmd.Start(); m.Warn(e != nil, ErrStart, cmd.Args, " ", e) {
		return
	}

	h := m.Cmdx(mdb.INSERT, DAEMON, "", mdb.HASH,
		kit.MDB_STATUS, Status.Start, kit.SSH_CMD, strings.Join(cmd.Args, " "),
		kit.SSH_DIR, cmd.Dir, kit.SSH_ENV, kit.Select("", cmd.Env), kit.SSH_PID, cmd.Process.Pid,
		CMD_STDOUT, out, CMD_STDERR, err,
	)
	m.Echo("%d", cmd.Process.Pid)

	m.Go(func() {
		if e := cmd.Wait(); m.Warn(e != nil, ErrStart, cmd.Args, " ", e) {
			m.Cmd(mdb.MODIFY, DAEMON, "", mdb.HASH, kit.MDB_HASH, h,
				kit.MDB_STATUS, Status.Error, kit.MDB_ERROR, e)
		} else {
			m.Cost("args", cmd.Args, "code", cmd.ProcessState.ExitCode())
			m.Cmd(mdb.MODIFY, DAEMON, "", mdb.HASH, kit.MDB_HASH, h,
				kit.MDB_STATUS, Status.Stop)
		}
		if w, ok := m.Optionv(CMD_STDOUT).(io.Closer); ok {
			w.Close()
		}
		if w, ok := m.Optionv(CMD_STDERR).(io.Closer); ok {
			w.Close()
		}
	})
}

const ErrStart = "daemon start: "

var Status = struct{ Error, Start, Stop string }{
	Error: "error",
	Start: "start",
	Stop:  "stop",
}

const (
	ENV = "env"
	CMD = "cmd"
	ARG = "arg"
	DIR = "dir"
)
const (
	RESTART = "restart"
	START   = "start"
	STOP    = "stop"
)

const DAEMON = "daemon"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			DAEMON: {Name: DAEMON, Help: "守护进程", Value: kit.Data(kit.MDB_PATH, "var/daemon")},
		},
		Commands: map[string]*ice.Command{
			DAEMON: {Name: "daemon hash auto 添加 清理", Help: "守护进程", Action: map[string]*ice.Action{
				RESTART: {Name: "restart", Help: "重启", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(DAEMON, STOP)
					m.Sleep("1s")
					m.Cmdy(DAEMON, START)
				}},
				START: {Name: "start cmd env dir", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Option(CMD_TYPE, DAEMON)
					m.Option(CMD_DIR, m.Option("dir"))
					m.Option(CMD_ENV, kit.Split(m.Option("env"), " ="))
					m.Cmdy(SYSTEM, kit.Split(m.Option("cmd")))
				}},
				STOP: {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,hash,status,pid,cmd,dir,env")
					m.Cmd(mdb.SELECT, DAEMON, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH)).Table(func(index int, value map[string]string, head []string) {
						m.Cmd(mdb.MODIFY, DAEMON, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH), kit.MDB_STATUS, Status.Stop)
						m.Debug("what %v", value)
						m.Cmdy(SYSTEM, "kill", "-9", value[kit.SSH_PID])
					})
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, DAEMON, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,hash,status,pid,cmd,dir,env")
					m.Cmdy(mdb.PRUNES, DAEMON, "", mdb.HASH, kit.MDB_STATUS, Status.Error)
					m.Cmdy(mdb.PRUNES, DAEMON, "", mdb.HASH, kit.MDB_STATUS, Status.Stop)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,hash,status,pid,cmd,dir,env")
					m.Cmdy(mdb.SELECT, DAEMON, "", mdb.HASH)
					m.Table(func(index int, value map[string]string, head []string) {
						switch value[kit.MDB_STATUS] {
						case Status.Start:
							m.PushButton(RESTART, STOP)
						default:
							m.PushButton(mdb.REMOVE)
						}
					})

				} else if m.Richs(DAEMON, "", arg[0], nil) != nil {
					m.Option(mdb.FIELDS, mdb.DETAIL)
					m.Cmdy(mdb.SELECT, DAEMON, "", mdb.HASH, kit.MDB_HASH, arg)

				} else {
					m.Option(CMD_TYPE, DAEMON)
					m.Cmdy(SYSTEM, arg)
				}
			}},
		},
	}, nil)
}
