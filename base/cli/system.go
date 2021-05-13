package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _system_show(m *ice.Message, cmd *exec.Cmd) {
	if r, ok := m.Optionv(CMD_INPUT).(io.Reader); ok {
		cmd.Stdin = r
	} else if r, ok := m.Optionv("input").(io.Reader); ok {
		cmd.Stdin = r
	}

	if w, ok := m.Optionv(CMD_ERRPUT).(io.WriteCloser); ok {
		cmd.Stderr = w
	}
	if w, ok := m.Optionv(CMD_OUTPUT).(io.WriteCloser); ok {
		cmd.Stdout = w

		if e := cmd.Run(); e != nil {
			m.Warn(e != nil, ErrRun, strings.Join(cmd.Args, " "), "\n", e.Error())
		} else {
			m.Cost("args", cmd.Args, "code", cmd.ProcessState.ExitCode())
		}
	} else if w, ok := m.Optionv("output").(io.WriteCloser); ok {
		cmd.Stderr = w
		cmd.Stdout = w

		if e := cmd.Run(); e != nil {
			m.Warn(e != nil, ErrRun, strings.Join(cmd.Args, " "), "\n", e.Error())
		} else {
			m.Cost("args", cmd.Args, "code", cmd.ProcessState.ExitCode())
		}
	} else {
		err := bytes.NewBuffer(make([]byte, 0, 1024))
		out := bytes.NewBuffer(make([]byte, 0, 1024))
		cmd.Stderr = err
		cmd.Stdout = out
		defer func() {
			m.Push(CMD_ERR, err.String())
			m.Push(CMD_OUT, out.String())
			m.Echo(kit.Select(err.String(), out.String()))
		}()

		if e := cmd.Run(); e != nil {
			m.Warn(e != nil, ErrRun, strings.Join(cmd.Args, " "), "\n", kit.Select(e.Error(), err.String()))
		} else {
			m.Cost("args", cmd.Args, "code", cmd.ProcessState.ExitCode(), "err", err.Len(), "out", out.Len())
		}
	}

	m.Push(kit.MDB_TIME, m.Time())
	m.Push(CMD_CODE, int(cmd.ProcessState.ExitCode()))
}

const ErrRun = "cli run err: "

const (
	CMD_INPUT  = "cmd_input"
	CMD_OUTPUT = "cmd_output"
	CMD_ERRPUT = "cmd_errput"

	CMD_STDERR = "cmd_stderr"
	CMD_STDOUT = "cmd_stdout"

	CMD_TYPE = "cmd_type"
	CMD_DIR  = "cmd_dir"
	CMD_ENV  = "cmd_env"

	CMD_OUT  = "cmd_out"
	CMD_ERR  = "cmd_err"
	CMD_CODE = "cmd_code"
)

const (
	LINUX   = "linux"
	DARWIN  = "darwin"
	WINDOWS = "windows"
)
const SYSTEM = "system"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SYSTEM: {Name: SYSTEM, Help: "系统命令", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			SYSTEM: {Name: "system id auto", Help: "系统命令", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {

				if len(arg) == 0 || kit.Int(arg[0]) > 0 {
					m.Option("_control", "_page")
					m.Option(mdb.FIELDS, kit.Select("time,id,cmd,dir,env", mdb.DETAIL, len(arg) > 0))
					m.Cmdy(mdb.SELECT, SYSTEM, "", mdb.LIST, kit.MDB_ID, arg)
					return
				}

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
					_daemon_show(m, cmd, m.Option(CMD_STDOUT), m.Option(CMD_STDERR))
				default:
					_system_show(m, cmd)
				}
			}},
		},
	})
}
