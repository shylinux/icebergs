package log

import (
	"bufio"
	"path"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
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
	f.p = make(chan *Log, ice.MOD_BUFS)
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

			file := kit.Select(BENCH, m.Conf(SHOW, kit.Keys(l.l, FILE)))
			view := m.Confm(VIEW, m.Conf(SHOW, kit.Keys(l.l, VIEW)))
			bio := m.Confv(FILE, kit.Keys(file, FILE)).(*bufio.Writer)

			bio.WriteString(l.p)
			bio.WriteString(ice.MOD_SP)
			if p, ok := view[PREFIX].(string); ok {
				bio.WriteString(p)
			}
			bio.WriteString(l.l)
			bio.WriteString(ice.MOD_SP)
			bio.WriteString(l.s)
			if p, ok := view[SUFFIX].(string); ok {
				bio.WriteString(p)
			}
			bio.WriteString(ice.MOD_NL)
			bio.Flush()
		}
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

const (
	PREFIX = "prefix"
	SUFFIX = "suffix"
)
const (
	GREEN  = "green"
	YELLOW = "yellow"
	RED    = "red"
)
const (
	WATCH = "watch"
	BENCH = "bench"
	ERROR = "error"
	TRACE = "trace"
)
const (
	FILE = "file"
	VIEW = "view"
	SHOW = "show"
)

var Index = &ice.Context{Name: "log", Help: "日志模块",
	Configs: map[string]*ice.Config{
		FILE: {Name: FILE, Help: "日志文件", Value: kit.Dict(
			WATCH, kit.Dict(kit.MDB_PATH, path.Join(ice.VAR_LOG, "watch.log"), kit.MDB_LIST, []string{
				ice.LOG_CREATE, ice.LOG_REMOVE,
				ice.LOG_INSERT, ice.LOG_DELETE,
				ice.LOG_SELECT, ice.LOG_MODIFY,
				ice.LOG_EXPORT, ice.LOG_IMPORT,
			}),
			BENCH, kit.Dict(kit.MDB_PATH, path.Join(ice.VAR_LOG, "bench.log"), kit.MDB_LIST, []string{}),
			ERROR, kit.Dict(kit.MDB_PATH, path.Join(ice.VAR_LOG, "error.log"), kit.MDB_LIST, []string{
				ice.LOG_WARN, ice.LOG_ERROR,
			}),
			TRACE, kit.Dict(kit.MDB_PATH, path.Join(ice.VAR_LOG, "trace.log"), kit.MDB_LIST, []string{}),
		)},
		VIEW: {Name: VIEW, Help: "日志格式", Value: kit.Dict(
			GREEN, kit.Dict(PREFIX, "\033[32m", SUFFIX, "\033[0m", kit.MDB_LIST, []string{
				ice.LOG_START, ice.LOG_SERVE,
				ice.LOG_CMDS,
			}),
			YELLOW, kit.Dict(PREFIX, "\033[33m", SUFFIX, "\033[0m", kit.MDB_LIST, []string{
				ice.LOG_AUTH, ice.LOG_COST,
			}),
			RED, kit.Dict(PREFIX, "\033[31m", SUFFIX, "\033[0m", kit.MDB_LIST, []string{
				ice.LOG_WARN, ice.LOG_CLOSE,
			}),
		)},
		SHOW: {Name: SHOW, Help: "日志分流", Value: kit.Dict()},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Confm(VIEW, nil, func(key string, value map[string]interface{}) {
				kit.Fetch(value[kit.MDB_LIST], func(index int, k string) {
					m.Conf(SHOW, kit.Keys(k, VIEW), key)
				})
			})
			m.Confm(FILE, nil, func(key string, value map[string]interface{}) {
				kit.Fetch(value[kit.MDB_LIST], func(index int, k string) {
					m.Conf(SHOW, kit.Keys(k, FILE), key)
				})
				// 日志文件
				if f, p, e := kit.Create(kit.Format(value[kit.MDB_PATH])); m.Assert(e) {
					m.Cap(ice.CTX_STREAM, path.Base(p))
					value[FILE] = bufio.NewWriter(f)
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
