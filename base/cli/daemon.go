package cli

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	StatusError = "error"
	StatusStart = "start"
	StatusClose = "close"
)

func _daemon_show(m *ice.Message, cmd *exec.Cmd, out, err string) {
	if f, p, e := kit.Create(out); m.Assert(e) {
		m.Log_EXPORT(kit.MDB_META, DAEMON, CMD_STDOUT, p)
		cmd.Stdout = f
		cmd.Stderr = f
	}
	if f, p, e := kit.Create(err); m.Assert(e) {
		m.Log_EXPORT(kit.MDB_META, DAEMON, CMD_STDERR, p)
		cmd.Stderr = f
	}

	cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", os.Getenv("PATH")))
	if e := cmd.Start(); m.Warn(e != nil, ice.ErrStart, cmd.Args, " ", e) {
		return
	}

	h := m.Rich(DAEMON, nil, kit.Dict(
		kit.MDB_TYPE, "shell", kit.MDB_NAME, strings.Join(cmd.Args, " "),
		kit.MDB_DIR, cmd.Dir, kit.MDB_PID, cmd.Process.Pid, kit.MDB_STATUS, StatusStart,
		kit.MDB_EXTRA, kit.Dict(CMD_STDOUT, out, CMD_STDERR, err),
	))
	m.Log_EXPORT(kit.MDB_META, DAEMON, kit.MDB_KEY, h, kit.MDB_PID, cmd.Process.Pid)
	m.Echo("%d", cmd.Process.Pid)

	m.Gos(m, func(m *ice.Message) {
		if e := cmd.Wait(); e != nil {
			m.Warn(e != nil, ice.ErrStart, cmd.Args, " ", e)
			m.Richs(DAEMON, nil, h, func(key string, value map[string]interface{}) {
				kit.Value(value, kit.MDB_STATUS, StatusError)
				kit.Value(value, kit.Keys(kit.MDB_EXTRA, kit.MDB_ERROR), e)
			})
		} else {
			m.Cost("%v exit: %v", cmd.Args, cmd.ProcessState.ExitCode())
			m.Richs(DAEMON, nil, h, func(key string, value map[string]interface{}) {
				kit.Value(value, kit.MDB_STATUS, StatusClose)
			})
		}
	})
}

const DAEMON = "daemon"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			DAEMON: {Name: DAEMON, Help: "守护进程", Value: kit.Data(kit.MDB_PATH, "var/daemon")},
		},
		Commands: map[string]*ice.Command{
			DAEMON: {Name: "daemon hash 查看:button=auto 清理:button", Help: "守护进程", Action: map[string]*ice.Action{
				"delete": {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(DAEMON, "", m.Option(kit.MDB_HASH), func(key string, value map[string]interface{}) {
						m.Conf(DAEMON, kit.Keys(kit.MDB_HASH, key), "")
						m.Log_DELETE(DAEMON, kit.Format(value))
					})
				}},
				"prune": {Name: "prune", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(DAEMON, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
						if value["status"] == "error" || strings.Count(m.Cmdx(SYSTEM, "ps", value[kit.MDB_PID]), "\n") == 1 {
							m.Conf(DAEMON, kit.Keys(kit.MDB_HASH, key), "")
							m.Log_DELETE(DAEMON, kit.Format(value))
						}
					})
				}},
				"stop": {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(DAEMON, "", m.Option(kit.MDB_HASH), func(key string, value map[string]interface{}) {
						m.Cmdy(SYSTEM, "kill", value[kit.MDB_PID])
						if strings.Count(m.Cmdx(SYSTEM, "ps", value[kit.MDB_PID]), "\n") == 1 {
							value[kit.MDB_STATUS] = StatusClose
						}
					})
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option("fields", "time,hash,status,pid,name,dir")
					m.Cmdy("mdb.select", DAEMON, "", kit.MDB_HASH)
					m.Sort("time", "time_r")
					m.PushAction("停止", "删除")
					return
				}

				m.Option(CMD_TYPE, DAEMON)
				m.Cmdy(SYSTEM, arg)
			}},
		},
	}, nil)
}
