package ice

import (
	kit "github.com/shylinux/toolkits"
	"github.com/shylinux/toolkits/conf"
	"github.com/shylinux/toolkits/log"

	"fmt"
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
	defer m.Cost("begin ice")

	list := map[*Context]*Message{m.target: m}
	m.Travel(func(p *Context, s *Context) {
		s.root = m.target
		if msg, ok := list[p]; ok && msg != nil {
			list[s] = msg.Spawns(s)
			s.Begin(list[s], arg...)
		}
	})
	return f
}
func (f *Frame) Start(m *Message, arg ...string) bool {
	m.Log(LOG_START, "ice")
	defer m.Cost("start ice")

	m.Cap(CTX_STATUS, "start")
	m.Cap(CTX_STREAM, strings.Split(m.Time(), " ")[1])

	m.Cmdy("init", arg)
	return true
}
func (f *Frame) Close(m *Message, arg ...string) bool {
	m.Log(LOG_CLOSE, "ice")
	defer m.Cost("close ice")

	list := map[*Context]*Message{m.target: m}
	m.Travel(func(p *Context, s *Context) {
		if msg, ok := list[p]; ok && msg != nil {
			list[s] = msg.Spawns(s)
			s.Close(list[s], arg...)
		}
	})

	m.TryCatch(m, true, func(m *Message) {
		m.target.wg.Wait()
	})
	return true
}

var Index = &Context{Name: "ice", Help: "冰山模块",
	Caches: map[string]*Cache{
		CTX_FOLLOW: {Value: ""},
		CTX_STREAM: {Value: "shy"},
		CTX_STATUS: {Value: "begin"},
	},
	Configs: map[string]*Config{
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
			defer m.Cost(ICE_INIT)
			m.Travel(func(p *Context, c *Context) {
				if cmd, ok := c.Commands[ICE_INIT]; ok && p != nil {
					c.Run(m.Spawns(c), cmd, ICE_INIT, arg...)
				}
			})
		}},
		"init": {Name: "init", Help: "启动", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.root.Cmd(ICE_INIT)

			m.target.root.wg = &sync.WaitGroup{}
			for _, k := range kit.Split(kit.Select("gdb,log,ssh,ctx", os.Getenv("ctx_mod"))) {
				m.Start(k)
			}

			m.Cmd(SSH_SOURCE, "etc/init.shy", "init.shy", "启动配置")
			m.Cmdy(arg)
		}},
		"help": {Name: "help", Help: "帮助", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.Echo(strings.Join(kit.Simple(m.Confv("help", "index")), "\n"))
		}},
		"exit": {Name: "exit", Help: "结束", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.root.target.server.(*Frame).code = kit.Int(kit.Select("0", arg, 0))
			m.Cmd(SSH_SOURCE, "etc/exit.shy", "exit.shy", "退出配置")

			m.root.Cmd(ICE_EXIT)
		}},
		ICE_EXIT: {Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			defer m.Cost(ICE_EXIT)
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

	source: Index, target: Index, Hand: true,
	messages: []*Message{}, message: nil, root: nil,
}

var Log func(*Message, string, string)

func Run(arg ...string) string {
	if len(arg) == 0 {
		arg = os.Args[1:]
	}
	if len(arg) == 0 {
		arg = append(arg, WEB_SPACE, "connect", "self")
	}

	log.Init(conf.New(nil))

	frame := &Frame{}
	Index.root = Index
	Index.server = frame

	Pulse.root = Pulse
	Pulse.Option("cache.limit", "30")
	Pulse.Option("begin_time", Pulse.Time())

	if frame.Begin(Pulse.Spawns(), arg...).Start(Pulse, arg...) {
		frame.Close(Pulse.Spawns(), arg...)
	}

	if Pulse.Result() == "" {
		Pulse.Table()
	}
	fmt.Printf(Pulse.Result())

	os.Exit(frame.code)
	return ""
}
