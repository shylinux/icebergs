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
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func Render(msg *ice.Message, cmd string, args ...interface{}) (res string) {
	switch arg := kit.Simple(args...); cmd {
	case ice.RENDER_VOID:
		return res
	case ice.RENDER_RESULT:
		if len(arg) > 0 {
			msg.Resultv(arg)
		}
		res = msg.Result()

	default:
		if res = msg.Result(); res == "" {
			res = msg.Table().Result()
		}
	}
	if fmt.Fprint(msg.O, res); !strings.HasSuffix(res, ice.NL) {
		fmt.Fprint(msg.O, ice.NL)
	}
	return res
}

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
		case kit.MDB_COUNT:
			fmt.Fprintf(f.stdout, "%d", f.count)
		case kit.MDB_TIME:
			fmt.Fprintf(f.stdout, time.Now().Format("15:04:05"))
		case TARGET:
			fmt.Fprintf(f.stdout, f.target.Name)
		default:
			fmt.Fprintf(f.stdout, v)
		}
	}
	return f
}
func (f *Frame) printf(m *ice.Message, res string, arg ...interface{}) *Frame {
	if len(arg) > 0 {
		fmt.Fprintf(f.stdout, res, arg...)
	} else {
		fmt.Fprint(f.stdout, res)
	}
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
			m.Log_SELECT(ctx.CONTEXT, s.Name)
			f.target = s
		})
	}
	return ls
}
func (f *Frame) alias(m *ice.Message, ls []string) []string {
	if alias, ok := m.Optionv(ice.MSG_ALIAS).(map[string]interface{}); ok {
		if len(ls) > 0 {
			if a := kit.Simple(alias[ls[0]]); len(a) > 0 {
				ls = append(append([]string{}, a...), ls[1:]...)
			}
		}
	}
	return ls
}
func (f *Frame) parse(m *ice.Message, line string) string {
	for _, one := range kit.Split(line, ";", ";", ";") {
		msg := m.Spawn(f.target)

		ls := f.change(msg, f.alias(msg, kit.Split(strings.TrimSpace(one))))
		if len(ls) == 0 {
			continue
		}

		if msg.Cmdy(ls[0], ls[1:]); msg.Result(1) == ice.ErrNotFound {
			msg.Set(ice.MSG_RESULT).Cmdy(cli.SYSTEM, ls)
		}

		_args, _ := msg.Optionv(ice.MSG_ARGS).([]interface{})
		f.res = Render(msg, msg.Option(ice.MSG_OUTPUT), _args...)
	}
	return ""
}
func (f *Frame) scan(m *ice.Message, h, line string) *Frame {
	f.ps1 = kit.Simple(m.Confv(PROMPT, kit.Keym(PS1)))
	f.ps2 = kit.Simple(m.Confv(PROMPT, kit.Keym(PS2)))
	ps := f.ps1

	m.Sleep300ms()
	m.I, m.O = f.stdin, f.stdout
	bio := bufio.NewScanner(f.stdin)
	for f.prompt(m, ps...); bio.Scan() && f.stdin != nil; f.prompt(m, ps...) {
		if h == STDIO && len(bio.Text()) == 0 {
			continue // 空行
		}

		m.Cmdx(mdb.INSERT, SOURCE, kit.Keys(kit.MDB_HASH, h), mdb.LIST, kit.MDB_TEXT, bio.Text())
		f.count++

		if len(bio.Text()) == 0 {
			if strings.Count(line, "`")%2 == 1 {
				line += ice.NL
			}
			continue // 空行
		}
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
			continue
		}
		// if line = strings.Split(line, " # ")[0]; len(line) == 0 {
		// 	continue // 注释
		// }
		if ps = f.ps1; f.stdout == os.Stdout {
			f.printf(m, "\033[0m") // 清空格式
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
		m.Cap(ice.CTX_STREAM, f.source)
		if f.target == nil {
			f.target = m.Target()
		}

		r, w, _ := os.Pipe()
		m.Go(func() { io.Copy(w, os.Stdin) })
		f.stdin, f.stdout = r, os.Stdout
		f.pipe = w

		aaa.UserRoot(m)
		m.Option(ice.MSG_OPTS, ice.MSG_USERNAME)

		m.Conf(SOURCE, kit.Keys(kit.MDB_HASH, STDIO, kit.Keym(kit.MDB_NAME)), STDIO)
		m.Conf(SOURCE, kit.Keys(kit.MDB_HASH, STDIO, kit.Keym(kit.MDB_TIME)), m.Time())

		f.count = kit.Int(m.Conf(SOURCE, kit.Keys(kit.MDB_HASH, STDIO, kit.Keym(kit.MDB_COUNT)))) + 1
		f.scan(m, STDIO, "")

	default: // 脚本文件
		if strings.Contains(m.Option(ice.MSG_SCRIPT), ice.PS) {
			f.source = path.Join(path.Dir(m.Option(ice.MSG_SCRIPT)), f.source)
		}
		m.Option(ice.MSG_SCRIPT, f.source)
		f.target = m.Source()

		if msg := m.Cmd(nfs.CAT, f.source); msg.Result(0) == ice.ErrWarn {
			return true // 查找失败
		} else {
			buf := bytes.NewBuffer(make([]byte, 0, ice.MOD_BUFS))
			defer func() { m.Echo(buf.String()) }()

			f.stdin, f.stdout = bytes.NewBuffer([]byte(msg.Result())), buf
		}

		f.count = 1
		f.scan(m, m.Cmdx(mdb.INSERT, SOURCE, "", mdb.HASH, kit.MDB_NAME, f.source), "")
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
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SOURCE: {Name: SOURCE, Help: "加载脚本", Value: kit.Data()},
		PROMPT: {Name: PROMPT, Help: "命令提示", Value: kit.Data(
			PS1, []interface{}{"\033[33;44m", kit.MDB_COUNT, "[", kit.MDB_TIME, "]", "\033[5m", TARGET, "\033[0m", "\033[44m", ">", "\033[0m ", "\033[?25h", "\033[32m"},
			PS2, []interface{}{kit.MDB_COUNT, " ", TARGET, "> "},
		)},
	}, Commands: map[string]*ice.Command{
		SOURCE: {Name: "source file", Help: "脚本解析", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.REPEAT: {Name: "repeat", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SCREEN, m.Option(kit.MDB_TEXT))
				m.ProcessInner()
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && kit.Ext(arg[0]) == ice.SHY {
				(&Frame{}).Start(m, arg...)
				return // 脚本解析
			}
		}},
		TARGET: {Name: "target name run:button", Help: "当前模块", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			f := c.Server().(*Frame)
			m.Search(arg[0]+ice.PT, func(p *ice.Context, s *ice.Context, key string) { f.target = s })
			f.prompt(m)
		}},
		PROMPT: {Name: "prompt arg run:button", Help: "命令提示", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			f := m.Optionv(FRAME).(*Frame)
			f.ps1 = arg
			f.prompt(m)
		}},
		PRINTF: {Name: "printf run:button text", Help: "输出显示", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			f := m.Optionv(FRAME).(*Frame)
			f.printf(m, arg[0])
		}},
		SCREEN: {Name: "screen run:button text", Help: "输出命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			f := m.Optionv(FRAME).(*Frame)
			for _, line := range kit.Split(arg[0], ice.NL, ice.NL) {
				fmt.Fprintf(f.pipe, line+ice.NL)
				f.printf(m, line+ice.NL)
				m.Sleep300ms()
			}
			m.Echo(f.res)
		}},
		RETURN: {Name: "return", Help: "结束脚本", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			f := m.Optionv(FRAME).(*Frame)
			f.Close(m, arg...)
		}},
	}})
}
