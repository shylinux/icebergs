package yac

import (
	"io/ioutil"
	"strconv"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/lex"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

type frame struct {
	pos  int
	key  string
	skip bool
	data map[string]string
}
type stack struct {
	fs  []*frame
	res []string
}

func (s *stack) push(f *frame) *stack {
	f.data = map[string]string{}
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
func (s *stack) echo(arg ...interface{}) {
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
	if arg == "true" {
		return true
	}
	if arg == "false" {
		return false
	}
	if n1, e1 := strconv.ParseInt(arg, 10, 64); e1 == nil {
		return n1 != 0
	}
	return false
}

const SCRIPT = "script"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		SCRIPT: {Name: "script name npage text:textarea auto create", Help: "脚本解析", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create name=shy text=etc/yac.txt", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(MATRIX, mdb.CREATE, m.Option(kit.MDB_NAME))
				if buf, err := ioutil.ReadFile(m.Option(kit.MDB_TEXT)); err == nil {
					m.Option(kit.MDB_TEXT, string(buf))
				}

				m.Option(kit.MDB_TEXT, strings.ReplaceAll(m.Option(kit.MDB_TEXT), "\\", "\\\\"))
				for _, line := range kit.Split(m.Option(kit.MDB_TEXT), "\n", "\n", "\n") {
					if strings.HasPrefix(strings.TrimSpace(line), "#") {
						continue
					}
					line = strings.ReplaceAll(line, "\\", "\\\\")
					if list := kit.Split(line, " ", " ", " "); len(list) > 2 {
						m.Cmdx(MATRIX, mdb.INSERT, m.Option(kit.MDB_NAME), list[0], list[1], strings.Join(list[2:], " "))
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
						m.Echo("%s", strings.ReplaceAll(arg[0], arg[2], ""))
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
						m.Echo("%s", strings.ReplaceAll(arg[0], arg[2], ""))
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
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				m.Cmdy(MATRIX, arg)
				return
			}

			stack := &stack{}
			stack.push(&frame{})
			m.Option("stack", stack)
			m.Cmdy(MATRIX, PARSE, arg[0], arg[1], arg[2], func(nhash string, hash int, word []string, begin int, stream *lex.Stream) (int, []string) {
				m.Option("stream", stream)
				if _, ok := c.Commands[SCRIPT].Action[nhash]; ok && stack.can_run(nhash) {
					msg := m.Cmd(SCRIPT, nhash, word, ice.Option{"begin", begin})
					return hash, msg.Resultv()
				}
				return hash, word
			})
			m.Resultv(stack.res)
		}},
	}})
}
