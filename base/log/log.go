package log

import (
	"bufio"
	"fmt"
	"path"
	"strings"
	"sync/atomic"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

type Log struct {
	c       bool
	p, l, s string
}
type Frame struct{ p chan *Log }

func (f *Frame) Begin(m *ice.Message, arg ...string) {
}
func (f *Frame) Start(m *ice.Message, arg ...string) {
	if !ice.HasVar() {
		return
	}
	mdb.Confm(m, FILE, nil, func(k string, v ice.Map) {
		if mdb.Conf(m, k, kit.Keym(mdb.DISABLE)) == ice.TRUE {
			return
		}
		if f, p, e := logs.CreateFile(kit.Format(v[nfs.PATH])); e == nil {
			m.Logs(nfs.SAVE, nfs.FILE, p)
			v[FILE] = bufio.NewWriter(f)
		}
	})
	f.p = make(chan *Log, ice.MOD_BUFS)
	ice.Info.Log = func(m *ice.Message, p, l, s string) {
		f.p <- &Log{c: m.Option(ice.LOG_DEBUG) == ice.TRUE, p: p, l: l, s: s}
	}
	for {
		select {
		case l, ok := <-f.p:
			if !ok {
				return
			}
			kit.For([]string{m.Conf(SHOW, kit.Keys(l.l, FILE)), BENCH}, func(file string) {
				if file == "" {
					return
				}
				conf := m.Confv(FILE, file)
				bio := kit.Value(conf, FILE).(*bufio.Writer)
				if bio == nil {
					return
				}
				defer bio.Flush()
				defer fmt.Fprintln(bio)
				fmt.Fprint(bio, l.p, lex.SP)
				view := mdb.Confm(m, VIEW, m.Conf(SHOW, kit.Keys(l.l, VIEW)))
				kit.If(ice.Info.Colors || l.c, func() { bio.WriteString(kit.Format(view[PREFIX])) })
				defer kit.If(ice.Info.Colors || l.c, func() { bio.WriteString(kit.Format(view[SUFFIX])) })
				fmt.Fprint(bio, l.l, lex.SP, l.s)
			})
		}
	}
}
func (f *Frame) Close(m *ice.Message, arg ...string) { ice.Info.Log = nil; close(f.p) }

const (
	PREFIX  = "prefix"
	SUFFIX  = "suffix"
	TRACEID = "traceid"
)
const (
	GREEN  = "green"
	YELLOW = "yellow"
	RED    = "red"
)
const (
	FILE = "file"
	VIEW = "view"
	SHOW = "show"
)
const (
	BENCH_LOG = "bench.log"
	DEBUG_LOG = "debug.log"
	ERROR_LOG = "error.log"
	WATCH_LOG = "watch.log"
)
const LOG = "log"

var Index = &ice.Context{Name: LOG, Help: "日志模块", Configs: ice.Configs{
	FILE: {Name: FILE, Help: "日志文件", Value: kit.Dict(
		BENCH, kit.Dict(nfs.PATH, path.Join(ice.VAR_LOG, BENCH_LOG), mdb.LIST, []string{}),
		DEBUG, kit.Dict(nfs.PATH, path.Join(ice.VAR_LOG, DEBUG_LOG), mdb.LIST, []string{ice.LOG_DEBUG}),
		ERROR, kit.Dict(nfs.PATH, path.Join(ice.VAR_LOG, ERROR_LOG), mdb.LIST, []string{ice.LOG_WARN, ice.LOG_ERROR}),
		WATCH, kit.Dict(nfs.PATH, path.Join(ice.VAR_LOG, WATCH_LOG), mdb.LIST, []string{mdb.CREATE, mdb.REMOVE, mdb.INSERT, mdb.DELETE, mdb.MODIFY, mdb.EXPORT, mdb.IMPORT}),
	)},
	VIEW: {Name: VIEW, Help: "日志格式", Value: kit.Dict(
		GREEN, kit.Dict(PREFIX, "\033[32m", SUFFIX, "\033[0m", mdb.LIST, []string{ice.CTX_START, ice.LOG_CMDS}),
		YELLOW, kit.Dict(PREFIX, "\033[33m", SUFFIX, "\033[0m", mdb.LIST, []string{ice.LOG_AUTH, ice.LOG_COST}),
		RED, kit.Dict(PREFIX, "\033[31m", SUFFIX, "\033[0m", mdb.LIST, []string{ice.CTX_CLOSE, ice.LOG_WARN}),
	)},
	SHOW: {Name: SHOW, Help: "日志分流", Value: kit.Dict()},
}, Commands: ice.Commands{
	ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		ice.Info.Load(m)
		mdb.Confm(m, FILE, nil, func(key string, value ice.Map) {
			kit.For(value[mdb.LIST], func(index int, k string) { m.Conf(SHOW, kit.Keys(k, FILE), key) })
		})
		mdb.Confm(m, VIEW, nil, func(key string, value ice.Map) {
			kit.For(value[mdb.LIST], func(index int, k string) { m.Conf(SHOW, kit.Keys(k, VIEW), key) })
		})
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {
		ice.Info.Save(m)
	}},
}}

func init() { ice.Index.Register(Index, &Frame{}, TAIL) }

func init() {
	ice.Info.Traceid = "short"
	ice.Pulse.Option("work.id", "0")
	ice.Pulse.Option("task.id", "0")
	ice.Pulse.Option(ice.LOG_TRACEID, Traceid(ice.Pulse))
}

var _trace_count int64

func Traceid(m *ice.Message) (traceid string) {
	ls := []string{}
	kit.For(kit.Split(ice.Info.Traceid), func(key string) {
		switch key {
		case "short":
			if len(ls) == 0 {
				ls = append(ls, kit.Hashs(mdb.UNIQ)[:6])
			}
		case "long":
			if len(ls) == 0 {
				ls = append(ls, kit.Hashs(mdb.UNIQ))
			}
		case "node":
			ls = append(ls, ice.Info.NodeName)
		case "hide":
			ls = ls[:0]
		}
	})
	kit.If(len(ls) > 0, func() { ls = append(ls, kit.Format(atomic.AddInt64(&_trace_count, 1))) })
	return strings.Join(ls, "-")
}
