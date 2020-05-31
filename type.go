// icebergs: 后端 冰山架 挨撕不可
// CMS: a cluster manager system

package ice

import (
	kit "github.com/shylinux/toolkits"

	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
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
type Action struct {
	Name string
	Help string
	Hand func(m *Message, arg ...string)
}
type Command struct {
	Name   interface{} // string []string
	Help   interface{} // string []string
	List   []interface{}
	Meta   map[string]interface{}
	Hand   func(m *Message, c *Context, key string, arg ...string)
	Action map[string]*Action
}
type Context struct {
	Name string
	Help interface{} // string []string
	Test interface{} // string []string

	Caches   map[string]*Cache
	Configs  map[string]*Config
	Commands map[string]*Command

	contexts map[string]*Context
	context  *Context
	root     *Context

	begin  *Message
	start  *Message
	server Server

	wg *sync.WaitGroup
	id int32
}
type Server interface {
	Spawn(m *Message, c *Context, arg ...string) Server
	Begin(m *Message, arg ...string) Server
	Start(m *Message, arg ...string) bool
	Close(m *Message, arg ...string) bool
}

func (c *Context) ID() int32 {
	return atomic.AddInt32(&c.id, 1)
}
func (c *Context) Cap(key string, arg ...interface{}) string {
	if len(arg) > 0 {
		c.Caches[key].Value = kit.Format(arg[0])
	}
	return c.Caches[key].Value
}
func (c *Context) Run(m *Message, cmd *Command, key string, arg ...string) *Message {
	action, args := m.Option("_action"), arg
	if len(arg) > 0 && arg[0] == "action" {
		action, args = arg[1], arg[2:]
	}
	m.Log(LOG_CMDS, "%s.%s %d %v", c.Name, key, len(arg), arg)
	if m.Hand = true; len(arg) > 1 && action != "" && cmd.Action != nil {
		if h, ok := cmd.Action[action]; ok {
			h.Hand(m, args...)
			return m
		}
		for _, h := range cmd.Action {
			if h.Name == action || h.Help == action {
				h.Hand(m, args...)
				return m
			}
		}
	}
	cmd.Hand(m, c, key, arg...)
	return m
}
func (c *Context) Runs(m *Message, cmd string, key string, arg ...string) {
	if s, ok := m.Target().Commands[key]; ok {
		c.Run(m, s, cmd, arg...)
	}
	return
}
func (c *Context) Server() Server {
	return c.server
}
func (c *Context) Register(s *Context, x Server) *Context {
	Pulse.Log("register", "%s <- %s", c.Name, s.Name)
	if c.contexts == nil {
		c.contexts = map[string]*Context{}
	}
	c.contexts[kit.Format(s.Name)] = s
	s.root = c.root
	s.context = c
	s.server = x
	return s
}
func (c *Context) Merge(s *Context, x Server) *Context {
	for k, v := range s.Commands {
		c.Commands[k] = v
	}
	for k, v := range s.Configs {
		c.Configs[k] = v
	}
	for k, v := range s.Caches {
		c.Caches[k] = v
	}
	s.server = x
	return c
}

func (c *Context) Spawn(m *Message, name string, help string, arg ...string) *Context {
	s := &Context{Name: name, Help: help, Caches: map[string]*Cache{}, Configs: map[string]*Config{}}
	if m.target.Server != nil {
		c.Register(s, m.target.server.Spawn(m, s, arg...))
	} else {
		c.Register(s, nil)
	}
	m.target = s
	return s
}
func (c *Context) Begin(m *Message, arg ...string) *Context {
	if c.Caches == nil {
		c.Caches = map[string]*Cache{}
	}
	if c.Configs == nil {
		c.Configs = map[string]*Config{}
	}
	c.Caches[CTX_FOLLOW] = &Cache{Name: CTX_FOLLOW, Value: ""}
	c.Caches[CTX_STREAM] = &Cache{Name: CTX_STREAM, Value: ""}
	c.Caches[CTX_STATUS] = &Cache{Name: CTX_STATUS, Value: ""}

	if c.context == Index {
		c.Cap(CTX_FOLLOW, c.Name)
	} else if c.context != nil {
		c.Cap(CTX_FOLLOW, kit.Keys(c.context.Cap(CTX_FOLLOW), c.Name))
	}
	m.Log(LOG_BEGIN, "%s", c.Cap(CTX_FOLLOW))
	c.Cap(CTX_STATUS, ICE_BEGIN)

	if c.begin = m; c.server != nil {
		m.TryCatch(m, true, func(m *Message) {
			// 初始化模块
			c.server.Begin(m, arg...)
		})
	}
	return c
}
func (c *Context) Start(m *Message, arg ...string) bool {
	c.start = m
	m.Hold(1)

	wait := make(chan bool)
	m.Gos(m, func(m *Message) {
		m.Log(LOG_START, "%s", c.Cap(CTX_FOLLOW))
		c.Cap(CTX_STATUS, ICE_START)
		wait <- true

		// 启动模块
		if c.server != nil {
			c.server.Start(m, arg...)
		}
		if m.Done(); m.wait != nil {
			m.wait <- true
		}
	})
	<-wait
	return true
}
func (c *Context) Close(m *Message, arg ...string) bool {
	m.Log(LOG_CLOSE, "%s", c.Cap(CTX_FOLLOW))
	c.Cap(CTX_STATUS, ICE_CLOSE)

	if c.server != nil {
		// 结束模块
		return c.server.Close(m, arg...)
	}
	return true
}

type Message struct {
	time time.Time
	code int
	Hand bool

	meta map[string][]string
	data map[string]interface{}

	messages []*Message
	message  *Message
	root     *Message

	source *Context
	target *Context
	frames interface{}

	cb   func(*Message) *Message
	W    http.ResponseWriter
	R    *http.Request
	O    io.Writer
	I    io.Reader
	wait chan bool
}

func (m *Message) Time(args ...interface{}) string {
	// [duration] [format [args...]]
	t := m.time
	if len(args) > 0 {
		switch arg := args[0].(type) {
		case string:
			if d, e := time.ParseDuration(arg); e == nil {
				// 时间偏移
				t, args = t.Add(d), args[1:]
			}
		}
	}
	f := ICE_TIME
	if len(args) > 0 {
		switch arg := args[0].(type) {
		case string:
			f = arg
			if len(args) > 1 {
				// 时间格式
				f = fmt.Sprintf(f, args[1:]...)
			}
		}
	}
	return t.Format(f)
}
func (m *Message) Target() *Context {
	return m.target
}
func (m *Message) Source() *Context {
	return m.source
}
func (m *Message) Format(key interface{}) string {
	switch key := key.(type) {
	case string:
		switch key {
		case "cost":
			return kit.FmtTime(kit.Int64(time.Now().Sub(m.time)))
		case "meta":
			return kit.Format(m.meta)
		case "append":
			if len(m.meta["append"]) == 0 {
				return fmt.Sprintf("%dx%d %s", 0, len(m.meta["append"]), kit.Format(m.meta["append"]))
			} else {
				return fmt.Sprintf("%dx%d %s", len(m.meta[m.meta["append"][0]]), len(m.meta["append"]), kit.Format(m.meta["append"]))
			}

		case "time":
			return m.Time()
		case "ship":
			return fmt.Sprintf("%s->%s", m.source.Name, m.target.Name)
		case "prefix":
			return fmt.Sprintf("%s %d %s->%s", m.Time(), m.code, m.source.Name, m.target.Name)

		case "chain":
			// 调用链
			ms := []*Message{}
			for msg := m; msg != nil; msg = msg.message {
				ms = append(ms, msg)
			}

			meta := append([]string{}, "\n\n")
			for i := len(ms) - 1; i >= 0; i-- {
				msg := ms[i]

				meta = append(meta, fmt.Sprintf("%s ", msg.Format("prefix")))
				if len(msg.meta[MSG_DETAIL]) > 0 {
					meta = append(meta, fmt.Sprintf("detail:%d %v", len(msg.meta[MSG_DETAIL]), msg.meta[MSG_DETAIL]))
				}

				if len(msg.meta[MSG_OPTION]) > 0 {
					meta = append(meta, fmt.Sprintf("option:%d %v\n", len(msg.meta[MSG_OPTION]), msg.meta[MSG_OPTION]))
					for _, k := range msg.meta[MSG_OPTION] {
						if v, ok := msg.meta[k]; ok {
							meta = append(meta, fmt.Sprintf("    %s: %d %v\n", k, len(v), v))
						}
					}
				} else {
					meta = append(meta, "\n")
				}

				if len(msg.meta[MSG_APPEND]) > 0 {
					meta = append(meta, fmt.Sprintf("  append:%d %v\n", len(msg.meta[MSG_APPEND]), msg.meta[MSG_APPEND]))
					for _, k := range msg.meta[MSG_APPEND] {
						if v, ok := msg.meta[k]; ok {
							meta = append(meta, fmt.Sprintf("    %s: %d %v\n", k, len(v), v))
						}
					}
				}
				if len(msg.meta[MSG_RESULT]) > 0 {
					meta = append(meta, fmt.Sprintf("  result:%d %v\n", len(msg.meta[MSG_RESULT]), msg.meta[MSG_RESULT]))
				}
			}
			return strings.Join(meta, "")
		case "stack":
			// 调用栈
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
	return m.time.Format(ICE_TIME)
}
func (m *Message) Formats(key string) string {
	switch key {
	case "meta":
		return kit.Formats(m.meta)
	}
	return m.Format(key)
}
func (m *Message) Spawns(arg ...interface{}) *Message {
	msg := m.Spawn(arg...)
	msg.code = int(m.target.root.ID())
	// m.messages = append(m.messages, msg)
	return msg
}
func (m *Message) Spawn(arg ...interface{}) *Message {
	msg := &Message{
		code: -1, time: time.Now(),

		meta: map[string][]string{},
		data: map[string]interface{}{},

		message: m, root: m.root,

		source: m.target,
		target: m.target,

		W: m.W, R: m.R,
		O: m.O, I: m.I,
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

func (m *Message) TryCatch(msg *Message, safe bool, hand ...func(msg *Message)) *Message {
	defer func() {
		switch e := recover(); e {
		case io.EOF:
		case nil:
		default:
			_, file, line, _ := runtime.Caller(3)
			if list := strings.Split(file, "/"); len(list) > 2 {
				file = strings.Join(list[len(list)-2:], "/")
			}
			m.Log(LOG_WARN, "catch: %s %s:%d", e, file, line)
			m.Log(LOG_INFO, "chain: %s", msg.Format("chain"))
			m.Log(LOG_WARN, "catch: %s %s:%d", e, file, line)
			m.Log(LOG_INFO, "stack: %s", msg.Format("stack"))
			if m.Log(LOG_WARN, "catch: %s %s:%d", e, file, line); len(hand) > 1 {
				// 捕获异常
				m.TryCatch(msg, safe, hand[1:]...)
			} else if !safe {
				// 抛出异常
				m.Assert(e)
			}
		}
	}()

	if len(hand) > 0 {
		// 运行函数
		hand[0](msg)
	}
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

	// 抛出异常
	panic(errors.New(fmt.Sprintf("error %v", arg)))
}
func (m *Message) Sleep(arg string) *Message {
	time.Sleep(kit.Duration(arg))
	return m
}
func (m *Message) Hold(n int) *Message {
	ctx := m.target.root
	if c := m.target; c.context != nil && c.context.wg != nil {
		ctx = c.context
	}

	ctx.wg.Add(n)
	m.Log(LOG_TRACE, "%s wait %s %v", ctx.Name, m.target.Name, ctx.wg)
	return m
}
func (m *Message) Done() bool {
	defer func() { recover() }()

	ctx := m.target.root
	if c := m.target; c.context != nil && c.context.wg != nil {
		ctx = c.context
	}

	m.Log(LOG_TRACE, "%s done %s %v", ctx.Name, m.target.Name, ctx.wg)
	ctx.wg.Done()
	return true
}
func (m *Message) Call(sync bool, cb func(*Message) *Message) *Message {
	wait := make(chan bool, 2)
	t := time.AfterFunc(kit.Duration("10s"), func() {
		m.Log(LOG_WARN, "timeout")
		m.Back(nil)
		wait <- false
	})
	m.cb = func(sub *Message) *Message {
		if sync {
			t.Stop()
			wait <- true
		}
		return cb(sub)
	}

	if sync {
		<-wait
	} else {
		t.Stop()
	}
	return m
}
func (m *Message) Back(res *Message) *Message {
	if m.cb != nil {
		// if res != nil {
		// 	m.Info("back %v", res.Format("prefix"))
		// } else {
		// 	m.Info("back %v", nil)
		// }
		if sub := m.cb(res); m.message != nil {
			m.message.Back(sub)
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
func (m *Message) Start(key string, arg ...string) *Message {
	m.Travel(func(p *Context, s *Context) {
		if s.Name == key {
			s.Start(m.Spawns(s), arg...)
		}
	})
	return m
}
func (m *Message) Starts(name string, help string, arg ...string) *Message {
	m.wait = make(chan bool)
	m.target.Spawn(m, name, help, arg...).Begin(m, arg...).Start(m, arg...)
	<-m.wait
	return m
}

func (m *Message) Right(arg ...interface{}) bool {
	return m.Option(MSG_USERROLE) == ROLE_ROOT || !m.Warn(m.Cmdx(AAA_ROLE, "right", m.Option(MSG_USERROLE), kit.Keys(arg...)) != "ok", "no right")
}
func (m *Message) Space(arg interface{}) []string {
	if arg == nil || kit.Format(arg) == m.Conf(CLI_RUNTIME, "node.name") {
		return nil
	}
	return []string{WEB_SPACE, kit.Format(arg)}
}
func (m *Message) Event(key string, arg ...string) *Message {
	m.Cmd(GDB_EVENT, "action", key, arg)
	return m
}
func (m *Message) Watch(key string, arg ...string) *Message {
	m.Cmd(GDB_EVENT, "listen", key, arg)
	return m
}

func (m *Message) Travel(cb interface{}) *Message {
	list := []*Context{m.target}
	for i := 0; i < len(list); i++ {
		switch cb := cb.(type) {
		case func(*Context, *Context):
			// 模块回调
			cb(list[i].context, list[i])
		case func(*Context, *Context, string, *Command):
			ls := []string{}
			for k := range list[i].Commands {
				ls = append(ls, k)
			}
			sort.Strings(ls)
			for _, k := range ls {
				// 命令回调
				cb(list[i].context, list[i], k, list[i].Commands[k])
			}
		case func(*Context, *Context, string, *Config):
			ls := []string{}
			for k := range list[i].Configs {
				ls = append(ls, k)
			}
			sort.Strings(ls)
			for _, k := range ls {
				// 配置回调
				cb(list[i].context, list[i], k, list[i].Configs[k])
			}
		}

		// 下级模块
		ls := []string{}
		for k := range list[i].contexts {
			ls = append(ls, k)
		}
		sort.Strings(ls)
		for _, k := range ls {
			list = append(list, list[i].contexts[k])
		}
	}
	return m
}
func (m *Message) Search(key interface{}, cb interface{}) *Message {
	switch key := key.(type) {
	case string:
		if k, ok := Alias[key]; ok {
			key = k
		}

		// 查找模块
		p := m.target.root
		if strings.Contains(key, ":") {

		} else if key == "." {
			if m.target.context != nil {
				p = m.target.context
			}
		} else if strings.Contains(key, ".") {
			list := strings.Split(key, ".")
			for _, p = range []*Context{m.target.root, m.target, m.source} {
				for _, v := range list[:len(list)-1] {
					if s, ok := p.contexts[v]; ok {
						p = s
					} else {
						p = nil
						break
					}
				}
				if p != nil {
					break
				}
			}
			if p == nil {
				m.Log(LOG_WARN, "not found %s", key)
				break
			}
			key = list[len(list)-1]
		} else {
			p = m.target
		}

		// 遍历命令
		switch cb := cb.(type) {
		case func(p *Context, s *Context, key string, cmd *Command):
			if key == "" {
				for k, v := range p.Commands {
					cb(p.context, p, k, v)
				}
				break
			}

			for _, p = range []*Context{p, m.target, m.source} {
				for c := p; c != nil; c = c.context {
					if cmd, ok := c.Commands[key]; ok {
						cb(c, p, key, cmd)
						return m
					}
				}
			}
		case func(p *Context, s *Context, key string, conf *Config):
			for _, p = range []*Context{p, m.target, m.source} {
				for c := p; c != nil; c = c.context {
					if cmd, ok := c.Configs[key]; ok {
						cb(c.context, c, key, cmd)
						return m
					}
				}
			}
		case func(p *Context, s *Context, key string):
			cb(p.context, p, key)
		}
	}
	return m
}
func (m *Message) Preview(arg string) (res string) {
	list := kit.Split(arg)
	m.Search(list[0], func(p *Context, s *Context, key string, cmd *Command) {
		res = kit.Format(kit.Dict("feature", cmd.Meta, "inputs", cmd.List))
	})
	return res
}

var count = int32(0)

func (m *Message) AddCmd(cmd *Command) string {
	name := fmt.Sprintf("_cb_%d", atomic.AddInt32(&count, 1))
	m.target.Commands[name] = cmd
	return kit.Keys(m.target.Cap(CTX_FOLLOW), name)
}
func (m *Message) Cmdy(arg ...interface{}) *Message {
	msg := m.Cmd(arg...)
	m.Copy(msg)
	return m
}
func (m *Message) Cmdx(arg ...interface{}) string {
	return kit.Select("", m.Cmd(arg...).meta[MSG_RESULT], 0)
}
func (m *Message) Cmds(arg ...interface{}) bool {
	return kit.Select("", m.Cmd(arg...).meta[MSG_RESULT], 0) != ""
}
func (m *Message) Cmd(arg ...interface{}) *Message {
	list := kit.Simple(arg...)
	if len(list) == 0 {
		list = m.meta[MSG_DETAIL]
	}
	if len(list) == 0 {
		return m
	}

	m.Search(list[0], func(p *Context, c *Context, key string, cmd *Command) {
		m.TryCatch(m.Spawns(c), true, func(msg *Message) {
			m.Hand, msg.Hand = true, true
			msg.meta[MSG_DETAIL] = list

			p.Run(msg, cmd, key, list[1:]...)
			m.Hand, msg.Hand, m = true, true, msg
		})
	})

	if m.Warn(m.Hand == false, "not found %v", list) {
		// return m.Set(MSG_RESULT).Cmd(CLI_SYSTEM, list)
	}
	return m
}
func (m *Message) Confv(arg ...interface{}) (val interface{}) {
	m.Search(arg[0], func(p *Context, s *Context, key string, conf *Config) {
		if len(arg) > 1 {
			if len(arg) > 2 {
				if arg[1] == nil {
					// 写配置
					conf.Value = arg[2]
				} else {
					// 写修改项
					kit.Value(conf.Value, arg[1:]...)
				}
			}
			// 读配置项
			val = kit.Value(conf.Value, arg[1])
		} else {
			// 读配置
			val = conf.Value
		}
	})
	return
}
func (m *Message) Confm(key string, chain interface{}, cbs ...interface{}) map[string]interface{} {
	val := m.Confv(key, chain)
	if len(cbs) > 0 {
		kit.Fetch(val, cbs[0])
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
				if len(arg) > 0 {
					// 写数据
					caps.Value = kit.Format(arg[0])
				}
				// 读数据
				return caps.Value
			}
		}
	}
	return nil
}
func (m *Message) Cap(arg ...interface{}) string {
	return kit.Format(m.Capv(arg...))
}
