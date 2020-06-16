package ssh

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

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
	in  io.ReadCloser
	out io.Writer

	target *ice.Context
	exit   bool
}

func Render(msg *ice.Message, cmd string, args ...interface{}) {
	msg.Log(ice.LOG_EXPORT, "%s: %v", cmd, args)
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
			res = msg.Table().Result()
		}

		// 输出结果
		if fmt.Fprintf(msg.O, res); !strings.HasSuffix(res, "\n") {
			fmt.Fprintf(msg.O, "\n")
		}
	}
	msg.Append(ice.MSG_OUTPUT, ice.RENDER_OUTPUT)
}

func (f *Frame) prompt(m *ice.Message) *Frame {
	if f.out == os.Stdout {
		fmt.Fprintf(f.out, "\r")
		for _, v := range kit.Simple(m.Optionv(ice.MSG_PROMPT)) {
			switch v {
			case "count":
				fmt.Fprintf(f.out, "%d", kit.Int(m.Conf("history", "meta.count"))+1)
			case "time":
				fmt.Fprintf(f.out, time.Now().Format("15:04:05"))
			case "target":
				fmt.Fprintf(f.out, f.target.Name)
			default:
				fmt.Fprintf(f.out, v)
			}
		}
	}
	return f
}
func (f *Frame) printf(m *ice.Message, res string, arg ...interface{}) *Frame {
	if len(arg) > 0 {
		fmt.Fprintf(f.out, res, arg...)
	} else {
		fmt.Fprint(f.out, res)
	}
	// if !strings.HasSuffix(res, "\n") {
	// 	fmt.Fprint(f.out, "\n")
	// }
	return f
}
func (f *Frame) parse(m *ice.Message, line string) *Frame {
	for _, one := range kit.Split(line, ";", ";", ";") {
		ls := kit.Split(one)
		m.Log(ice.LOG_IMPORT, "stdin: %d %v", len(ls), ls)

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
		if len(ls) == 0 {
			continue
		}
		if m.Option("scan_mode") == "scan" {
			f.printf(m, ls[0])
			f.printf(m, "`")
			f.printf(m, strings.Join(ls[1:], "` `"))
			f.printf(m, "`")
			f.printf(m, "\n")
			continue
		}

		// 命令替换
		if alias, ok := m.Optionv(ice.MSG_ALIAS).(map[string]interface{}); ok {
			if a := kit.Simple(alias[ls[0]]); len(a) > 0 {
				ls = append(append([]string{}, a...), ls[1:]...)
			}
		}

		// 解析选项
		ln := []string{}
		msg := m.Spawns(f.target)
		msg.Option("cache.limit", 10)
		for i := 0; i < len(ls); i++ {
			if ls[i] == "--" {
				ln = append(ln, ls[i+1:]...)
				break
			} else if strings.HasPrefix(ls[i], "-") {
				for j := i; j < len(ls); j++ {
					if j == len(ls)-1 || strings.HasPrefix(ls[j+1], "-") {
						if i == j {
							msg.Option(ls[i][1:], "true")
						} else {
							msg.Option(ls[i][1:], ls[i+1:j+1])
						}
						i = j
						break
					}
				}
			} else {
				ln = append(ln, ls[i])
			}
		}

		if ln[0] == "qrcode" {
			msg.Option(ice.MSG_OUTPUT, ice.RENDER_QRCODE)
			ln = ln[1:]
		}
		msg.Option("_cmd", one)

		// 执行命令
		msg.Cmdy(ln[0], ln[1:])

		// 渲染引擎
		_args, _ := msg.Optionv(ice.MSG_ARGS).([]interface{})
		Render(msg, msg.Option(ice.MSG_OUTPUT), _args...)

	}
	return f
}

