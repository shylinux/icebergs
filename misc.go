package ice

import (
	"reflect"
	"strings"

	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func (m *Message) Split(str string, arg ...string) *Message {
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
		if i == 0 && (field == "" || field == INDEX) {
			if fields = kit.Split(l, sp, sp); field == INDEX {
				if strings.HasPrefix(l, SP) || strings.HasPrefix(l, TB) {
					indexs = append(indexs, 0)
					for _, v := range fields {
						indexs = append(indexs, strings.Index(l, v)+len(v))
					}
					indexs = indexs[0 : len(indexs)-1]
				} else {
					for _, v := range fields {
						indexs = append(indexs, strings.Index(l, v))
					}
				}
			}
			continue
		}
		if len(indexs) > 0 {
			for i, v := range indexs {
				if v >= len(l) {
					m.Push(strings.TrimSpace(kit.Select(SP, fields, i)), "")
					continue
				}
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
func (m *Message) SplitIndex(str string, arg ...string) *Message {
	return m.Split(str, kit.Simple(INDEX, arg)...)
}
func (m *Message) PushRecord(value Any, arg ...string) *Message {
	return m.Push("", value, kit.Split(kit.Join(arg)))
}
func (m *Message) PushDetail(value Any, arg ...string) *Message {
	switch v := value.(type) {
	case string:
		value = kit.UnMarshal(v)
	}
	return m.Push(FIELDS_DETAIL, value, kit.Split(kit.Join(arg)))
}

func (m *Message) ToLowerAppend(arg ...string) *Message {
	for _, k := range m.meta[MSG_APPEND] {
		m.RenameAppend(k, strings.ToLower(k))
	}
	return m
}
func (m *Message) RenameOption(from, to string) *Message {
	m.Option(to, m.Option(from))
	m.Option(from, "")
	return m
}
func (m *Message) RenameAppend(arg ...string) *Message {
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
			key = append(key, m.meta[KEY]...)
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
func (m *Message) SetAppend(arg ...string) *Message {
	if len(arg) == 0 {
		m.OptionFields("")
	}
	return m.Set(MSG_APPEND, arg...)
}
func (m *Message) SetResult(arg ...string) *Message {
	return m.Set(MSG_RESULT, arg...)
}

func (m *Message) Design(action Any, help string, input ...Any) {
	list := kit.List()
	for _, input := range input {
		switch input := input.(type) {
		case string:
			list = append(list, SplitCmd("action "+input, nil)...)
		case Map:
			if kit.Format(input[TYPE]) != "" && kit.Format(input[NAME]) != "" {
				list = append(list, input)
				continue
			}
			kit.For(kit.KeyValue(nil, "", input), func(k string, v Any) {
				list = append(list, kit.Dict(NAME, k, TYPE, TEXT, VALUE, v))
			})
		default:
			m.ErrorNotImplement(input)
		}
	}
	k := kit.Format(action)
	if a, ok := m._cmd.Actions[k]; ok {
		m._cmd.Meta[k], a.List = list, list
		kit.Value(m._cmd.Meta, kit.Keys("_trans", k), help)
	}
}
func (m *Message) _fileline() string {
	switch m.target.Name {
	case MDB, GDB, AAA:
		return m._source
	default:
		return m._target
	}
}
func (m *Message) ActionHand(cmd *Command, key, sub string, arg ...string) *Message {
	if action, ok := cmd.Actions[sub]; !m.Warn(!ok, ErrNotFound, sub, cmd.GetFileLine()) {
		return m.Target()._action(m, cmd, key, sub, action, arg...)
	}
	return m
}
func (m *Message) CmdHand(cmd *Command, key string, arg ...string) *Message {
	if m._key, m._cmd = key, cmd; cmd == nil {
		return m
	}
	m._target = kit.Join(kit.Slice(kit.Split(cmd.GetFileLines(), PS), -3), PS)
	if fileline := m._fileline(); key == SELECT {
		m.Log(LOG_CMDS, "%s.%s %d %v %v", m.Target().Name, key, len(arg), arg, m.Optionv(MSG_FIELDS), logs.FileLineMeta(fileline))
	} else {
		m.Log(LOG_CMDS, "%s.%s %d %v", m.Target().Name, key, len(arg), arg, logs.FileLineMeta(fileline))
	}
	if cmd.Hand != nil {
		cmd.Hand(m, arg...)
	} else if cmd.Actions != nil && cmd.Actions[SELECT] != nil {
		cmd.Actions[SELECT].Hand(m, arg...)
	}
	return m
}
func (m *Message) _command(arg ...Any) *Message {
	args, opts, cbs, _source := []Any{}, Map{}, kit.Value(nil), logs.FileLine(3)
	for _, v := range arg {
		switch val := v.(type) {
		case string:
			args = append(args, v)
		case Maps:
			for k, v := range val {
				opts[k] = v
			}
		case Map:
			for k, v := range kit.KeyValue(nil, "", val) {
				opts[k] = v
			}
		case Option:
			opts[val.Name] = val.Value
		case *Option:
			opts[val.Name] = val.Value
		case logs.Meta:
			if val.Key == "fileline" {
				_source = val.Value
			}
		case func(int, Maps, []string):
			defer func() { m.Table(val) }()
		case func(Maps):
			defer func() { m.Tables(val) }()
		case nil:
		default:
			if reflect.TypeOf(val).Kind() == reflect.Func {
				cbs = val
			} else {
				args = append(args, v)
			}
		}
	}
	list := kit.Simple(args...)
	if len(list) == 0 && !m.Hand {
		list = m.meta[MSG_DETAIL]
	}
	if len(list) == 0 {
		return m
	}
	ok := false
	run := func(msg *Message, ctx *Context, cmd *Command, key string, arg ...string) {
		key = kit.Slice(strings.Split(key, PT), -1)[0]
		if ok = true; cbs != nil {
			msg.OptionCB(key, cbs)
		}
		for k, v := range opts {
			msg.Option(k, v)
		}
		msg._source = _source
		m.TryCatch(msg, true, func(msg *Message) { m = ctx._command(msg, cmd, key, arg...) })
	}
	if list[0] == "" {
		run(m.Spawn(), m.target, m._cmd, m._key, list[1:]...)
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
func (c *Context) _command(m *Message, cmd *Command, key string, arg ...string) *Message {
	if m._key, m._sub, m._cmd = key, SELECT, cmd; cmd == nil {
		return m
	}
	if m.Hand, m.meta[MSG_DETAIL] = true, kit.Simple(m.PrefixKey(), arg); cmd.Actions != nil {
		if len(arg) > 1 && arg[0] == ACTION {
			if h, ok := cmd.Actions[arg[1]]; ok {
				return c._action(m, cmd, key, arg[1], h, arg[2:]...)
			}
		}
		if len(arg) > 0 {
			if h, ok := cmd.Actions[arg[0]]; ok {
				return c._action(m, cmd, key, arg[0], h, arg[1:]...)
			}
		}
	}
	if len(arg) > 0 && arg[0] == ACTION {
		if arg[1] == "inputs" {
			return m
		}
	}
	return m.CmdHand(cmd, key, arg...)
}
func (c *Context) _action(m *Message, cmd *Command, key string, sub string, h *Action, arg ...string) *Message {
	if h.Hand == nil {
		return m.Cmdy(kit.Split(kit.Select(sub, h.Name)), arg)
	}
	if m._key, m._cmd, m._sub = key, cmd, sub; len(h.List) > 0 && sub != SEARCH {
		order := false
		for i, v := range h.List {
			name := kit.Format(kit.Value(v, NAME))
			if i == 0 {
				if len(arg) > 0 && arg[0] == name {
					for i := 0; i < len(arg)-1; i += 2 {
						if strings.HasPrefix(arg[i], PS) {
							break
						}
						// if arg[i+1] != "" {
						// 	m.Option(arg[i], arg[i+1])
						// }
						m.Option(arg[i], arg[i+1])
					}
				} else {
					order = true
				}
			}
			if order {
				// if value := kit.Select("", arg, i); value != "" {
				// 	m.Option(name, value)
				// }
				if i < len(arg) {
					m.Option(name, arg[i])
				}
			}
			if m.Warn(m.OptionDefault(name, kit.Format(kit.Value(v, VALUE))) == "" && kit.Value(v, "need") == "must", ErrNotValid, name) {
				return m
			}
		}
	}
	if m._target = logs.FileLine(h.Hand); cmd.RawHand != nil {
		m._target = kit.Join(kit.Slice(kit.Split(kit.Format(cmd.RawHand), PS), -3), PS)
	}
	m.Log(LOG_CMDS, "%s.%s %s %d %v", c.Name, key, sub, len(arg), arg, logs.FileLineMeta(m._fileline()))
	h.Hand(m, arg...)
	return m
}
func MergeActions(arg ...Any) Actions {
	if len(arg) == 0 {
		return nil
	}
	list := arg[0].(Actions)
	for _, from := range arg[1:] {
		switch from := from.(type) {
		case Actions:
			for k, v := range from {
				if h, ok := list[k]; !ok {
					list[k] = v
				} else if k == CTX_INIT {
					last := h.Hand
					hand := v.Hand
					h.Hand = func(m *Message, arg ...string) {
						hand(m, arg...)
						last(m, arg...)
					}
				} else if k == CTX_EXIT {
					last := h.Hand
					hand := v.Hand
					h.Hand = func(m *Message, arg ...string) {
						hand(m, arg...)
						last(m, arg...)
					}
				} else if h.Name = kit.Select(v.Name, h.Name); h.Hand == nil {
					h.Hand = v.Hand
				}
			}
		case string:
			h := list[CTX_INIT]
			list[CTX_INIT] = &Action{Hand: func(m *Message, arg ...string) {
				m.Search(from, func(p *Context, s *Context, key string, cmd *Command) {
					for k, v := range cmd.Actions {
						func(k string) {
							if h, ok := list[k]; !ok {
								list[k] = &Action{Name: v.Name, Help: v.Help, Hand: func(m *Message, arg ...string) { m.Cmdy(from, k, arg) }}
							} else if h.Hand == nil {
								h.Hand = func(m *Message, arg ...string) { m.Cmdy(from, k, arg) }
							}
						}(k)
					}
				})
				if h != nil {
					h.Hand(m, arg...)
				}
			}}
		default:
			Pulse.ErrorNotImplement(from)
		}
	}
	return list
}
func SplitCmd(name string, actions Actions) (list []Any) {
	const (
		TEXT     = "text"
		CONTENT  = "content"
		TEXTAREA = "textarea"
		PASSWORD = "password"
		SELECT   = "select"
		BUTTON   = "button"
	)
	const (
		RUN     = "run"
		REFRESH = "refresh"
		LIST    = "list"
		BACK    = "back"
		AUTO    = "auto"
		PAGE    = "page"
		ARGS    = "args"
	)
	item, button := kit.Dict(), false
	push := func(arg ...string) {
		button = kit.Select("", arg, 0) == BUTTON
		item = kit.Dict(TYPE, kit.Select("", arg, 0), NAME, kit.Select("", arg, 1), ACTION, kit.Select("", arg, 2))
		list = append(list, item)
	}
	ls := kit.Split(name, SP, "*:=@")
	for i := 1; i < len(ls); i++ {
		switch ls[i] {
		case RUN:
			push(BUTTON, ls[i])
		case REFRESH:
			push(BUTTON, ls[i], AUTO)
		case LIST:
			push(BUTTON, ls[i], AUTO)
		case AUTO:
			push(BUTTON, LIST, AUTO)
			push(BUTTON, BACK)
		case PAGE:
			push(TEXT, "limit")
			push(TEXT, "offend")
			push(BUTTON, "prev")
			push(BUTTON, "next")
		case ARGS, TEXT, TEXTAREA, CONTENT:
			push(TEXTAREA, ls[i])
		case PASSWORD:
			push(PASSWORD, ls[i])
		case "time":
			push(TEXT, ls[i], "date")
		case "*":
			item["need"] = "must"
		case DF:
			if item[TYPE] = kit.Select("", ls, i+1); item[TYPE] == BUTTON {
				button = true
			}
			i++
		case EQ:
			if value := kit.Select("", ls, i+1); strings.Contains(value, FS) {
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
		case AT:
			item[ACTION] = kit.Select("", ls, i+1)
			i++
		default:
			push(kit.Select(TEXT, BUTTON, button || actions != nil && actions[ls[i]] != nil), ls[i])
		}
	}
	return list
}
func MergeHand(hand ...Handler) Handler {
	if len(hand) == 0 {
		return nil
	}
	if len(hand) == 1 {
		return hand[0]
	}
	return func(m *Message, arg ...string) {
		for _, h := range hand {
			if h != nil {
				h(m, arg...)
			}
		}
	}
}
