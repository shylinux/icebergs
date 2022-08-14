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
	if len(list) == 0 {
		list = append(list, f.ps1...)
	}

	fmt.Fprintf(f.stdout, "\r")
	for _, v := range list {
		switch v {
		case mdb.COUNT:
			fmt.Fprintf(f.stdout, "%d", f.count)
		case mdb.TIME:
			fmt.Fprintf(f.stdout, time.Now().Format("15:04:05"))
		case TARGET:
			fmt.Fprintf(f.stdout, f.target.Name)
		default:
			if ice.Info.Colors || v[0] != '\033' {
				fmt.Fprintf(f.stdout, v)
			}
		}
	}
	return f
}
func (f *Frame) printf(m *ice.Message, str string, arg ...ice.Any) *Frame {
	fmt.Fprint(f.stdout, kit.Format(str, arg...))
	return f
}
func (f *Frame) change(m *ice.Message, ls []string) []string {
	if len(ls) == 1 && ls[0] == "~" { // 模块列表
		ls = []string{ctx.CONTEXT}

	} else if len(ls) > 0 && strings.HasPrefix(ls[0], "~") { // 切换模块
		target := ls[0][1:]
		if ls = ls[1:]; len(target) == 0 && len(ls) > 0 {
			target, ls = ls[0], ls[1:]
		}
		if target == "~" {
			target = ""
		}
		m.Spawn(f.target).Search(target+ice.PT, func(p *ice.Context, s *ice.Context, key string) {
			m.Logs(mdb.SELECT, ctx.CONTEXT, s.Name)
			f.target = s
		})
	}
	return ls
}
func (f *Frame) alias(m *ice.Message, ls []string) []string {
	if len(ls) == 0 {
		return ls
	}
	if alias := kit.Simple(kit.Value(m.Optionv(ice.MSG_ALIAS), ls[0])); len(alias) > 0 {
		ls = append(alias, ls[1:]...)
	}
	return ls
}
func (f *Frame) parse(m *ice.Message, line string) string {
	msg := m.Spawn(f.target)
	ls := f.change(msg, f.alias(msg, kit.Split(strings.TrimSpace(line))))
	if len(ls) == 0 {
		return ""
	}

	msg.Render("", kit.List())
	if msg.Cmdy(ls); msg.IsErrNotFound() {
		msg.SetResult().Cmdy(cli.SYSTEM, ls)
	}

	f.res = Render(msg, msg.Option(ice.MSG_OUTPUT), msg.Optionv(ice.MSG_ARGS).([]ice.Any)...)
	return ""
}
func (f *Frame) scan(m *ice.Message, h, line string) *Frame {
	f.ps1 = kit.Simple(m.Confv(PROMPT, kit.Keym(PS1)))
	f.ps2 = kit.Simple(m.Confv(PROMPT, kit.Keym(PS2)))
	ps := f.ps1

	if h == STDIO {
		m.Sleep("800ms")
		pwd := ice.Render(m, ice.RENDER_QRCODE, m.Cmdx("space", "domain"))
		m.Sleep("100ms")
		f.printf(m, pwd+ice.NL)
	}

	m.I, m.O = f.stdin, f.stdout
	bio := bufio.NewScanner(f.stdin)
	for f.prompt(m, ps...); f.stdin != nil && bio.Scan(); f.prompt(m, ps...) {
		if len(bio.Text()) == 0 && h == STDIO {
			continue // 空行
		}

		f.count++

		if strings.HasSuffix(bio.Text(), "\\") {
			line += bio.Text()[:len(bio.Text())-1]
			ps = f.ps2
			continue // 续行
		}
		if line += bio.Text(); strings.Count(line, "`")%2 == 1 {
			line += ice.NL
			ps = f.ps2
			continue // 多行
		}
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			line = ""
			continue // 注释
		}
		if ps = f.ps1; f.stdout == os.Stdout {
			if ice.Info.Colors {
				f.printf(m, "\033[0m") // 清空格式
			}
		}
		line = f.parse(m, line)
	}
	return f
}

