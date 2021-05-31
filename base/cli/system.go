package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

func _system_show(m *ice.Message, cmd *exec.Cmd) {
	// 输入流
	if r, ok := m.Optionv(CMD_INPUT).(io.Reader); ok {
		cmd.Stdin = r
	}

	// 输出流
	if w, ok := m.Optionv(CMD_OUTPUT).(io.Writer); ok {
		cmd.Stdout, cmd.Stderr = w, w
		if w, ok := m.Optionv(CMD_ERRPUT).(io.Writer); ok {
			cmd.Stderr = w
		}

	} else {
		out := bytes.NewBuffer(make([]byte, 0, ice.MOD_BUFS))
		err := bytes.NewBuffer(make([]byte, 0, ice.MOD_BUFS))
		defer func() {
			m.Push(CMD_ERR, err.String())
			m.Push(CMD_OUT, out.String())
			m.Echo(kit.Select(out.String(), err.String()))
		}()

		cmd.Stdout, cmd.Stderr = out, err
	}

	// 执行命令
	if e := cmd.Run(); e != nil {
		m.Warn(e != nil, cmd.Args, " ", e.Error())
	} else {
		m.Cost("code", cmd.ProcessState.ExitCode(), "args", cmd.Args)
	}

	m.Push(kit.MDB_TIME, m.Time())
	m.Push(CMD_CODE, int(cmd.ProcessState.ExitCode()))
}

const (
	CMD_DIR  = "cmd_dir"
	CMD_ENV  = "cmd_env"
	CMD_TYPE = "cmd_type"

	CMD_INPUT  = "cmd_input"
	CMD_OUTPUT = "cmd_output"
	CMD_ERRPUT = "cmd_errput"

	CMD_CODE = "cmd_code"
	CMD_ERR  = "cmd_err"
	CMD_OUT  = "cmd_out"
)

const (
	LINUX   = "linux"
	DARWIN  = "darwin"
	WINDOWS = "windows"
	SOURCE  = "source"

	USER = "USER"
	HOME = "HOME"
	PATH = "PATH"
)
const SYSTEM = "system"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SYSTEM: {Name: SYSTEM, Help: "系统命令", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			SYSTEM: {Name: "system cmd= 执行:button", Help: "系统命令", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
				if len(arg) == 1 {
					arg = kit.Split(arg[0])
				}
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
					_daemon_show(m, cmd, m.Option(CMD_OUTPUT), m.Option(CMD_ERRPUT))
				default:
					_system_show(m, cmd)
				}
			}},
		},
	})
}
