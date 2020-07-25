package cli

import (
	"strings"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"bytes"
	"fmt"
	"os"
	"os/exec"
)

const (
	CMD_STDOUT = "cmd_stdout"
	CMD_STDERR = "cmd_stderr"

	CMD_TYPE = "cmd_type"
	CMD_DIR  = "cmd_dir"
	CMD_ENV  = "cmd_env"

	CMD_OUT  = "cmd_out"
	CMD_ERR  = "cmd_err"
	CMD_CODE = "cmd_code"
)

const ErrRun = "run err "

func _system_show(m *ice.Message, cmd *exec.Cmd) {
	out := bytes.NewBuffer(make([]byte, 0, 1024))
	err := bytes.NewBuffer(make([]byte, 0, 1024))
	cmd.Stdout = out
	cmd.Stderr = err
	defer func() {
		m.Cost("%v exit: %v out: %v err: %v ",
			cmd.Args, cmd.ProcessState.ExitCode(), out.Len(), err.Len())
	}()

	if e := cmd.Run(); e != nil {
		m.Warn(e != nil, ErrRun, strings.Join(cmd.Args, " "), "\n", kit.Select(e.Error(), err.String()))
	}

	m.Push(kit.MDB_TIME, m.Time())
	m.Push(CMD_CODE, int(cmd.ProcessState.ExitCode()))
	m.Push(CMD_ERR, err.String())
	m.Push(CMD_OUT, out.String())
	m.Echo(out.String())
}

const SYSTEM = "system"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SYSTEM: {Name: "system", Help: "系统命令", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			SYSTEM: {Name: "system cmd arg", Help: "系统命令", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
				cmd := exec.Command(arg[0], arg[1:]...)

				// 运行目录
				if cmd.Dir = m.Option(CMD_DIR); len(cmd.Dir) > 0 {
					m.Log_EXPORT(kit.MDB_META, SYSTEM, CMD_DIR, cmd.Dir)
					if _, e := os.Stat(cmd.Dir); e != nil && os.IsNotExist(e) {
						os.MkdirAll(cmd.Dir, ice.MOD_DIR)
					}
				}

				// 环境变量
				env := kit.Simple(m.Optionv(CMD_ENV))
				for i := 0; i < len(env)-1; i += 2 {
					cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", env[i], env[i+1]))
				}
				if len(cmd.Env) > 0 {
					m.Log_EXPORT(kit.MDB_META, SYSTEM, CMD_ENV, cmd.Env)
				}

				switch m.Option(CMD_TYPE) {
				case DAEMON:
					_daemon_show(m, cmd, m.Option(CMD_STDOUT), m.Option(CMD_STDERR))
				default:
					_system_show(m, cmd)
				}
			}},
			"ssh_user": {Name: "ssh_user", Help: "ssh_user", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
				msg := m.Cmd(SYSTEM, "who")
				msg.Split(msg.Result(), "name term begin", " \t", "\n")
				msg.Table(func(index int, value map[string]string, head []string) {
					m.Push("name", value["name"])
					m.Push("term", value["term"])
					ls := strings.Split(value["begin"], " (")
					t, _ := time.Parse("Jan 2 15:04", ls[0])
					m.Push("begin", t.Format("2006-01-02 15:04:05"))

					m.Push("ip", value["ip"])
					m.Push("duration", value["duration"])
				})
			}},
		},
	}, nil)
}
