package ice

import (
	kit "github.com/shylinux/toolkits"

	"os"
	"strings"
	"sync"
	"time"
)

var Info = struct {
	HostName string
	PathName string
	UserName string
	PassWord string
	NodeType string
	NodeName string

	Build struct {
		Time     string
		Hash     string
		Remote   string
		Branch   string
		Version  string
		HostName string
		UserName string
	}
}{}

type Frame struct{ code int }

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
			list[s] = msg.Spawn(s)
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
	m.TryCatch(m, true, func(m *Message) {
		m.target.wg.Wait()
	})

	m.Log(LOG_CLOSE, "ice")
	defer m.Cost("close ice")

	list := map[*Context]*Message{m.target: m}
	m.Travel(func(p *Context, s *Context) {
		if msg, ok := list[p]; ok && msg != nil {
			list[s] = msg.Spawn(s)
			s.Close(list[s], arg...)
		}
	})
	return true
}

var Index = &Context{Name: "ice", Help: "冰山模块",
	Caches: map[string]*Cache{
		CTX_FOLLOW: {Value: "ice"},
		CTX_STREAM: {Value: "shy"},
		CTX_STATUS: {Value: "begin"},
	},
	Configs: map[string]*Config{
		"help": {Value: map[string]interface{}{
			"index": []interface{}{
				"^_^      欢迎使用冰山框架       ^_^",
				"^_^  Welcome to Icebergs World  ^_^",
				"",
				"Report: shylinuxc@gmail.com",
				"Public: https://shylinux.com",
				"Source: https://github.com/shylinux/icebergs",
				"",
			},
		}},
	},
	Commands: map[string]*Command{
		CTX_INIT: {Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			defer m.Cost("_init ice")
			m.Travel(func(p *Context, c *Context) {
				if cmd, ok := c.Commands[CTX_INIT]; ok && p != nil {
					c.cmd(m.Spawn(c), cmd, CTX_INIT, arg...)
				}
			})
		}},
		"init": {Name: "init", Help: "启动", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			if m.target != m.target.root {
				return
			}
			m.root.Cmd(CTX_INIT)
			m.target.root.wg = &sync.WaitGroup{}
			for _, k := range kit.Split(kit.Select("gdb,log,ssh,mdb")) {
				m.Start(k)
			}
			m.Cmd("ssh.source", "etc/init.shy", "init.shy", "启动配置")
			m.Cmdy(arg)
		}},
		"help": {Name: "help", Help: "帮助", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.Echo(strings.Join(kit.Simple(m.Confv("help", "index")), "\n"))
		}},
		"exit": {Name: "exit restart:button", Help: "结束", Action: map[string]*Action{
			"restart": {Name: "restart", Help: "重启", Hand: func(m *Message, arg ...string) {
			}},
		}, Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.root.target.server.(*Frame).code = kit.Int(kit.Select("0", arg, 0))
			m.Cmd("ssh.source", "etc/exit.shy", "exit.shy", "退出配置")
			m.root.Cmd(CTX_EXIT)
		}},
		CTX_EXIT: {Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			defer m.Cost(CTX_EXIT)
			m.root.Travel(func(p *Context, c *Context) {
				if cmd, ok := c.Commands[CTX_EXIT]; ok && p != nil {
					m.TryCatch(m.Spawn(c), true, func(msg *Message) {
						c.cmd(msg, cmd, CTX_EXIT, arg...)
					})
				}
			})
			wait <- true
		}},
	},
}
var Pulse = &Message{
	time: time.Now(), code: 0,
	meta: map[string][]string{},
	data: map[string]interface{}{},

	source: Index, target: Index, Hand: true,
}
var wait = make(chan bool, 1)

func Run(arg ...string) string {
	if len(arg) == 0 {
		arg = os.Args[1:]
	}
	if len(arg) == 0 {
		arg = append(arg, "help")
	}

	frame := &Frame{}
	Index.root = Index
	Index.server = frame
	Index.Merge(Index)

	Pulse.root = Pulse
	Pulse.Option("name", "")
	Pulse.Option("cache.limit", "30")
	Pulse.Option("begin_time", Pulse.Time())
	switch kit.Select("", arg, 0) {
	case "space", "serve":
		if _log_disable = false; frame.Begin(Pulse.Spawn(), arg...).Start(Pulse, arg...) {
			frame.Close(Pulse.Spawn(), arg...)
		}

		<-wait
		os.Exit(frame.code)

	default:
		_log_disable = os.Getenv("ctx_debug") != "true"
		if Pulse.Cmdy(arg); Pulse.Result() == "" {
			Pulse.Table()
		}
		if strings.TrimSpace(Pulse.Result()) == "" {
			Pulse.Set(MSG_RESULT).Cmdy("cli.system", arg)
		}
		Pulse.Sleep("10ms")
	}

	return Pulse.Result()
}

var BinPack = map[string][]byte{}
var names = map[string]interface{}{}
var ErrNameExists = "name already exists: "

func Name(name string, value interface{}) string {
	if s, ok := names[name]; ok {
		last := ""
		switch s := s.(type) {
		case *Context:
			last = s.Name
		}
		panic(kit.Format("name already exits: %s %v", name, last))
	}

	names[name] = value
	return name
}
