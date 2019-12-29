package log

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"fmt"
	"os"
	"path"
)

type Log struct {
	m *ice.Message
	l string
	s string
}

type Frame struct {
	p chan *Log
}

func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	// 日志管道
	f.p = make(chan *Log, 1024)
	ice.Log = func(msg *ice.Message, level string, str string) {
		f.p <- &Log{m: msg, l: level, s: str}
	}
	return f
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	for {
		if l, ok := <-f.p; !ok {
			break
		} else {
			// 日志文件
			file := kit.Select("bench", m.Conf("show", l.l+".file"))
			f := m.Confv("file", file+".file").(*os.File)

			// 日志内容
			ls := []string{l.m.Format("prefix"), " "}
			ls = append(ls, m.Conf("show", l.l+".prefix"),
				l.l, " ", l.s, m.Conf("show", l.l+".suffix"), "\n")

			// 输出日志
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
		"file": &ice.Config{Name: "file", Help: "日志文件", Value: kit.Dict(
			"watch", kit.Dict("path", "var/log/watch.log"),
			"bench", kit.Dict("path", "var/log/bench.log"),
			"error", kit.Dict("path", "var/log/error.log"),
			"trace", kit.Dict("path", "var/log/trace.log"),
		)},
		"show": &ice.Config{Name: "show", Help: "日志格式", Value: kit.Dict(
			ice.LOG_ENABLE, kit.Dict("file", "watch"),
			ice.LOG_IMPORT, kit.Dict("file", "watch"),
			ice.LOG_CREATE, kit.Dict("file", "watch"),
			ice.LOG_INSERT, kit.Dict("file", "watch"),
			ice.LOG_EXPORT, kit.Dict("file", "watch"),

			ice.LOG_LISTEN, kit.Dict("file", "bench"),
			ice.LOG_SIGNAL, kit.Dict("file", "bench"),
			ice.LOG_TIMERS, kit.Dict("file", "bench"),
			ice.LOG_EVENTS, kit.Dict("file", "bench"),

			ice.LOG_BEGIN, kit.Dict("file", "bench"),
			ice.LOG_START, kit.Dict("file", "bench", "prefix", "\033[32m", "suffix", "\033[0m"),
			ice.LOG_SERVE, kit.Dict("file", "bench", "prefix", "\033[32m", "suffix", "\033[0m"),
			ice.LOG_CLOSE, kit.Dict("file", "bench", "prefix", "\033[31m", "suffix", "\033[0m"),

			ice.LOG_CMDS, kit.Dict("file", "bench", "prefix", "\033[32m", "suffix", "\033[0m"),
			ice.LOG_COST, kit.Dict("file", "bench", "prefix", "\033[33m", "suffix", "\033[0m"),
			ice.LOG_INFO, kit.Dict("file", "bench"),
			ice.LOG_WARN, kit.Dict("file", "bench", "prefix", "\033[31m", "suffix", "\033[0m"),
			ice.LOG_ERROR, kit.Dict("file", "error"),
			ice.LOG_TRACE, kit.Dict("file", "trace"),
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Confm("file", nil, func(key string, value map[string]interface{}) {
				// 日志文件
				if f, p, e := kit.Create(kit.Format(value["path"])); m.Assert(e) {
					m.Cap(ice.CTX_STREAM, path.Base(p))
					m.Log("create", "%s: %s", key, p)
					value["file"] = f
				}
			})
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, ok := m.Target().Server().(*Frame); ok {
				// 关闭日志
				ice.Log = nil
				close(f.p)
			}
		}},
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
