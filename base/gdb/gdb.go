package gdb

import (
	"github.com/shylinux/icebergs"
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
			m.Log(ice.LOG_SIGNAL, "%v: %v", s, m.Confv(ice.GDB_SIGNAL, kit.Keys(kit.MDB_HASH, s)))
			m.Cmd(m.Confv(ice.GDB_SIGNAL, kit.Keys(kit.MDB_HASH, s)))

		case t, ok := <-f.t:
			if !ok {
				return true
			}
			break
			stamp := int(t.Unix())
			m.Confm(ice.GDB_TIMER, kit.MDB_HASH, func(key string, value map[string]interface{}) {
				if kit.Int(value["next"]) <= stamp {
					m.Log(ice.LOG_INFO, "timer %v %v", key, value["next"])
					value["next"] = stamp + int(kit.Duration(value["interval"]))/int(time.Second)
					m.Cmd(value["cmd"])
					m.Grow(ice.GDB_TIMER, nil, map[string]interface{}{
						"create_time": kit.Format(t), "interval": value["interval"],
						"cmd": value["cmd"], "key": key,
					})
				}
			})

		case d, ok := <-f.d:
			if !ok {
				return true
			}
			m.Log(ice.LOG_EVENTS, "%s: %v", d[0], d[1:])
			m.Grows(ice.GDB_EVENT, d[0], "", "", func(index int, value map[string]interface{}) {
				m.Cmd(value["cmd"], d[1:]).Cost("event %v", d)
			})
		}
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

var Index = &ice.Context{Name: "gdb", Help: "事件模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ice.GDB_SIGNAL: {Name: "signal", Help: "信号器", Value: kit.Dict(
			kit.MDB_META, kit.Dict("pid", "var/run/ice.pid"),
			kit.MDB_LIST, kit.List(),
			kit.MDB_HASH, kit.Dict(
				"2", []interface{}{"exit"},
				"3", []interface{}{"exit", "1"},
				"15", []interface{}{"exit"},
				"30", []interface{}{"exit"},
				"31", []interface{}{"exit", "1"},
				"28", "WINCH",
			),
		)},
		ice.GDB_TIMER: {Name: "timer", Help: "定时器", Value: kit.Data("tick", "100ms")},
		ice.GDB_EVENT: {Name: "event", Help: "触发器", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 进程标识
			m.Cmd("nfs.save", kit.Select(m.Conf(ice.GDB_SIGNAL, "meta.pid"),
				m.Conf(ice.CLI_RUNTIME, "conf.ctx_pid")), m.Conf(ice.CLI_RUNTIME, "host.pid"))

			if f, ok := m.Target().Server().(*Frame); ok {
				// 注册信号
				f.s = make(chan os.Signal, ice.ICE_CHAN)
				m.Richs(ice.GDB_SIGNAL, nil, "*", func(key string, value string) {
					m.Log(ice.LOG_LISTEN, "%s: %s", key, value)
					signal.Notify(f.s, syscall.Signal(kit.Int(key)))
				})
				// 启动心跳
				f.t = time.Tick(kit.Duration(m.Cap(ice.CTX_STREAM, m.Conf(ice.GDB_TIMER, "meta.tick"))))
				// 分发事件
				f.d = make(chan []string, ice.ICE_CHAN)
			}
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, ok := m.Target().Server().(*Frame); ok {
				// 停止心跳
				close(f.s)
				// 停止事件
				close(f.d)
			}
		}},
		ice.GDB_SIGNAL: {Name: "signal", Help: "信号器", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "listen":
				// 监听信号
				m.Rich(ice.GDB_SIGNAL, arg[0], arg[1:])
			}
		}},
		ice.GDB_TIMER: {Name: "timer", Help: "定时器", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "repeat":
				// 周期命令
				m.Rich(ice.GDB_TIMER, nil, kit.Dict(
					"next", time.Now().Add(kit.Duration(arg[1])).Unix(),
					"interval", arg[1], "cmd", arg[2:],
				))
			}
		}},
		ice.GDB_EVENT: {Name: "event", Help: "触发器", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "listen":
				// 监听事件
				m.Grow(ice.GDB_EVENT, arg[1], kit.Dict("cmd", arg[2:]))
				m.Log(ice.LOG_LISTEN, "%s: %v", arg[1], arg[2:])
			case "action":
				// 触发事件
				if f, ok := m.Target().Server().(*Frame); ok {
					f.d <- arg[1:]
				}
			}
		}},
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
