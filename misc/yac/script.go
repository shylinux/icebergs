package yac

import (
	"io/ioutil"
	"strconv"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

type frame struct {
	pos  int
	key  string
	skip bool
	data ice.Maps
}
type stack struct {
	fs  []*frame
	res []string
}

func (s *stack) push(f *frame) *stack {
	f.data = ice.Maps{}
	s.fs = append(s.fs, f)
	return s
}
func (s *stack) pop() *frame {
	last := s.fs[len(s.fs)-1]
	s.fs = s.fs[:len(s.fs)-1]
	return last
}
func (s *stack) can_run(nhash string) bool {
	switch nhash {
	case "if", "for", "end":
		return true
	}

	return !s.fs[len(s.fs)-1].skip
}

func (s *stack) define(key, value string) {
	if len(s.fs) > 0 {
		s.fs[len(s.fs)-1].data[key] = value
	}
}
func (s *stack) value(key string) string {
	for i := len(s.fs) - 1; i >= 0; i-- {
		if value, ok := s.fs[i].data[key]; ok {
			return value
		}
	}
	return ""
}
func (s *stack) let(key, value string) string {
	for i := len(s.fs) - 1; i >= 0; i-- {
		if val, ok := s.fs[i].data[key]; ok {
			s.fs[i].data[key] = value
			return val
		}
	}
	return ""
}
func (s *stack) echo(arg ...ice.Any) {
	s.res = append(s.res, kit.Simple(arg...)...)
}

func _get_stack(m *ice.Message) *stack {
	return m.Optionv("stack").(*stack)
}
func _push_stack(m *ice.Message, f *frame) {
	f.pos = kit.Int(m.Option("begin"))
	_get_stack(m).push(f)
}
func _pop_stack(m *ice.Message) *frame {
	return _get_stack(m).pop()
}

func _exp_true(m *ice.Message, arg string) bool {
	if arg == ice.TRUE {
		return true
	}
	if arg == ice.FALSE {
		return false
	}
	if n1, e1 := strconv.ParseInt(arg, 10, 64); e1 == nil {
		return n1 != 0
	}
	return false
}

const SCRIPT = "script"

func init() {
	Index.MergeCommands(ice.Commands{
		SCRIPT: {Name: "script name npage text auto create", Help: "脚本解析", Actions: ice.Actions{
			mdb.CREATE: {Name: "create name=shy text=etc/yac.txt", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(MATRIX, mdb.CREATE, m.Option(mdb.NAME))
				if buf, err := ioutil.ReadFile(m.Option(mdb.TEXT)); err == nil {
					m.Option(mdb.TEXT, string(buf))
				}

				m.Option(mdb.TEXT, strings.Replace(m.Option(mdb.TEXT), "\\", "\\\\", -1))
				for _, line := range kit.Split(m.Option(mdb.TEXT), "\n", "\n", "\n") {
					if strings.HasPrefix(strings.TrimSpace(line), "#") {
						continue
					}
					line = strings.Replace(line, "\\", "\\\\", -1)
					if list := kit.Split(line, " ", " ", " "); len(list) > 2 {
						m.Cmdx(MATRIX, mdb.INSERT, m.Option(mdb.NAME), list[0], list[1], strings.Join(list[2:], " "))
					}
				}
			}},
			"exp": {Name: "exp num op2 num", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				stack := m.Optionv("stack").(*stack)

				arg[0] = kit.Select(arg[0], stack.value(arg[0]))
				if len(arg) == 1 {
					m.Echo(arg[0])
					return
				}
				arg[2] = kit.Select(arg[2], stack.value(arg[2]))

				n1, e1 := strconv.ParseInt(arg[0], 10, 64)
				n2, e2 := strconv.ParseInt(arg[2], 10, 64)
				switch arg[1] {
				case ">":
					if e1 == nil && e2 == nil {
						m.Echo("%t", n1 > n2)
					} else {
						m.Echo("%t", arg[0] > arg[2])
					}
				case "<":
					if e1 == nil && e2 == nil {
						m.Echo("%t", n1 < n2)
					} else {
						m.Echo("%t", arg[0] < arg[2])
					}
				case "+":
					if e1 == nil && e2 == nil {
						m.Echo("%d", n1+n2)
					} else {
						m.Echo("%s", arg[0]+arg[2])
					}
				case "-":
					if e1 == nil && e2 == nil {
						m.Echo("%d", n1-n2)
					} else {
						m.Echo("%s", strings.Replace(arg[0], arg[2], "", -1))
					}
				case "*":
					if e1 == nil && e2 == nil {
						m.Echo("%d", n1*n2)
					} else {
						m.Echo(arg[0])
					}
				case "/":
					if e1 == nil && e2 == nil {
						m.Echo("%d", n1/n2)
					} else {
						m.Echo("%s", strings.Replace(arg[0], arg[2], "", -1))
					}
				case "%":
					if e1 == nil && e2 == nil {
						m.Echo("%d", n1%n2)
					} else {
						m.Echo(arg[0])
					}
				default:
					m.Echo(arg[0], arg[1], arg[2])
				}
			}},
			"var": {Name: "var key = exp", Help: "变量", Hand: func(m *ice.Message, arg ...string) {
				_get_stack(m).define(arg[1], arg[3])
			}},
			"let": {Name: "let key = exp", Help: "变量", Hand: func(m *ice.Message, arg ...string) {
				_get_stack(m).let(arg[1], arg[3])
			}},
			"cmd": {Name: "cmd", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				_get_stack(m).echo(m.Cmdx(cli.SYSTEM, arg))
			}},
			"if": {Name: "if exp", Help: "判断", Hand: func(m *ice.Message, arg ...string) {
				_push_stack(m, &frame{key: arg[0], skip: !_exp_true(m, arg[1])})
			}},
			"for": {Name: "for exp", Help: "循环", Hand: func(m *ice.Message, arg ...string) {
				_push_stack(m, &frame{key: arg[0], skip: !_exp_true(m, arg[1])})
			}},
			"end": {Name: "end", Help: "结束", Hand: func(m *ice.Message, arg ...string) {
				frame := _pop_stack(m)
				if frame.key == "for" && !frame.skip {
					stream := m.Optionv("stream").(*lex.Stream)
					stream.P = frame.pos
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 2 {
				m.Cmdy(MATRIX, arg)
				return
			}

			stack := &stack{}
			stack.push(&frame{})
			m.Option("stack", stack)
			m.Cmdy(MATRIX, PARSE, arg[0], arg[1], arg[2], func(nhash string, hash int, word []string, begin int, stream *lex.Stream) (int, []string) {
				m.Option("stream", stream)
				if _, ok := m.Target().Commands[SCRIPT].Actions[nhash]; ok && stack.can_run(nhash) {
					msg := m.Cmd(SCRIPT, nhash, word, ice.Option{"begin", begin})
					return hash, msg.Resultv()
				}
				return hash, word
			})
			m.Resultv(stack.res)
		}},
	})
}
