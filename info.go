package ice

import (
	"os"
	"path"
	"reflect"
	"strings"

	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

type MakeInfo struct {
	Username string
	Hostname string
	Path     string
	Time     string
	Git      string
	Go       string

	Remote  string
	Branch  string
	Version string
	Forword string
	Author  string
	Email   string
	Hash    string
	When    string
	Message string

	System string
	Domain string
	Module string
}

func (s MakeInfo) Versions() string {
	if s.Hash == "" {
		return ""
	} else if s.Version == "" {
		return s.Hash[:6]
	} else if kit.IsIn(s.Forword, "", "0") {
		return s.Version
	} else {
		return kit.Format("%s-%s-%s", s.Version, s.Forword, s.Hash[:6])
	}
}

var Info = struct {
	Make MakeInfo
	Time string
	Size string
	Hash string

	Username string
	Hostname string
	Pathname string
	NodeName string
	NodeType string

	Pwd       string
	Lang      string
	System    string
	Domain    string
	PidPath   string
	Traceid   string
	Localhost bool
	Important bool
	Colors    bool

	File  Maps
	Gomod Maps
	Route Maps
	Index Map

	merges []Any
	render map[string]func(*Message, ...Any) string
	Stack  map[string]func(m *Message, key string, arg ...Any) Any
	Inputs []func(m *Message, arg ...string)

	PushStream func(m *Message) *Message
	PushNotice func(m *Message, arg ...Any)
	SystemCmd  func(m *Message, arg ...Any) *Message
	AdminCmd   func(m *Message, cmd string, arg ...Any) *Message
	Template   func(m *Message, p string, data ...Any) string
	Save       func(m *Message, key ...string) *Message
	Load       func(m *Message, key ...string) *Message
	Log        func(m *Message, p, l, s string)
}{
	Localhost: true,

	File:  Maps{},
	Gomod: Maps{},
	Route: Maps{},
	Index: Map{},

	render: map[string]func(*Message, ...Any) string{},
	Stack:  map[string]func(m *Message, key string, arg ...Any) Any{},
}

func init() { Info.Pwd = kit.Path(""); Info.Traceid = os.Getenv(LOG_TRACE) }

func AddMergeAction(h ...Any) { Info.merges = append(Info.merges, h...) }
func MergeHand(hand ...Handler) Handler {
	if len(hand) == 0 {
		return nil
	} else if len(hand) == 1 {
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
	kit.If(list == nil, func() { list = Actions{} })
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
				} else if h.Name, h.Help, h.Icon = kit.Select(v.Name, h.Name), kit.Select(v.Help, h.Help), kit.Select(v.Icon, h.Icon); h.Hand == nil {
					h.Hand = v.Hand
				}
			}
		case string:
			h := list[CTX_INIT]
			kit.If(h == nil, func() { list[CTX_INIT] = &Action{}; h = list[CTX_INIT] })
			h.Hand = MergeHand(h.Hand, func(m *Message, arg ...string) {
				_cmd := m._cmd
				m.Search(from, func(p *Context, s *Context, key string, cmd *Command) {
					kit.For(kit.Value(cmd.Meta, kit.Keys(CTX_TRANS, html.INPUT)), func(k, v string) {
						if kit.Format(kit.Value(_cmd.Meta, kit.Keys(CTX_TRANS, html.INPUT, k))) == "" {
							kit.Value(_cmd.Meta, kit.Keys(CTX_TRANS, html.INPUT, k), v)
						}
					})
					for k, v := range cmd.Actions {
						func(k string) {
							h, ok := list[k]
							if !ok {
								list[k] = &Action{Name: v.Name, Help: v.Help, Icon: v.Icon, Hand: func(m *Message, arg ...string) { m.Cmdy(from, k, arg) }}
							} else if h.Hand == nil {
								h.Hand = func(m *Message, arg ...string) { m.Cmdy(from, k, arg) }
							}
							if !ok || h.Icon == "" {
								kit.Value(_cmd.Meta, kit.Keys(CTX_ICONS, k), v.Icon)
							}
							if !ok || h.Help == "" {
								if help := kit.Split(v.Help, " :："); len(help) > 0 {
									if kit.Value(_cmd.Meta, kit.Keys(CTX_TRANS, strings.TrimPrefix(k, "_")), help[0]); len(help) > 1 {
										kit.Value(_cmd.Meta, kit.Keys(CTX_TITLE, k), help[1])
									}
								}
							}
							kit.If((!ok || len(h.List) == 0) && len(v.List) > 0, func() { _cmd.Meta[k] = v.List })
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
		RUN     = "run"
		REFRESH = "refresh"
		LIST    = "list"
		BACK    = "back"
		AUTO    = "auto"
		PAGE    = "page"
		ARGS    = "args"
		CONTENT = "content"
		FILTER  = "filter"
	)
	item, button := kit.Dict(), false
	push := func(arg ...string) {
		button = kit.Select("", arg, 0) == html.BUTTON
		item = kit.Dict(TYPE, kit.Select("", arg, 0), NAME, kit.Select("", arg, 1), ACTION, kit.Select("", arg, 2))
		list = append(list, item)
	}
	ls := kit.Split(name, SP, "*:=@")
	for i := 1; i < len(ls); i++ {
		switch ls[i] {
		case RUN:
			push(html.BUTTON, ls[i])
		case REFRESH:
			push(html.BUTTON, ls[i], AUTO)
		case LIST:
			push(html.BUTTON, ls[i], AUTO)
		case AUTO:
			push(html.BUTTON, LIST, AUTO)
			push(html.BUTTON, BACK)
		case PAGE:
			push(html.BUTTON, "prev")
			push(html.BUTTON, "next")
			push(html.TEXT, "offend")
			push(html.TEXT, "limit")
		case ARGS, CONTENT, html.TEXTAREA, html.TEXT, "extra":
			push(html.TEXTAREA, ls[i])
		case html.PASSWORD:
			push(html.PASSWORD, ls[i])
		case FILTER:
			push(html.TEXT, ls[i])
		case "*":
			item["need"] = "must"
		case DF:
			if item[TYPE] = kit.Select("", ls, i+1); item[TYPE] == html.BUTTON {
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
			push(kit.Select(html.TEXT, html.BUTTON, button || actions != nil && actions[ls[i]] != nil), ls[i])
		}
	}
	return list
}
func Module(prefix string, arg ...Any) {
	list := map[string]Any{}
	for _, v := range arg {
		list[kit.FuncName(v)] = v
	}
	Info.Stack[prefix] = func(m *Message, key string, arg ...Any) Any {
		if len(arg) > 0 {
			switch v := arg[0].(type) {
			case *Message:
				m, arg = v, arg[1:]
			}
		}
		if v, ok := list[kit.LowerCapital(key)]; ok {
			switch v := v.(type) {
			case func(m *Message):
				v(m)
			case func(m *Message, arg ...Any):
				v(m, arg...)
			case func(m *Message, arg ...Any) string:
				return v(m, arg...)
			case func(m *Message, arg ...Any) *Message:
				return v(m, arg...)
			case func(m *Message, arg ...string) *Message:
				return v(m, kit.Simple(arg...)...)
			default:
				cb, args := reflect.ValueOf(v), []reflect.Value{reflect.ValueOf(m)}
				kit.For(arg, func(v Any) { args = append(args, reflect.ValueOf(v)) })
				if res := cb.Call(args); len(res) > 0 && res[0].CanInterface() {
					return res[0].Interface()
				}
			}
		} else {
			m.ErrorNotImplement(key)
		}
		return m
	}
}
func (m *Message) FileURI(dir string) string {
	if dir == "" || kit.HasPrefix(dir, HTTP) {
		return dir
	}
	if strings.Contains(dir, "/pkg/mod/") {
		dir = strings.Split(dir, "/pkg/mod/")[1]
	} else if Info.Make.Path != "" && strings.HasPrefix(dir, Info.Make.Path) {
		dir = strings.TrimPrefix(dir, Info.Make.Path)
	} else if strings.HasPrefix(dir, kit.Path("")+PS) {
		dir = strings.TrimPrefix(dir, kit.Path("")+PS)
	} else if strings.HasPrefix(dir, ISH_PLUGED) {
		dir = strings.TrimPrefix(dir, ISH_PLUGED)
	}
	if strings.HasPrefix(dir, PS) {

	} else if strings.HasPrefix(dir, USR_VOLCANOS) {
		dir = strings.TrimPrefix(dir, USR)
	} else {
		dir = kit.MergeURL(path.Join(PS, REQUIRE, dir), POD, m.Option(MSG_USERPOD))
	}
	if m.Option(MSG_USERWEB0) != "" {
		dir = kit.MergeURL2(m.Option(MSG_USERWEB), dir)
	}
	return dir
}
