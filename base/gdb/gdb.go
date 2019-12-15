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
	signal chan os.Signal
}

func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	if f, p, e := kit.Create(m.Conf("logpid")); m.Assert(e) {
		defer f.Close()
		f.WriteString(kit.Format(os.Getpid()))
		m.Log("info", "pid %d %s", os.Getpid(), p)
	}

	f.signal = make(chan os.Signal, 10)
	m.Confm("signal", nil, func(sig string, action string) {
		m.Log("signal", "add %s: %s", sig, action)
		signal.Notify(f.signal, syscall.Signal(kit.Int(sig)))
	})

	m.Gos(m, func(m *ice.Message) {
		if s, ok := <-f.signal; ok {
			m.Log("info", "signal %v", s)
			m.Cmd(m.Confv("signal", kit.Format(s)))
		}
	})
	m.Gos(m, func(m *ice.Message) {
		for {
			time.Sleep(100 * time.Millisecond)
			now := int(time.Now().Unix())
			m.Confm("timer", "hash", func(key string, value map[string]interface{}) {
				if kit.Int(value["next"]) <= now {
					m.Log("info", "timer %v %v", key, value["next"])
					m.Cmd(value["cmd"])
					value["next"] = now + int(kit.Duration(value["interval"]))/int(time.Second)
				}
			})
		}
	})
	return f
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

var Index = &ice.Context{Name: "gdb", Help: "调试模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"logpid": &ice.Config{Name: "logpid", Value: "var/run/shy.pid", Help: ""},
		"signal": &ice.Config{Name: "signal", Value: map[string]interface{}{
			"2": []interface{}{"exit"}, "3": "QUIT", "15": "TERM",
			"28": "WINCH",
			"30": "restart",
			"31": "upgrade",
		}, Help: "信号"},
		"timer": {Name: "定时器", Value: map[string]interface{}{
			"meta": map[string]interface{}{},
			"hash": map[string]interface{}{},
			"list": map[string]interface{}{},
		}},
	},
	Commands: map[string]*ice.Command{
		"timer": {Name: "timer", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "start":
				h := kit.ShortKey(m.Confm("timer", "hash"), 6)

				next := time.Now().Add(kit.Duration(arg[1])).Unix()
				m.Conf("timer", "hash."+h, map[string]interface{}{
					"interval": arg[1],
					"next":     next,
					"cmd":      arg[2:],
				})
				m.Echo(h)

			case "stop":
				m.Conf("timer", "hash."+arg[1], "")
			}
		}},
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
