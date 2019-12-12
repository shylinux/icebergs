package ice

import (
	"github.com/shylinux/toolkits"

	"fmt"
	"os"
	"strings"
	"time"
)

type Cache struct {
	Name  string
	Help  string
	Value string
}
type Config struct {
	Name  string
	Help  string
	Value interface{}
}
type Command struct {
	Name string
	Help interface{}
	Form map[string]int
	Hand func(m *Message, c *Context, key string, arg ...string)
}
type Context struct {
	Name string
	Help string

	Caches   map[string]*Cache
	Configs  map[string]*Config
	Commands map[string]*Command

	contexts map[string]*Context
	context  *Context
	root     *Context

	exit   chan bool
	server Server
}
type Server interface {
	Spawn(m *Message, c *Context, arg ...string) Server
	Begin(m *Message, arg ...string) Server
	Start(m *Message, arg ...string) bool
	Close(m *Message, arg ...string) bool
}

type Message struct {
	time time.Time
	code int

	meta map[string][]string
	data map[string]interface{}

	messages []*Message
	message  *Message
	root     *Message

	source *Context
	target *Context
	Hand   bool
}

func (m *Message) Spawn(arg ...interface{}) *Message {
	msg := &Message{
		time: time.Now(),
		code: -1,

		meta: map[string][]string{},
		data: map[string]interface{}{},

		message: m,
		root:    m.root,

		source: m.target,
		target: m.target,
	}

	if len(arg) > 0 {
		switch val := arg[0].(type) {
		case *Context:
			msg.target = val
		}
	}
	return msg
}
func (m *Message) Log(level string, str string, arg ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s %s %s\n", time.Now().Format("2006-01-02 15:04:05"), level, fmt.Sprintf(str, arg...))
}
func (m *Message) Confv(arg ...interface{}) interface{} {
	key := ""
	switch val := arg[0].(type) {
	case string:
		key, arg = val, arg[1:]
	}

	for _, s := range []*Context{m.target} {
		for c := s; c != nil; c = c.context {
			if conf, ok := c.Configs[key]; ok {
				m.Log("conf", "%s.%s", c.Name, key)
				return kit.Value(conf.Value, key)
			}
		}
	}
	return nil
}
func (m *Message) Conf(arg ...interface{}) string {
	return kit.Format(m.Confv(arg...))
}
func (m *Message) Cmd(arg ...interface{}) *Message {
	list := kit.Trans(arg...)
	if len(list) == 0 {
		return m
	}
	for _, s := range []*Context{m.target} {
		for c := s; c != nil; c = c.context {
			if cmd, ok := c.Commands[list[0]]; ok {
				m.Log("cmd", "%s.%s", c.Name, list[0])
				cmd.Hand(m, c, list[0], list[1:]...)
			}
		}
	}
	return m
}
func (m *Message) Echo(str string, arg ...interface{}) *Message {
	m.meta["result"] = append(m.meta["result"], fmt.Sprintf(str, arg...))
	return m
}
func (m *Message) Result(arg ...interface{}) string {
	return strings.Join(m.meta["result"], "")
}

var Pulse = &Message{
	time: time.Now(), code: 0,
	meta: map[string][]string{},
	data: map[string]interface{}{},

	messages: []*Message{}, message: nil, root: nil,
	source: Index, target: Index, Hand: true,
}
var Index = &Context{Name: "root", Help: "元始模块",
	Caches:  map[string]*Cache{},
	Configs: map[string]*Config{},
	Commands: map[string]*Command{
		"hi": {Name: "hi", Help: "hello", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.Echo("hello world")
		}},
	},
}

func (c *Context) Register(s *Context, x Server) *Context {
	if c.contexts == nil {
		c.contexts = map[string]*Context{}
	}
	c.contexts[s.Name] = s
	s.root = c.root
	s.context = c
	s.server = x
	return s
}
func init() {
	Index.root = Index
	Pulse.root = Pulse
}
