package ice

import (
	"github.com/shylinux/toolkits"

	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
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
	begin    *Message
	start    *Message

	exit   chan bool
	server Server
	id     int
}
type Server interface {
	Spawn(m *Message, c *Context, arg ...string) Server
	Begin(m *Message, arg ...string) Server
	Start(m *Message, arg ...string) bool
	Close(m *Message, arg ...string) bool
}

func (c *Context) ID() int {
	c.id++
	return c.id
}
func (c *Context) Server() Server {
	return c.server
}
func (c *Context) Register(s *Context, x Server) *Context {
	Pulse.Log("register", "%s <- %s", c.Name, s.Name)
	if c.contexts == nil {
		c.contexts = map[string]*Context{}
	}
	c.contexts[s.Name] = s
	s.root = c.root
	s.context = c
	s.server = x
	return s
}

func (c *Context) Begin(m *Message, arg ...string) *Context {
	c.begin = m
	m.Log("begin", "%s", c.Name)
	if c.server != nil {
		c.server.Begin(m, arg...)
	}
	return c
}
func (c *Context) Start(m *Message, arg ...string) bool {
	c.start = m
	m.Log("start", "%s", c.Name)
	return c.server.Start(m, arg...)
}
func (c *Context) Close(m *Message, arg ...string) bool {
	m.Log("close", "%s", c.Name)
	if c.server != nil {
		return c.server.Close(m, arg...)
	}
	return true
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
	cb     func(*Message) *Message
}

func (m *Message) Time() string {
	return m.time.Format("2006-01-02 15:04:05")
}
func (m *Message) Target() *Context {
	return m.target
}
func (m *Message) Format(key interface{}) string {
	switch key := key.(type) {
	case string:
		switch key {
		case "cost":
			return time.Now().Sub(m.time).String()
		case "meta":
			return kit.Format(m.meta)
		case "stack":
			pc := make([]uintptr, 100)
			pc = pc[:runtime.Callers(5, pc)]
			frames := runtime.CallersFrames(pc)

			meta := []string{}
			for {
				frame, more := frames.Next()
				file := strings.Split(frame.File, "/")
				name := strings.Split(frame.Function, "/")
				meta = append(meta, fmt.Sprintf("\n%s:%d\t%s", file[len(file)-1], frame.Line, name[len(name)-1]))
				if !more {
					break
				}
			}
			return strings.Join(meta, "")
		}
	case []byte:
		json.Unmarshal(key, &m.meta)
	}
	return m.time.Format("2006-01-02 15:04:05")
}
func (m *Message) Formats(key string) string {
	switch key {
	case "meta":
		return kit.Formats(m.meta)
	default:
		return m.Format(key)
	}
	return m.time.Format("2006-01-02 15:04:05")
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
		case []byte:
			json.Unmarshal(val, &msg.meta)
		}
	}
	return msg
}
func (m *Message) Spawns(arg ...interface{}) *Message {
	msg := m.Spawn(arg...)
	msg.code = Index.ID()
	m.messages = append(m.messages, msg)
	return msg
}

func (m *Message) Add(key string, arg ...string) *Message {
	switch key {
	case "detail", "result":
		m.meta[key] = append(m.meta[key], arg...)

	case "option", "append":
		if len(arg) > 0 {
			if kit.IndexOf(m.meta[key], arg[0]) == -1 {
				m.meta[key] = append(m.meta[key], arg[0])
			}
			m.meta[arg[0]] = append(m.meta[arg[0]], arg[1:]...)
		}
	}
	return m
}
func (m *Message) Set(key string, arg ...string) *Message {
	switch key {
	case "detail", "result":
		delete(m.meta, key)
	case "option", "append":
		if len(arg) > 0 {
			delete(m.meta, arg[0])
		} else {
			for _, k := range m.meta[key] {
				delete(m.meta, k)
			}
			delete(m.meta, key)
			return m
		}
	}
	return m.Add(key, arg...)
}
func (m *Message) Copy(msg *Message) *Message {
	for _, k := range msg.meta["append"] {
		if kit.IndexOf(m.meta["append"], k) == -1 {
			m.meta["append"] = append(m.meta["append"], k)
		}
		for _, v := range msg.meta[k] {
			m.meta[k] = append(m.meta[k], v)
		}
	}
	for _, v := range msg.meta["result"] {
		m.meta["result"] = append(m.meta["result"], v)
	}
	return m
}
func (m *Message) Push(key string, value interface{}) *Message {
	return m.Add("append", key, kit.Format(value))
}
func (m *Message) Echo(str string, arg ...interface{}) *Message {
	m.meta["result"] = append(m.meta["result"], fmt.Sprintf(str, arg...))
	return m
}
func (m *Message) Option(key string, arg ...interface{}) string {
	return kit.Select("", kit.Simple(m.Optionv(key, arg...)), 0)
}
func (m *Message) Optionv(key string, arg ...interface{}) interface{} {
	if len(arg) > 0 {
		if kit.IndexOf(m.meta["option"], key) == -1 {
			m.meta["option"] = append(m.meta["option"], key)
		}

		switch arg := arg[0].(type) {
		case string:
			m.meta[key] = []string{arg}
		case []string:
			m.meta[key] = arg
		default:
			m.data[key] = arg
		}
	}

	for msg := m; msg != nil; msg = msg.message {
		if list, ok := m.meta[key]; ok {
			return list
		}
		if list, ok := m.data[key]; ok {
			return list
		}
	}
	return nil
}
func (m *Message) Resultv(arg ...interface{}) []string {
	return m.meta["result"]
}
func (m *Message) Result(arg ...interface{}) string {
	return strings.Join(m.Resultv(), "")
}

