package cli

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/file"
)

func _path_split(ps string) []string {
	ps = strings.ReplaceAll(ps, "\\", "/")
	return kit.Split(ps, "\n"+kit.Select(":", ";", strings.Contains(ps, ";")), "\n")
}
func _system_cmd(m *ice.Message, arg ...string) *exec.Cmd {
	bin, env := "", kit.Simple(m.Optionv(CMD_ENV))
	for i := 0; i < len(env)-1; i += 2 {
		if env[i] == PATH {
			if bin = _system_find(m, arg[0], _path_split(env[i+1])...); bin != "" {
				m.Logs(mdb.SELECT, "envpath cmd", bin)
			}
		}
	}
	if bin == "" {
		if text := kit.ReadFile(ice.ETC_PATH); len(text) > 0 {
			if bin = _system_find(m, arg[0], strings.Split(text, ice.NL)...); bin != "" {
				m.Logs(mdb.SELECT, "etcpath cmd", bin)
			}
		}
	}
	if bin == "" {
		if bin = _system_find(m, arg[0], ice.BIN, m.Option(CMD_DIR)); bin != "" {
			m.Logs(mdb.SELECT, "contexts cmd", bin)
		}
	}
	if bin == "" {
		if bin = _system_find(m, arg[0], ice.BIN, nfs.PWD); bin != "" {
			m.Logs(mdb.SELECT, "contexts cmd", bin)
		}
	}
	if bin == "" && !strings.Contains(arg[0], ice.PS) {
		if bin = _system_find(m, arg[0]); bin != "" {
			m.Logs(mdb.SELECT, "systems cmd", bin)
		}
	}
	if bin == "" && !strings.Contains(arg[0], ice.PS) {
		m.Cmd(MIRRORS, CMD, arg[0])
		if bin = _system_find(m, arg[0]); bin != "" {
			m.Logs(mdb.SELECT, "mirrors cmd", bin)
		}
	}
	if bin == "" && runtime.GOOS == WINDOWS {
		if bin = _system_find(m, arg[0], path.Join(os.Getenv("ProgramFiles"), "Git/bin"), path.Join(os.Getenv("ProgramFiles"), "Go/bin")); bin != "" {
			m.Logs(mdb.SELECT, "systems cmd", bin)
		}
	}
	if bin == "" {
		bin = arg[0]
	}
	cmd := exec.Command(bin, arg[1:]...)
	if cmd.Dir = kit.TrimPath(m.Option(CMD_DIR)); len(cmd.Dir) > 0 {
		if m.Logs(mdb.PARAMS, CMD_DIR, cmd.Dir); !nfs.ExistsFile(m, cmd.Dir) {
			file.MkdirAll(cmd.Dir, ice.MOD_DIR)
		}
	}
	for i := 0; i < len(env)-1; i += 2 {
		cmd.Env = append(cmd.Env, kit.Format("%s=%s", env[i], env[i+1]))
	}
	if len(cmd.Env) > 0 {
		m.Logs(mdb.PARAMS, CMD_ENV, kit.Format(cmd.Env))
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
func _system_exec(m *ice.Message, cmd *exec.Cmd) {
	if r, ok := m.Optionv(CMD_INPUT).(io.Reader); ok {
		cmd.Stdin = r
	}
	if w := _system_out(m, CMD_OUTPUT); w != nil {
		cmd.Stdout, cmd.Stderr = w, w
		if w := _system_out(m, CMD_ERRPUT); w != nil {
			cmd.Stderr = w
		}
	} else {
		out := bytes.NewBuffer(make([]byte, 0, ice.MOD_BUFS))
		err := bytes.NewBuffer(make([]byte, 0, ice.MOD_BUFS))
		cmd.Stdout, cmd.Stderr = out, err
		defer func() {
			m.Push(CMD_OUT, out.String()).Push(CMD_ERR, err.String())
			m.Echo(strings.TrimRight(out.String(), ice.NL))
			if m.IsErr() {
				m.Option(ice.MSG_ARGS, kit.Simple(http.StatusBadRequest, cmd.Args, err.String()))
				m.Echo(strings.TrimRight(err.String(), ice.NL))
			}
		}()
	}
	if e := cmd.Run(); !m.Warn(e, ice.ErrNotValid, cmd.Args) {
		m.Cost(CODE, _system_code(cmd), ctx.ARGS, cmd.Args)
	}
	m.Push(mdb.TIME, m.Time()).Push(CODE, _system_code(cmd))
}
func _system_code(cmd *exec.Cmd) string {
	return kit.Select("1", "0", cmd.ProcessState != nil && cmd.ProcessState.Success())
}
func _system_find(m Message, bin string, dir ...string) string {
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
		dir = append(dir, _path_split(kit.Env(PATH))...)
	}
	for _, p := range dir {
		if nfs.ExistsFile(m, path.Join(p, bin)) {
			return kit.Path(p, bin)
		}
		if nfs.ExistsFile(m, path.Join(p, bin)+".exe") {
			return kit.Path(p, bin) + ".exe"
		}
	}
	if nfs.ExistsFile(m, bin) {
		return kit.Path(bin)
	}
	return ""
}

const (
	CMD_DIR = "cmd_dir"
	CMD_ENV = "cmd_env"

	CMD_INPUT  = "cmd_input"
	CMD_OUTPUT = "cmd_output"
	CMD_ERRPUT = "cmd_errput"

	CMD_ERR = "cmd_err"
	CMD_OUT = "cmd_out"

	MAN  = "man"
	GREP = "grep"
)

const SYSTEM = "system"

func init() {
	Index.MergeCommands(ice.Commands{
		SYSTEM: {Name: "system cmd", Help: "系统命令", Actions: ice.MergeActions(ice.Actions{
			nfs.PUSH: {Hand: func(m *ice.Message, arg ...string) {
				for _, p := range arg {
					if !strings.Contains(m.Cmdx(nfs.CAT, ice.ETC_PATH), p) {
						m.Cmd(nfs.PUSH, ice.ETC_PATH, strings.TrimSpace(p)+ice.NL)
					}
				}
				m.Cmdy(nfs.CAT, ice.ETC_PATH)
			}},
			"find": {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(_system_find(m, arg[0], arg[1:]...))
			}},
			MAN: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 1 {
					arg = append(arg, "")
				}
				m.Option(CMD_ENV, "COLUMNS", kit.Int(kit.Select("1920", m.Option(ice.MSG_WIDTH)))/12)
				m.Echo(SystemCmds(m, "man %s %s|col -b", kit.Select("", arg[1], arg[1] != "1"), arg[0]))
			}},
		}, mdb.HashAction(mdb.SHORT, "cmd", mdb.FIELD, "time,cmd,arg")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m)
				return
			}
			if _system_exec(m, _system_cmd(m, arg...)); IsSuccess(m) && m.Append(CMD_ERR) == "" {
				m.SetAppend()
			}
		}},
	})
}

type Message interface {
	Append(key string, arg ...ice.Any) string
	Optionv(key string, arg ...ice.Any) ice.Any
}

func SystemFind(m Message, bin string, dir ...string) string {
	if text := kit.ReadFile(ice.ETC_PATH); len(text) > 0 {
		dir = append(dir, strings.Split(text, ice.NL)...)
	}
	dir = append(dir, _path_split(kit.Env(PATH))...)
	return _system_find(m, bin, dir...)
}
func IsSuccess(m Message) bool                        { return m.Append(CODE) == "" || m.Append(CODE) == "0" }
func SystemExec(m *ice.Message, arg ...string) string { return strings.TrimSpace(m.Cmdx(SYSTEM, arg)) }
func SystemCmds(m *ice.Message, cmds string, args ...ice.Any) string {
	return strings.TrimRight(m.Cmdx(SYSTEM, "sh", "-c", kit.Format(cmds, args...), ice.Option{CMD_OUTPUT, ""}), ice.NL)
}
