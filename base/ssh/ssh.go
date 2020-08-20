package ssh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type Frame struct {
	source string
	target *ice.Context
	stdout io.Writer

	count int
	ps1   []string
	ps2   []string

	exit bool
}

func Render(msg *ice.Message, cmd string, args ...interface{}) {
	defer func() { msg.Log_EXPORT(mdb.RENDER, cmd, kit.MDB_TEXT, args) }()

	switch arg := kit.Simple(args...); cmd {
	case ice.RENDER_OUTPUT:

	case ice.RENDER_DOWNLOAD:
		os.Link(kit.Select(path.Base(arg[0]), arg, 2), arg[0])

	case ice.RENDER_RESULT:
		fmt.Fprintf(msg.O, msg.Result())

	case ice.RENDER_QRCODE:
		if len(args) > 0 {
			fmt.Println(msg.Cmdx("cli.python", "qrcode", kit.Format(args[0], args[1:]...)))
		} else {
			fmt.Println(msg.Cmdx("cli.python", "qrcode", kit.Format(kit.Dict(
				kit.MDB_TYPE, "cmd", kit.MDB_NAME, msg.Option("_cmd"), kit.MDB_TEXT, strings.TrimSpace(msg.Result()),
			))))
		}
	default:
		// 转换结果
		res := msg.Result()
		if res == "" {
			res = msg.Table(nil).Result()
		}
		args = append(args, "length:", len(res))

		// 输出结果
		if fmt.Fprintf(msg.O, res); !strings.HasSuffix(res, "\n") {
			fmt.Fprintf(msg.O, "\n")
		}
	}
	msg.Append(ice.MSG_OUTPUT, ice.RENDER_OUTPUT)
}