func (m *Message) Log(level string, str string, arg ...interface{}) *Message {
	fmt.Fprintf(os.Stderr, "%s %d %s->%s %s %s\n", time.Now().Format("2006-01-02 15:04:05"), m.code, m.source.Name, m.target.Name, level, fmt.Sprintf(str, arg...))
	return m
}
func (m *Message) Assert(arg interface{}) bool {
	switch arg := arg.(type) {
	case nil:
		return true
	case bool:
		if arg == true {
			return true
		}
	}

	panic(errors.New(fmt.Sprintf("error %v", arg)))
}
func (m *Message) TryCatch(msg *Message, safe bool, hand ...func(msg *Message)) *Message {
	defer func() {
		switch e := recover(); e {
		case io.EOF:
		case nil:
		default:
			m.Log("bench", "chain: %s", msg.Format("chain"))
			m.Log("bench", "catch: %s", e)
			m.Log("bench", "stack: %s", msg.Format("stack"))

			if m.Log("error", "catch: %s", e); len(hand) > 1 {
				m.TryCatch(msg, safe, hand[1:]...)
			} else if !safe {
				m.Assert(e)
			}
		}
	}()

	if len(hand) > 0 {
		hand[0](msg)
	}
	return m
}
func (m *Message) Travel(cb func(p *Context, s *Context)) *Message {
	list := []*Context{m.target}
	for i := 0; i < len(list); i++ {
		cb(list[i].context, list[i])
		for _, v := range list[i].contexts {
			list = append(list, v)
		}
	}
	return m
}
func (m *Message) Search(key interface{}, cb func(p *Context, s *Context, key string)) *Message {
	switch key := key.(type) {
	case string:
		if strings.Contains(key, ":") {

		} else if strings.Contains(key, ".") {
			list := strings.Split(key, ".")

			p := m.target.root
			for _, v := range list[:len(list)-1] {
				if s, ok := p.contexts[v]; ok {
					p = s
				} else {
					p = nil
					break
				}
			}
			if p != nil {
				cb(p.context, p, list[len(list)-1])
			}
		} else {
			cb(m.target.context, m.target, key)
		}
	}
	return m
}
func (m *Message) Gos(msg *Message, cb func(*Message)) *Message {
	go func() { msg.TryCatch(msg, true, func(msg *Message) { cb(msg) }) }()
	return m
}
func (m *Message) Run(arg ...string) *Message {
	m.target.server.Start(m, arg...)
	return m
}
func (m *Message) Call(sync bool, cb func(*Message) *Message) *Message {
	if sync {
		wait := make(chan bool)
		m.cb = func(sub *Message) *Message {
			wait <- true
			return cb(sub)
		}
		<-wait
	}
	return m
}
func (m *Message) Back(sub *Message) *Message {
	if m.cb != nil {
		m.cb(sub)
	}
	return m
}

func (m *Message) Cmdy(arg ...interface{}) *Message {
	msg := m.Cmd(arg...)
	m.Copy(msg)
	return m
}
func (m *Message) Cmd(arg ...interface{}) *Message {
	list := kit.Simple(arg...)
	if len(list) == 0 {
		list = m.meta["detail"]
	}
	if len(list) == 0 {
		return m
	}

	msg := m
	m.Search(list[0], func(p *Context, s *Context, key string) {
		for c := s; c != nil; c = c.context {
			if cmd, ok := c.Commands[key]; ok {
				msg = m.Spawns(s).Log("cmd", "%s.%s %v", c.Name, key, list[1:])
				msg.TryCatch(msg, true, func(msg *Message) {
					cmd.Hand(msg, c, key, list[1:]...)
				})
				break
			}
		}
	})
	return msg
}
func (m *Message) Confv(arg ...interface{}) (val interface{}) {
	m.Search(arg[0], func(p *Context, s *Context, key string) {
		for c := s; c != nil; c = c.context {
			if conf, ok := c.Configs[key]; ok {
				if len(arg) > 0 {
					val = kit.Value(conf.Value, arg[1:]...)
				} else {
					val = conf.Value
				}
			}
		}
	})
	return
}
func (m *Message) Confm(key string, chain interface{}, cbs ...interface{}) map[string]interface{} {
	val := m.Confv(key, chain)
	if len(cbs) > 0 {
		switch val := val.(type) {
		case map[string]interface{}:
			switch cb := cbs[0].(type) {
			case func(string, map[string]interface{}):
				for k, v := range val {
					cb(k, v.(map[string]interface{}))
				}
			}
		}
	}
	value, _ := val.(map[string]interface{})
	return value
}
func (m *Message) Confs(arg ...interface{}) bool {
	return kit.Format(m.Confv(arg...)) != ""
}
func (m *Message) Confi(arg ...interface{}) int {
	return kit.Int(m.Confv(arg...))
}
func (m *Message) Conf(arg ...interface{}) string {
	return kit.Format(m.Confv(arg...))
}
func (m *Message) Capv(arg ...interface{}) interface{} {
	key := ""
	switch val := arg[0].(type) {
	case string:
		key, arg = val, arg[1:]
	}

	for _, s := range []*Context{m.target} {
		for c := s; c != nil; c = c.context {
			if caps, ok := c.Caches[key]; ok {
				return kit.Value(caps.Value, arg[0])
			}
		}
	}
	return nil
}
func (m *Message) Cap(arg ...interface{}) string {
	return kit.Format(m.Capv(arg...))
}
