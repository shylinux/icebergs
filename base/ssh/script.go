package ssh

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

type Frame struct {
	source string
	target *ice.Context
	stdout io.Writer
	stdin  io.Reader
	pipe   io.Writer

	ps1   []string
	ps2   []string
	res   string
	count int
}

func (f *Frame) prompt(m *ice.Message, list ...string) *Frame {
	if f.source != STDIO {
		return f
	}
	kit.If(len(list) == 0, func() { list = append(list, f.ps1...) })
	fmt.Fprintf(f.stdout, kit.Select("\r", "\r\033[2K", ice.Info.Colors))
	for _, v := range list {
		switch v {
		case mdb.COUNT:
			fmt.Fprintf(f.stdout, "%d", f.count)
		case tcp.HOSTNAME:
			fmt.Fprintf(f.stdout, "%s", kit.Slice(kit.Split(ice.Info.Hostname, " -/."), -1)[0])
		case mdb.TIME:
			fmt.Fprintf(f.stdout, time.Now().Format("15:04:05"))
		case TARGET:
			fmt.Fprintf(f.stdout, f.target.Name)
		default:
			kit.If(ice.Info.Colors || v[0] != '\033', func() { fmt.Fprintf(f.stdout, v) })
		}
	}
	return f
}
func (f *Frame) printf(m *ice.Message, str string, arg ...ice.Any) *Frame {
	if f.source != STDIO {
		return f
	}
	fmt.Fprint(f.stdout, kit.Format(str, arg...))
	return f
}
func (f *Frame) change(m *ice.Message, ls []string) []string {
	if len(ls) == 1 && ls[0] == "~" {
		ls = []string{ctx.CONTEXT}
	} else if len(ls) > 0 && strings.HasPrefix(ls[0], "~") {
		target := ls[0][1:]
		if ls = ls[1:]; len(target) == 0 && len(ls) > 0 {
			target, ls = ls[0], ls[1:]
		}
		if target == "~" {
			target = ""
		}
		m.Spawn(f.target).Search(target+ice.PT, func(p *ice.Context, s *ice.Context) { f.target = s })
	}
	return ls
}
func (f *Frame) alias(m *ice.Message, ls []string) []string {
	if len(ls) == 0 {
		return ls
	} else if alias := kit.Simple(kit.Value(m.Optionv(ice.MSG_ALIAS), ls[0])); len(alias) > 0 {
		ls = append(alias, ls[1:]...)
	}
	return ls
}
func (f *Frame) parse(m *ice.Message, h, line string) string {
	msg := m.Spawn(f.target)
	ls := kit.Split(strings.TrimSpace(line))
	for i, v := range ls {
		if v == "#" {
			ls = ls[:i]
			break
		}
	}
	if ls = f.change(msg, f.alias(msg, ls)); len(ls) == 0 {
		return ""
	}
	if msg.Cmdy(ls); h == STDIO && msg.IsErrNotFound() {
		msg.SetResult().Cmdy(cli.SYSTEM, ls)
	}
	f.res = Render(msg, msg.Option(ice.MSG_OUTPUT), kit.List(msg.Optionv(ice.MSG_ARGS))...)
	return ""
}
func (f *Frame) scan(m *ice.Message, h, line string) *Frame {
	f.ps1 = kit.Simple(mdb.Confv(m, PROMPT, kit.Keym(PS1)))
	f.ps2 = kit.Simple(mdb.Confv(m, PROMPT, kit.Keym(PS2)))
	// m.Options(MESSAGE, m, ice.LOG_DISABLE, ice.TRUE)
	m.I, m.O = f.stdin, f.stdout
	ps, bio := f.ps1, bufio.NewScanner(f.stdin)
	for f.prompt(m, ps...); f.stdin != nil && bio.Scan(); f.prompt(m, ps...) {
		if len(bio.Text()) == 0 && h == STDIO {
			continue
		}
		if f.count++; len(bio.Text()) == 0 {
			continue
		}
		if strings.HasSuffix(bio.Text(), "\\") {
			line += bio.Text()[:len(bio.Text())-1]
			ps = f.ps2
			continue
		}
		if line += bio.Text(); strings.Count(line, "`")%2 == 1 {
			line += ice.NL
			ps = f.ps2
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			line = ""
			continue
		}
		if ps = f.ps1; f.stdout == os.Stdout && ice.Info.Colors {
			f.printf(m, "\033[0m")
		}
		line = f.parse(m, h, line)
	}
	return f
}

