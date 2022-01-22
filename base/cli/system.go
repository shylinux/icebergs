package cli

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _system_cmd(m *ice.Message, arg ...string) *exec.Cmd {
	cmd := exec.Command(arg[0], arg[1:]...)

	// 运行目录
	if cmd.Dir = m.Option(CMD_DIR); len(cmd.Dir) > 0 {
		m.Log_EXPORT(CMD_DIR, cmd.Dir)
		if _, e := os.Stat(cmd.Dir); e != nil && os.IsNotExist(e) {
			os.MkdirAll(cmd.Dir, ice.MOD_DIR)
		}
	}

	// 环境变量
	env := kit.Simple(m.Optionv(CMD_ENV))
	for i := 0; i < len(env)-1; i += 2 {
		cmd.Env = append(cmd.Env, kit.Format("%s=%s", env[i], env[i+1]))
	}
	if len(cmd.Env) > 0 {
		m.Log_EXPORT(CMD_ENV, cmd.Env)
	}
	return cmd
}
func _system_out(m *ice.Message, out string) io.Writer {
	if w, ok := m.Optionv(out).(io.Writer); ok {
		return w
	} else if m.Option(out) == "" {
		return nil
	} else if f, p, e := kit.Create(m.Option(out)); m.Assert(e) {
		m.Log_EXPORT(out, p)
		m.Optionv(out, f)
		return f
	}
	return nil
}
func _system_exec(m *ice.Message, cmd *exec.Cmd) {
	// 输入流
	if r, ok := m.Optionv(CMD_INPUT).(io.Reader); ok {
		cmd.Stdin = r
	}

	// 输出流
	if w := _system_out(m, CMD_OUTPUT); w != nil {
		cmd.Stdout, cmd.Stderr = w, w
		if w := _system_out(m, CMD_ERRPUT); w != nil {
			cmd.Stderr = w
		}
	} else {
		out := bytes.NewBuffer(make([]byte, 0, ice.MOD_BUFS))
		err := bytes.NewBuffer(make([]byte, 0, ice.MOD_BUFS))
		defer func() {
			m.Push(CMD_OUT, out.String())
			m.Push(CMD_ERR, err.String())
			m.Echo(strings.TrimSpace(kit.Select(out.String(), err.String())))
		}()
		cmd.Stdout, cmd.Stderr = out, err
	}

	// 执行命令
	if e := cmd.Run(); !m.Warn(e, ice.ErrNotFound, cmd.Args) {
		m.Cost(CODE, cmd.ProcessState.ExitCode(), ctx.ARGS, cmd.Args)
	}

	m.Push(mdb.TIME, m.Time())
	m.Push(CODE, int(cmd.ProcessState.ExitCode()))
}
func IsSuccess(m *ice.Message) bool {
	return m.Append(CODE) == "0" || m.Append(CODE) == ""
}

const (
	CMD_DIR = "cmd_dir"
	CMD_ENV = "cmd_env"

	CMD_INPUT  = "cmd_input"
	CMD_OUTPUT = "cmd_output"
	CMD_ERRPUT = "cmd_errput"

	CMD_OUT = "cmd_out"
	CMD_ERR = "cmd_err"
)

const SYSTEM = "system"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SYSTEM: {Name: SYSTEM, Help: "系统命令", Value: kit.Data(mdb.FIELD, "time,id,cmd")},
	}, Commands: map[string]*ice.Command{
		SYSTEM: {Name: "system cmd run:button", Help: "系统命令", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) == 0 {
				mdb.ListSelect(m, arg...)
				return
			}
			m.Grow(SYSTEM, "", kit.Dict(mdb.TIME, m.Time(), ice.CMD, kit.Join(arg, ice.SP)))

			if len(arg) == 1 {
				arg = kit.Split(arg[0])
			}
			_system_exec(m, _system_cmd(m, arg...))
		}},
	}})
}
