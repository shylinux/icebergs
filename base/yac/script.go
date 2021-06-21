package yac

import (
	"strconv"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

type frame struct {
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

func _script_push_stack(m *ice.Message, f *frame) {
	stack := m.Optionv("stack").(*stack)
	stack.push(f)
}
func _script_define(m *ice.Message, key, value string) {
	stack := m.Optionv("stack").(*stack)
	stack.define(key, value)
}
func _script_runing(m *ice.Message, nhash string) bool {
	switch nhash {
	case "if", "for", "end":
		return true
	}

	stack := m.Optionv("stack").(*stack)
	return !stack.fs[len(stack.fs)-1].skip
}
func _script_pop_stack(m *ice.Message) {
	stack := m.Optionv("stack").(*stack)
	stack.fs = stack.fs[:len(stack.fs)-1]
}
func _script_res(m *ice.Message, arg ...interface{}) {
	stack := m.Optionv("stack").(*stack)
	stack.res = append(stack.res, kit.Simple(arg...)...)
}

const SCRIPT = "script"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		SCRIPT: {Name: "script name npage text:textarea auto create", Help: "脚本解析", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create name=shy", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(MATRIX, mdb.CREATE, m.Option(kit.MDB_NAME))
				for _, p := range [][]string{
					[]string{"num", "num", "[0-9]+"},
					[]string{"key", "key", "[abc]+"},
					[]string{"op2", "op2", "[+\\\\-*/%]"},
					[]string{"op2", "op2", "[>=<]"},
					[]string{"val", "val", "mul{ num key }"},
					[]string{"exp", "exp", "val"},
					[]string{"exp", "exp", "val op2 val"},
					[]string{"stm", "var", "var key = exp"},
					[]string{"stm", "for", "for exp"},
					[]string{"stm", "if", "if exp"},
					[]string{"stm", "cmd", "pwd"},
					[]string{"stm", "end", "end"},
					[]string{"script", "script", "rep{ stm }"},
				} {
					m.Cmdx(MATRIX, mdb.INSERT, m.Option(kit.MDB_NAME), p)
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

				m.Debug(" %v %v %v", arg[0], arg[1], arg[2])
				n1, e1 := strconv.ParseInt(arg[0], 10, 64)
				n2, e2 := strconv.ParseInt(arg[2], 10, 64)
				m.Debug(" %v %v %v", n1, arg[1], n2)
				switch arg[1] {
				case ">":
					if e1 == nil && e2 == nil {
						m.Echo("%t", n1 > n2)
					} else {
						m.Echo("%t", arg[0] > arg[2])
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
			"if": {Name: "if exp", Help: "判断", Hand: func(m *ice.Message, arg ...string) {
				if arg[1] == "true" {
					_script_push_stack(m, &frame{key: arg[0], skip: false})
					return
				}
				if arg[1] == "false" {
					_script_push_stack(m, &frame{key: arg[0], skip: true})
					return
				}
				if n1, e1 := strconv.ParseInt(arg[1], 10, 64); e1 == nil {
					_script_push_stack(m, &frame{key: arg[0], skip: n1 == 0})
					m.Echo("%t", n1 != 0)
				} else {
					_script_push_stack(m, &frame{skip: len(arg[1]) == 0})
					m.Echo("%t", len(arg[1]) > 0)
				}
			}},
			"var": {Name: "var key = exp", Help: "变量", Hand: func(m *ice.Message, arg ...string) {
				_script_define(m, arg[1], arg[3])
			}},
			"for": {Name: "for exp", Help: "循环", Hand: func(m *ice.Message, arg ...string) {
			}},
			"cmd": {Name: "cmd", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				_script_res(m, m.Cmdx(cli.SYSTEM, arg))
			}},
			"end": {Name: "end", Help: "结束", Hand: func(m *ice.Message, arg ...string) {
				_script_pop_stack(m)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				m.Cmdy(MATRIX, arg)
				return
			}

			stack := &stack{}
			stack.push(&frame{})
			m.Option("stack", stack)
			m.Cmdy(MATRIX, PARSE, arg[0], arg[1], arg[2], func(nhash string, hash int, word []string, rest []byte) (int, []string, []byte) {
				m.Debug("script %v %v", nhash, word)
				if _, ok := c.Commands[SCRIPT].Action[nhash]; ok && _script_runing(m, nhash) {
					msg := m.Cmd(SCRIPT, nhash, word)
					return hash, msg.Resultv(), rest
				}
				return hash, word, rest
			})
			m.Resultv(stack.res)
		}},
	}})
}
