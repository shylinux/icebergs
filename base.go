package ice

import (
	"github.com/shylinux/toolkits"
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
	return true
}
func (f *Frame) Close(m *Message, arg ...string) bool {
	// 保存配置
	m.Travel(func(p *Context, s *Context) {
		if cmd, ok := s.Commands["_exit"]; ok {
			msg := m.Spawns(s)
			msg.Log("_exit", "some")
			cmd.Hand(msg, s, "_exit", arg...)
		}
	})

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
	Caches: map[string]*Cache{},
	Configs: map[string]*Config{
		"cache": {Name: "数据缓存", Value: map[string]interface{}{
			"store": "var/data",
			"limit": "30",
			"least": "10",
		}},
	},
	Commands: map[string]*Command{
		"_init": {Name: "_init", Help: "hello", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
		}},
		"exit": {Name: "exit", Help: "hello", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			c.Close(m.Spawn(c), arg...)
			os.Exit(kit.Int(kit.Select("0", arg, 0)))
		}},
		"_exit": {Name: "_init", Help: "hello", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
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
	if len(arg) == 0 {
		arg = append(arg, os.Getenv("ice_serve"))
	}

	if Index.Begin(Pulse.Spawns(), arg...).Start(Index.begin.Spawns(), arg...) {
		Index.Close(Index.start.Spawns(), arg...)
	}
	return Pulse.Result()
}
