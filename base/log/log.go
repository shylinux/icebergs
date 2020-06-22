package log

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"bufio"
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
		select {
		case l, ok := <-f.p:
			if !ok {
				break
			}
			// 日志文件
			file := kit.Select("bench", m.Conf("show", kit.Keys(l.l, "file")))
			// 日志格式
			view := m.Confm("view", m.Conf("show", kit.Keys(l.l, "view")))

			// 输出日志
			bio := m.Confv("file", file+".file").(*bufio.Writer)
			bio.WriteString(l.m.Format("prefix"))
			bio.WriteString(" ")
			if p, ok := view["prefix"].(string); ok {
				bio.WriteString(p)
			}
			bio.WriteString(l.l)
			bio.WriteString(" ")
			bio.WriteString(l.s)
			if p, ok := view["suffix"].(string); ok {
				bio.WriteString(p)
			}
			bio.WriteString("\n")
			bio.Flush()
		}
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

const (
	ERROR = "error"
	TRACE = "trace"
)

var Index = &ice.Context{Name: "log", Help: "日志模块",
	Configs: map[string]*ice.Config{
		"file": &ice.Config{Name: "file", Help: "日志文件", Value: kit.Dict(
			"watch", kit.Dict("path", "var/log/watch.log"),
			"bench", kit.Dict("path", "var/log/bench.log"),
			"error", kit.Dict("path", "var/log/error.log"),
			"trace", kit.Dict("path", "var/log/trace.log"),
		)},
		"view": &ice.Config{Name: "view", Help: "日志格式", Value: kit.Dict(
			"red", kit.Dict("prefix", "\033[31m", "suffix", "\033[0m"),
			"green", kit.Dict("prefix", "\033[32m", "suffix", "\033[0m"),
			"yellow", kit.Dict("prefix", "\033[33m", "suffix", "\033[0m"),
		)},
		"show": &ice.Config{Name: "show", Help: "日志分流", Value: kit.Dict(
			// 数据
			ice.LOG_ENABLE, kit.Dict("file", "watch"),
			ice.LOG_IMPORT, kit.Dict("file", "watch"),
			ice.LOG_EXPORT, kit.Dict("file", "watch"),
			ice.LOG_CREATE, kit.Dict("file", "watch"),
			ice.LOG_REMOVE, kit.Dict("file", "watch"),
			ice.LOG_INSERT, kit.Dict("file", "watch"),
			ice.LOG_DELETE, kit.Dict("file", "watch"),
			ice.LOG_MODIFY, kit.Dict("file", "watch"),
			ice.LOG_SELECT, kit.Dict("file", "watch"),

			// 事件
			ice.LOG_LISTEN, kit.Dict("file", "bench"),
			ice.LOG_ACCEPT, kit.Dict("file", "bench", "view", "green"),
			ice.LOG_FINISH, kit.Dict("file", "bench", "view", "red"),
			ice.LOG_SIGNAL, kit.Dict("file", "bench"),
			ice.LOG_EVENTS, kit.Dict("file", "bench"),
			ice.LOG_TIMERS, kit.Dict("file", "bench"),

			// 状态
			ice.LOG_BEGIN, kit.Dict("file", "bench"),
			ice.LOG_START, kit.Dict("file", "bench", "view", "green"),
			ice.LOG_SERVE, kit.Dict("file", "bench", "view", "green"),
			ice.LOG_CLOSE, kit.Dict("file", "bench", "view", "red"),

			// 分类
			ice.LOG_AUTH, kit.Dict("file", "bench", "view", "yellow"),
			ice.LOG_CMDS, kit.Dict("file", "bench", "view", "green"),
			ice.LOG_COST, kit.Dict("file", "bench", "view", "yellow"),
			ice.LOG_INFO, kit.Dict("file", "bench"),
			ice.LOG_WARN, kit.Dict("file", "bench", "view", "red"),
			ice.LOG_ERROR, kit.Dict("file", "error"),
			ice.LOG_TRACE, kit.Dict("file", "trace"),
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if os.Getenv("ctx_mod") != "" {
				m.Confm("file", nil, func(key string, value map[string]interface{}) {
					// 日志文件
					if f, p, e := kit.Create(kit.Format(value["path"])); m.Assert(e) {
						m.Cap(ice.CTX_STREAM, path.Base(p))
						value["file"] = bufio.NewWriter(f)
						m.Log_CREATE(kit.MDB_FILE, p)
					}
				})
			}
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, ok := m.Target().Server().(*Frame); ok {
				// 关闭日志
				ice.Log = nil
				close(f.p)
			}
		}},
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
