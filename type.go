package ice

import (
	"github.com/shylinux/toolkits"

	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Cache struct {
	Name  string
	Help  string
	Value string
}
type Config struct {
	Name  string
	Help  string
	Value interface{}
}
type Command struct {
	Name string
	Help interface{}
	List []interface{}
	Meta map[string]interface{}
	Hand func(m *Message, c *Context, key string, arg ...string)
}
type Context struct {
	Name string
	Help string

	Caches   map[string]*Cache
	Configs  map[string]*Config
	Commands map[string]*Command

	contexts map[string]*Context
	context  *Context
	root     *Context

	begin  *Message
	start  *Message
	server Server

	wg *sync.WaitGroup
	id int
}
type Server interface {
	Spawn(m *Message, c *Context, arg ...string) Server
	Begin(m *Message, arg ...string) Server
	Start(m *Message, arg ...string) bool
	Close(m *Message, arg ...string) bool
}

func (c *Context) ID() int {
	c.id++
	return c.id
}
func (c *Context) Cap(key string, arg ...interface{}) string {
	if len(arg) > 0 {
		c.Caches[key].Value = kit.Format(arg[0])
	}
	return c.Caches[key].Value
}
func (c *Context) Run(m *Message, cmd *Command, key string, arg ...string) *Message {
	m.Hand = true
	m.Log(LOG_CMDS, "%s.%s %d %v", c.Name, key, len(arg), arg)
	cmd.Hand(m, c, key, arg...)
	return m
}
func (c *Context) Runs(m *Message, cmd string, key string, arg ...string) {
	if s, ok := m.Target().Commands[key]; ok {
		c.Run(m, s, cmd, arg...)
	}
	return
}
func (c *Context) Server() Server {
	return c.server
}
func (c *Context) Register(s *Context, x Server) *Context {
	Pulse.Log("register", "%s <- %s", c.Name, s.Name)
	if c.contexts == nil {
		c.contexts = map[string]*Context{}
	}
	c.contexts[s.Name] = s
	s.root = c.root
	s.context = c
	s.server = x
	return s
}

func (c *Context) Spawn(m *Message, name string, help string, arg ...string) *Context {
	s := &Context{Name: name, Help: help, Caches: map[string]*Cache{}, Configs: map[string]*Config{}}
	if m.target.Server != nil {
		c.Register(s, m.target.server.Spawn(m, s, arg...))
	} else {
		c.Register(s, nil)
	}
	m.target = s
	return s
}
func (c *Context) Begin(m *Message, arg ...string) *Context {
	c.Caches[CTX_FOLLOW] = &Cache{Name: CTX_FOLLOW, Value: ""}
	c.Caches[CTX_STREAM] = &Cache{Name: CTX_STREAM, Value: ""}
	c.Caches[CTX_STATUS] = &Cache{Name: CTX_STATUS, Value: ""}

	if c.context == Index {
		c.Cap(CTX_FOLLOW, c.Name)
	} else if c.context != nil {
		c.Cap(CTX_FOLLOW, kit.Keys(c.context.Cap(CTX_FOLLOW), c.Name))
	}
	m.Log(LOG_BEGIN, "%s", c.Cap(CTX_FOLLOW))
	c.Cap(CTX_STATUS, ICE_BEGIN)

	if c.begin = m; c.server != nil {
		m.TryCatch(m, true, func(m *Message) {
			// 初始化模块
			c.server.Begin(m, arg...)
		})
	}
	return c
}
func (c *Context) Start(m *Message, arg ...string) bool {
	c.start = m
	m.Hold(1)

	wait := make(chan bool)
	m.Gos(m, func(m *Message) {
		m.Log(LOG_START, "%s", c.Cap(CTX_FOLLOW))
		c.Cap(CTX_STATUS, ICE_START)
		wait <- true

		// 启动模块
		if c.server != nil {
			c.server.Start(m, arg...)
		}
		if m.Done(); m.wait != nil {
			m.wait <- true
		}
	})
	<-wait
	return true
}
func (c *Context) Close(m *Message, arg ...string) bool {
	m.Log(LOG_CLOSE, "%s", c.Cap(CTX_FOLLOW))
	c.Cap(CTX_STATUS, ICE_CLOSE)

	if c.server != nil {
		// 结束模块
		return c.server.Close(m, arg...)
	}
	return true
}

type Message struct {
	time time.Time
	code int
	Hand bool

	meta map[string][]string
	data map[string]interface{}

	messages []*Message
	message  *Message
	root     *Message

	source *Context
	target *Context
	frames interface{}

	cb   func(*Message) *Message
	W    http.ResponseWriter
	R    *http.Request
	O    io.Writer
	I    io.Reader
	wait chan bool
}

func (m *Message) Time(args ...interface{}) string {
	// [duration] [format [args...]]
	t := m.time
	if len(args) > 0 {
		switch arg := args[0].(type) {
		case string:
			if d, e := time.ParseDuration(arg); e == nil {
				// 时间偏移
				t, args = t.Add(d), args[1:]
			}
		}
	}
	f := ICE_TIME
	if len(args) > 0 {
		switch arg := args[0].(type) {
		case string:
			f = arg
			if len(args) > 1 {
				// 时间格式
				f = fmt.Sprintf(f, args[1:]...)
			}
		}
	}
	return t.Format(f)
}
func (m *Message) Target() *Context {
	return m.target
}
func (m *Message) Source() *Context {
	return m.source
}
func (m *Message) Format(key interface{}) string {
	switch key := key.(type) {
	case string:
		switch key {
		case "cost":
			return kit.FmtTime(kit.Int64(time.Now().Sub(m.time)))
		case "meta":
			return kit.Format(m.meta)
		case "append":
			if len(m.meta["append"]) == 0 {
				return fmt.Sprintf("%dx%d %s", 0, len(m.meta["append"]), kit.Format(m.meta["append"]))
			} else {
				return fmt.Sprintf("%dx%d %s", len(m.meta[m.meta["append"][0]]), len(m.meta["append"]), kit.Format(m.meta["append"]))
			}

		case "time":
			return m.Time()
		case "ship":
			return fmt.Sprintf("%s->%s", m.source.Name, m.target.Name)
		case "prefix":
			return fmt.Sprintf("%s %d %s->%s", m.Time(), m.code, m.source.Name, m.target.Name)

		case "chain":
			// 调用链
			ms := []*Message{}
			for msg := m; msg != nil; msg = msg.message {
				ms = append(ms, msg)
			}

			meta := append([]string{}, "\n\n")
			for i := len(ms) - 1; i >= 0; i-- {
				msg := ms[i]

				meta = append(meta, fmt.Sprintf("%s ", msg.Format("prefix")))
				if len(msg.meta[MSG_DETAIL]) > 0 {
					meta = append(meta, fmt.Sprintf("detail:%d %v", len(msg.meta[MSG_DETAIL]), msg.meta[MSG_DETAIL]))
				}

				if len(msg.meta[MSG_OPTION]) > 0 {
					meta = append(meta, fmt.Sprintf("option:%d %v\n", len(msg.meta[MSG_OPTION]), msg.meta[MSG_OPTION]))
					for _, k := range msg.meta[MSG_OPTION] {
						if v, ok := msg.meta[k]; ok {
							meta = append(meta, fmt.Sprintf("    %s: %d %v\n", k, len(v), v))
						}
					}
				} else {
					meta = append(meta, "\n")
				}

				if len(msg.meta[MSG_APPEND]) > 0 {
					meta = append(meta, fmt.Sprintf("  append:%d %v\n", len(msg.meta[MSG_APPEND]), msg.meta[MSG_APPEND]))
					for _, k := range msg.meta[MSG_APPEND] {
						if v, ok := msg.meta[k]; ok {
							meta = append(meta, fmt.Sprintf("    %s: %d %v\n", k, len(v), v))
						}
					}
				}
				if len(msg.meta[MSG_RESULT]) > 0 {
					meta = append(meta, fmt.Sprintf("  result:%d %v\n", len(msg.meta[MSG_RESULT]), msg.meta[MSG_RESULT]))
				}
			}
			return strings.Join(meta, "")
		case "stack":
			// 调用栈
			pc := make([]uintptr, 100)
			pc = pc[:runtime.Callers(5, pc)]
			frames := runtime.CallersFrames(pc)

			meta := []string{}
			for {
				frame, more := frames.Next()
				file := strings.Split(frame.File, "/")
				name := strings.Split(frame.Function, "/")
				meta = append(meta, fmt.Sprintf("\n%s:%d\t%s", file[len(file)-1], frame.Line, name[len(name)-1]))
				if !more {
					break
				}
			}
			return strings.Join(meta, "")
		}
	case []byte:
		json.Unmarshal(key, &m.meta)
	}
	return m.time.Format(ICE_TIME)
}
func (m *Message) Formats(key string) string {
	switch key {
	case "meta":
		return kit.Formats(m.meta)
	}
	return m.Format(key)
}
func (m *Message) Spawns(arg ...interface{}) *Message {
	msg := m.Spawn(arg...)
	msg.code = m.target.root.ID()
	m.messages = append(m.messages, msg)
	return msg
}
func (m *Message) Spawn(arg ...interface{}) *Message {
	msg := &Message{
		code: -1, time: time.Now(),

		meta: map[string][]string{},
		data: map[string]interface{}{},

		message: m, root: m.root,

		source: m.target,
		target: m.target,

		W: m.W, R: m.R,
		O: m.O, I: m.I,
	}

	if len(arg) > 0 {
		switch val := arg[0].(type) {
		case *Context:
			msg.target = val
		case []byte:
			json.Unmarshal(val, &msg.meta)
		}
	}
	return msg
}

func (m *Message) Add(key string, arg ...string) *Message {
	switch key {
	case MSG_DETAIL, MSG_RESULT:
		m.meta[key] = append(m.meta[key], arg...)

	case MSG_OPTION, MSG_APPEND:
		if len(arg) > 0 {
			if kit.IndexOf(m.meta[key], arg[0]) == -1 {
				m.meta[key] = append(m.meta[key], arg[0])
			}
			m.meta[arg[0]] = append(m.meta[arg[0]], arg[1:]...)
		}
	}
	return m
}
func (m *Message) Set(key string, arg ...string) *Message {
	switch key {
	case MSG_DETAIL, MSG_RESULT:
		delete(m.meta, key)
	case MSG_OPTION, MSG_APPEND:
		if len(arg) > 0 {
			if delete(m.meta, arg[0]); len(arg) == 1 {
				return m
			}
		} else {
			for _, k := range m.meta[key] {
				delete(m.meta, k)
			}
			delete(m.meta, key)
			return m
		}
	default:
		delete(m.meta, key)
	}
	return m.Add(key, arg...)
}
func (m *Message) Push(key string, value interface{}, arg ...interface{}) *Message {
	switch value := value.(type) {
	case map[string]string:
	case map[string]interface{}:
		if key == "detail" {
			// 格式转换
			value = kit.KeyValue(map[string]interface{}{}, "", value)
		}

		// 键值排序
		list := []string{}
		if len(arg) > 0 {
			list = kit.Simple(arg[0])
		} else {
			for k := range value {
				list = append(list, k)
			}
			sort.Strings(list)
		}

		// 追加数据
		for _, k := range list {
			switch key {
			case "detail":
				m.Add(MSG_APPEND, "key", k)
				m.Add(MSG_APPEND, "value", kit.Format(value[k]))
			default:
				if k == "key" {
					m.Add(MSG_APPEND, k, key)
				} else {
					m.Add(MSG_APPEND, k, kit.Format(kit.Value(value, k)))
				}
			}
		}
		return m
	}

	for _, v := range kit.Simple(value) {
		m.Add(MSG_APPEND, key, v)
	}
	return m
}
func (m *Message) Echo(str string, arg ...interface{}) *Message {
	if len(arg) > 0 {
		str = fmt.Sprintf(str, arg...)
	}
	m.meta[MSG_RESULT] = append(m.meta[MSG_RESULT], str)
	return m
}
func (m *Message) Copy(msg *Message, arg ...string) *Message {
	if m == msg {
		return m
	}
	if m == nil {
		return m
	}
	if len(arg) > 0 {
		// 精确复制
		for _, k := range arg[1:] {
			if kit.IndexOf(m.meta[arg[0]], k) == -1 {
				m.meta[arg[0]] = append(m.meta[arg[0]], k)
			}
			m.meta[k] = append(m.meta[k], msg.meta[k]...)
		}
		return m
	}

	// 复制选项
	for _, k := range msg.meta[MSG_OPTION] {
		if kit.IndexOf(m.meta[MSG_OPTION], k) == -1 {
			m.meta[MSG_OPTION] = append(m.meta[MSG_OPTION], k)
		}
		if _, ok := msg.meta[k]; ok {
			m.meta[k] = msg.meta[k]
		} else {
			m.data[k] = msg.data[k]
		}
	}

	// 复制数据
	for _, k := range msg.meta[MSG_APPEND] {
		if kit.IndexOf(m.meta[MSG_OPTION], k) > -1 && len(m.meta[k]) > 0 {
			m.meta[k] = m.meta[k][:0]
		}
		if kit.IndexOf(m.meta[MSG_APPEND], k) == -1 {
			m.meta[MSG_APPEND] = append(m.meta[MSG_APPEND], k)
		}
		m.meta[k] = append(m.meta[k], msg.meta[k]...)
	}

	// 复制文本
	m.meta[MSG_RESULT] = append(m.meta[MSG_RESULT], msg.meta[MSG_RESULT]...)
	return m
}
func (m *Message) Sort(key string, arg ...string) *Message {
	// 排序方法
	cmp := "str"
	if len(arg) > 0 && arg[0] != "" {
		cmp = arg[0]
	} else {
		cmp = "int"
		for _, v := range m.meta[key] {
			if _, e := strconv.Atoi(v); e != nil {
				cmp = "str"
			}
		}
	}

	// 排序因子
	number := map[int]int{}
	table := []map[string]string{}
	m.Table(func(index int, line map[string]string, head []string) {
		table = append(table, line)
		switch cmp {
		case "int":
			number[index] = kit.Int(line[key])
		case "int_r":
			number[index] = -kit.Int(line[key])
		case "time":
			number[index] = kit.Time(line[key])
		case "time_r":
			number[index] = -kit.Time(line[key])
		}
	})

	// 排序数据
	for i := 0; i < len(table)-1; i++ {
		for j := i + 1; j < len(table); j++ {
			result := false
			switch cmp {
			case "", "str":
				if table[i][key] > table[j][key] {
					result = true
				}
			case "str_r":
				if table[i][key] < table[j][key] {
					result = true
				}
			default:
				if number[i] > number[j] {
					result = true
				}
			}

			if result {
				table[i], table[j] = table[j], table[i]
				number[i], number[j] = number[j], number[i]
			}
		}
	}

	// 输出数据
	for _, k := range m.meta[MSG_APPEND] {
		delete(m.meta, k)
	}
	for _, v := range table {
		for _, k := range m.meta[MSG_APPEND] {
			m.Add(MSG_APPEND, k, v[k])
		}
	}
	return m
}
func (m *Message) Table(cbs ...interface{}) *Message {
	if len(cbs) > 0 {
		switch cb := cbs[0].(type) {
		case func(int, map[string]string, []string):
			nrow := 0
			for _, k := range m.meta[MSG_APPEND] {
				if len(m.meta[k]) > nrow {
					nrow = len(m.meta[k])
				}
			}

			for i := 0; i < nrow; i++ {
				line := map[string]string{}
				for _, k := range m.meta[MSG_APPEND] {
					line[k] = kit.Select("", m.meta[k], i)
				}
				// 依次回调
				cb(i, line, m.meta[MSG_APPEND])
			}
		}
		return m
	}

	//计算列宽
	space := kit.Select(" ", m.Option("table.space"))
	depth, width := 0, map[string]int{}
	for _, k := range m.meta[MSG_APPEND] {
		if len(m.meta[k]) > depth {
			depth = len(m.meta[k])
		}
		width[k] = kit.Width(k, len(space))
		for _, v := range m.meta[k] {
			if kit.Width(v, len(space)) > width[k] {
				width[k] = kit.Width(v, len(space))
			}
		}
	}

	// 回调函数
	rows := kit.Select("\n", m.Option("table.row_sep"))
	cols := kit.Select(" ", m.Option("table.col_sep"))
	compact := m.Option("table.compact") == "true"
	cb := func(maps map[string]string, lists []string, line int) bool {
		for i, v := range lists {
			if k := m.meta[MSG_APPEND][i]; compact {
				v = maps[k]
			}

			if m.Echo(v); i < len(lists)-1 {
				m.Echo(cols)
			}
		}
		m.Echo(rows)
		return true
	}

	// 输出表头
	row := map[string]string{}
	wor := []string{}
	for _, k := range m.meta[MSG_APPEND] {
		row[k], wor = k, append(wor, k+strings.Repeat(space, width[k]-kit.Width(k, len(space))))
	}
	if !cb(row, wor, -1) {
		return m
	}

	// 输出数据
	for i := 0; i < depth; i++ {
		row := map[string]string{}
		wor := []string{}
		for _, k := range m.meta[MSG_APPEND] {
			data := ""
			if i < len(m.meta[k]) {
				data = m.meta[k][i]
			}

			row[k], wor = data, append(wor, data+strings.Repeat(space, width[k]-kit.Width(data, len(space))))
		}
		// 依次回调
		if !cb(row, wor, i) {
			break
		}
	}
	return m
}
func (m *Message) Render(cmd string, args ...interface{}) *Message {
	m.Log(LOG_EXPORT, "%s: %v", cmd, args)
	m.Optionv(MSG_OUTPUT, cmd)
	m.Optionv(MSG_ARGS, args)

	switch cmd {
	case RENDER_TEMPLATE:
		if len(args) == 1 {
			args = append(args, m)
		}
		if res, err := kit.Render(args[0].(string), args[1]); m.Assert(err) {
			m.Echo(string(res))
		}
	}
	return m
}
func (m *Message) Parse(meta string, key string, arg ...string) *Message {
	list := []string{}
	for _, line := range kit.Split(strings.Join(arg, " "), "\n") {
		ls := kit.Split(line)
		for i := 0; i < len(ls); i++ {
			if strings.HasPrefix(ls[i], "#") {
				ls = ls[:i]
				break
			}
		}
		list = append(list, ls...)
	}

	switch data := kit.Parse(nil, "", list...); meta {
	case MSG_OPTION:
		m.Option(key, data)
	case MSG_APPEND:
		m.Append(key, data)
	}
	return m
}
func (m *Message) Split(str string, field string, space string, enter string) *Message {
	indexs := []int{}
	fields := kit.Split(field, space, "{}")
	for i, l := range kit.Split(str, enter, "{}") {
		if i == 0 && (field == "" || field == "index") {
			// 表头行
			fields = kit.Split(l, space)
			if field == "index" {
				for _, v := range fields {
					indexs = append(indexs, strings.Index(l, v))
				}
			}
			continue
		}

		if len(indexs) > 0 {
			// 数据行
			for i, v := range indexs {
				if i == len(indexs)-1 {
					m.Push(kit.Select("some", fields, i), l[v:])
				} else {
					m.Push(kit.Select("some", fields, i), l[v:indexs[i+1]])
				}
			}
			continue
		}

		for i, v := range kit.Split(l, space) {
			m.Push(kit.Select("some", fields, i), v)
		}
	}
	return m
}
func (m *Message) CSV(text string) *Message {
	bio := bytes.NewBufferString(text)
	r := csv.NewReader(bio)
	heads, _ := r.Read()
	for {
		lines, e := r.Read()
		if e != nil {
			break
		}
		for i, k := range heads {
			m.Push(k, kit.Select("", lines, i))
		}
	}
	return m
}

func (m *Message) Detail(arg ...interface{}) string {
	return kit.Select("", m.meta[MSG_DETAIL], 0)
}
func (m *Message) Detailv(arg ...interface{}) []string {
	return m.meta[MSG_DETAIL]
}
func (m *Message) Optionv(key string, arg ...interface{}) interface{} {
	if len(arg) > 0 {
		// 写数据
		if kit.IndexOf(m.meta[MSG_OPTION], key) == -1 {
			m.meta[MSG_OPTION] = append(m.meta[MSG_OPTION], key)
		}

		switch str := arg[0].(type) {
		case string:
			m.meta[key] = kit.Simple(arg)
		case []string:
			m.meta[key] = str
		default:
			m.data[key] = str
		}
	}

	for msg := m; msg != nil; msg = msg.message {
		if list, ok := msg.data[key]; ok {
			// 读数据
			return list
		}
		if list, ok := msg.meta[key]; ok {
			// 读选项
			return list
		}
	}
	return nil
}
func (m *Message) Options(key string, arg ...interface{}) bool {
	return kit.Select("", kit.Simple(m.Optionv(key, arg...)), 0) != ""
}
func (m *Message) Option(key string, arg ...interface{}) string {
	return kit.Select("", kit.Simple(m.Optionv(key, arg...)), 0)
}
func (m *Message) Append(key string, arg ...interface{}) string {
	return kit.Select("", m.Appendv(key, arg...), 0)
}
func (m *Message) Appendv(key string, arg ...interface{}) []string {
	if key == "_index" {
		max := 0
		for _, k := range m.meta[MSG_APPEND] {
			if len(m.meta[k]) > max {
				max = len(m.meta[k])
			}
		}
		index := []string{}
		for i := 0; i < max; i++ {
			index = append(index, kit.Format(i))
		}
		return index
	}
	if len(arg) > 0 {
		m.meta[key] = kit.Simple(arg...)
	}
	return m.meta[key]
}
func (m *Message) Resultv(arg ...interface{}) []string {
	if len(arg) > 0 {
		m.meta[MSG_RESULT] = kit.Simple(arg...)
	}
	return m.meta[MSG_RESULT]
}
func (m *Message) Result(arg ...interface{}) string {
	if len(arg) > 0 {
		switch v := arg[0].(type) {
		case int:
			return kit.Select("", m.meta[MSG_RESULT], v)
		}
	}
	return strings.Join(m.Resultv(arg...), "")
}

func (m *Message) Logs(level string, arg ...interface{}) *Message {
	list := []string{}
	for i := 0; i < len(arg)-1; i++ {
		list = append(list, fmt.Sprintf("%v: %v", arg[i], arg[i+1]))
	}
	m.Log(level, strings.Join(list, " "))
	return m
}
func (m *Message) Log(level string, str string, arg ...interface{}) *Message {
	if str = strings.TrimSpace(fmt.Sprintf(str, arg...)); Log != nil {
		// 日志模块
		Log(m, level, str)
	}

	// 日志颜色
	prefix, suffix := "", ""
	switch level {
	case LOG_ENABLE, LOG_IMPORT, LOG_CREATE, LOG_INSERT, LOG_EXPORT:
		prefix, suffix = "\033[36;44m", "\033[0m"

	case LOG_LISTEN, LOG_SIGNAL, LOG_TIMERS, LOG_EVENTS:
		prefix, suffix = "\033[33m", "\033[0m"

	case LOG_CMDS, LOG_START, LOG_SERVE:
		prefix, suffix = "\033[32m", "\033[0m"
	case LOG_COST:
		prefix, suffix = "\033[33m", "\033[0m"
	case LOG_WARN, LOG_ERROR, LOG_CLOSE:
		prefix, suffix = "\033[31m", "\033[0m"
	}

	if os.Getenv("ctx_mod") != "" {
		// 输出日志
		fmt.Fprintf(os.Stderr, "%s %02d %9s %s%s %s%s\n",
			m.time.Format(ICE_TIME), m.code, fmt.Sprintf("%s->%s", m.source.Name, m.target.Name),
			prefix, level, str, suffix)
	}
	return m
}
func (m *Message) Cost(str string, arg ...interface{}) *Message {
	return m.Log(LOG_COST, "%s: %s", m.Format("cost"), kit.Format(str, arg...))
}
func (m *Message) Info(str string, arg ...interface{}) *Message {
	return m.Log(LOG_INFO, str, arg...)
}
func (m *Message) Warn(err bool, str string, arg ...interface{}) bool {
	if err {
		m.Echo("warn: ").Echo(str, arg...)
		return m.Log(LOG_WARN, str, arg...) != nil
	}
	return false
}
func (m *Message) Error(err bool, str string, arg ...interface{}) bool {
	if err {
		m.Echo("error: ").Echo(str, arg...)
		m.Log(LOG_ERROR, m.Format("stack"))
		m.Log(LOG_ERROR, str, arg...)
		m.Log(LOG_ERROR, m.Format("chain"))
		return true
	}
	return false
}
func (m *Message) Trace(key string, str string, arg ...interface{}) *Message {
	if m.Options(key) {
		m.Echo("trace: ").Echo(str, arg...)
		return m.Log(LOG_TRACE, str, arg...)
	}
	return m
}

func (m *Message) TryCatch(msg *Message, safe bool, hand ...func(msg *Message)) *Message {
	defer func() {
		switch e := recover(); e {
		case io.EOF:
		case nil:
		default:
			m.Log(LOG_WARN, "catch: %s", e)
			m.Log(LOG_INFO, "chain: %s", msg.Format("chain"))
			m.Log(LOG_WARN, "catch: %s", e)
			m.Log(LOG_INFO, "stack: %s", msg.Format("stack"))
			if m.Log(LOG_WARN, "catch: %s", e); len(hand) > 1 {
				// 捕获异常
				m.TryCatch(msg, safe, hand[1:]...)
			} else if !safe {
				// 抛出异常
				m.Assert(e)
			}
		}
	}()

	if len(hand) > 0 {
		// 运行函数
		hand[0](msg)
	}
	return m
}
func (m *Message) Assert(arg interface{}) bool {
	switch arg := arg.(type) {
	case nil:
		return true
	case bool:
		if arg == true {
			return true
		}
	}

	// 抛出异常
	panic(errors.New(fmt.Sprintf("error %v", arg)))
}
func (m *Message) Sleep(arg string) *Message {
	time.Sleep(kit.Duration(arg))
	return m
}
func (m *Message) Hold(n int) *Message {
	ctx := m.target.root
	if c := m.target; c.context != nil && c.context.wg != nil {
		ctx = c.context
	}

	ctx.wg.Add(n)
	m.Log(LOG_TRACE, "%s wait %s %v", ctx.Name, m.target.Name, ctx.wg)
	return m
}
func (m *Message) Done() bool {
	defer func() { recover() }()

	ctx := m.target.root
	if c := m.target; c.context != nil && c.context.wg != nil {
		ctx = c.context
	}

	m.Log(LOG_TRACE, "%s done %s %v", ctx.Name, m.target.Name, ctx.wg)
	ctx.wg.Done()
	return true
}
func (m *Message) Call(sync bool, cb func(*Message) *Message) *Message {
	if sync {
		wait := make(chan bool)
		m.cb = func(sub *Message) *Message {
			wait <- true
			return cb(sub)
		}
		<-wait
	}
	return m
}
func (m *Message) Back(sub *Message) *Message {
	if m.cb != nil {
		m.cb(sub)
	}
	return m
}
func (m *Message) Gos(msg *Message, cb func(*Message)) *Message {
	go func() { msg.TryCatch(msg, true, func(msg *Message) { cb(msg) }) }()
	return m
}

func (m *Message) Run(arg ...string) *Message {
	m.target.server.Start(m, arg...)
	return m
}
func (m *Message) Start(key string, arg ...string) *Message {
	m.Travel(func(p *Context, s *Context) {
		if s.Name == key {
			s.Start(m.Spawns(s), arg...)
		}
	})
	return m
}
func (m *Message) Starts(name string, help string, arg ...string) *Message {
	m.wait = make(chan bool)
	m.target.Spawn(m, name, help, arg...).Begin(m, arg...).Start(m, arg...)
	<-m.wait
	return m
}

func (m *Message) Right(arg ...interface{}) bool {
	return m.Option(MSG_USERROLE) == ROLE_ROOT || !m.Warn(m.Cmdx(AAA_ROLE, "right", m.Option(MSG_USERROLE), kit.Keys(arg...)) != "ok", "no right")
}
func (m *Message) Space(arg interface{}) []string {
	if arg == nil || kit.Format(arg) == m.Conf(CLI_RUNTIME, "node.name") {
		return nil
	}
	return []string{WEB_SPACE, kit.Format(arg)}
}
func (m *Message) Event(key string, arg ...string) *Message {
	m.Cmd(GDB_EVENT, "action", key, arg)
	return m
}
func (m *Message) Watch(key string, arg ...string) *Message {
	m.Cmd(GDB_EVENT, "listen", key, arg)
	return m
}

func (m *Message) Travel(cb interface{}) *Message {
	list := []*Context{m.target}
	for i := 0; i < len(list); i++ {
		switch cb := cb.(type) {
		case func(*Context, *Context):
			// 模块回调
			cb(list[i].context, list[i])
		case func(*Context, *Context, string, *Command):
			ls := []string{}
			for k := range list[i].Commands {
				ls = append(ls, k)
			}
			sort.Strings(ls)
			for _, k := range ls {
				// 命令回调
				cb(list[i].context, list[i], k, list[i].Commands[k])
			}
		case func(*Context, *Context, string, *Config):
			ls := []string{}
			for k := range list[i].Configs {
				ls = append(ls, k)
			}
			sort.Strings(ls)
			for _, k := range ls {
				// 配置回调
				cb(list[i].context, list[i], k, list[i].Configs[k])
			}
		}

		// 下级模块
		ls := []string{}
		for k := range list[i].contexts {
			ls = append(ls, k)
		}
		sort.Strings(ls)
		for _, k := range ls {
			list = append(list, list[i].contexts[k])
		}
	}
	return m
}
func (m *Message) Search(key interface{}, cb interface{}) *Message {
	switch key := key.(type) {
	case string:
		if k, ok := Alias[key]; ok {
			key = k
		}

		// 查找模块
		p := m.target.root
		if strings.Contains(key, ":") {

		} else if key == "." {
			if m.target.context != nil {
				p = m.target.context
			}
		} else if strings.Contains(key, ".") {
			list := strings.Split(key, ".")
			for _, p = range []*Context{m.target.root, m.target, m.source} {
				for _, v := range list[:len(list)-1] {
					if s, ok := p.contexts[v]; ok {
						p = s
					} else {
						p = nil
						break
					}
				}
				if p != nil {
					break
				}
			}
			if p == nil {
				m.Log(LOG_WARN, "not found %s", key)
				break
			}
			key = list[len(list)-1]
		} else {
			p = m.target
		}

		// 遍历命令
		switch cb := cb.(type) {
		case func(p *Context, s *Context, key string, cmd *Command):
			if key == "" {
				for k, v := range p.Commands {
					cb(p.context, p, k, v)
				}
				break
			}

			for _, p = range []*Context{p, m.target, m.source} {
				for c := p; c != nil; c = c.context {
					if cmd, ok := c.Commands[key]; ok {
						cb(c, p, key, cmd)
						return m
					}
				}
			}
		case func(p *Context, s *Context, key string, conf *Config):
			for _, p = range []*Context{p, m.target, m.source} {
				for c := p; c != nil; c = c.context {
					if cmd, ok := c.Configs[key]; ok {
						cb(c.context, c, key, cmd)
						return m
					}
				}
			}
		case func(p *Context, s *Context, key string):
			cb(p.context, p, key)
		}
	}
	return m
}
func (m *Message) Preview(arg string) (res string) {
	list := kit.Split(arg)
	m.Search(list[0], func(p *Context, s *Context, key string, cmd *Command) {
		res = kit.Format(kit.Dict("feature", cmd.Meta, "inputs", cmd.List))
	})
	return res
}

func (m *Message) Prefile(favor string, id string) map[string]string {
	res := map[string]string{}
	m.Option("render", "")
	m.Option("_action", "")
	m.Cmd(WEB_FAVOR, kit.Select(m.Option("favor"), favor), id).Table(func(index int, value map[string]string, head []string) {
		res[value["key"]] = value["value"]
	})

	res["content"] = m.Cmdx(CLI_SYSTEM, "sed", "-n", kit.Format("%d,%dp", kit.Int(res["extra.row"]), kit.Int(res["extra.row"])+3), res["extra.buf"])
	return res
}
func (m *Message) Prefix(arg ...string) string {
	return kit.Keys(m.Cap(CTX_FOLLOW), arg)
}
func (m *Message) Save(arg ...string) *Message {
	list := []string{}
	for _, k := range arg {
		list = append(list, kit.Keys(m.Cap(CTX_FOLLOW), k))
	}
	m.Cmd(CTX_CONFIG, "save", kit.Keys(m.Cap(CTX_FOLLOW), "json"), list)
	return m
}
func (m *Message) Load(arg ...string) *Message {
	list := []string{}
	for _, k := range arg {
		list = append(list, kit.Keys(m.Cap(CTX_FOLLOW), k))
	}
	m.Cmd(CTX_CONFIG, "load", kit.Keys(m.Cap(CTX_FOLLOW), "json"), list)
	return m
}

func (m *Message) Richs(key string, chain interface{}, raw interface{}, cb interface{}) (res map[string]interface{}) {
	// 数据结构
	cache := m.Confm(key, chain)
	if cache == nil {
		return nil
	}
	meta, ok := cache[kit.MDB_META].(map[string]interface{})
	hash, ok := cache[kit.MDB_HASH].(map[string]interface{})
	if !ok {
		return nil
	}

	h := kit.Format(raw)
	switch h {
	case "*":
		// 全部遍历
		switch cb := cb.(type) {
		case func(string, string):
			for k, v := range hash {
				cb(k, kit.Format(v))
			}
		case func(string, map[string]interface{}):
			for k, v := range hash {
				res = v.(map[string]interface{})
				cb(k, res)
			}
		}
		return res
	case "%":
		// 随机选取
		if len(hash) > 0 {
			list := []string{}
			for k := range hash {
				list = append(list, k)
			}
			h = list[rand.Intn(len(list))]
			res, _ = hash[h].(map[string]interface{})
		}
	default:
		// 单个查询
		if res, ok = hash[h].(map[string]interface{}); !ok {
			switch kit.Format(kit.Value(meta, kit.MDB_SHORT)) {
			case "", "uniq":
			default:
				hh := kit.Hashs(h)
				if res, ok = hash[hh].(map[string]interface{}); ok {
					h = hh
					break
				}

				prefix := path.Join(kit.Select(m.Conf(WEB_CACHE, "meta.store"), kit.Format(meta["store"])), key)
				for _, k := range []string{h, hh} {
					if f, e := os.Open(path.Join(prefix, kit.Keys(k, "json"))); e == nil {
						defer f.Close()
						if b, e := ioutil.ReadAll(f); e == nil {
							if json.Unmarshal(b, &res) == e {
								h = k
								m.Log(LOG_IMPORT, "%s/%s.json", prefix, k)
								break
							}
						}
					}
				}
			}
		}
	}

	// 返回数据
	if res != nil {
		switch cb := cb.(type) {
		case func(map[string]interface{}):
			cb(res)
		case func(string, map[string]interface{}):
			cb(h, res)
		}
	}
	return res
}
func (m *Message) Rich(key string, chain interface{}, data interface{}) string {
	// 数据结构
	cache := m.Confm(key, chain)
	if cache == nil {
		cache = map[string]interface{}{}
		m.Confv(key, chain, cache)
	}
	meta, ok := cache[kit.MDB_META].(map[string]interface{})
	if !ok {
		meta = map[string]interface{}{}
		cache[kit.MDB_META] = meta
	}
	hash, ok := cache[kit.MDB_HASH].(map[string]interface{})
	if !ok {
		hash = map[string]interface{}{}
		cache[kit.MDB_HASH] = hash
	}

	// 通用数据
	prefix := kit.Select("", "meta.", kit.Value(data, "meta") != nil)
	if kit.Value(data, prefix+kit.MDB_TIME) == nil {
		kit.Value(data, prefix+kit.MDB_TIME, m.Time())
	}

	// 生成键值
	h := ""
	switch short := kit.Format(kit.Value(meta, kit.MDB_SHORT)); short {
	case "":
		h = kit.ShortKey(hash, 6)
	case "uniq":
		h = kit.Hashs("uniq")
	case "data":
		h = kit.Hashs(kit.Format(data))
	default:
		if kit.Value(data, "meta") != nil {
			h = kit.Hashs(kit.Format(kit.Value(data, "meta."+short)))
		} else {
			h = kit.Hashs(kit.Format(kit.Value(data, short)))
		}
	}

	// 添加数据
	if hash[h] = data; len(hash) >= kit.Int(kit.Select(m.Conf(WEB_CACHE, "meta.limit"), kit.Format(meta["limit"]))) {
		least := kit.Int(kit.Select(m.Conf(WEB_CACHE, "meta.least"), kit.Format(meta["least"])))

		// 时间淘汰
		list := []int{}
		for _, v := range hash {
			list = append(list, kit.Time(kit.Format(kit.Value(v, "time"))))
		}
		sort.Ints(list)
		dead := list[len(list)-1-least]

		prefix := path.Join(kit.Select(m.Conf(WEB_CACHE, "meta.store"), kit.Format(meta["store"])), key)
		for k, v := range hash {
			if kit.Time(kit.Format(kit.Value(v, "time"))) > dead {
				break
			}

			name := path.Join(prefix, kit.Keys(k, "json"))
			if f, p, e := kit.Create(name); m.Assert(e) {
				defer f.Close()
				// 保存数据
				if n, e := f.WriteString(kit.Format(v)); m.Assert(e) {
					m.Log(LOG_EXPORT, "%s: %d", p, n)
					delete(hash, k)
				}
			}
		}
	}

	return h
}
func (m *Message) Grow(key string, chain interface{}, data interface{}) int {
	// 数据结构
	cache := m.Confm(key, chain)
	if cache == nil {
		cache = map[string]interface{}{}
		m.Confv(key, chain, cache)
	}
	meta, ok := cache[kit.MDB_META].(map[string]interface{})
	if !ok {
		meta = map[string]interface{}{}
		cache[kit.MDB_META] = meta
	}
	list, _ := cache[kit.MDB_LIST].([]interface{})

	// 通用数据
	id := kit.Int(meta["count"]) + 1
	prefix := kit.Select("", "meta.", kit.Value(data, "meta") != nil)
	if kit.Value(data, prefix+kit.MDB_ID, id); kit.Value(data, prefix+kit.MDB_TIME) == nil {
		kit.Value(data, prefix+kit.MDB_TIME, m.Time())
	}

	// 添加数据
	list = append(list, data)
	cache[kit.MDB_LIST] = list
	meta["count"] = id

	// 保存数据
	if len(list) >= kit.Int(kit.Select(m.Conf(WEB_CACHE, "meta.limit"), kit.Format(meta["limit"]))) {
		least := kit.Int(kit.Select(m.Conf(WEB_CACHE, "meta.least"), kit.Format(meta["least"])))

		record, _ := meta["record"].([]interface{})

		// 文件命名
		prefix := path.Join(kit.Select(m.Conf(WEB_CACHE, "meta.store"), kit.Format(meta["store"])), key)
		name := path.Join(prefix, kit.Keys(kit.Select("list", chain), "csv"))
		if len(record) > 0 {
			name = kit.Format(kit.Value(record, kit.Keys(len(record)-1, "file")))
			if s, e := os.Stat(name); e == nil {
				if s.Size() > kit.Int64(kit.Select(m.Conf(WEB_CACHE, "meta.fsize"), kit.Format(meta["fsize"]))) {
					name = fmt.Sprintf("%s/%s_%d.csv", prefix, kit.Select("list", chain), kit.Int(meta["offset"]))
				}
			}
		}

		// 打开文件
		f, e := os.OpenFile(name, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if e != nil {
			f, _, e = kit.Create(name)
			m.Info("%s.%v create: %s", key, chain, name)
		} else {
			m.Info("%s.%v append: %s", key, chain, name)
		}
		defer f.Close()
		s, e := f.Stat()
		m.Assert(e)

		// 保存表头
		keys := []string{}
		w := csv.NewWriter(f)
		if s.Size() == 0 {
			for k := range list[0].(map[string]interface{}) {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			w.Write(keys)
			m.Info("write head: %v", keys)
			w.Flush()
			s, e = f.Stat()
		} else {
			r := csv.NewReader(f)
			keys, e = r.Read()
			m.Info("read head: %v", keys)
		}

		// 创建索引
		count := len(list) - least
		offset := kit.Int(meta["offset"])
		meta["record"] = append(record, map[string]interface{}{
			"time": m.Time(), "offset": offset, "count": count,
			"file": name, "position": s.Size(),
		})

		// 保存数据
		for i, v := range list {
			if i >= count {
				break
			}

			val := v.(map[string]interface{})

			values := []string{}
			for _, k := range keys {
				values = append(values, kit.Format(val[k]))
			}
			w.Write(values)

			if i < least {
				list[i] = list[count+i]
			}
		}

		m.Log(LOG_INFO, "%s.%v save %s offset %v+%v", key, chain, name, offset, count)
		meta["offset"] = offset + count
		list = list[count:]
		cache[kit.MDB_LIST] = list
		w.Flush()
	}
	return id
}
func (m *Message) Grows(key string, chain interface{}, match string, value string, cb interface{}) map[string]interface{} {
	// 数据结构
	cache := m.Confm(key, chain)
	if cache == nil {
		return nil
	}
	meta, ok := cache[kit.MDB_META].(map[string]interface{})
	list, ok := cache[kit.MDB_LIST].([]interface{})
	if !ok {
		return nil
	}

	// 数据范围
	offend := kit.Int(kit.Select("0", m.Option("cache.offend")))
	limit := kit.Int(kit.Select("10", m.Option("cache.limit")))
	current := kit.Int(meta["offset"])
	end := current + len(list) - offend
	begin := end - limit
	switch limit {
	case -1:
		begin = current
	case -2:
		begin = 0
	}

	if match == kit.MDB_ID {
		begin, end = kit.Int(value)-1, kit.Int(value)
		match, value = "", ""
	}

	order := 0
	if begin < current {
		// 读取磁盘
		m.Log(LOG_INFO, "%s.%v read %v-%v from %v-%v", key, chain, begin, end, current, current+len(list))
		store, _ := meta["record"].([]interface{})
		for s := len(store) - 1; s > -1; s-- {
			// 查找索引
			item, _ := store[s].(map[string]interface{})
			line := kit.Int(item["offset"])
			m.Log(LOG_INFO, "check history %v %v %v", s, line, item)
			if begin < line && s > 0 {
				continue
			}

			for ; s < len(store); s++ {
				if begin >= end {
					break
				}

				// 查找偏移
				item, _ := store[s].(map[string]interface{})
				name := kit.Format(item["file"])
				pos := kit.Int(item["position"])
				offset := kit.Int(item["offset"])
				if offset+kit.Int(item["count"]) <= begin {
					m.Log(LOG_INFO, "skip store %v %d", item, begin)
					continue
				}

				// 打开文件
				m.Log(LOG_IMPORT, "load history %v %v %v", s, offset, item)
				if f, e := os.Open(name); m.Assert(e) {
					defer f.Close()
					r := csv.NewReader(f)
					heads, _ := r.Read()
					m.Log(LOG_IMPORT, "load head %v", heads)

					f.Seek(int64(pos), os.SEEK_SET)
					r = csv.NewReader(f)
					for i := offset; i < end; i++ {
						lines, e := r.Read()
						if e != nil {
							m.Log(LOG_IMPORT, "load head %v", e)
							break
						}

						if i >= begin {
							item := map[string]interface{}{}
							for i := range heads {
								item[heads[i]] = lines[i]
							}
							m.Log(LOG_INFO, "load line %v %v %v", i, order, item)
							if match == "" || strings.Contains(kit.Format(item[match]), value) {
								// 读取文件
								switch cb := cb.(type) {
								case func(int, map[string]interface{}):
									cb(order, item)
								case func(int, map[string]interface{}) bool:
									if cb(order, item) {
										return meta
									}
								}
								order++
							}
							begin = i + 1
						} else {
							m.Log(LOG_INFO, "skip line %v", i)
						}
					}
				}
			}
			break
		}
	}

	if begin < current {
		begin = current
	}
	for i := begin - current; i < end-current; i++ {
		if match == "" || strings.Contains(kit.Format(kit.Value(list[i], match)), value) {
			// 读取缓存
			switch cb := cb.(type) {
			case func(int, map[string]interface{}):
				cb(order, list[i].(map[string]interface{}))
			case func(int, map[string]interface{}) bool:
				if cb(order, list[i].(map[string]interface{})) {
					return meta
				}
			}
			order++
		}
	}
	return meta
}

func (m *Message) Cmdy(arg ...interface{}) *Message {
	msg := m.Cmd(arg...)
	m.Copy(msg)
	return m
}
func (m *Message) Cmdx(arg ...interface{}) string {
	return kit.Select("", m.Cmd(arg...).meta[MSG_RESULT], 0)
}
func (m *Message) Cmds(arg ...interface{}) bool {
	return kit.Select("", m.Cmd(arg...).meta[MSG_RESULT], 0) != ""
}
func (m *Message) Cmd(arg ...interface{}) *Message {
	list := kit.Simple(arg...)
	if len(list) == 0 {
		list = m.meta[MSG_DETAIL]
	}
	if len(list) == 0 {
		return m
	}

	m.Search(list[0], func(p *Context, c *Context, key string, cmd *Command) {
		m.TryCatch(m.Spawns(c), true, func(msg *Message) {
			m.Hand, msg.Hand = true, true
			msg.meta[MSG_DETAIL] = list

			// _key := kit.Format(kit.Value(cmd.Meta, "remote"))
			// if you := m.Option(_key); you != "" {
			// 	// 远程命令
			// 	msg.Option(_key, "")
			// 	msg.Option("_option", m.Optionv("option"))
			// 	msg.Copy(msg.Spawns(c).Cmd(WEB_LABEL, you, list[0], list[1:]))
			// } else {
			// 	// 本地命令
			// 	p.Run(msg, cmd, key, list[1:]...)
			// }

			p.Run(msg, cmd, key, list[1:]...)
			m.Hand, msg.Hand, m = true, true, msg
		})
	})

	if m.Warn(m.Hand == false, "not found %v", list) {
		return m.Set(MSG_RESULT).Cmd(CLI_SYSTEM, list)
	}
	return m
}
func (m *Message) Confv(arg ...interface{}) (val interface{}) {
	m.Search(arg[0], func(p *Context, s *Context, key string, conf *Config) {
		if len(arg) > 1 {
			if len(arg) > 2 {
				if arg[1] == nil {
					// 写配置
					conf.Value = arg[2]
				} else {
					// 写修改项
					kit.Value(conf.Value, arg[1:]...)
				}
			}
			// 读配置项
			val = kit.Value(conf.Value, arg[1])
		} else {
			// 读配置
			val = conf.Value
		}
	})
	return
}
func (m *Message) Confm(key string, chain interface{}, cbs ...interface{}) map[string]interface{} {
	val := m.Confv(key, chain)
	if len(cbs) > 0 {
		kit.Fetch(val, cbs[0])
	}
	value, _ := val.(map[string]interface{})
	return value
}
func (m *Message) Confs(arg ...interface{}) bool {
	return kit.Format(m.Confv(arg...)) != ""
}
func (m *Message) Confi(arg ...interface{}) int {
	return kit.Int(m.Confv(arg...))
}
func (m *Message) Conf(arg ...interface{}) string {
	return kit.Format(m.Confv(arg...))
}
func (m *Message) Capv(arg ...interface{}) interface{} {
	key := ""
	switch val := arg[0].(type) {
	case string:
		key, arg = val, arg[1:]
	}

	for _, s := range []*Context{m.target} {
		for c := s; c != nil; c = c.context {
			if caps, ok := c.Caches[key]; ok {
				if len(arg) > 0 {
					// 写数据
					caps.Value = kit.Format(arg[0])
				}
				// 读数据
				return caps.Value
			}
		}
	}
	return nil
}
func (m *Message) Cap(arg ...interface{}) string {
	return kit.Format(m.Capv(arg...))
}
