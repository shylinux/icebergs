package ice

import (
	"strings"

	kit "shylinux.com/x/toolkits"
)

type MakeInfo struct {
	Hash     string
	Time     string
	Path     string
	Module   string
	Domain   string
	Remote   string
	Branch   string
	Version  string
	Hostname string
	Username string
	Email    string
}

var Info = struct {
	Make MakeInfo

	Hostname string
	Pathname string
	Username string
	PidPath  string
	Colors   bool

	Domain    string
	NodeType  string
	NodeName  string
	Localhost bool
	Important bool

	File  Maps
	Gomod Maps
	Route Maps
	Index Map

	merges []Any
	render map[string]func(*Message, ...Any) string
	Load   func(m *Message, key ...string) *Message
	Save   func(m *Message, key ...string) *Message
	Log    func(m *Message, p, l, s string)
}{
	File:  Maps{},
	Gomod: Maps{},
	Route: Maps{},
	Index: Map{},

	render: map[string]func(*Message, ...Any) string{},
	Load:   func(m *Message, key ...string) *Message { return m },
	Save:   func(m *Message, key ...string) *Message { return m },
	Log:    func(m *Message, p, l, s string) {},
}

func AddMerges(h ...Any) { Info.merges = append(Info.merges, h...) }

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
					h.Hand = MergeHand(v.Hand, h.Hand)
				} else if k == CTX_EXIT {
					h.Hand = MergeHand(h.Hand, v.Hand)
				} else if h.Name = kit.Select(v.Name, h.Name); h.Hand == nil {
					h.Hand = v.Hand
				}
			}
		case string:
			h := list[CTX_INIT]
			if h == nil {
				list[CTX_INIT] = &Action{}
				h = list[CTX_INIT]
			}
			h.Hand = MergeHand(h.Hand, func(m *Message, arg ...string) {
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
			})
		default:
			Pulse.ErrorNotImplement(from)
		}
	}
	return list
}
func SplitCmd(name string, actions Actions) (list []Any) {
	const (
		TEXT     = "text"
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
		CONTENT = "content"
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
			push(BUTTON, "prev")
			push(BUTTON, "next")
			push(TEXT, "offend")
			push(TEXT, "limit")
		case ARGS, CONTENT, TEXTAREA, TEXT:
			push(TEXTAREA, ls[i])
		case PASSWORD:
			push(PASSWORD, ls[i])
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