func (f *Frame) Begin(m *ice.Message, arg ...string) {
	switch strings.Split(os.Getenv(cli.TERM), "-")[0] {
	case "xterm", "screen":
		ice.Info.Colors = true
	default:
		ice.Info.Colors = false
	}
}
func (f *Frame) Start(m *ice.Message, arg ...string) {
	m.Optionv(FRAME, f)
	switch f.source = kit.Select(STDIO, arg, 0); f.source {
	case STDIO:
		if f.target == nil {
			f.target = m.Target()
		}
		r, w, _ := os.Pipe()
		go func() { io.Copy(w, os.Stdin) }()
		f.pipe, f.stdin, f.stdout = w, r, os.Stdout
		m.Optionv(ice.MSG_OPTS, ice.MSG_USERNAME, ice.MSG_USERROLE)
		f.scan(m, STDIO, "")
	default:
		if m.Option(ice.MSG_SCRIPT) != "" {
			ls := kit.Split(m.Option(ice.MSG_SCRIPT), ice.PS)
			for i := len(ls) - 1; i > 0; i-- {
				if p := path.Join(path.Join(ls[:i]...), f.source); nfs.ExistsFile(m, p) {
					f.source = p
				}
			}
		}
		m.Option(ice.MSG_SCRIPT, f.source)
		f.target = m.Source()
		if msg := m.Cmd(nfs.CAT, f.source); msg.IsErr() {
			return
		} else {
			buf := bytes.NewBuffer(make([]byte, 0, ice.MOD_BUFS))
			f.stdin, f.stdout = bytes.NewBufferString(msg.Result()), buf
			defer func() { m.Echo(buf.String()) }()
		}
		f.scan(m, "", "")
	}
}
func (f *Frame) Close(m *ice.Message, arg ...string) {
	if stdin, ok := f.stdin.(io.Closer); ok {
		stdin.Close()
	}
	f.stdin = nil
}
func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}

const (
	MESSAGE = "message"
	SHELL   = "shell"
	FRAME   = "frame"
	STDIO   = "stdio"
	PS1     = "PS1"
	PS2     = "PS2"
)
const (
	SOURCE = "source"
	RETURN = "return"
	TARGET = "target"
	PROMPT = "prompt"
	PRINTF = "printf"
	SCREEN = "screen"
)

func init() {
	Index.MergeCommands(ice.Commands{
		SHELL: {Name: "shell", Help: "交互命令", Actions: mdb.HashAction(), Hand: func(m *ice.Message, arg ...string) {
			if f, ok := m.Target().Server().(*Frame); ok {
				f.Spawn(m, m.Target()).Start(m, arg...)
			}
		}},
		SOURCE: {Name: "source file run", Help: "脚本解析", Actions: mdb.HashAction(), Hand: func(m *ice.Message, arg ...string) {
			if f, ok := m.Target().Server().(*Frame); ok {
				f.Spawn(m, m.Target()).Start(m, arg...)
			}
		}},
		RETURN: {Name: "return run", Help: "结束脚本", Hand: func(m *ice.Message, arg ...string) {
			if f, ok := m.Optionv(FRAME).(*Frame); ok {
				f.Close(m, arg...)
			}
		}},
		TARGET: {Name: "target name run", Help: "当前模块", Hand: func(m *ice.Message, arg ...string) {
			if f, ok := m.Target().Server().(*Frame); ok {
				m.Search(arg[0]+ice.PT, func(p *ice.Context, s *ice.Context) { f.target = s })
				f.prompt(m)
			}
		}},
		PROMPT: {Name: "prompt arg run", Help: "命令提示", Actions: ctx.ConfAction(
			PS1, ice.List{"\033[33;44m", mdb.COUNT, "@", tcp.HOSTNAME, "[", mdb.TIME, "]", "\033[5m", TARGET, "\033[0m", "\033[44m", ">", "\033[0m ", "\033[?25h", "\033[32m"},
			PS2, ice.List{mdb.COUNT, " ", TARGET, "> "},
		), Hand: func(m *ice.Message, arg ...string) {
			if f, ok := m.Target().Server().(*Frame); ok {
				f.prompt(m, arg...)
			}
		}},
		PRINTF: {Name: "printf run text", Help: "输出显示", Hand: func(m *ice.Message, arg ...string) {
			if f, ok := m.Target().Server().(*Frame); ok {
				f.printf(m, kit.Select(m.Option(nfs.CONTENT), arg, 0))
			}
		}},
		SCREEN: {Name: "screen run text", Help: "输出命令", Hand: func(m *ice.Message, arg ...string) {
			if f, ok := m.Target().Server().(*Frame); ok {
				for _, line := range kit.Split(arg[0], ice.NL, ice.NL) {
					fmt.Fprintf(f.pipe, line+ice.NL)
					f.printf(m, line+ice.NL)
					m.Sleep300ms()
				}
				m.Echo(f.res)
			}
		}},
	})
}
