package xterm

import (
	"bytes"
	"os"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

type idata struct {
	cut  string
	bak  string
	arg  string
	app  string
	tip  string
	due  string
	args []string
	cmds []string
	list []string
	pos  int
	pipe *os.File
}
type iterm struct {
	m *ice.Message
	r *os.File
	w *os.File
	*idata
}

func newiterm(m *ice.Message) (XTerm, error) {
	r, w, e := os.Pipe()
	return &iterm{m: m, r: r, w: w, idata: &idata{cmds: kit.Simple(
		kit.SortedKey(ice.Info.Index),
		m.Cmd(ctx.COMMAND).Appendv(ctx.INDEX),
		m.Cmd(nfs.DIR, "/bin", mdb.NAME).Appendv(mdb.NAME),
	)}}, e
}
func (s iterm) Setsize(rows, cols string) error {
	s.w.Write([]byte(s.prompt()))
	return nil
}
func (s iterm) Writeln(data string, arg ...ice.Any) { s.Write(kit.Format(data, arg...) + lex.NL) }
func (s iterm) Write(data string) (int, error) {
	if s.pipe != nil {
		return s.pipe.Write([]byte(data))
	}
	res, ctrl := "", ""
	for _, c := range data {
		switch c := string(c); c {
		case SOH: // Ctrl+A
			res += s.repeat(s.arg)
			s.arg, s.app = "", s.arg+s.app
		case STX: // Ctrl+B
			res = s.left(res)
		case ETX: // Ctrl+C
			s.m.Cmd("web.code.vimer", "compile")
		case EOT: // Ctrl+D
			if len(s.app) > 0 {
				s.app = s.app[1:]
				res = s.rest(res)
			} else {
				s.w.Close()
			}
		case ENQ: // Ctrl+E
			res += s.repeat(s.app, ESC_C)
			s.arg, s.app = s.arg+s.app, ""
		case ACK: // Ctrl+F
			res = s.right(res)
		case BEL: // Ctrl+G
			res += c
		case BS: // Ctrl+H
			res = s.dels(res)
		case NL: // Ctrl+J
			res = s.exec(s.m, res)
		case VT: // Ctrl+K
			if len(s.app) > 0 {
				s.cut, s.app = s.app, ""
				res = s.rest(res)
			}
		case NP: // Ctrl+L
			res = s.rest(res + ESC_H + ESC_2J + s.prompt() + s.arg)
		case CR: // Ctrl+M
			res = s.exec(s.m, res)
		case SO: // Ctrl+N
			res = s.hist(res, 1)
		case SI: // Ctrl+O
			if s.arg == "" {
				s.arg = kit.Select("", s.list, -1)
				res = s.exec(s.m, res+s.arg)
			} else {
				res = s.exec(s.m, res)
			}
		case DLE: // Ctrl+P
			res = s.hist(res, -1)
		case DC1: // Ctrl+Q
		case DC2: // Ctrl+R
		case DC3: // Ctrl+S
		case DC4: // Ctrl+T
			if n := len(s.arg); n > 1 {
				s.arg = s.arg[:n-2] + string(s.arg[n-1]) + string(s.arg[n-2])
				res = s.rest(res + BS + BS + s.arg[n-2:])
			}
		case NAK: // Ctrl+U
			if len(s.arg) > 0 {
				res = res + s.repeat(s.arg)
				s.cut, s.arg = s.arg, ""
				res = s.rest(res)
			}
		case SYN: // Ctrl+V
		case ETB: // Ctrl+W
			arg := s.arg
			for len(s.arg) > 0 {
				if c := s.arg[len(s.arg)-1]; c == ' ' {
					s.arg = s.arg[:len(s.arg)-1]
					res += s.repeat(string(c))
					continue
				}
				break
			}
			for len(s.arg) > 0 {
				if c := s.arg[len(s.arg)-1]; len(s.arg) == 1 || c != ' ' {
					s.arg = s.arg[:len(s.arg)-1]
					res += s.repeat(string(c))
					continue
				}
				break
			}
			s.cut = strings.TrimPrefix(arg, s.arg)
			res = s.rest(res)
		case CAN: // Ctrl+X
		case EM: // Ctrl+Y
			s.arg += s.cut
			res = s.rest(res + s.cut)
		case SUB: // Ctrl+Z
		case DEL:
			res = s.dels(res)
		case ESC: // Ctrl+[
			ctrl = c
		default:
			if ctrl == "" {
				if c == HT { // Ctrl+I
					if s.tip != "" {
						s.arg += s.app + s.tip
						res += s.app + s.tip
						s.app = ""
						break
					}
				}
				s.arg += c
				res = s.rest(res + c)
				break
			} else if ctrl == ESC && c == "[" {
				ctrl += c
				break
			} else if ctrl == ESC+"[" {
				switch c {
				case "A": // ArrowUp
					res = s.hist(res, -1)
				case "B": // ArrowDown
					res = s.hist(res, 1)
				case "C": // ArrowRight
					res = s.right(res)
				case "D": // ArrowLeft
					res = s.left(res)
				}
			}
			ctrl = ""
		}
	}
	s.w.Write([]byte(res))
	return len(data), nil
}
func (s iterm) Read(buf []byte) (int, error) {
	return s.r.Read(buf)
}
func (s iterm) Close() error { return nil }
func (s iterm) style(style string, str string) string {
	return kit.Format("\033[%sm%s\033[0m", style, str)
}
func (s iterm) prompt() string {
	return kit.Format("%s%d[%s]$ ", CR+ESC_K, len(s.list), time.Now().Format("15:04:05"))
}
func (s iterm) repeat(str string, arg ...string) string {
	count := 0
	for _, c := range str {
		switch c {
		case '\t':
			count += 4
		default:
			count++
		}
	}
	return strings.Repeat(kit.Select(BS, arg, 0), count)
}
func (s iterm) left(res string) string {
	if n := len(s.arg); n > 0 {
		s.arg, s.app = s.arg[:n-1], s.arg[n-1:]+s.app
		res += BS
	}
	return res
}
func (s iterm) right(res string) string {
	if len(s.app) > 0 {
		s.arg, s.app = s.arg+s.app[:1], s.app[1:]
		res += ESC_C
	}
	return res
}
func (s iterm) dels(res string) string {
	if len(s.arg) > 0 {
		s.arg = s.arg[:len(s.arg)-1]
		res = s.rest(res + BS)
	}
	return res
}
func (s iterm) tips(arg string) (tip string) {
	if kit.HasSuffix(arg, lex.SP, nfs.PS) {
		s.args = s.m.CmdList(kit.Split(arg)...)
		kit.For(s.args, func(i int, v string) { s.args[i] = strings.TrimPrefix(v, kit.Select("", kit.Split(arg), -1)) })
		s.due = CRNL + ESC_K + strings.Join(s.args, lex.SP)
	} else if len(s.args) > 0 {
		args, key := []string{}, kit.Select("", kit.Split(arg, "\t ./"), -1)
		kit.If(kit.HasSuffix(arg, ".", "/"), func() { key = "" })
		kit.For(s.args, func(arg string) { kit.If(strings.HasPrefix(arg, key), func() { args = append(args, arg) }) })
		s.due = CRNL + ESC_K + strings.Join(args, lex.SP)
		return strings.TrimPrefix(kit.Select("", args, 0), key)
	} else {
		s.due = ""
	}
	if strings.HasSuffix(arg, nfs.PT) {
		s.args = nil
		kit.For(s.cmds, func(cmd string) {
			kit.If(strings.HasPrefix(cmd, arg), func() { s.args = append(s.args, strings.TrimPrefix(cmd, arg)) })
		})
		s.due = CRNL + ESC_K + strings.Join(s.args, lex.SP)
	}
	for i := len(s.list) - 1; i >= 0; i-- {
		if v := s.list[i]; strings.HasPrefix(v, arg) {
			return strings.TrimPrefix(v, arg)
		}
	}
	for _, v := range s.cmds {
		if strings.HasPrefix(v, arg) {
			return strings.TrimPrefix(v, arg)
		}
	}
	return tip
}
func (s iterm) rest(res string) string {
	s.tip = s.tips(s.arg + s.app)
	return res + ESC_K + ESC_s + s.app + s.style("2", s.tip+s.due) + ESC_u
}
func (s iterm) exec(m *ice.Message, res string) string {
	defer func() { s.arg, s.app, s.args, s.pipe = "", "", nil, nil }()
	arg := kit.Split(s.arg + s.app)
	if len(arg) == 0 {
		return CRNL + s.prompt()
	} else if len(s.list) == 0 || s.arg+s.app != s.list[len(s.list)-1] {
		s.list, s.pos = append(s.list, s.arg+s.app), len(s.list)+1
	}
	msg, end := m.Cmd(arg, kit.Dict(ice.MSG_USERUA, "ish")), false
	if res += CRNL + ESC_K; msg.IsErrNotFound() {
		s.w.Write([]byte(res))
		r, w, _ := os.Pipe()
		res, s.pipe = "", w
		env := kit.EnvList(
			"TERM", "xterm",
			"LINES", m.Option("rows"),
			"COLUMNS", m.Option("cols"),
			"SHELL", "/bin/ish",
			"USER", m.Option(ice.MSG_USERNAME),
		)
		m.Cmd(cli.SYSTEM, arg, kit.Dict(cli.CMD_INPUT, r, cli.CMD_OUTPUT, nfs.Pipe(m, func(buf []byte) {
			s.w.Write(bytes.ReplaceAll(buf, []byte(lex.NL), []byte(CRNL)))
			end = bytes.HasSuffix(buf, []byte(lex.NL))
		}), cli.CMD_ENV, env))
	} else {
		kit.If(msg.Result() == "", func() { msg.TableEcho() })
		res += strings.ReplaceAll(msg.Result(), lex.NL, CRNL)
		end = strings.HasSuffix(res, lex.NL)
	}
	kit.If(!end, func() { res += CRNL })
	return res + s.prompt()
}
func (s iterm) hist(res string, skip int) string {
	res += s.repeat(s.arg) + ESC_K
	kit.If(s.pos == len(s.list), func() { s.bak = s.arg })
	for s.history(skip); s.pos < len(s.list); s.history(skip) {
		if strings.Contains(s.list[s.pos], s.bak) {
			s.arg, s.app = s.list[s.pos], ""
			res += s.list[s.pos]
			return res
		}
	}
	s.arg, s.app = s.bak, ""
	res += s.bak
	return res
}
func (s iterm) history(n int) { s.pos = (s.pos + n + len(s.list) + 1) % (len(s.list) + 1) }

// AEBF UKHD ITWY JMCZ NPRS LGVQ XO
const (
	NUL    = "\000"
	SOH    = "\001" // Ctrl+A
	STX    = "\002" // Ctrl+B
	ETX    = "\003" // Ctrl+C
	EOT    = "\004" // Ctrl+D
	ENQ    = "\005" // Ctrl+E
	ACK    = "\006" // Ctrl+F
	BEL    = "\007" // Ctrl+G
	BS     = "\010" // Ctrl+H
	HT     = "\011" // Ctrl+I
	NL     = "\012" // Ctrl+J
	VT     = "\013" // Ctrl+K
	NP     = "\014" // Ctrl+L
	CR     = "\015" // Ctrl+M
	SO     = "\016" // Ctrl+N
	SI     = "\017" // Ctrl+O
	DLE    = "\020" // Ctrl+P
	DC1    = "\021" // Ctrl+Q
	DC2    = "\022" // Ctrl+R
	DC3    = "\023" // Ctrl+S
	DC4    = "\024" // Ctrl+T
	NAK    = "\025" // Ctrl+U
	SYN    = "\026" // Ctrl+V
	ETB    = "\027" // Ctrl+W
	CAN    = "\030" // Ctrl+X
	EM     = "\031" // Ctrl+Y
	SUB    = "\032" // Ctrl+Z
	ESC    = "\033" // Ctrl+[
	ESC_C  = "\033[C"
	ESC_H  = "\033[H"
	ESC_K  = "\033[K"
	ESC_2J = "\033[2J"
	ESC_s  = "\033[s"
	ESC_u  = "\033[u"
	DEL    = "\177"
	CRNL   = "\r\n"
)
