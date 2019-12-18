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
	list := map[*Context]*Message{m.target: m}
	m.Travel(func(p *Context, s *Context) {
		if msg, ok := list[p]; ok && msg != nil {
			sub := msg.Spawns(s)
			s.Begin(sub, arg...)
			list[s] = sub
		}
	})
	m.target.wg = &sync.WaitGroup{}
	return f
}
func (f *Frame) Start(m *Message, arg ...string) bool {
	m.Cmd("_init").Cmd("init", arg)
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
		"status": {Value: "begin"},
		"stream": {Value: "shy"},
	},
	Configs: map[string]*Config{
		"table": {Name: "数据缓存", Value: map[string]interface{}{
			"space":   " ",
			"col_sep": " ",
			"row_sep": "\n",
			"compact": "false",
		}},
		"cache": {Name: "数据缓存", Value: map[string]interface{}{
			"store": "var/data",
			"limit": "30",
			"least": "10",
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
		"_init": {Name: "_init", Help: "hello", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.Travel(func(p *Context, s *Context) {
				if _, ok := s.Commands["_init"]; ok && p != nil {
					m.Spawns(s).Runs("_init", "_init", arg...)
				}
			})
		}},
		"init": {Name: "init", Help: "hello", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.Start("log", arg...)
			m.Start("gdb", arg...)
			m.Start("ssh", arg...)
			m.Cmd(arg)
		}},
		"help": {Name: "help", Help: "帮助", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.Echo(strings.Join(kit.Simple(m.Confv("help", "index")), "\n"))
		}},
		"exit": {Name: "exit", Help: "hello", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			Code = kit.Int(kit.Select("0", arg, 0))
			m.root.Cmd("_exit")
		}},
		"_exit": {Name: "_init", Help: "hello", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.root.Travel(func(p *Context, s *Context) {
				if _, ok := s.Commands["_exit"]; ok && p != nil {
					m.TryCatch(m.Spawns(s), true, func(msg *Message) {
						msg.Runs("_exit", "_exit", arg...)
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

func init() {
	Index.root = Index
	Pulse.root = Pulse
}

var Code = 0

func Run(arg ...string) string {
	if len(arg) == 0 {
		arg = os.Args[1:]
	}
	if len(arg) == 0 {
		arg = append(arg, os.Getenv("ice_serve"))
	}

	frame := &Frame{}
	Index.server = frame

	if frame.Begin(Pulse.Spawns(), arg...).Start(Pulse.Spawns(), arg...) {
		frame.Close(Pulse.Spawns(), arg...)
	}
	time.Sleep(time.Second)
	os.Exit(Code)
	return Pulse.Result()
}
