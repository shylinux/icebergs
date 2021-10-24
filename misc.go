package ice

import (
	"bytes"
	"encoding/csv"
	"net/url"
	"path"
	"reflect"
	"strings"

	kit "shylinux.com/x/toolkits"
)

func (m *Message) Length() (max int) {
	for _, k := range m.meta[MSG_APPEND] {
		if l := len(m.meta[k]); l > max {
			max = l
		}
	}
	return max
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
func (m *Message) Split(str string, field string, sp string, nl string) *Message {
	fields, indexs := kit.Split(field, sp, sp, sp), []int{}
	for i, l := range kit.Split(str, nl, nl, nl) {
		if strings.HasPrefix(l, "Binary") {
			continue
		}
		if strings.TrimSpace(l) == "" {
			continue
		}
		if i == 0 && (field == "" || field == "index") { // 表头行
			if fields = kit.Split(l, sp, sp); field == "index" {
				for _, v := range fields {
					indexs = append(indexs, strings.Index(l, v))
				}
			}
			continue
		}

		if len(indexs) > 0 { // 按位切分
			for i, v := range indexs {
				if i == len(indexs)-1 {
					m.Push(kit.Select(SP, fields, i), l[v:])
				} else {
					m.Push(kit.Select(SP, fields, i), l[v:indexs[i+1]])
				}
			}
			continue
		}

		ls := kit.Split(l, sp, sp)
		for i, v := range ls {
			if i == len(fields)-1 {
				m.Push(kit.Select(SP, fields, i), strings.Join(ls[i:], sp))
				break
			}
			m.Push(kit.Select(SP, fields, i), v)
		}
	}
	return m
}

func (m *Message) FieldsIsDetail() bool {
	if m.OptionFields() == "detail" {
		return true
	}
	if len(m.meta[MSG_APPEND]) == 2 && m.meta[MSG_APPEND][0] == kit.MDB_KEY && m.meta[MSG_APPEND][1] == kit.MDB_VALUE {
		return true
	}
	return false
}
func (m *Message) OptionUserWeb() *url.URL {
	return kit.ParseURL(m.Option(MSG_USERWEB))
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
func (m *Message) AppendSimple(key ...string) (res []string) {
	if len(key) == 0 {
		key = append(key, m.Appendv(MSG_APPEND)...)
	}
	for _, k := range key {
		res = append(res, k, m.Append(k))
	}
	return
}
func (m *Message) AppendTrans(cb func(value string, key string, index int) string) {
	for _, k := range m.meta[MSG_APPEND] {
		for i, v := range m.meta[k] {
			m.meta[k][i] = cb(v, k, i)
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

		case map[string]interface{}:
			for k, v := range val {
				opts[k] = v
			}
		case map[string]string:
			for k, v := range val {
				opts[k] = v
			}

		case *Option:
			opts[val.Name] = val.Value
		case Option:
			opts[val.Name] = val.Value

		case string:
			args = append(args, v)
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
			msg.Option(kit.Keycb(list[0]), cbs)
		}
		for k, v := range opts {
			msg.Option(k, v)
		}

		// 执行命令
		m.TryCatch(msg, true, func(msg *Message) { m = ctx.cmd(msg, cmd, key, arg...) })
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
		return m.Set(MSG_RESULT).Cmdy("system", list)
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

	if k == RUN && m.Warn(!m.Right(arg), ErrNotRight, arg) {
		return m
	}

	m.Log(LOG_CMDS, "%s.%s %s %d %v %s", c.Name, key, k, len(arg), arg, kit.FileLine(h.Hand, 3))
	if len(h.List) > 0 && k != "search" {
		order := false
		for i, v := range h.List {
			name := kit.Format(kit.Value(v, kit.MDB_NAME))
			value := kit.Format(kit.Value(v, kit.MDB_VALUE))

			if i == 0 && len(arg) > 0 && arg[0] != name {
				order = true
			}
			if order {
				m.Option(name, kit.Select(value, arg, i))
			} else {
				if m.Option(name) == "" {
					m.Option(name, value)
				}
			}
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
func (c *Context) split(name string) (list []interface{}) {
	const (
		TEXT     = "text"
		TEXTAREA = "textarea"
		SELECT   = "select"
		BUTTON   = "button"
	)

	item, button := kit.Dict(), false
	ls := kit.Split(name, SP, ":=@")
	for i := 1; i < len(ls); i++ {
		switch ls[i] {
		case "text":
			list = append(list, kit.List(kit.MDB_TYPE, TEXTAREA, kit.MDB_NAME, "text")...)
		case "auto":
			list = append(list, kit.List(kit.MDB_TYPE, BUTTON, kit.MDB_NAME, "list", kit.MDB_VALUE, AUTO)...)
			list = append(list, kit.List(kit.MDB_TYPE, BUTTON, kit.MDB_NAME, "back")...)
			button = true
		case "page":
			list = append(list, kit.List(kit.MDB_TYPE, TEXT, kit.MDB_NAME, "limit")...)
			list = append(list, kit.List(kit.MDB_TYPE, TEXT, kit.MDB_NAME, "offend")...)
			list = append(list, kit.List(kit.MDB_TYPE, BUTTON, kit.MDB_NAME, "prev")...)
			list = append(list, kit.List(kit.MDB_TYPE, BUTTON, kit.MDB_NAME, "next")...)

		case ":":
			if item[kit.MDB_TYPE] = kit.Select("", ls, i+1); item[kit.MDB_TYPE] == BUTTON {
				button = true
			}
			i++
		case "=":
			if value := kit.Select("", ls, i+1); strings.Contains(value, ",") {
				vs := kit.Split(value)
				if strings.Count(value, vs[0]) > 1 {
					item["values"] = vs[1:]
				} else {
					item["values"] = vs
				}
				item[kit.MDB_VALUE] = vs[0]
				item[kit.MDB_TYPE] = SELECT
			} else {
				item[kit.MDB_VALUE] = value
			}
			i++
		case "@":
			item[kit.MDB_ACTION] = kit.Select("", ls, i+1)
			i++

		default:
			item = kit.Dict(kit.MDB_TYPE, kit.Select(TEXT, BUTTON, button), kit.MDB_NAME, ls[i])
			list = append(list, item)
		}
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
func SelectAction(list map[string]*Action, fields ...string) map[string]*Action {
	if len(fields) == 0 {
		return list
	}

	res := map[string]*Action{}
	for _, field := range fields {
		res[field] = list[field]
	}
	return res
}
