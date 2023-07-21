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
	"shylinux.com/x/icebergs/base/lex"
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

func (f *Frame) prompt(m *ice.Message, arg ...string) *Frame {
	if f.source != STDIO {
		return f
	}
	kit.If(len(arg) == 0, func() { arg = append(arg, f.ps1...) })
	fmt.Fprintf(f.stdout, kit.Select("\r", "\r\033[2K", ice.Info.Colors))
	kit.For(arg, func(k string) {
		switch k {
		case mdb.COUNT:
			fmt.Fprintf(f.stdout, "%d", f.count)
		case tcp.HOSTNAME:
			fmt.Fprintf(f.stdout, ice.Info.NodeName)
		case mdb.TIME:
			fmt.Fprintf(f.stdout, kit.Slice(kit.Split(time.Now().Format(ice.MOD_TIME)), -1)[0])
		case TARGET:
			fmt.Fprintf(f.stdout, f.target.Name)
		default:
			kit.If(ice.Info.Colors || k[0] != '\033', func() { fmt.Fprintf(f.stdout, k) })
		}
	})
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
		kit.If(target == "~", func() { target = "" })
		m.Spawn(f.target).Search(target+nfs.PT, func(p *ice.Context, s *ice.Context) { f.target = s })
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
	ls := kit.Split(strings.TrimSpace(line))
	for i, v := range ls {
		if v == "#" {
			ls = ls[:i]
			break
		}
	}
	if ls = f.change(m, f.alias(m, ls)); len(ls) == 0 {
		return ""
	}
	msg := m.Spawn(f.target)
	if msg.Cmdy(ls); h == STDIO && msg.IsErrNotFound() {
		msg.SetResult().Cmdy(cli.SYSTEM, ls)
	}
	f.res = Render(msg, msg.Option(ice.MSG_OUTPUT), kit.List(msg.Optionv(ice.MSG_ARGS))...)
	return ""
}
func (f *Frame) scan(m *ice.Message, h, line string) *Frame {
	kit.If(f.source == STDIO, func() { m.Option(ice.LOG_DISABLE, ice.TRUE) })
	f.ps1 = kit.Simple(mdb.Confv(m, PROMPT, kit.Keym(PS1)))
	f.ps2 = kit.Simple(mdb.Confv(m, PROMPT, kit.Keym(PS2)))
	ps, bio := f.ps1, bufio.NewScanner(f.stdin)
	m.I, m.O = f.stdin, f.stdout
	for f.prompt(m, ps...); f.stdin != nil && bio.Scan(); f.prompt(m, ps...) {
		if len(bio.Text()) == 0 && h == STDIO {
			continue
		}
		f.count++
		if line += bio.Text(); strings.Count(line, "`")%2 == 1 {
			line += lex.NL
			ps = f.ps2
			continue
		} else if len(bio.Text()) == 0 {
			continue
		} else if strings.HasSuffix(bio.Text(), "\\") {
			line += bio.Text()[:len(bio.Text())-1]
			ps = f.ps2
			continue
		} else if strings.HasPrefix(strings.TrimSpace(line), "#") {
			line = ""
			continue
		} else if ps = f.ps1; f.stdout == os.Stdout && ice.Info.Colors {
			f.printf(m, "\033[0m")
		}
		line = f.parse(m, h, line)
	}
	return f
}

func (f *Frame) Begin(m *ice.Message, arg ...string) {
	ice.Info.Colors = kit.IsIn(strings.Split(os.Getenv(cli.TERM), "-")[0], "xterm", "screen")
}
func (f *Frame) Start(m *ice.Message, arg ...string) {
	m.Optionv(FRAME, f)
	switch f.source = kit.Select(STDIO, arg, 0); f.source {
	case STDIO:
		r, w, _ := os.Pipe()
		go func() { io.Copy(w, os.Stdin) }()
		f.pipe, f.stdin, f.stdout = w, r, os.Stdout
		kit.If(f.target == nil, func() { f.target = m.Target() })
		m.Optionv(ice.MSG_OPTS, ice.MSG_USERNAME, ice.MSG_USERROLE)
		f.scan(m, STDIO, "")
	default:
		if m.Option(ice.MSG_SCRIPT) != "" {
			ls := kit.Split(m.Option(ice.MSG_SCRIPT), nfs.PS)
			for i := len(ls) - 1; i > 0; i-- {
				if p := path.Join(path.Join(ls[:i]...), f.source); nfs.Exists(m, p) {
					f.source = p
				}
			}
		}
		if msg := m.Cmd(nfs.CAT, m.Option(ice.MSG_SCRIPT, f.source)); msg.IsErr() {
			return
		} else {
			buf := bytes.NewBuffer(make([]byte, 0, ice.MOD_BUFS))
			f.stdin, f.stdout = bytes.NewBufferString(msg.Result()), buf
			defer func() { m.Echo(buf.String()) }()
		}
		f.target = m.Source()
		f.scan(m, "", "")
	}
}
func (f *Frame) Close(m *ice.Message, arg ...string) {
	nfs.Close(m, f.stdin)
	f.stdin = nil
}
func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server { return &Frame{} }

const (
	FRAME = "frame"
	SHELL = "shell"
	WEBIO = "webio"
	STDIO = "stdio"
	PS1   = "PS1"
	PS2   = "PS2"
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
				m.Search(arg[0]+nfs.PT, func(p *ice.Context, s *ice.Context) { f.target = s })
				f.prompt(m)
			}
		}},
		PROMPT: {Name: "prompt arg run", Help: "命令提示", Actions: ctx.ConfAction(
			PS1, ice.List{"\033[33;44m", mdb.COUNT, mdb.AT, tcp.HOSTNAME, "[", mdb.TIME, "]", "\033[5m", TARGET, "\033[0m", "\033[44m", ">", "\033[0m ", "\033[?25h", "\033[32m"},
			PS2, ice.List{mdb.COUNT, lex.SP, TARGET, "> "},
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
				for _, line := range kit.Split(arg[0], lex.NL, lex.NL) {
					fmt.Fprintf(f.pipe, line+lex.NL)
					f.printf(m, line+lex.NL)
					m.Sleep300ms()
				}
				m.Echo(f.res)
			}
		}},
	})
}

func PrintQRCode(m *ice.Message, url string) {
	m.Spawn(ice.OptionSilent()).Cmd(PRINTF, kit.Dict(nfs.CONTENT, lex.NL+ice.Render(m, ice.RENDER_QRCODE, url))).Cmd(PROMPT)
}
