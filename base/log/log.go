package log

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"bufio"
	"path"
)

type Log struct {
	m *ice.Message
	p string
	l string
	s string
}

type Frame struct{ p chan *Log }

func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	f.p = make(chan *Log, 4096)
	ice.Log = func(msg *ice.Message, p, l, s string) {
		f.p <- &Log{m: msg, p: p, l: l, s: s}
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

			file := kit.Select("bench", m.Conf("show", kit.Keys(l.l, "file")))
			view := m.Confm("view", m.Conf("show", kit.Keys(l.l, "view")))
			bio := m.Confv("file", file+".file").(*bufio.Writer)

			bio.WriteString(l.p)
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

var Index = &ice.Context{Name: "log", Help: "日志模块",
	Configs: map[string]*ice.Config{
		"file": {Name: "file", Help: "日志文件", Value: kit.Dict(
			"watch", kit.Dict("path", "var/log/watch.log", "list", []string{
				ice.LOG_CREATE, ice.LOG_REMOVE,
				ice.LOG_INSERT, ice.LOG_DELETE,
				ice.LOG_SELECT, ice.LOG_MODIFY,
				ice.LOG_EXPORT, ice.LOG_IMPORT,
				ice.LOG_ENABLE,
			}),
			"bench", kit.Dict("path", "var/log/bench.log", "list", []string{}),
			"error", kit.Dict("path", "var/log/error.log", "list", []string{
				ice.LOG_WARN, ice.LOG_ERROR,
			}),
			"trace", kit.Dict("path", "var/log/trace.log", "list", []string{}),
		)},
		"view": {Name: "view", Help: "日志格式", Value: kit.Dict(
			"green", kit.Dict("prefix", "\033[32m", "suffix", "\033[0m", "list", []string{
				ice.LOG_START, ice.LOG_SERVE,
				ice.LOG_CMDS,
			}),
			"yellow", kit.Dict("prefix", "\033[33m", "suffix", "\033[0m", "list", []string{
				ice.LOG_AUTH, ice.LOG_COST,
			}),
			"red", kit.Dict("prefix", "\033[31m", "suffix", "\033[0m", "list", []string{
				ice.LOG_WARN, ice.LOG_CLOSE,
			}),
		)},
		"show": {Name: "show", Help: "日志分流", Value: kit.Dict()},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Confm("view", nil, func(key string, value map[string]interface{}) {
				kit.Fetch(value["list"], func(index int, k string) {
					m.Conf("show", kit.Keys(k, "view"), key)
				})
			})
			m.Confm("file", nil, func(key string, value map[string]interface{}) {
				kit.Fetch(value["list"], func(index int, k string) {
					m.Conf("show", kit.Keys(k, "file"), key)
				})
				// 日志文件
				if f, p, e := kit.Create(kit.Format(value["path"])); m.Assert(e) {
					m.Cap(ice.CTX_STREAM, path.Base(p))
					value["file"] = bufio.NewWriter(f)
					m.Log_CREATE(kit.MDB_FILE, p)
				}
			})
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