func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	return f
}
func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	m.Optionv(FRAME, f)
	switch f.source = kit.Select(STDIO, arg, 0); f.source {
	case STDIO: // 终端交互
		if m.Cap(ice.CTX_STREAM, f.source); f.target == nil {
			f.target = m.Target()
		}

		r, w, _ := os.Pipe()
		go func() { io.Copy(w, os.Stdin) }()
		f.pipe, f.stdin, f.stdout = w, r, os.Stdout

		m.Option(ice.MSG_OPTS, ice.MSG_USERNAME)
		f.scan(m, STDIO, "")

	default: // 脚本文件
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
			return true // 查找失败
		} else {
			buf := bytes.NewBuffer(make([]byte, 0, ice.MOD_BUFS))
			f.stdin, f.stdout = bytes.NewBufferString(msg.Result()), buf
			defer func() { m.Echo(buf.String()) }()
		}

		f.scan(m, "", "")
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	if stdin, ok := f.stdin.(io.Closer); ok {
		stdin.Close()
	}
	f.stdin = nil
	return true
}

const (
	FRAME = "frame"
	STDIO = "stdio"
	PS1   = "PS1"
	PS2   = "PS2"
)
const (
	SCRIPT = "script"
	SOURCE = "source"
	TARGET = "target"
	PROMPT = "prompt"
	PRINTF = "printf"
	SCREEN = "screen"
	RETURN = "return"
)

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		SOURCE: {Name: SOURCE, Help: "加载脚本", Value: kit.Data()},
		PROMPT: {Name: PROMPT, Help: "命令提示", Value: kit.Data(
			PS1, []ice.Any{"\033[33;44m", mdb.COUNT, "[", mdb.TIME, "]", "\033[5m", TARGET, "\033[0m", "\033[44m", ">", "\033[0m ", "\033[?25h", "\033[32m"},
			PS2, []ice.Any{mdb.COUNT, " ", TARGET, "> "},
		)},
	}, Commands: ice.Commands{
		SOURCE: {Name: "source file", Help: "脚本解析", Hand: func(m *ice.Message, arg ...string) {
			if f, ok := m.Target().Server().(*Frame); ok {
				f.Spawn(m, m.Target()).Start(m, arg...)
			}
		}},
		TARGET: {Name: "target name run", Help: "当前模块", Hand: func(m *ice.Message, arg ...string) {
			f := m.Target().Server().(*Frame)
			m.Search(arg[0]+ice.PT, func(p *ice.Context, s *ice.Context, key string) { f.target = s })
			f.prompt(m)
		}},
		PROMPT: {Name: "prompt arg run", Help: "命令提示", Hand: func(m *ice.Message, arg ...string) {
			if f, ok := m.Optionv(FRAME).(*Frame); ok {
				f.ps1 = arg
				f.prompt(m)
			}
		}},
		PRINTF: {Name: "printf run text", Help: "输出显示", Hand: func(m *ice.Message, arg ...string) {
			if f, ok := m.Optionv(FRAME).(*Frame); ok {
				f.printf(m, arg[0])
			}
		}},
		SCREEN: {Name: "screen run text", Help: "输出命令", Hand: func(m *ice.Message, arg ...string) {
			if f, ok := m.Optionv(FRAME).(*Frame); ok {
				for _, line := range kit.Split(arg[0], ice.NL, ice.NL) {
					fmt.Fprintf(f.pipe, line+ice.NL)
					f.printf(m, line+ice.NL)
					m.Sleep300ms()
				}
				m.Echo(f.res)
			}
		}},
		RETURN: {Name: "return", Help: "结束脚本", Hand: func(m *ice.Message, arg ...string) {
			if f, ok := m.Optionv(FRAME).(*Frame); ok {
				f.Close(m, arg...)
			}
		}},
	}})
}