package cli

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _system_cmd(m *ice.Message, arg ...string) *exec.Cmd {
	cmd := exec.Command(arg[0], arg[1:]...)

	// 运行目录
	if cmd.Dir = m.Option(CMD_DIR); len(cmd.Dir) > 0 {
		if m.Log_EXPORT(CMD_DIR, cmd.Dir); !kit.FileExists(cmd.Dir) {
			os.MkdirAll(cmd.Dir, ice.MOD_DIR)
		}
	}

	// 环境变量
	env := kit.Simple(m.Optionv(CMD_ENV))
	for i := 0; i < len(env)-1; i += 2 {
		if cmd.Env = append(cmd.Env, kit.Format("%s=%s", env[i], env[i+1])); env[i] == PATH {
			if file := _system_find(m, arg[0], strings.Split(env[i+1], ice.DF)...); file != "" {
				m.Debug("cmd: %v", file)
				cmd.Path = file
			}
		}
	}

	// 定制目录
	if text := kit.ReadFile(ice.ETC_PATH); len(text) > 0 {
		if file := _system_find(m, arg[0], strings.Split(text, ice.NL)...); file != "" {
			m.Debug("cmd: %v", file)
			cmd.Path = file
		}
	}
	m.Debug("cmd: %v", cmd.Path)

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
func _system_find(m *ice.Message, bin string, dir ...string) string {
	if strings.HasPrefix(bin, ice.PS) {
		return bin
	}
	if strings.HasPrefix(bin, nfs.PWD) {
		return kit.Path(bin)
	}
	if len(dir) == 0 {
		dir = append(dir, strings.Split(kit.Env(PATH), ice.DF)...)
	}
	for _, p := range dir {
		if _, err := os.Stat(path.Join(p, bin)); err == nil {
			return kit.Path(path.Join(p, bin))
		}
	}
	return ""
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
func SystemFind(m *ice.Message, bin string, dir ...string) string {
	return _system_find(m, bin, dir...)
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
		SYSTEM: {Name: "system cmd run", Help: "系统命令", Action: map[string]*ice.Action{
			nfs.FIND: {Name: "find", Help: "查找", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(_system_find(m, arg[0], arg[1:]...))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
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
