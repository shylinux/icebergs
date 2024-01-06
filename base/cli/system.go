package cli

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/file"
)

func _path_split(ps string) []string {
	ps = kit.ReplaceAll(ps, "\\", nfs.PS)
	return kit.Split(ps, lex.NL+kit.Select(nfs.DF, ";", strings.Contains(ps, ";")), lex.NL)
}
func _system_cmd(m *ice.Message, arg ...string) *exec.Cmd {
	bin, env := "", kit.Simple(m.Optionv(CMD_ENV))
	kit.For(env, func(k, v string) {
		if k == PATH {
			if bin = _system_find(m, arg[0], _path_split(v)...); bin != "" {
				m.Logs(FIND, "envpath cmd", bin)
			}
		}
	})
	if bin == "" {
		if bin = _system_find(m, arg[0], EtcPath(m)...); bin != "" {
			m.Logs(FIND, "etcpath cmd", bin)
		}
	}
	if bin == "" {
		if bin = _system_find(m, arg[0], m.Option(CMD_DIR), ice.BIN, nfs.PWD); bin != "" {
			m.Logs(FIND, "contexts cmd", bin)
		}
	}
	if bin == "" && !strings.Contains(arg[0], nfs.PS) {
		if bin = _system_find(m, arg[0]); bin != "" {
			m.Logs(FIND, "systems cmd", bin)
		}
	}
	if bin == "" && !strings.Contains(arg[0], nfs.PS) {
		m.Cmd(MIRRORS, CMD, arg[0])
		if bin = _system_find(m, arg[0]); bin != "" {
			m.Logs(FIND, "mirrors cmd", bin)
		}
	}
	cmd := exec.Command(kit.Select(arg[0], bin), arg[1:]...)
	if cmd.Dir = kit.TrimPath(m.Option(CMD_DIR)); len(cmd.Dir) > 0 {
		if m.Logs(EXEC, CMD_DIR, cmd.Dir); !nfs.Exists(m, cmd.Dir) {
			file.MkdirAll(cmd.Dir, ice.MOD_DIR)
		}
	}
	kit.For(env, func(k, v string) { cmd.Env = append(cmd.Env, kit.Format("%s=%s", k, v)) })
	kit.If(len(cmd.Env) > 0, func() { m.Logs(EXEC, CMD_ENV, kit.Format(cmd.Env)) })
	return cmd
}
func _system_out(m *ice.Message, out string) io.Writer {
	if w, ok := m.Optionv(out).(io.Writer); ok {
		return w
	} else if m.Option(out) == "" {
		return nil
	} else if f, p, e := file.CreateFile(m.Option(out)); m.Assert(e) {
		m.Logs(nfs.SAVE, out, p).Optionv(out, f)
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
			if m.Echo(out.String()).Echo(err.String()); m.IsErr() {
				m.Option(ice.MSG_ARGS, kit.Simple(http.StatusBadRequest, cmd.Args, err.String()))
				m.Echo(strings.TrimRight(err.String(), lex.NL))
			}
		}()
	}
	if e := cmd.Run(); !m.Warn(e, ice.ErrNotValid, cmd.Args) {
		m.Cost(CODE, _system_code(cmd), EXEC, cmd.Args)
	}
	m.Push(mdb.TIME, m.Time()).Push(CODE, _system_code(cmd)).StatusTime()
}
func _system_code(cmd *exec.Cmd) string {
	return kit.Select("1", "0", cmd.ProcessState != nil && cmd.ProcessState.Success())
}
func _system_find(m *ice.Message, bin string, dir ...string) string {
	if strings.Contains(bin, nfs.DF) {
		return bin
	}
	if strings.HasPrefix(bin, nfs.PS) {
		return bin
	}
	if strings.HasPrefix(bin, nfs.PWD) {
		return bin
	}
	kit.If(len(dir) == 0, func() { dir = append(dir, _path_split(kit.Env(PATH))...) })
	for _, p := range dir {
		if nfs.Exists(m, path.Join(p, bin)) {
			return kit.Path(p, bin)
		}
		if IsWindows() && nfs.Exists(m, path.Join(p, bin)+".exe") {
			return kit.Path(p, bin) + ".exe"
		}
	}
	if nfs.Exists(m, bin) {
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

	SH    = "sh"
	MAN   = "man"
	RUN   = "run"
	KILL  = "kill"
	FIND  = "find"
	GREP  = "grep"
	SUDO  = "sudo"
	EXEC  = "exec"
	EXIT  = "exit"
	ECHO  = "echo"
	REST  = "rest"
	OPENS = "opens"
	PARAM = "param"
)

const SYSTEM = "system"

func init() {
	Index.MergeCommands(ice.Commands{
		SYSTEM: {Name: "system cmd", Help: "系统命令", Actions: ice.MergeActions(ice.Actions{
			nfs.PUSH: {Hand: func(m *ice.Message, arg ...string) {
				kit.For(arg, func(p string) {
					kit.If(!kit.IsIn(p, EtcPath(m)...), func() {
						m.Cmd(nfs.PUSH, ice.ETC_PATH, strings.TrimSpace(p)+lex.NL)
					})
				})
				m.Cmdy(nfs.CAT, ice.ETC_PATH)
			}},
			FIND: {Hand: func(m *ice.Message, arg ...string) { m.Echo(_system_find(m, arg[0], arg[1:]...)) }},
			MAN: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(len(arg) == 1, func() { arg = append(arg, "") })
				m.Echo(SystemCmds(m, "man %s %s|col -b", kit.Select("", arg[1], arg[1] != "1"), arg[0]))
			}},
			OPENS: {Hand: func(m *ice.Message, arg ...string) { Opens(m, arg...) }},
		}, mdb.HashAction(mdb.SHORT, CMD, mdb.FIELD, "time,cmd,arg")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m)
			} else if _system_exec(m, _system_cmd(m, arg...)); IsSuccess(m) && m.Append(CMD_ERR) == "" {
				m.SetAppend()
			}
		}},
	})
}

func SystemFind(m *ice.Message, bin string, dir ...string) string {
	dir = append(dir, EtcPath(m)...)
	return _system_find(m, bin, append(dir, _path_split(kit.Env(PATH))...)...)
}
func SystemExec(m *ice.Message, arg ...string) string { return strings.TrimSpace(m.Cmdx(SYSTEM, arg)) }
func SystemCmds(m *ice.Message, cmds string, args ...ice.Any) string {
	return strings.TrimRight(m.Cmdx(SYSTEM, "sh", "-c", kit.Format(cmds, args...), ice.Option{CMD_OUTPUT, ""}), lex.NL)
}
func IsSuccess(m *ice.Message) bool { return m.Append(CODE) == "" || m.Append(CODE) == "0" }

var _cache_path []string

func Shell(m *ice.Message) string {
	return kit.Select("/bin/sh", os.Getenv("SHELL"))
}
func EtcPath(m *ice.Message) (res []string) {
	if len(_cache_path) > 0 {
		return _cache_path
	}
	nfs.Exists(m, ice.ETC_PATH, func(p string) {
		kit.For(strings.Split(m.Cmdx(nfs.CAT, p, kit.Dict(aaa.UserRole, aaa.ROOT)), lex.NL), func(p string) {
			kit.If(p != "" && !strings.HasPrefix(p, "# "), func() {
				res = append(res, p)
			})
		})
	})
	_cache_path = res
	return
}