func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	return f
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	m.Option(ice.MSG_PROMPT, m.Confv("prompt", "meta.PS1"))
	f.target = m.Source()

	switch kit.Select("stdio", arg, 0) {
	case "stdio":
		// 解析终端
		f.in, f.out = os.Stdin, os.Stdout
		m.Cap(ice.CTX_STREAM, "stdio")
		f.target = m.Target()
	default:
		if s, e := os.Open(arg[0]); m.Warn(e != nil, "%s", e) {
			// 打开失败
			return true
		} else {
			// 解析脚本
			defer s.Close()
			if f.in, f.out = s, os.Stdout; m.Optionv(ice.MSG_STDOUT) != nil {
				f.out = m.Optionv(ice.MSG_STDOUT).(io.Writer)
			}
			m.Cap(ice.CTX_STREAM, arg[0])
		}
	}
	m.I, m.O = f.in, f.out

	line := ""
	bio := bufio.NewScanner(f.in)
	for f.prompt(m); bio.Scan() && !f.exit; f.prompt(m) {
		if len(bio.Text()) == 0 {
			// 空行
			continue
		}
		if strings.HasSuffix(bio.Text(), "\\") {
			// 续行
			m.Option(ice.MSG_PROMPT, m.Confv("prompt", "meta.PS2"))
			line += bio.Text()[:len(bio.Text())-1]
			continue
		}
		if line += bio.Text(); strings.Count(line, "`")%2 == 1 {
			// 多行
			m.Option(ice.MSG_PROMPT, m.Confv("prompt", "meta.PS2"))
			line += "\n"
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			// 注释
			line = ""
			continue
		}

		if strings.HasPrefix(strings.TrimSpace(line), "!") {
			if len(line) == 1 {
				// 历史记录
				f.printf(m, m.Cmd("history").Table().Result())
				line = ""
				continue
			}
			if i, e := strconv.Atoi(line[1:]); e == nil {
				// 历史命令
				m.Grows("history", nil, "id", kit.Format(i), func(index int, value map[string]interface{}) {
					line = kit.Format(value["line"])
				})
			} else {
				f.printf(m, m.Cmd("history", "search", line[1:]).Table().Result())
				line = ""
				continue
			}
		} else {
			// 记录历史
			m.Grow("history", nil, kit.Dict("line", line))
		}

		// 清空格式
		if m.Option(ice.MSG_PROMPT, m.Confv("prompt", "meta.PS1")); f.out == os.Stdout {
			f.printf(m, "\033[0m")
		}
		if m.Cap(ice.CTX_STREAM) == "stdio" {
			m.Cmd(ice.WEB_FAVOR, "cmd.history", "cmd", kit.Select("stdio", arg, 0), line)
		}

		// 解析命令
		f.parse(m, line)
		m.Cost("stdin: %v", line)
		line = ""
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

var Index = &ice.Context{Name: "ssh", Help: "终端模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"history": {Name: "history", Help: "history", Value: kit.Data("limit", "200", "least", "100")},
		"prompt": {Name: "prompt", Help: "prompt", Value: kit.Data(
			"PS1", []interface{}{"\033[33;44m", "count", "[", "time", "]", "\033[5m", "target", "\033[0m", "\033[44m", ">", "\033[0m ", "\033[?25h", "\033[32m"},
			"PS2", []interface{}{"count", " ", "target", "> "},
		)},

		"super": {Name: "super", Help: "super", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("history")

			if f, ok := m.Target().Server().(*Frame); ok {
				// 关闭终端
				f.in.Close()
				m.Done()
			}
		}},

		"history": {Name: "history", Help: "历史", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Grows("history", nil, "", "", func(index int, value map[string]interface{}) {
					m.Push("", value, []string{"id", "time", "line"})
				})
				return
			}

			switch arg[0] {
			case "search":
				m.Grows("history", nil, "", "", func(index int, value map[string]interface{}) {
					if strings.Contains(kit.Format(value["line"]), arg[1]) {
						m.Push("", value, []string{"id", "time", "line"})
					}
				})
			default:
				m.Grows("history", nil, "id", arg[0], func(index int, value map[string]interface{}) {
					m.Push("", value, []string{"id", "time", "line"})
				})
			}
		}},
		"return": {Name: "return", Help: "解析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(ice.MSG_PROMPT, m.Confv("prompt", "meta.PS1"))
			f := m.Target().Server().(*Frame)
			f.exit = true
		}},
		"target": {Name: "target", Help: "目标", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			f := m.Target().Server().(*Frame)
			m.Search(arg[0], func(p *ice.Context, s *ice.Context, key string) {
				f.target = s
			})
			f.prompt(m)
		}},
		"source": {Name: "source file", Help: "解析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			buf := bytes.NewBuffer(make([]byte, 0, 4096))
			m.Optionv(ice.MSG_STDOUT, buf)

			m.Starts(strings.Replace(arg[0], ".", "_", -1), arg[0], arg[0:]...)
			m.Echo(buf.String())
		}},
		"show": {Name: "show", Help: "解析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			f, e := os.Open("usr/local/what/hi.shy")
			m.Assert(e)

			bio := bufio.NewScanner(f)
			for bio.Scan() {
				ls := kit.Split(bio.Text())
				m.Echo("%d: %v\n", len(ls), ls)
				m.Info("%v", ls)
			}
		}},

		"super": {Name: "super user remote port local", Help: "上位机", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			key := m.Rich("super", nil, kit.Dict(
				"user", arg[0], "remote", arg[1], "port", arg[2], "local", arg[3],
			))
			m.Echo(key)
			m.Info(key)

			m.Gos(m, func(m *ice.Message) {
				for {
					m.Cmd(ice.CLI_SYSTEM, "ssh", "-CNR", kit.Format("%s:%s:22", arg[2], arg[3]), kit.Format("%s@%s", arg[0], arg[1]))
					m.Info("reconnect after 10s")
					time.Sleep(time.Second * 10)
				}
			})
		}},

		"what": {Name: "what", Help: "上位机", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			ls := kit.Split("window:=auto", " ", ":=")
			m.Echo("%v %v", len(ls), ls)
		}},
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
