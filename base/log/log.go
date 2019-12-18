package log

import (
	"fmt"
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
	"os"
	"path"
)

type Log struct {
	m     *ice.Message
	level string
	str   string
}

type Frame struct {
	p chan *Log
}

func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	f.p = make(chan *Log, 100)
	ice.Log = func(msg *ice.Message, level string, str string) {
		f.p <- &Log{m: msg, level: level, str: str}
	}
	return f
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	for {
		if l, ok := <-f.p; !ok {
			break
		} else {
			file := kit.Select("bench", m.Conf("show", l.level+".file"))
			f := m.Confv("file", file+".file").(*os.File)

			ls := []string{l.m.Format("prefix"), " "}
			ls = append(ls, m.Conf("show", l.level+".prefix"),
				l.level, " ", l.str, m.Conf("show", l.level+".suffix"), "\n")

			for _, v := range ls {
				fmt.Fprintf(f, v)
			}
		}
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

var Index = &ice.Context{Name: "log", Help: "日志模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"file": &ice.Config{Name: "file", Value: map[string]interface{}{
			"bench": map[string]interface{}{"path": "var/log/bench.log"},
		}, Help: "信号"},
		"show": &ice.Config{Name: "file", Value: map[string]interface{}{
			"bench": map[string]interface{}{"file": "bench"},

			"cmd":   map[string]interface{}{"file": "bench", "prefix": "\033[32m", "suffix": "\033[0m"},
			"start": map[string]interface{}{"file": "bench", "prefix": "\033[32m", "suffix": "\033[0m"},
			"serve": map[string]interface{}{"file": "bench", "prefix": "\033[32m", "suffix": "\033[0m"},

			"cost": map[string]interface{}{"file": "bench", "prefix": "\033[33m", "suffix": "\033[0m"},

			"warn":  map[string]interface{}{"file": "bench", "prefix": "\033[31m", "suffix": "\033[0m"},
			"close": map[string]interface{}{"file": "bench", "prefix": "\033[31m", "suffix": "\033[0m"},
		}, Help: "信号"},
	},
	Commands: map[string]*ice.Command{
		"_init": {Name: "_init", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Confm("file", nil, func(key string, value map[string]interface{}) {
				if f, p, e := kit.Create(kit.Format(value["path"])); m.Assert(e) {
					m.Cap("stream", path.Base(p))
					m.Log("info", "log %s %s", key, p)
					value["file"] = f
				}
			})
		}},
		"_exit": {Name: "_exit", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			f := m.Target().Server().(*Frame)
			ice.Log = nil
			close(f.p)
		}},
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
