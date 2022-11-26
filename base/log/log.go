package log

import (
	"bufio"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

type Log struct {
	m *ice.Message
	p string
	l string
	s string
}

type Frame struct{ p chan *Log }

func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	f.p = make(chan *Log, ice.MOD_BUFS)
	ice.Info.Log = func(m *ice.Message, p, l, s string) { f.p <- &Log{m: m, p: p, l: l, s: s} }
	return f
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	for {
		select {
		case l, ok := <-f.p:
			if !ok {
				return true
			}
			for _, file := range []string{m.Conf(SHOW, kit.Keys(l.l, FILE)), BENCH} {
				if file == "" {
					continue
				}
				view := m.Confm(VIEW, m.Conf(SHOW, kit.Keys(l.l, VIEW)))
				bio := m.Confv(FILE, kit.Keys(file, FILE)).(*bufio.Writer)
				if bio == nil {
					continue
				}
				bio.WriteString(l.p)
				bio.WriteString(ice.SP)
				if ice.Info.Colors == true {
					if p, ok := view[PREFIX].(string); ok {
						bio.WriteString(p)
					}
				}
				bio.WriteString(l.l)
				bio.WriteString(ice.SP)
				bio.WriteString(l.s)
				if ice.Info.Colors == true {
					if p, ok := view[SUFFIX].(string); ok {
						bio.WriteString(p)
					}
				}
				bio.WriteString(ice.NL)
				bio.Flush()
			}
		}
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	ice.Info.Log = nil
	close(f.p)
	return true
}
func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server { return &Frame{} }

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
	BENCH = "bench"
	WATCH = "watch"
	ERROR = "error"
	TRACE = "trace"
)
const (
	FILE = "file"
	VIEW = "view"
	SHOW = "show"
)
const LOG = "log"

var Index = &ice.Context{Name: LOG, Help: "日志模块", Configs: ice.Configs{
	FILE: {Name: FILE, Help: "日志文件", Value: kit.Dict(
		BENCH, kit.Dict(nfs.PATH, path.Join(ice.VAR_LOG, "bench.log"), mdb.LIST, []string{}),
		WATCH, kit.Dict(nfs.PATH, path.Join(ice.VAR_LOG, "watch.log"), mdb.LIST, []string{
			mdb.CREATE, mdb.REMOVE, mdb.INSERT, mdb.DELETE, mdb.MODIFY, mdb.SELECT, mdb.EXPORT, mdb.IMPORT,
		}),
		ERROR, kit.Dict(nfs.PATH, path.Join(ice.VAR_LOG, "error.log"), mdb.LIST, []string{ice.LOG_WARN, ice.LOG_ERROR}),
		TRACE, kit.Dict(nfs.PATH, path.Join(ice.VAR_LOG, "trace.log"), mdb.LIST, []string{ice.LOG_DEBUG}),
	)},
	VIEW: {Name: VIEW, Help: "日志格式", Value: kit.Dict(
		GREEN, kit.Dict(PREFIX, "\033[32m", SUFFIX, "\033[0m", mdb.LIST, []string{ice.CTX_START, ice.LOG_CMDS}),
		YELLOW, kit.Dict(PREFIX, "\033[33m", SUFFIX, "\033[0m", mdb.LIST, []string{ice.LOG_AUTH, ice.LOG_COST}),
		RED, kit.Dict(PREFIX, "\033[31m", SUFFIX, "\033[0m", mdb.LIST, []string{ice.CTX_CLOSE, ice.LOG_WARN}),
	)},
	SHOW: {Name: SHOW, Help: "日志分流", Value: kit.Dict()},
}, Commands: ice.Commands{
	ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		m.Confm(VIEW, nil, func(key string, value ice.Map) {
			kit.Fetch(value[mdb.LIST], func(index int, k string) { m.Conf(SHOW, kit.Keys(k, VIEW), key) })
		})
		m.Confm(FILE, nil, func(key string, value ice.Map) {
			kit.Fetch(value[mdb.LIST], func(index int, k string) { m.Conf(SHOW, kit.Keys(k, FILE), key) })
			if f, p, e := logs.CreateFile(kit.Format(value[nfs.PATH])); e == nil {
				m.Cap(ice.CTX_STREAM, path.Base(p))
				value[FILE] = bufio.NewWriter(f)
				m.Logs(mdb.CREATE, nfs.FILE, p)
			}
		})
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {}},
}}

func init() { ice.Index.Register(Index, &Frame{}) }
