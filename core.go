package ice

import (
	"os"
	"time"
)

type Frame struct {
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
	return f
}
func (f *Frame) Start(m *Message, arg ...string) bool {
	// 加载配置
	m.Travel(func(p *Context, s *Context) {
		if cmd, ok := s.Commands["_init"]; ok {
			msg := m.Spawns(s)
			msg.Log("_init", s.Name)
			cmd.Hand(msg, s, "_init", arg...)
		}
	})

	// 启动服务
	Index.begin.Cmd(arg)

	m.Travel(func(p *Context, s *Context) {
		if cmd, ok := s.Commands["_exit"]; ok {
			msg := m.Spawns(s)
			msg.Log("_exit", "some")
			cmd.Hand(msg, s, "_exit", arg...)
		}
	})
	// 保存配置
	return true
}
func (f *Frame) Close(m *Message, arg ...string) bool {
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
	Caches:  map[string]*Cache{},
	Configs: map[string]*Config{},
	Commands: map[string]*Command{
		"_init": {Name: "_init", Help: "hello", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.Echo("hello %s world", c.Name)
		}},
		"hi": {Name: "hi", Help: "hello", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.Echo("hello %s world", c.Name)
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
	Index.server = &Frame{}
}

func Run(arg ...string) string {
	if len(arg) == 0 {
		arg = os.Args[1:]
	}

	if Index.Begin(Pulse.Spawns(), arg...).Start(Index.begin.Spawns(), arg...) {
		Index.Close(Index.start.Spawns(), arg...)
	}
	return Pulse.Result()
}
