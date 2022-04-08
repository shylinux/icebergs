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
func (m *Message) SplitIndex(str string, arg ...string) *Message {
	return m.Split(str, kit.Simple("index", arg)...)
}
func (m *Message) Split(str string, arg ...string) *Message { // field sp nl
	m.Set(MSG_APPEND).Set(MSG_RESULT)
	field := kit.Select("", arg, 0)
	sp := kit.Select(SP, arg, 1)
	nl := kit.Select(NL, arg, 2)
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
					m.Push(strings.TrimSpace(kit.Select(SP, fields, i)), strings.TrimSpace(l[v:]))
				} else {
					m.Push(strings.TrimSpace(kit.Select(SP, fields, i)), strings.TrimSpace(l[v:indexs[i+1]]))
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
	if m.OptionFields() == CACHE_DETAIL {
		return true
	}
	if len(m.meta[MSG_APPEND]) == 2 && m.meta[MSG_APPEND][0] == KEY && m.meta[MSG_APPEND][1] == VALUE {
		return true
	}
	return false
}

func (m *Message) PushDetail(value interface{}, arg ...interface{}) *Message {
	return m.Push(CACHE_DETAIL, value, arg...)
}
func (m *Message) IsErr(arg ...string) bool {
	return len(arg) > 0 && m.Result(1) == arg[0] || m.Result(0) == ErrWarn
}
func (m *Message) IsErrNotFound() bool { return m.Result(1) == ErrNotFound }
func (m *Message) OptionCB(key string, cb ...interface{}) interface{} {
	if len(cb) > 0 {
		return m.Optionv(kit.Keycb(key), cb...)
	}
	return m.Optionv(kit.Keycb(key))
}
func (m *Message) OptionUserWeb() *url.URL {
	return kit.ParseURL(m.Option(MSG_USERWEB))
}
func (m *Message) SetResult(arg ...string) *Message {
	return m.Set(MSG_RESULT, arg...)
}
func (m *Message) SetAppend(arg ...string) *Message {
	return m.Set(MSG_APPEND, arg...)
}
func (m *Message) ToLowerAppend(arg ...string) *Message {
	for _, k := range m.meta[MSG_APPEND] {
		m.RenameAppend(k, strings.ToLower(k))
	}
	return m
}
func (m *Message) RenameAppend(arg ...string) *Message { // [from to]...
	for i := 0; i < len(arg)-1; i += 2 {
		if arg[i] == arg[i+1] {
			continue
		}
		for j, v := range m.meta[MSG_APPEND] {
			if v == arg[i] {
				m.meta[MSG_APPEND][j] = arg[i+1]
				m.meta[arg[i+1]] = m.meta[arg[i]]
				delete(m.meta, arg[i])
			}
		}
	}
	return m
}
func (m *Message) AppendSimple(key ...string) (res []string) {
	if len(key) == 0 {
		if m.FieldsIsDetail() {
			key = append(key, m.Appendv(KEY)...)
		} else {
			key = append(key, m.Appendv(MSG_APPEND)...)
		}
	}
	for _, k := range key {
		res = append(res, k, m.Append(k))
	}
	return
}
func (m *Message) AppendTrans(cb func(value string, key string, index int) string) *Message {
	if m.FieldsIsDetail() {
		for i, v := range m.meta[VALUE] {
			k := m.meta[KEY][i]
			m.meta[VALUE][i] = cb(v, k, 0)
		}
		return m
	}
	for _, k := range m.meta[MSG_APPEND] {
		for i, v := range m.meta[k] {
			m.meta[k][i] = cb(v, k, i)
		}
	}
	return m
}
func (m *Message) MergeLink(url string, arg ...interface{}) string {
	return strings.Split(kit.MergeURL2(m.Option(MSG_USERWEB), url, arg...), "?")[0]
}
func (m *Message) MergeURL2(url string, arg ...interface{}) string {
	return kit.MergeURL2(m.Option(MSG_USERWEB), url, arg...)
}
func (m *Message) MergePod(name string, arg ...interface{}) string {
	return kit.MergePOD(kit.Select(Info.Domain, m.Option(MSG_USERWEB)), name, arg...)
}
func (m *Message) MergeCmd(name string, arg ...interface{}) string {
	if name == "" {
		name = m.PrefixKey()
	}
	if m.Option(MSG_USERPOD) == "" {
		return kit.MergeURL2(kit.Select(Info.Domain, m.Option(MSG_USERWEB)), path.Join("/chat/cmd", name))
	}
	return kit.MergeURL2(kit.Select(Info.Domain, m.Option(MSG_USERWEB)), path.Join("cmd", name), arg...)
}
func (m *Message) MergeWebsite(name string, arg ...interface{}) string {
	if m.Option(MSG_USERPOD) == "" {
		return kit.MergeURL2(kit.Select(Info.Domain, m.Option(MSG_USERWEB)), path.Join("/chat/website", name))
	}
	return kit.MergeURL2(kit.Select(Info.Domain, m.Option(MSG_USERWEB)), path.Join("website", name), arg...)
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
			msg.OptionCB(kit.Slice(kit.Split(list[0], PT), -1)[0], cbs)
		}
		for k, v := range opts {
			msg.Option(k, v)
		}

		// 执行命令
		key = kit.Slice(strings.Split(key, PT), -1)[0]
		m.TryCatch(msg, true, func(msg *Message) { m = ctx.cmd(msg, cmd, key, arg...) })
	}

	// 查找命令
	if list[0] == "" {
		run(m.Spawn(), m.target, m._cmd, list[0], list[1:]...)
	} else if cmd, ok := m.target.Commands[strings.TrimPrefix(list[0], m.target.Cap(CTX_FOLLOW)+PT)]; ok {
		run(m.Spawn(), m.target, cmd, list[0], list[1:]...)
	} else if cmd, ok := m.source.Commands[strings.TrimPrefix(list[0], m.source.Cap(CTX_FOLLOW)+PT)]; ok {
		run(m.Spawn(m.source), m.source, cmd, list[0], list[1:]...)
	} else {
		m.Search(list[0], func(p *Context, s *Context, key string, cmd *Command) {
			run(m.Spawn(s), s, cmd, key, list[1:]...)
		})
	}

	m.Warn(!ok, ErrNotFound, kit.Format(list))
	return m
}
func (c *Context) cmd(m *Message, cmd *Command, key string, arg ...string) *Message {
	if m._key, m._cmd = key, cmd; cmd == nil {
		return m
	}

	m.meta[MSG_DETAIL] = kit.Simple(key, arg)
	if m.Hand = true; len(arg) > 1 && arg[0] == ACTION && cmd.Action != nil {
		if h, ok := cmd.Action[arg[1]]; ok {
			return c._cmd(m, cmd, key, arg[1], h, arg[2:]...)
		}
	}
	if len(arg) > 0 && arg[0] != COMMAND && cmd.Action != nil {
		if h, ok := cmd.Action[arg[0]]; ok {
			return c._cmd(m, cmd, key, arg[0], h, arg[1:]...)
		}
	}

	m.Log(LOG_CMDS, "%s.%s %d %v %s", c.Name, key, len(arg), arg,
		kit.Select(kit.FileLine(cmd.Hand, 3), kit.FileLine(9, 3), m.target.Name == MDB))

	if cmd.Hand != nil {
		cmd.Hand(m, c, key, arg...)
	} else if cmd.Action != nil && cmd.Action["select"] != nil {
		cmd.Action["select"].Hand(m, arg...)
	}
	return m
}
func (c *Context) _cmd(m *Message, cmd *Command, key string, sub string, h *Action, arg ...string) *Message {
	if h.Hand == nil {
		m.Cmdy(kit.Split(h.Name), arg)
		return m
	}

	m.Log(LOG_CMDS, "%s.%s %s %d %v %s", c.Name, key, sub, len(arg), arg, kit.FileLine(h.Hand, 3))
	if m._sub = sub; len(h.List) > 0 && sub != "search" {
		order := false
		for i, v := range h.List {
			name := kit.Format(kit.Value(v, NAME))
			value := kit.Format(kit.Value(v, VALUE))

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

	h.Hand(m, arg...)
	return m
}
func (c *Context) split(name string) (list []interface{}) {
	const (
		TEXT     = "text"
		ARGS     = "args"
		TEXTAREA = "textarea"
		SELECT   = "select"
		BUTTON   = "button"
	)

	item, button := kit.Dict(), false
	ls := kit.Split(name, SP, ":=@")
	for i := 1; i < len(ls); i++ {
		switch ls[i] {
		case "run":
			item = kit.Dict(TYPE, BUTTON, NAME, "run")
			list = append(list, item)
			button = true
		case "text", "args":
			item = kit.Dict(TYPE, TEXTAREA, NAME, ls[i])
			list = append(list, item)
		case "auto":
			list = append(list, kit.List(TYPE, BUTTON, NAME, "list", ACTION, AUTO)...)
			list = append(list, kit.List(TYPE, BUTTON, NAME, "back")...)
			button = true
		case "page":
			list = append(list, kit.List(TYPE, TEXT, NAME, "limit")...)
			list = append(list, kit.List(TYPE, TEXT, NAME, "offend")...)
			list = append(list, kit.List(TYPE, BUTTON, NAME, "prev")...)
			list = append(list, kit.List(TYPE, BUTTON, NAME, "next")...)

		case ":":
			if item[TYPE] = kit.Select("", ls, i+1); item[TYPE] == BUTTON {
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
				item[VALUE] = vs[0]
				item[TYPE] = SELECT
			} else {
				item[VALUE] = value
			}
			i++
		case "@":
			item[ACTION] = kit.Select("", ls, i+1)
			i++

		default:
			item = kit.Dict(TYPE, kit.Select(TEXT, BUTTON, button), NAME, ls[i])
			list = append(list, item)
		}
	}
	return list
}

func MergeAction(list ...interface{}) map[string]*Action {
	if len(list) == 0 {
		return nil
	}
	base := list[0].(map[string]*Action)
	for _, item := range list[1:] {
		switch item := item.(type) {
		case map[string]*Action:
			for k, v := range item {
				if h, ok := base[k]; !ok {
					base[k] = v
				} else if h.Hand == nil {
					h.Hand = v.Hand
				} else if k == CTX_INIT {
					last := base[k].Hand
					prev := v.Hand
					base[k].Hand = func(m *Message, arg ...string) {
						prev(m, arg...)
						last(m, arg...)
					}
				}
			}
		case string:
			base[CTX_INIT] = &Action{Hand: func(m *Message, arg ...string) {
				m.Search(item, func(p *Context, s *Context, key string, cmd *Command) {
					MergeAction(base, cmd.Action)
					m.target.Merge(m.target)
				})
			}}
		}
	}
	return base
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
