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
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func Render(msg *ice.Message, cmd string, args ...interface{}) string {
	switch arg := kit.Simple(args...); cmd {
	case ice.RENDER_VOID:
	case ice.RENDER_RESULT:
		// 转换结果
		if len(arg) > 0 {
			msg.Resultv(arg)
		}
		res := msg.Result()

		// 输出结果
		if fmt.Fprint(msg.O, res); !strings.HasSuffix(res, ice.NL) {
			fmt.Fprint(msg.O, ice.NL)
		}
		return res

	default:
		// 转换结果
		res := msg.Result()
		if res == "" {
			res = msg.Table().Result()
		}

		// 输出结果
		if fmt.Fprint(msg.O, res); !strings.HasSuffix(res, ice.NL) {
			fmt.Fprint(msg.O, ice.NL)
		}
		return res
	}
	return ""
}
func Script(m *ice.Message, name string) io.Reader {
	if strings.Contains(m.Option(ice.MSG_SCRIPT), "/") {
		name = path.Join(path.Dir(m.Option(ice.MSG_SCRIPT)), name)
	}
	m.Option(ice.MSG_SCRIPT, name)

	// 本地文件
	back := kit.Split(m.Option(nfs.DIR_ROOT))
	for i := len(back) - 1; i >= 0; i-- {
		if s, e := os.Open(path.Join(path.Join(back[:i]...), name)); e == nil {
			return s
		}
	}
	if s, e := os.Open(name); e == nil {
		return s
	}

	switch strings.Split(name, "/")[0] {
	case kit.SSH_ETC, kit.SSH_VAR:
		m.Warn(true, ice.ErrNotFound)
		return nil
	}

	// 打包文件
	if b, ok := ice.Info.BinPack["/"+name]; ok {
		m.Info("binpack %v %v", len(b), name)
		return bytes.NewReader(b)
	}

	// 远程文件
	if msg := m.Cmd("web.spide", "dev", "GET", path.Join("/share/local/", name)); msg.Result(0) != ice.ErrWarn {
		return bytes.NewBuffer([]byte(msg.Result()))
	}

	// 源码文件
	if strings.HasPrefix(name, kit.SSH_USR) {
		ls := strings.Split(name, "/")
		m.Cmd("web.code.git.repos", ls[1], path.Join(kit.SSH_USR, ls[1]))
		if s, e := os.Open(name); e == nil {
			return s
		}
	}
	return nil
}

type Frame struct {
	source string
	target *ice.Context
	stdout io.Writer
	stdin  io.Reader
	pipe   io.Writer

	count int
	last  string
	ps1   []string
	ps2   []string
}

func (f *Frame) prompt(m *ice.Message, list ...string) *Frame {
	if f.source != STDIO {
		return f
	}
	if len(list) == 0 {
		list = append(list, f.ps1...)
	}

	m.Sleep("30ms")
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
func (f *Frame) option(m *ice.Message, ls []string) []string {
	ln := []string{}
	m.Option(mdb.CACHE_LIMIT, 10)
	for i := 0; i < len(ls); i++ {
		if ls[i] == "--" {
			ln = append(ln, ls[i+1:]...)
			break
		}

		if strings.HasPrefix(ls[i], "-") {
			for j := i; j < len(ls); j++ {
				if j == len(ls)-1 || strings.HasPrefix(ls[j+1], "-") {
					if i < j {
						m.Option(ls[i][1:], ls[i+1:j+1])
					} else {
						m.Option(ls[i][1:], ice.TRUE)
					}
					i = j
					break
				}
			}
		} else {
			ln = append(ln, ls[i])
		}
	}
	return ln
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
		m.Spawn(f.target).Search(target+".", func(p *ice.Context, s *ice.Context, key string) {
			m.Info("choice: %s", s.Name)
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
	if strings.HasPrefix(line, "<") {
		fmt.Fprintf(m.O, line)
		return ""
	}

	for _, one := range kit.Split(line, ";", ";", ";") {
		async, one := false, strings.TrimSpace(one)
		if strings.TrimSuffix(one, "&") != one {
			async, one = true, strings.TrimSuffix(one, "&")
		}

		msg := m.Spawn(f.target)
		msg.Option("_cmd", one)

		ls := kit.Split(one)
		ls = f.alias(msg, ls)
		ls = f.change(msg, ls)
		ls = f.option(msg, ls)
		if len(ls) == 0 {
			continue
		}

		if async {
			msg.Gos(msg, func(msg *ice.Message) { msg.Cmd(ls[0], ls[1:]) })
			continue
		} else {
			msg.Cmdy(ls[0], ls[1:])
		}

		if strings.HasPrefix(msg.Result(), ice.ErrWarn) && m.Option(ice.MSG_RENDER) == ice.RENDER_RAW {
			fmt.Fprintf(msg.O, line)
			continue
		}

		// 渲染引擎
		_args, _ := msg.Optionv(ice.MSG_ARGS).([]interface{})
		f.last = Render(msg, msg.Option(ice.MSG_OUTPUT), _args...)
	}
	return ""
}
func (f *Frame) scan(m *ice.Message, h, line string) *Frame {
	m.Option(kit.Keycb(RETURN), func() { f.close() })
	f.ps1 = kit.Simple(m.Confv(PROMPT, kit.Keym(PS1)))
	f.ps2 = kit.Simple(m.Confv(PROMPT, kit.Keym(PS2)))
	ps := f.ps1

	m.Sleep("300ms")
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
				line += "\n"
			}
			continue // 空行
		}

		if strings.HasSuffix(bio.Text(), "\\") {
			line += bio.Text()[:len(bio.Text())-1]
			ps = f.ps2
			continue // 续行
		}
		if line += bio.Text(); strings.Count(line, "`")%2 == 1 {
			line += "\n"
			ps = f.ps2
			continue // 多行
		}
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			line = ""
			continue // 注释
		}
		if ps = f.ps1; f.stdout == os.Stdout {
			// 清空格式
			f.printf(m, "\033[0m")
		}
		line = f.parse(m, line)
	}
	return f
}
func (f *Frame) close() {
	if stdin, ok := f.stdin.(io.Closer); ok {
		stdin.Close()
		f.stdin = nil
	}
}

