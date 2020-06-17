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
	if e := cmd.Start(); m.Warn(e != nil, "%v start: %s", cmd.Args, e) {
		return
	}

	h := m.Rich(DAEMON, nil, kit.Dict(
		kit.MDB_TYPE, ice.TYPE_SHELL, kit.MDB_NAME, cmd.Process.Pid, kit.MDB_TEXT, strings.Join(cmd.Args, " "),
		kit.MDB_EXTRA, kit.Dict(
			kit.MDB_STATUS, StatusStart,
			CMD_STDOUT, out,
			CMD_STDERR, err,
		),
	))
	m.Log_EXPORT(kit.MDB_META, DAEMON, kit.MDB_KEY, h, kit.MDB_PID, cmd.Process.Pid)
	m.Echo("%d", cmd.Process.Pid)

	m.Gos(m, func(m *ice.Message) {
		defer m.Cost("%v exit: %v", cmd.Args, 0)
		if e := cmd.Wait(); e != nil {
			m.Warn(e != nil, "%v wait: %s", cmd.Args, e)
			m.Richs(DAEMON, nil, h, func(key string, value map[string]interface{}) {
				kit.Value(value, kit.Keys(kit.MDB_EXTRA, kit.MDB_STATUS), StatusError)
				kit.Value(value, kit.Keys(kit.MDB_EXTRA, kit.MDB_ERROR), e)
			})
		} else {
			m.Richs(DAEMON, nil, h, func(key string, value map[string]interface{}) {
				kit.Value(value, kit.Keys(kit.MDB_EXTRA, kit.MDB_STATUS), StatusClose)
			})
		}
	})
}

func Daemon(m *ice.Message, key string, arg ...string) {
	cmd := exec.Command(key, arg...)
	_daemon_show(m, cmd, m.Option(CMD_STDOUT), m.Option(CMD_STDERR))
}
func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			DAEMON: {Name: "daemon", Help: "守护进程", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			DAEMON: {Name: "daemon cmd arg...", Help: "守护进程", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(CMD_TYPE, DAEMON)
				m.Cmdy(SYSTEM, arg)
			}},
		},
	}, nil)
}
