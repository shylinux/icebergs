package ice

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"reflect"
	"runtime"
	"strings"
	"time"

	kit "shylinux.com/x/toolkits"
)

func (m *Message) Cut(fields ...string) *Message {
	m.meta[MSG_APPEND] = strings.Split(strings.Join(fields, ","), ",")
	return m
}
func (m *Message) CSV(text string, head ...string) *Message {
	bio := bytes.NewBufferString(text)
	r := csv.NewReader(bio)

	if len(head) == 0 {
		head, _ = r.Read()
	}
	for {
		line, e := r.Read()
		if e != nil {
			break
		}
		for i, k := range head {
			m.Push(k, kit.Select("", line, i))
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
	fields := kit.Split(field, space, space, space)
	for i, l := range kit.Split(str, enter, enter, enter) {
		if strings.HasPrefix(l, "Binary") {
			continue
		}
		if strings.TrimSpace(l) == "" {
			continue
		}
		if i == 0 && (field == "" || field == "index") {
			// 表头行
			fields = kit.Split(l, space, space)
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

		ls := kit.Split(l, space, space)
		for i, v := range ls {
			if i == len(fields)-1 {
				m.Push(kit.Select("some", fields, i), strings.Join(ls[i:], space))
				break
			}
			m.Push(kit.Select("some", fields, i), v)
		}
	}
	return m
}
func (m *Message) FormatStack() string {
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
func (m *Message) FormatChain() string {
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
}
func (m *Message) FormatTime() string {
	return m.Format("time")
}
func (m *Message) Format(key interface{}) string {
	switch key := key.(type) {
	case []byte:
		json.Unmarshal(key, &m.meta)
	case string:
		switch key {
		case "cost":
			return kit.FmtTime(kit.Int64(time.Since(m.time)))
		case "meta":
			return kit.Format(m.meta)
		case "size":
			if len(m.meta["append"]) == 0 {
				return fmt.Sprintf("%dx%d", 0, 0)
			} else {
				return fmt.Sprintf("%dx%d", len(m.meta[m.meta["append"][0]]), len(m.meta["append"]))
			}
		case "append":
			if len(m.meta["append"]) == 0 {
				return fmt.Sprintf("%dx%d %s", 0, 0, "[]")
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
			return m.FormatChain()
		case "stack":
			return m.FormatStack()
		}
	}
	return m.time.Format(MOD_TIME)
}
func (m *Message) Formats(key string) string {
	switch key {
	case "meta":
		return kit.Formats(m.meta)
	}
	return m.Format(key)
}
func (m *Message) RenameAppend(from, to string) {
	for i, v := range m.meta[MSG_APPEND] {
		if v == from {
			m.meta[MSG_APPEND][i] = to
			m.meta[to] = m.meta[from]
			delete(m.meta, from)
		}
	}
}

func (m *Message) cmd(arg ...interface{}) *Message {
	opts := map[string]interface{}{}
	args := []interface{}{}
	var cbs interface{}

	// 解析参数
	for _, v := range arg {
		switch val := v.(type) {
		case func(int, map[string]string, []string):
			defer func() { m.Table(val) }()

		case map[string]string:
			for k, v := range val {
				opts[k] = v
			}

		case *Option:
			opts[val.Name] = val.Value
		case Option:
			opts[val.Name] = val.Value

		case *Sort:
			defer func() { m.Sort(val.Fields, val.Method) }()
		case Sort:
			defer func() { m.Sort(val.Fields, val.Method) }()

		default:
			if reflect.Func == reflect.TypeOf(val).Kind() {
				cbs = val
			} else {
				args = append(args, v)
			}
		}
	}

	// 解析命令
	list := kit.Simple(args...)
	if len(list) == 0 && !m.Hand {
		list = m.meta[MSG_DETAIL]
	}
	if len(list) == 0 {
		return m
	}

	ok := false
	run := func(msg *Message, ctx *Context, cmd *Command, key string, arg ...string) {
		if ok = true; cbs != nil {
			msg.Option(list[0]+".cb", cbs)
		}
		for k, v := range opts {
			msg.Option(k, v)
		}

		// 执行命令
		m.TryCatch(msg, true, func(msg *Message) {
			m = ctx.cmd(msg, cmd, key, arg...)
		})
	}

	// 查找命令
	if cmd, ok := m.target.Commands[list[0]]; ok {
		run(m.Spawn(), m.target, cmd, list[0], list[1:]...)
	} else if cmd, ok := m.source.Commands[list[0]]; ok {
		run(m.Spawn(m.source), m.source, cmd, list[0], list[1:]...)
	} else {
		m.Search(list[0], func(p *Context, s *Context, key string, cmd *Command) {
			run(m.Spawn(s), s, cmd, key, list[1:]...)
		})
	}

	// 系统命令
	if m.Warn(!ok, ErrNotFound, list) {
		return m.Set(MSG_RESULT).Cmdy("cli.system", list)
	}
	return m
}
func (c *Context) cmd(m *Message, cmd *Command, key string, arg ...string) *Message {
	if m._key, m._cmd = key, cmd; cmd == nil {
		return m
	}

	m.meta[MSG_DETAIL] = kit.Simple(key, arg)
	if m.Hand = true; len(arg) > 1 && arg[0] == "action" && cmd.Action != nil {
		if h, ok := cmd.Action[arg[1]]; ok {
			return c._cmd(m, cmd, key, arg[1], h, arg[2:]...)
		}
	}
	if len(arg) > 0 && arg[0] != "command" && cmd.Action != nil {
		if h, ok := cmd.Action[arg[0]]; ok {
			return c._cmd(m, cmd, key, arg[0], h, arg[1:]...)
		}
	}

	m.Log(LOG_CMDS, "%s.%s %d %v %s", c.Name, key, len(arg), arg,
		kit.Select(kit.FileLine(cmd.Hand, 3), kit.FileLine(9, 3), m.target.Name == "mdb"))
	cmd.Hand(m, c, key, arg...)
	return m
}
func (c *Context) _cmd(m *Message, cmd *Command, key string, k string, h *Action, arg ...string) *Message {
	if k == "command" && h.Hand == nil {
		for _, cmd := range arg {
			m.Cmdy("command", cmd)
		}
		return m
	}

	if k == "run" && m.Warn(!m.Right(arg), ErrNotRight, arg) {
		return m
	}

	m.Log(LOG_CMDS, "%s.%s %s %d %v %s", c.Name, key, k, len(arg), arg, kit.FileLine(h.Hand, 3))
	if len(h.List) > 0 && k != "search" {
		order := false
		for i, v := range h.List {
			name := kit.Format(kit.Value(v, "name"))
			value := kit.Format(kit.Value(v, "value"))

			if i == 0 && len(arg) > 0 && arg[0] != name {
				order = true
			}
			if order {
				value = kit.Select(value, arg, i)
			}

			m.Option(name, kit.Select(m.Option(name), value, !strings.HasPrefix(value, "@")))
		}
		if !order {
			for i := 0; i < len(arg)-1; i += 2 {
				m.Option(arg[i], arg[i+1])
			}
		}
	}

	if h.Hand == nil {
		m.Cmdy(kit.Split(h.Name), arg)
	} else {
		h.Hand(m, arg...)
	}
	return m
}
func (c *Context) split(key string, cmd *Command, name string) []interface{} {
	const (
		BUTTON   = "button"
		SELECT   = "select"
		TEXT     = "text"
		TEXTAREA = "textarea"
	)

	button, list := false, []interface{}{}
	for _, v := range kit.Split(kit.Select("key", name), " ", " ")[1:] {
		switch v {
		case "text":
			list = append(list, kit.List(kit.MDB_INPUT, TEXTAREA, kit.MDB_NAME, "text", kit.MDB_VALUE, "@key")...)
			continue
		case "page":
			list = append(list, kit.List(kit.MDB_INPUT, TEXT, kit.MDB_NAME, "limit")...)
			list = append(list, kit.List(kit.MDB_INPUT, TEXT, kit.MDB_NAME, "offend")...)
			list = append(list, kit.List(kit.MDB_INPUT, BUTTON, kit.MDB_NAME, "prev")...)
			list = append(list, kit.List(kit.MDB_INPUT, BUTTON, kit.MDB_NAME, "next")...)
			continue
		case "auto":
			list = append(list, kit.List(kit.MDB_INPUT, BUTTON, kit.MDB_NAME, "list", kit.MDB_VALUE, "auto")...)
			list = append(list, kit.List(kit.MDB_INPUT, BUTTON, kit.MDB_NAME, "back")...)
			button = true
			continue
		}

		ls, value := kit.Split(v, " ", ":=@"), ""
		item := kit.Dict(kit.MDB_INPUT, kit.Select(TEXT, BUTTON, button))
		if kit.Value(item, kit.MDB_NAME, ls[0]); item[kit.MDB_INPUT] == TEXT {
			kit.Value(item, kit.MDB_VALUE, kit.Select("@key", "auto", strings.Contains(name, "auto")))
		}

		for i := 1; i < len(ls); i += 2 {
			switch ls[i] {
			case ":":
				switch kit.Value(item, kit.MDB_INPUT, ls[i+1]); ls[i+1] {
				case TEXTAREA:
					kit.Value(item, "style.width", "360")
					kit.Value(item, "style.height", "60")
				case BUTTON:
					kit.Value(item, kit.MDB_VALUE, "")
					button = true
				}
			case "=":
				if value = kit.Select("", ls, i+1); len(ls) > i+1 && strings.Contains(ls[i+1], ",") {
					vs := strings.Split(ls[i+1], ",")
					kit.Value(item, "values", vs)
					kit.Value(item, kit.MDB_VALUE, vs[0])
					kit.Value(item, kit.MDB_INPUT, SELECT)
					if kit.Value(item, kit.MDB_NAME) == "scale" {
						kit.Value(item, kit.MDB_VALUE, "week")
					}
				} else {
					kit.Value(item, kit.MDB_VALUE, value)
				}
			case "@":
				if len(ls) > i+1 {
					if kit.Value(item, kit.MDB_INPUT) == BUTTON {
						kit.Value(item, "action", ls[i+1])
					} else {
						kit.Value(item, kit.MDB_VALUE, "@"+ls[i+1]+"="+value)
					}
				}
			}
		}
		list = append(list, item)
	}
	return list
}

func Display(file string, arg ...string) map[string]string {
	if file != "" && !strings.HasPrefix(file, "/") {
		ls := strings.Split(kit.FileLine(2, 100), "usr")
		file = kit.Select(file+".js", file, strings.HasSuffix(file, ".js"))
		file = path.Join("/require/shylinux.com/x", path.Dir(ls[len(ls)-1]), file)
	}
	return map[string]string{"display": file, kit.MDB_STYLE: kit.Select("", arg, 0)}
}
func MergeAction(list ...map[string]*Action) map[string]*Action {
	if len(list) == 0 {
		return nil
	}
	for _, item := range list[1:] {
		for k, v := range item {
			if h, ok := list[0][k]; !ok {
				list[0][k] = v
			} else if h.Hand == nil {
				h.Hand = v.Hand
			}
		}
	}
	return list[0]
}

func (m *Message) AppendSimple(key ...string) (res []string) {
	if len(key) == 0 {
		key = append(key, m.Appendv(MSG_APPEND)...)
	}
	for _, k := range key {
		res = append(res, k, m.Append(k))
	}
	return
}

func (m *Message) SetResult() {
	m.Set(MSG_RESULT)
	return
}

func (m *Message) AppendTrans(cb func(value string, key string, index int) string) {
	for _, k := range m.meta[MSG_APPEND] {
		for i, v := range m.meta[k] {
			m.meta[k][i] = cb(v, k, i)
		}
	}
}

func (m *Message) OptionUserWeb() *url.URL {
	return kit.ParseURL(m.Option(MSG_USERWEB))
}