func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	return f
}
func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	f.source, f.target = kit.Select(STDIO, arg, 0), m.Target()

	switch m.Cap(ice.CTX_STREAM, f.source) {
	case STDIO: // 终端交互
		r, w, _ := os.Pipe()
		m.Go(func() { io.Copy(w, os.Stdin) })
		f.stdin, f.stdout = r, os.Stdout
		f.pipe = w

		m.Option(ice.MSG_OPTS, ice.MSG_USERNAME)
		aaa.UserRoot(m)

	default: // 脚本文件
		f.target = m.Source()

		if strings.HasPrefix(f.source, "/dev") {
			f.stdin, f.stdout = m.I, m.O
			break
		}

		buf := bytes.NewBuffer(make([]byte, 0, ice.MOD_BUFS))
		defer func() { m.Echo(buf.String()) }()

		if s := Script(m, f.source); s != nil {
			f.stdin, f.stdout = s, buf
			break
		}

		// 查找失败
		return true
	}

	// 解析脚本
	if f.count = 1; f.source == STDIO {
		m.Conf(SOURCE, kit.Keys(kit.MDB_HASH, STDIO, kit.MDB_META, kit.MDB_NAME), STDIO)
		m.Conf(SOURCE, kit.Keys(kit.MDB_HASH, STDIO, kit.MDB_META, kit.MDB_TIME), m.Time())

		f.count = kit.Int(m.Conf(SOURCE, kit.Keys(kit.MDB_HASH, STDIO, kit.Keym(kit.MDB_COUNT)))) + 1
		f.scan(m, STDIO, "")
	} else {
		h := m.Cmdx(mdb.INSERT, SOURCE, "", mdb.HASH, kit.MDB_NAME, f.source)
		m.Conf(SOURCE, kit.Keys(kit.MDB_HASH, h, kit.Keym(kit.MDB_COUNT)), 0)
		m.Conf(SOURCE, kit.Keys(kit.MDB_HASH, h, kit.MDB_LIST), "")

		f.scan(m, h, "")
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

const (
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
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SOURCE: {Name: SOURCE, Help: "加载脚本", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
			PROMPT: {Name: PROMPT, Help: "命令提示", Value: kit.Data(
				PS1, []interface{}{"\033[33;44m", kit.MDB_COUNT, "[", kit.MDB_TIME, "]", "\033[5m", TARGET, "\033[0m", "\033[44m", ">", "\033[0m ", "\033[?25h", "\033[32m"},
				PS2, []interface{}{kit.MDB_COUNT, " ", TARGET, "> "},
			)},
		},
		Commands: map[string]*ice.Command{
			SOURCE: {Name: "source hash id limit offend auto", Help: "脚本解析", Action: map[string]*ice.Action{
				mdb.REPEAT: {Name: "repeat", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SCREEN, m.Option(kit.MDB_TEXT))
					m.ProcessInner()
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 && kit.Ext(arg[0]) == "shy" { // 解析脚本
					m.Starts(strings.Replace(arg[0], ".", "_", -1), arg[0], arg[0:]...)
					return
				}

				if len(arg) == 0 { // 脚本列表
					m.Fields(len(arg), "time,hash,name,count")
					m.Cmdy(mdb.SELECT, SOURCE, "", mdb.HASH)
					m.Sort(kit.MDB_NAME)
					return
				}

				if m.Option(mdb.CACHE_OFFEND, kit.Select("0", arg, 3)); arg[0] == STDIO {
					m.Option(mdb.CACHE_LIMIT, kit.Select("10", arg, 2))
				} else {
					m.Option(mdb.CACHE_LIMIT, kit.Select("-1", arg, 2))
				}

				// 命令列表
				m.Fields(len(arg[1:]), "time,id,text")
				m.Cmdy(mdb.SELECT, SOURCE, kit.Keys(kit.MDB_HASH, arg[0]), mdb.LIST, kit.MDB_ID, arg[1:])
				m.PushAction(mdb.REPEAT)
			}},
			TARGET: {Name: "target name 执行:button", Help: "当前模块", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				f := m.Target().Server().(*Frame)
				m.Search(arg[0]+".", func(p *ice.Context, s *ice.Context, key string) { f.target = s })
				f.prompt(m)
			}},
			PROMPT: {Name: "prompt arg 执行:button", Help: "命令提示", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				f := m.Target().Server().(*Frame)
				f.ps1 = arg
				f.prompt(m)
			}},
			PRINTF: {Name: "printf 执行:button text:textarea", Help: "输出显示", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				f := m.Target().Server().(*Frame)
				f.printf(m, arg[0])
			}},
			SCREEN: {Name: "screen 执行:button text:textarea", Help: "输出命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				f := m.Target().Server().(*Frame)
				for _, line := range kit.Split(arg[0], "\n", "\n", "\n") {
					f.printf(m, line+"\n")
					fmt.Fprintf(f.pipe, line+"\n")
					m.Sleep("300ms")
				}
				m.Echo(f.last)
			}},
			RETURN: {Name: "return", Help: "结束脚本", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch cb := m.Optionv(kit.Keycb(RETURN)).(type) {
				case func():
					cb()
				}
			}},
		},
	})
}
