package ssh

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Frame struct {
	in     io.ReadCloser
	out    io.WriteCloser
	target *ice.Context
	count  int
}

func (f *Frame) prompt(m *ice.Message) *ice.Message {
	prompt := "[15:04:05]%s> "
	fmt.Fprintf(f.out, kit.Format("%d", f.count))
	fmt.Fprintf(f.out, m.Time(prompt, f.target.Name))
	return m
}
func (f *Frame) printf(m *ice.Message, res string) *ice.Message {
	fmt.Fprint(f.out, res)
	if !strings.HasSuffix(res, "\n") {
		fmt.Fprint(f.out, "\n")
	}
	return m
}

func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	return f
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	switch kit.Select("stdio", arg, 0) {
	case "stdio":
		f.in = os.Stdin
		f.out = os.Stdout
		m.Cap(ice.CTX_STREAM, "stdio")
	default:
		if n, e := os.Open(arg[0]); m.Warn(e != nil, "%s", e) {
			return true
		} else {
			f.in = n
			f.out = os.Stderr
			m.Cap(ice.CTX_STREAM, arg[0])
		}
	}

	f.count = 0
	f.target = m.Target()
	bio := bufio.NewScanner(f.in)
	for f.prompt(m); bio.Scan(); f.prompt(m) {
		ls := kit.Split(bio.Text())
		m.Log(ice.LOG_IMPORT, "stdin: %v", ls)

		if len(ls) > 0 && strings.HasPrefix(ls[0], "~") {
			// 切换模块
			target := ls[0][1:]
			if ls = ls[1:]; len(target) == 0 {
				target, ls = ls[0], ls[1:]
			}
			ice.Pulse.Search(target+".", func(p *ice.Context, s *ice.Context, key string) {
				m.Info("choice: %s", s.Name)
				f.target = s
			})
		}

		if len(ls) == 0 {
			continue
		}

		// 执行命令
		msg := m.Spawns(f.target)
		if msg.Cmdy(ls); !msg.Hand {
			msg = msg.Set("result").Cmdy(ice.CLI_SYSTEM, ls)
		}

		// 生成结果
		res := msg.Result()
		if res == "" {
			res = msg.Table().Result()
		}

		// 输出结果
		msg.Cost("stdin:%v", ls)
		f.printf(msg, res)
		f.count++
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

var Index = &ice.Context{Name: "ssh", Help: "终端模块",
	Caches:  map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, ok := m.Target().Server().(*Frame); ok {
				// 关闭终端
				f.in.Close()
				m.Done()
			}
		}},
		"scan": {Name: "scan", Help: "解析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Starts(arg[0], arg[1], arg[2:]...)
		}},
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
