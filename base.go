package ice

import (
	"github.com/shylinux/toolkits"

	"os"
	"strings"
	"sync"
	"time"
)

type Frame struct {
	code int
}

func (f *Frame) Spawn(m *Message, c *Context, arg ...string) Server {
	return &Frame{}
}
func (f *Frame) Begin(m *Message, arg ...string) Server {
	m.Log(LOG_BEGIN, "ice")
	list := map[*Context]*Message{m.target: m}
	m.Travel(func(p *Context, s *Context) {
		s.root = m.target
		if msg, ok := list[p]; ok && msg != nil {
			sub := msg.Spawns(s)
			s.Begin(sub, arg...)
			list[s] = sub
		}
	})
	m.target.wg = &sync.WaitGroup{}
	m.root.Cost("begin")
	return f
}
func (f *Frame) Start(m *Message, arg ...string) bool {
	m.Log(LOG_START, "ice")
	m.Cmd(ICE_INIT).Cmd("init", arg)
	m.root.Cost("start")
	return true
}
func (f *Frame) Close(m *Message, arg ...string) bool {
	m.target.wg.Wait()
	list := map[*Context]*Message{m.target: m}
	m.Travel(func(p *Context, s *Context) {
		if msg, ok := list[p]; ok && msg != nil {
			sub := msg.Spawns(s)
			s.Close(sub, arg...)
			list[s] = sub
		}
	})
	return true
}

var Index = &Context{Name: "ice", Help: "冰山模块",
	Caches: map[string]*Cache{
		CTX_STATUS: {Value: "begin"},
		CTX_STREAM: {Value: "shy"},
	},
	Configs: map[string]*Config{
		"table": {Name: "数据缓存", Value: map[string]interface{}{
			"space":   " ",
			"col_sep": " ",
			"row_sep": "\n",
			"compact": "false",
		}},
		"help": {Value: map[string]interface{}{
			"index": []interface{}{
				"^_^      欢迎使用冰山框架       ^_^",
				"^_^  Welcome to Icebergs world  ^_^",
				"",
				"Meet: shylinuxc@gmail.com",
				"More: https://shylinux.com",
				"More: https://github.com/shylinux/icebergs",
				"",
			},
		}},
	},
	Commands: map[string]*Command{
		ICE_INIT: {Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.Travel(func(p *Context, c *Context) {
				if cmd, ok := c.Commands[ICE_INIT]; ok && p != nil {
					c.Run(m.Spawns(c), cmd, ICE_INIT, arg...)
				}
			})
		}},
		"init": {Name: "init", Help: "hello", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.root.Cost("_init")
			m.Start("log", arg...)
			m.Start("gdb", arg...)
			m.Start("ssh", arg...)
			m.Cmd(arg)
		}},
		"help": {Name: "help", Help: "帮助", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.Echo(strings.Join(kit.Simple(m.Confv("help", "index")), "\n"))
		}},
		"exit": {Name: "exit", Help: "hello", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			f := m.root.target.server.(*Frame)
			f.code = kit.Int(kit.Select("0", arg, 0))
			m.root.Cmd(ICE_EXIT)
		}},
		ICE_EXIT: {Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.root.Travel(func(p *Context, c *Context) {
				if cmd, ok := c.Commands[ICE_EXIT]; ok && p != nil {
					m.TryCatch(m.Spawns(c), true, func(msg *Message) {
						c.Run(msg, cmd, ICE_EXIT, arg...)
					})
				}
			})
		}},
	},
}

var Pulse = &Message{
	time: time.Now(), code: 0,
	meta: map[string][]string{},
	data: map[string]interface{}{},

	messages: []*Message{}, message: nil, root: nil,
	source: Index, target: Index, Hand: true,
}

var Log func(*Message, string, string)

func Run(arg ...string) string {
	Index.root = Index
	Pulse.root = Pulse

	if len(arg) == 0 {
		arg = os.Args[1:]
	}
	if len(arg) == 0 {
		arg = append(arg, WEB_SERVE)
	}

	frame := &Frame{}
	Index.server = frame
	Pulse.Option("begin_time", Pulse.Time())

	if frame.Begin(Pulse.Spawns(), arg...).Start(Pulse.Spawns(), arg...) {
		frame.Close(Pulse.Spawns(), arg...)
	}
	time.Sleep(time.Second)
	os.Exit(frame.code)
	return Pulse.Result()
}
