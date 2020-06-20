package gdb

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/toolkits"

	"os"
	"os/signal"
	"syscall"
	"time"
)

type Frame struct {
	s chan os.Signal
	t <-chan time.Time
	d chan []string
}

func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	return f
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	for {
		select {
		case s, ok := <-f.s:
			if !ok {
				return true
			}
			// 信号事件
			m.Logs(EVENT, SIGNAL, s)
			m.Cmd(m.Confv(SIGNAL, kit.Keys(kit.MDB_HASH, s)), kit.Keys(s))

		case t, ok := <-f.t:
			if !ok {
				return true
			}

			// 定时事件
			stamp := int(t.Unix())
			m.Confm(TIMER, kit.MDB_HASH, func(key string, value map[string]interface{}) {
				if kit.Int(value["next"]) <= stamp {
					m.Logs(EVENT, TIMER, key, kit.MDB_TIME, value["next"])
					value["next"] = stamp + int(kit.Duration(value["interval"]))/int(time.Second)
					m.Cmd(value["cmd"])
					m.Grow(TIMER, nil, map[string]interface{}{
						"create_time": kit.Format(t), "interval": value["interval"],
						"cmd": value["cmd"], "key": key,
					})
				}
			})

		case d, ok := <-f.d:
			if !ok {
				return true
			}
			// 异步事件
			m.Logs(EVENT, d[0], d[1:])
			m.Grows(EVENT, d[0], "", "", func(index int, value map[string]interface{}) {
				m.Cmd(value["cmd"], d[1:]).Cost("event %v", d)
			})
		}
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

const (
	SIGNAL = "signal"
	TIMER  = "timer"
	EVENT  = "event"
)

const (
	LISTEN = "listen"
	ACTION = "action"
)
const (
	SYSTEM_INIT = "system.init"

	SERVE_START = "serve.start"
	SERVE_CLOSE = "serve.close"
	SPACE_START = "space.start"
	SPACE_CLOSE = "space.close"
	DREAM_START = "dream.start"
	DREAM_CLOSE = "dream.close"

	USER_CREATE = "user.create"
	CHAT_CREATE = "chat.create"
	MISS_CREATE = "miss.create"
	MIND_CREATE = "mind.create"
)

var Index = &ice.Context{Name: "gdb", Help: "事件模块",
	Configs: map[string]*ice.Config{
		SIGNAL: {Name: "signal", Help: "信号器", Value: kit.Dict(
			kit.MDB_META, kit.Dict("pid", "var/run/ice.pid"),
			kit.MDB_LIST, kit.List(),
			kit.MDB_HASH, kit.Dict(
				"2", []interface{}{"exit", "0"},
				"3", []interface{}{"exit", "1"},
				"15", []interface{}{"exit", "1"},
				// "20", []interface{}{"void"},
				"30", []interface{}{"exit"},
				"31", []interface{}{"exit", "1"},
				// "28", []interface{}{"void"},
			),
		)},
		TIMER: {Name: "timer", Help: "定时器", Value: kit.Data("tick", "100ms")},
		EVENT: {Name: "event", Help: "触发器", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if os.Getenv("ctx_mod") != "" {
				m.Cmd(nfs.SAVE, kit.Select(m.Conf(SIGNAL, "meta.pid"),
					m.Conf(cli.RUNTIME, "conf.ctx_pid")), m.Conf(cli.RUNTIME, "host.pid"))
			}
			// 进程标识
			if f, ok := m.Target().Server().(*Frame); ok {
				// 注册信号
				f.s = make(chan os.Signal, ice.MOD_CHAN)
				m.Richs(SIGNAL, nil, "*", func(key string, value string) {
					m.Logs(LISTEN, key, "cmd", value)
					signal.Notify(f.s, syscall.Signal(kit.Int(key)))
				})
				// 启动心跳
				f.t = time.Tick(kit.Duration(m.Cap(ice.CTX_STREAM, m.Conf(TIMER, "meta.tick"))))
				// 分发事件
				f.d = make(chan []string, ice.MOD_CHAN)
			}
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, ok := m.Target().Server().(*Frame); ok {
				// 停止心跳
				close(f.s)
				// 停止事件
				close(f.d)
			}
		}},

		SIGNAL: {Name: "signal", Help: "信号器", Action: map[string]*ice.Action{
			LISTEN: {Name: "listen signal cmd...", Help: "监听事件", Hand: func(m *ice.Message, arg ...string) {
				m.Rich(SIGNAL, arg[0], arg[1:])
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		TIMER: {Name: "timer", Help: "定时器", Action: map[string]*ice.Action{
			LISTEN: {Name: "listen delay interval cmd...", Help: "监听事件", Hand: func(m *ice.Message, arg ...string) {
				m.Rich(TIMER, nil, kit.Dict(
					"next", time.Now().Add(kit.Duration(arg[0])).Unix(),
					"interval", arg[1], "cmd", arg[2:],
				))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		EVENT: {Name: "event", Help: "触发器", Action: map[string]*ice.Action{
			LISTEN: {Name: "listen event cmd...", Help: "监听事件", Hand: func(m *ice.Message, arg ...string) {
				m.Grow(EVENT, arg[0], kit.Dict("cmd", arg[1:]))
				m.Logs(LISTEN, arg[0], "cmd", arg[1:])
			}},
			ACTION: {Name: "action event arg...", Help: "触发事件", Hand: func(m *ice.Message, arg ...string) {
				if f, ok := m.Target().Server().(*Frame); ok {
					m.Logs(ACTION, arg[0], "arg", arg[1:])
					f.d <- arg
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		"void": {Name: "void", Help: "空命令", Action: map[string]*ice.Action{}},
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
