package log

import (
	"bufio"
	"fmt"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

type Log struct {
	p string
	l string
	s string
}

type Frame struct{ p chan *Log }

func (f *Frame) Begin(m *ice.Message, arg ...string) {
	f.p = make(chan *Log, ice.MOD_BUFS)
	ice.Info.Log = func(m *ice.Message, p, l, s string) { f.p <- &Log{p: p, l: l, s: s} }
}
func (f *Frame) Start(m *ice.Message, arg ...string) {
	m.Option("_lock", m.PrefixKey())
	mdb.Confm(m, FILE, nil, func(key string, value ice.Map) {
		if f, p, e := logs.CreateFile(kit.Format(value[nfs.PATH])); e == nil {
			value[FILE] = bufio.NewWriter(f)
			m.Logs(nfs.SAVE, nfs.FILE, p)
		}
	})
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
				bio := m.Confv(FILE, kit.Keys(file, FILE)).(*bufio.Writer)
				if bio == nil {
					return
				}
				defer bio.Flush()
				defer fmt.Fprintln(bio)
				fmt.Fprint(bio, l.p, ice.SP)
				view := mdb.Confm(m, VIEW, m.Conf(SHOW, kit.Keys(l.l, VIEW)))
				kit.If(ice.Info.Colors, func() { bio.WriteString(kit.Format(view[PREFIX])) })
				defer kit.If(ice.Info.Colors, func() { bio.WriteString(kit.Format(view[SUFFIX])) })
				fmt.Fprint(bio, l.l, ice.SP, l.s)
			})
		}
	}
}
func (f *Frame) Close(m *ice.Message, arg ...string) {
	ice.Info.Log = nil
	close(f.p)
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
	BENCH = "bench"
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
		DEBUG, kit.Dict(nfs.PATH, path.Join(ice.VAR_LOG, "debug.log"), mdb.LIST, []string{ice.LOG_DEBUG}),
		ERROR, kit.Dict(nfs.PATH, path.Join(ice.VAR_LOG, "error.log"), mdb.LIST, []string{ice.LOG_WARN, ice.LOG_ERROR}),
		WATCH, kit.Dict(nfs.PATH, path.Join(ice.VAR_LOG, "watch.log"), mdb.LIST, []string{mdb.CREATE, mdb.REMOVE, mdb.INSERT, mdb.DELETE, mdb.MODIFY, mdb.EXPORT, mdb.IMPORT}),
	)},
	VIEW: {Name: VIEW, Help: "日志格式", Value: kit.Dict(
		GREEN, kit.Dict(PREFIX, "\033[32m", SUFFIX, "\033[0m", mdb.LIST, []string{ice.CTX_START, ice.LOG_CMDS}),
		YELLOW, kit.Dict(PREFIX, "\033[33m", SUFFIX, "\033[0m", mdb.LIST, []string{ice.LOG_AUTH, ice.LOG_COST}),
		RED, kit.Dict(PREFIX, "\033[31m", SUFFIX, "\033[0m", mdb.LIST, []string{ice.CTX_CLOSE, ice.LOG_WARN}),
	)},
	SHOW: {Name: SHOW, Help: "日志分流", Value: kit.Dict()},
}, Commands: ice.Commands{
	ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		mdb.Confm(m, FILE, nil, func(key string, value ice.Map) {
			kit.For(value[mdb.LIST], func(index int, k string) { m.Conf(SHOW, kit.Keys(k, FILE), key) })
		})
		mdb.Confm(m, VIEW, nil, func(key string, value ice.Map) {
			kit.For(value[mdb.LIST], func(index int, k string) { m.Conf(SHOW, kit.Keys(k, VIEW), key) })
		})
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {}},
}}

func init() { ice.Index.Register(Index, &Frame{}, TAIL) }