func (f *Frame) history(m *ice.Message, line string) string {
	favor := m.Conf(SOURCE, kit.Keys(kit.MDB_META, web.FAVOR))
	if strings.HasPrefix(strings.TrimSpace(line), "!!") {
		if len(line) == 2 {
			line = m.Cmd(web.FAVOR, favor).Append(kit.MDB_TEXT)
		}
	} else if strings.HasPrefix(strings.TrimSpace(line), "!") {
		if len(line) == 1 {
			// 历史记录
			msg := m.Cmd(web.FAVOR, favor)
			msg.Sort(kit.MDB_ID)
			msg.Appendv(ice.MSG_APPEND, kit.MDB_TIME, kit.MDB_ID, kit.MDB_TEXT)
			f.printf(m, msg.Table().Result())
			return ""
		}
		if i, e := strconv.Atoi(line[1:]); e == nil {
			// 历史命令
			line = kit.Format(kit.Value(m.Cmd(web.FAVOR, favor, i).Optionv("value"), kit.MDB_TEXT))
		} else {
			f.printf(m, m.Cmd("history", "search", line[1:]).Table().Result())
			return ""
		}
	} else if strings.TrimSpace(line) != "" && f.source == STDIO {
		// 记录历史
		m.Cmd(web.FAVOR, favor, "cmd", f.source, line)
	}
	return line
}
func (f *Frame) printf(m *ice.Message, res string, arg ...interface{}) *Frame {
	if len(arg) > 0 {
		fmt.Fprintf(f.stdout, res, arg...)
	} else {
		fmt.Fprint(f.stdout, res)
	}
	return f
}
func (f *Frame) prompt(m *ice.Message, list ...string) *Frame {
	if f.stdout != os.Stdout {
		return f
	}
	if len(list) == 0 {
		list = append(list, f.ps1...)
	}

	fmt.Fprintf(f.stdout, "\r")
	for _, v := range list {
		switch v {
		case "count":
			fmt.Fprintf(f.stdout, "%d", kit.Int(m.Conf("history", "meta.count"))+1)
		case "time":
			fmt.Fprintf(f.stdout, time.Now().Format("15:04:05"))
		case "target":
			fmt.Fprintf(f.stdout, f.target.Name)
		default:
			fmt.Fprintf(f.stdout, v)
		}
	}
	return f
}
func (f *Frame) option(m *ice.Message, ls []string) []string {
	// 解析选项
	ln := []string{}
	m.Option("cache.limit", 10)
	for i := 0; i < len(ls); i++ {
		if ls[i] == "--" {
			ln = append(ln, ls[i+1:]...)
			break
		} else if strings.HasPrefix(ls[i], "-") {
			for j := i; j < len(ls); j++ {
				if j == len(ls)-1 || strings.HasPrefix(ls[j+1], "-") {
					if i == j {
						m.Option(ls[i][1:], "true")
					} else {
						m.Option(ls[i][1:], ls[i+1:j+1])
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
	if len(ls) == 1 && ls[0] == "~" {
		// 模块列表
		ls = []string{"context"}
	} else if len(ls) > 0 && strings.HasPrefix(ls[0], "~") {
		// 切换模块
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
	// 命令替换
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
		m.Log_IMPORT("stdin", one, "length", len(one))

		ls := kit.Split(one)
		if m.Option("scan_mode") == "scan" {
			f.printf(m, ls[0])
			f.printf(m, "`")
			f.printf(m, strings.Join(ls[1:], "` `"))
			f.printf(m, "`")
			f.printf(m, "\n")
			continue
		}

		// 解析引擎
		msg := m.Spawns(f.target)
		msg.Option("_cmd", one)
		ls = f.alias(msg, ls)
		ls = f.change(msg, ls)
		ls = f.option(msg, ls)
		if len(ls) == 0 {
			continue
		}

		if strings.HasPrefix(line, "<") {
			msg.Resultv(line)
		} else if msg.Cmdy(ls[0], ls[1:]); strings.HasPrefix(msg.Result(), "warn: ") && m.Option("render") == "raw" {
			msg.Resultv(line)
		}

		// 渲染引擎
		_args, _ := msg.Optionv(ice.MSG_ARGS).([]interface{})
		Render(msg, msg.Option(ice.MSG_OUTPUT), _args...)
	}
	return ""
}
func (f *Frame) scan(m *ice.Message, file, line string, r io.Reader) *Frame {
	f.ps1 = kit.Simple(m.Confv("prompt", "meta.PS1"))
	f.ps2 = kit.Simple(m.Confv("prompt", "meta.PS2"))
	ps := f.ps1

	m.I, m.O = r, f.stdout
	bio := bufio.NewScanner(r)
	for f.prompt(m, ps...); bio.Scan() && !f.exit; f.prompt(m, ps...) {
		if len(bio.Text()) == 0 {
			// 空行
			continue
		}
		if strings.HasSuffix(bio.Text(), "\\") {
			// 续行
			line += bio.Text()[:len(bio.Text())-1]
			ps = f.ps2
			continue
		}
		if line += bio.Text(); strings.Count(line, "`")%2 == 1 {
			// 多行
			line += "\n"
			ps = f.ps2
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			// 注释
			line = ""
			continue
		}
		if line = f.history(m, line); line == "" {
			// 历史命令
			continue
		}
		if ps = f.ps1; f.stdout == os.Stdout {
			// 清空格式
			f.printf(m, "\033[0m")
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
	var r io.Reader
	switch kit.Select(STDIO, arg, 0) {
	case STDIO:
		// 终端交互
		f.source = STDIO
		r, f.stdout = os.Stdin, os.Stdout
		m.Cap(ice.CTX_STREAM, STDIO)
		f.target = m.Target()
		m.Option("_option", ice.MSG_USERNAME)
		m.Option(ice.MSG_USERNAME, cli.UserName)
		m.Option(ice.MSG_USERROLE, aaa.ROOT)
		m.Option(ice.MSG_USERZONE, "boot")
		aaa.UserRoot(m)
	default:
		if b, ok := ice.BinPack[arg[0]]; ok {
			m.Debug("binpack %v %v", arg[0], len(b))
			buf := bytes.NewBuffer(make([]byte, 0, 4096))
			defer func() { m.Echo(buf.String()) }()

			// 脚本解析
			f.source = arg[0]
			r, f.stdout = bytes.NewReader(b), buf
			m.Cap(ice.CTX_STREAM, arg[0])
			f.target = m.Source()
			break
		}
		if s, e := os.Open(arg[0]); !m.Warn(e != nil, "%s", e) {
			defer s.Close()

			buf := bytes.NewBuffer(make([]byte, 0, 4096))
			defer func() { m.Echo(buf.String()) }()

			// 脚本解析
			f.source = arg[0]
			r, f.stdout = s, buf
			m.Cap(ice.CTX_STREAM, arg[0])
			f.target = m.Source()
			break
		}
		return true
	}

	f.scan(m, kit.Select(STDIO, arg, 0), "", r)
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

const (
	SOURCE = "source"
	TARGET = "target"
	PROMPT = "prompt"
	QRCODE = "qrcode"
	RETURN = "return"
	REMOTE = "remote"
)
const (
	STDIO = "stdio"
)

var Index = &ice.Context{Name: "ssh", Help: "终端模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		SOURCE: {Name: "prompt", Help: "命令提示", Value: kit.Data(
			web.FAVOR, "cmd.history",
		)},
		PROMPT: {Name: "prompt", Help: "命令提示", Value: kit.Data(
			"PS1", []interface{}{"\033[33;44m", "count", "[", "time", "]", "\033[5m", "target", "\033[0m", "\033[44m", ">", "\033[0m ", "\033[?25h", "\033[32m"},
			"PS2", []interface{}{"count", " ", "target", "> "},
		)},
		REMOTE: {Name: "remote", Help: "远程连接", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if _, ok := m.Target().Server().(*Frame); ok {
				m.Done()
			}
		}},

		SOURCE: {Name: "source file", Help: "脚本解析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if _, e := os.Stat(arg[0]); e != nil {
				arg[0] = path.Join(path.Dir(m.Option("_script")), arg[0])
			}
			m.Option("_script", arg[0])
			m.Starts(strings.Replace(arg[0], ".", "_", -1), arg[0], arg[0:]...)
		}},
		TARGET: {Name: "target name", Help: "当前模块", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Search(arg[0], func(p *ice.Context, s *ice.Context, key string) {
				f := m.Target().Server().(*Frame)
				f.target = s
				f.prompt(m)
			})
		}},
		PROMPT: {Name: "prompt arg...", Help: "命令提示", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			f := m.Target().Server().(*Frame)
			f.ps1 = arg
			f.prompt(m)
		}},
		QRCODE: {Name: "qrcode arg...", Help: "命令提示", Action: map[string]*ice.Action{
			"json": {Name: "json [key val]...", Help: "json", Hand: func(m *ice.Message, arg ...string) {
				val := map[string]interface{}{}
				for i := 0; i < len(arg)-1; i += 2 {
					kit.Value(val, arg[i], arg[i+1])
				}
				f := m.Target().Server().(*Frame)
				f.printf(m, m.Cmdx(cli.PYTHON, "qrcode", kit.Format(val)))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			f := m.Target().Server().(*Frame)
			f.printf(m, m.Cmdx(cli.PYTHON, "qrcode", strings.Join(arg, "")))
		}},
		RETURN: {Name: "return", Help: "结束脚本", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			f := m.Target().Server().(*Frame)
			f.exit = true
		}},

		REMOTE: {Name: "remote user remote port local", Help: "远程连接", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			key := m.Rich(REMOTE, nil, kit.Dict(
				"user", arg[0], "remote", arg[1], "port", arg[2], "local", arg[3],
			))
			m.Echo(key)
			m.Info(key)

			m.Gos(m, func(m *ice.Message) {
				for {
					m.Cmd(cli.SYSTEM, "ssh", "-CNR", kit.Format("%s:%s:22", arg[2], kit.Select("localhost", arg, 3)),
						kit.Format("%s@%s", arg[0], arg[1]))
					m.Info("reconnect after 10s")
					time.Sleep(time.Second * 10)
				}
			})
		}},
	},
}

func init() { ice.Index.Register(Index, &Frame{}, SOURCE, QRCODE) }
