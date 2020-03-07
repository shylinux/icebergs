package ssh

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Frame struct {
	in  io.ReadCloser
	out io.Writer

	target *ice.Context
	count  int
	exit   bool
}

func (f *Frame) prompt(m *ice.Message) *Frame {
	if f.out == os.Stdout {
		for _, v := range kit.Simple(m.Optionv(ice.MSG_PROMPT)) {
			switch v {
			case "count":
				fmt.Fprintf(f.out, kit.Format("%d", f.count))
			case "time":
				fmt.Fprintf(f.out, m.Time("15:04:05"))
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
		fmt.Fprintf(f.out, res, arg)
	} else {
		fmt.Fprint(f.out, res)
	}
	// if !strings.HasSuffix(res, "\n") {
	// 	fmt.Fprint(f.out, "\n")
	// }
	return f
}
func (f *Frame) parse(m *ice.Message, line string) *Frame {
	for _, one := range kit.Split(line, ";") {
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

		// 执行命令
		msg.Cmdy(ln[0], ln[1:])

		// 转换结果
		res := msg.Result()
		if res == "" {
			res = msg.Table().Result()
		}

		// 输出结果
		if f.printf(msg, res); !strings.HasSuffix(res, "\n") {
			f.printf(msg, "\n")
		}
	}
	return f
}

func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	m.Target().Configs["history"] = &ice.Config{Name: "history", Help: "历史", Value: kit.Data()}
	return f
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	m.Option(ice.MSG_PROMPT, m.Confv("prompt", "meta.PS1"))
	f.count, f.target = kit.Int(m.Conf("history", "meta.count"))+1, m.Source()

	switch kit.Select("stdio", arg, 0) {
	case "stdio":
		// 解析终端
		f.in, f.out = os.Stdin, os.Stdout
		m.Cap(ice.CTX_STREAM, "stdio")
		f.target = m.Target()
	default:
		// 解析脚本
		if s, e := os.Open(arg[0]); m.Warn(e != nil, "%s", e) {
			return true
		} else {
			defer s.Close()
			if f.in, f.out = s, os.Stdout; m.Optionv(ice.MSG_STDOUT) != nil {
				f.out = m.Optionv(ice.MSG_STDOUT).(io.Writer)
			}
			m.Cap(ice.CTX_STREAM, arg[0])
		}
	}

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

		m.Grow("history", nil, kit.Dict("line", line))
		m.Option(ice.MSG_PROMPT, m.Confv("prompt", "meta.PS1"))

		if f.out == os.Stdout {
			f.printf(m, "\033[0m")
		}
		f.parse(m, line)
		m.Cost("stdin: %v", line)
		f.count++
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

			f := m.Target().Server().(*Frame)
			m.Grows("history", nil, "id", arg[0], func(index int, value map[string]interface{}) {
				f.parse(m, kit.Format(value["line"]))
			})
		}},
		"source": {Name: "source file", Help: "解析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			buf := bytes.NewBuffer(make([]byte, 0, 4096))
			m.Optionv(ice.MSG_STDOUT, buf)

			m.Starts(strings.Replace(arg[0], ".", "_", -1), arg[0], arg[0:]...)
			m.Echo(buf.String())
		}},
		"print": {Name: "print", Help: "解析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			f := m.Target().Server().(*Frame)
			f.printf(m, arg[0])
			f.printf(m, arg[0])
			f.printf(m, arg[0])
			f.printf(m, arg[0])
		}},
		"prompt": {Name: "print", Help: "解析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(ice.MSG_PROMPT, m.Confv("prompt", "meta.PS1"))
			f := m.Target().Server().(*Frame)
			f.prompt(m)
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
		"return": {Name: "return", Help: "解析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(ice.MSG_PROMPT, m.Confv("prompt", "meta.PS1"))
			f := m.Target().Server().(*Frame)
			f.exit = true
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
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
