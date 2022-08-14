package cli

import (
	"bytes"
	"io"
	"os/exec"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/file"
)

func _system_cmd(m *ice.Message, arg ...string) *exec.Cmd {
	if text := kit.ReadFile(ice.ETC_PATH); len(text) > 0 {
		if file := _system_find(m, arg[0], strings.Split(text, ice.NL)...); file != "" {
			m.Logs(mdb.SELECT, "etc path cmd", file)
			arg[0] = file // 配置目录
		}
	}
	env := kit.Simple(m.Optionv(CMD_ENV))
	for i := 0; i < len(env)-1; i += 2 {
		if env[i] == PATH {
			if file := _system_find(m, arg[0], strings.Split(env[i+1], ice.DF)...); file != "" {
				m.Logs(mdb.SELECT, "env path cmd", file)
				arg[0] = file // 环境变量
			}
		}
	}
	if _system_find(m, arg[0]) == "" {
		m.Cmd(MIRRORS, CMD, arg[0])
		if file := _system_find(m, arg[0]); file != "" {
			m.Logs(mdb.SELECT, "mirrors cmd", file)
			arg[0] = file // 软件镜像
		}
	}
	cmd := exec.Command(arg[0], arg[1:]...)

	// 运行目录
	if cmd.Dir = m.Option(CMD_DIR); len(cmd.Dir) > 0 {
		if m.Logs(mdb.EXPORT, CMD_DIR, cmd.Dir); !nfs.ExistsFile(m, cmd.Dir) {
			file.MkdirAll(cmd.Dir, ice.MOD_DIR)
		}
	}
	// 环境变量
	for i := 0; i < len(env)-1; i += 2 {
		cmd.Env = append(cmd.Env, kit.Format("%s=%s", env[i], env[i+1]))
	}
	if len(cmd.Env) > 0 {
		m.Logs(mdb.EXPORT, CMD_ENV, kit.Format(cmd.Env))
	}
	return cmd
}
func _system_out(m *ice.Message, out string) io.Writer {
	if w, ok := m.Optionv(out).(io.Writer); ok {
		return w
	} else if m.Option(out) == "" {
		return nil
	} else if f, p, e := file.CreateFile(m.Option(out)); m.Assert(e) {
		m.Logs(mdb.EXPORT, out, p)
		m.Optionv(out, f)
		return f
	}
	return nil
}
func _system_find(m *ice.Message, bin string, dir ...string) string {
	if strings.Contains(bin, ice.DF) {
		return bin
	}
	if strings.HasPrefix(bin, ice.PS) {
		return bin
	}
	if strings.HasPrefix(bin, nfs.PWD) {
		return bin
	}
	if len(dir) == 0 {
		dir = append(dir, strings.Split(kit.Env(PATH), ice.DF)...)
	}
	for _, p := range dir {
		if nfs.ExistsFile(m, path.Join(p, bin)) {
			return kit.Path(path.Join(p, bin))
		}
	}
	return ""
}
func _system_exec(m *ice.Message, cmd *exec.Cmd) {
	if r, ok := m.Optionv(CMD_INPUT).(io.Reader); ok {
		cmd.Stdin = r // 输入流
	}
	if w := _system_out(m, CMD_OUTPUT); w != nil {
		cmd.Stdout, cmd.Stderr = w, w // 输出流
		if w := _system_out(m, CMD_ERRPUT); w != nil {
			cmd.Stderr = w
		}
	} else {
		out := bytes.NewBuffer(make([]byte, 0, ice.MOD_BUFS))
		err := bytes.NewBuffer(make([]byte, 0, ice.MOD_BUFS))
		cmd.Stdout, cmd.Stderr = out, err
		defer func() {
			m.Push(CMD_OUT, out.String())
			m.Push(CMD_ERR, err.String())
			if m.Echo(strings.TrimSpace(kit.Select(out.String(), err.String()))); IsSuccess(m) {
				m.SetAppend()
			}
		}()
	}

	// 执行命令
	if e := cmd.Run(); !m.Warn(e, ice.ErrNotFound, cmd.Args) {
		m.Cost(CODE, cmd.ProcessState.ExitCode(), ctx.ARGS, cmd.Args)
	}

	m.Push(mdb.TIME, m.Time()).Push(CODE, int(cmd.ProcessState.ExitCode()))
}

const (
	CMD_DIR = "cmd_dir"
	CMD_ENV = "cmd_env"

	CMD_INPUT  = "cmd_input"
	CMD_OUTPUT = "cmd_output"
	CMD_ERRPUT = "cmd_errput"

	CMD_ERR = "cmd_err"
	CMD_OUT = "cmd_out"
)

const SYSTEM = "system"

func init() {
	Index.MergeCommands(ice.Commands{
		SYSTEM: {Name: "system cmd run", Help: "系统命令", Actions: ice.Actions{
			nfs.FIND: {Name: "find", Help: "查找", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(_system_find(m, arg[0], arg[1:]...))
			}},
			nfs.PUSH: {Name: "push", Help: "查找", Hand: func(m *ice.Message, arg ...string) {
				for _, p := range arg {
					if !strings.Contains(m.Cmdx(nfs.CAT, ice.ETC_PATH), p) {
						m.Cmd(nfs.PUSH, ice.ETC_PATH, strings.TrimSpace(p)+ice.NL)
					}
				}
				m.Cmdy(nfs.CAT, ice.ETC_PATH)
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				return
			}
			if len(arg) == 1 {
				arg = kit.Split(arg[0])
			}
			_system_exec(m, _system_cmd(m, arg...))
		}},
	})
}

func IsSuccess(m *ice.Message) bool {
	return m.Append(CODE) == "0" || m.Append(CODE) == ""
}
func SystemFind(m *ice.Message, bin string, dir ...string) string {
	if text := kit.ReadFile(ice.ETC_PATH); len(text) > 0 {
		dir = append(dir, strings.Split(text, ice.NL)...)
	}
	dir = append(dir, strings.Split(kit.Env(PATH), ice.DF)...)
	return _system_find(m, bin, dir...)
}
