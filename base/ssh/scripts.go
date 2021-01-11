package ssh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

func Render(msg *ice.Message, cmd string, args ...interface{}) {
	switch arg := kit.Simple(args...); cmd {
	case ice.RENDER_VOID:
	case ice.RENDER_RESULT:
		fmt.Fprintf(msg.O, msg.Result())

	case ice.RENDER_QRCODE:
		fmt.Fprintf(msg.O, msg.Cmdx(cli.PYTHON, "qrcode", kit.Format(args[0], args[1:]...)))

	case ice.RENDER_DOWNLOAD:
		if f, e := os.Open(arg[0]); e == nil {
			defer f.Close()

			io.Copy(msg.O, f)
		}

	default:
		// 转换结果
		res := msg.Result()
		if res == "" {
			res = msg.Table().Result()
		}
		args = append(args, "length:", len(res))

		// 输出结果
		if fmt.Fprintf(msg.O, res); !strings.HasSuffix(res, "\n") {
			fmt.Fprintf(msg.O, "\n")
		}
	}
}
func Script(m *ice.Message, name string) io.Reader {
	if strings.Contains(m.Option("_script"), "/") {
		name = path.Join(path.Dir(m.Option("_script")), name)
	}
	m.Option("_script", name)

	back := kit.Split(m.Option(nfs.DIR_ROOT))
	for i := len(back) - 1; i >= 0; i-- {
		if s, e := os.Open(path.Join(path.Join(back[:i]...), name)); e == nil {
			return s
		}
	}
	if s, e := os.Open(name); e == nil {
		return s
	}

	switch strings.Split(name, "/")[0] {
	case "etc", "var":
		m.Warn(true, ice.ErrNotFound)
		return nil
	}

	if b, ok := ice.BinPack["/"+name]; ok {
		m.Info("binpack %v %v", len(b), name)
		return bytes.NewReader(b)
	}

	if msg := m.Cmd("web.spide", "dev", "GET", path.Join("/share/local/", name)); msg.Result(0) != ice.ErrWarn {
		bio := bytes.NewBuffer([]byte(msg.Result()))
		return bio
	}

	if strings.HasPrefix(name, "usr") {
		ls := strings.Split(name, "/")
		m.Cmd("web.code.git.repos", ls[1], "usr/"+ls[1])
		if s, e := os.Open(name); e == nil {
			return s
		}
	}
	return nil
}

type Frame struct {
	source string
	target *ice.Context
	stdout io.Writer

	count int
	ps1   []string
	ps2   []string

	exit bool
}

func (f *Frame) prompt(m *ice.Message, list ...string) *Frame {
	if f.source != STDIO {
		return f
	}
	if len(list) == 0 {
		list = append(list, f.ps1...)
	}

	fmt.Fprintf(f.stdout, "\r")
	for _, v := range list {
		switch v {
		case "count":
			fmt.Fprintf(f.stdout, "%d", f.count)
		case "time":
			fmt.Fprintf(f.stdout, time.Now().Format("15:04:05"))
		case "target":
			fmt.Fprintf(f.stdout, f.target.Name)
		default:
			fmt.Fprintf(f.stdout, v)
		}
	}
	return f
}
func (f *Frame) printf(m *ice.Message, res string, arg ...interface{}) *Frame {
	if len(arg) > 0 {
		fmt.Fprintf(f.stdout, res, arg...)
	} else {
		fmt.Fprint(f.stdout, res)
	}
	return f
}
func (f *Frame) option(m *ice.Message, ls []string) []string {
	ln := []string{}
	m.Option("cache.limit", 10)
	for i := 0; i < len(ls); i++ {
		if ls[i] == "--" {
			ln = append(ln, ls[i+1:]...)
			break
		}

		if strings.HasPrefix(ls[i], "-") {
			for j := i; j < len(ls); j++ {
				if j == len(ls)-1 || strings.HasPrefix(ls[j+1], "-") {
					if i < j {
						m.Option(ls[i][1:], ls[i+1:j+1])
					} else {
						m.Option(ls[i][1:], "true")
					}
					i = j
					break
				}
			}
		} else {
			ln = append(ln, ls[i])
		}
	}
	return ln
}
func (f *Frame) change(m *ice.Message, ls []string) []string {
	if len(ls) == 1 && ls[0] == "~" {
		// 模块列表
		ls = []string{"context"}
	} else if len(ls) > 0 && strings.HasPrefix(ls[0], "~") {
		// 切换模块
		target := ls[0][1:]
		if ls = ls[1:]; len(target) == 0 && len(ls) > 0 {
			target, ls = ls[0], ls[1:]
		}
		if target == "~" {
			target = ""
		}
		m.Spawn(f.target).Search(target+".", func(p *ice.Context, s *ice.Context, key string) {
			m.Info("choice: %s", s.Name)
			f.target = s
		})
	}
	return ls
}
func (f *Frame) alias(m *ice.Message, ls []string) []string {
	if alias, ok := m.Optionv(ice.MSG_ALIAS).(map[string]interface{}); ok {
		if len(ls) > 0 {
			if a := kit.Simple(alias[ls[0]]); len(a) > 0 {
				ls = append(append([]string{}, a...), ls[1:]...)
			}
		}
	}
	return ls
}
func (f *Frame) parse(m *ice.Message, line string) string {
	if strings.HasPrefix(line, "<") {
		fmt.Fprintf(m.O, line)
		return ""
	}

	for _, one := range kit.Split(line, ";", ";", ";") {
		async, one := false, strings.TrimSpace(one)
		if strings.TrimSuffix(one, "&") != one {
			async, one = true, strings.TrimSuffix(one, "&")
		}

		msg := m.Spawn(f.target)
		msg.Option("_cmd", one)

		ls := kit.Split(one)
		ls = f.alias(msg, ls)
		ls = f.change(msg, ls)
		ls = f.option(msg, ls)
		if len(ls) == 0 {
			continue
		}

		if async {
			msg.Gos(msg, func(msg *ice.Message) { msg.Cmd(ls[0], ls[1:]) })
			continue
		} else {
			msg.Cmdy(ls[0], ls[1:])
		}

		if strings.HasPrefix(msg.Result(), ice.ErrWarn) && m.Option("render") == "raw" {
			fmt.Fprintf(msg.O, line)
			continue
		}

		// 渲染引擎
		_args, _ := msg.Optionv(ice.MSG_ARGS).([]interface{})
		Render(msg, msg.Option(ice.MSG_OUTPUT), _args...)
	}
	return ""
}
func (f *Frame) scan(m *ice.Message, h, line string, r io.Reader) *Frame {
	m.Option("ssh.return", func() { f.exit = true })
	f.ps1 = kit.Simple(m.Confv("prompt", "meta.PS1"))
	f.ps2 = kit.Simple(m.Confv("prompt", "meta.PS2"))
	ps := f.ps1

	m.I, m.O = r, f.stdout
	bio := bufio.NewScanner(r)
	for f.prompt(m, ps...); bio.Scan() && !f.exit; f.prompt(m, ps...) {
		if h == STDIO && len(bio.Text()) == 0 {
			continue // 空行
		}

		// m.Cmdx(mdb.INSERT, SOURCE, kit.Keys(kit.MDB_HASH, h), mdb.LIST, kit.MDB_TEXT, bio.Text())
		f.count++

		if len(bio.Text()) == 0 {
			continue // 空行
		}

		if strings.HasSuffix(bio.Text(), "\\") {
			line += bio.Text()[:len(bio.Text())-1]
			ps = f.ps2
			continue // 续行
		}
		if line += bio.Text(); strings.Count(line, "`")%2 == 1 {
			line += "\n"
			ps = f.ps2
			continue // 多行
		}
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			line = ""
			continue // 注释
		}
		if ps = f.ps1; f.stdout == os.Stdout {
			// 清空格式
			f.printf(m, "\033[0m")
		}
		line = f.parse(m, line)
	}
	return f
}

func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	return f
}
func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	f.source, f.target = kit.Select(STDIO, arg, 0), m.Target()

	var r io.Reader
	switch m.Cap(ice.CTX_STREAM, f.source) {
	case STDIO: // 终端交互
		r, f.stdout = os.Stdin, os.Stdout

		m.Option(ice.MSG_OPTS, ice.MSG_USERNAME)
		m.Option(ice.MSG_USERNAME, ice.Info.UserName)
		m.Option(ice.MSG_USERROLE, aaa.ROOT)
		m.Option(ice.MSG_USERZONE, "boot")
		aaa.UserRoot(m)
	default:
		f.target = m.Source()

		if strings.HasPrefix(f.source, "/dev") {
			r, f.stdout = m.I, m.O
			break
		}

		buf := bytes.NewBuffer(make([]byte, 0, 4096))
		defer func() { m.Echo(buf.String()) }()

		if s := Script(m, f.source); s != nil {
			r, f.stdout = s, buf
			break
		}

		return true
	}

	if f.count = 1; f.source == STDIO {
		m.Option("_disable_log", "true")
		f.count = kit.Int(m.Conf(SOURCE, "hash.stdio.meta.count")) + 1
		f.scan(m, STDIO, "", r)
	} else {
		h := m.Cmdx(mdb.INSERT, SOURCE, "", mdb.HASH, kit.MDB_NAME, f.source)
		f.scan(m, h, "", r)
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

const (
	STDIO = "stdio"
)
const (
	SOURCE = "source"
	TARGET = "target"
	PROMPT = "prompt"
	RETURN = "return"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SOURCE: {Name: SOURCE, Help: "加载脚本", Value: kit.Dict()},
			PROMPT: {Name: PROMPT, Help: "命令提示", Value: kit.Data(
				"PS1", []interface{}{"\033[33;44m", "count", "[", "time", "]", "\033[5m", "target", "\033[0m", "\033[44m", ">", "\033[0m ", "\033[?25h", "\033[32m"},
				"PS2", []interface{}{"count", " ", "target", "> "},
			)},
		},
		Commands: map[string]*ice.Command{
			SOURCE: {Name: "source hash id auto", Help: "脚本解析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 && strings.HasSuffix(arg[0], ".shy") {
					m.Starts(strings.Replace(arg[0], ".", "_", -1), arg[0], arg[0:]...)
					return
				}

				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,hash,name,count")
					m.Cmdy(mdb.SELECT, SOURCE, "", mdb.HASH)
					m.Sort(kit.MDB_NAME)
					return
				}

				if arg[0] == STDIO {
					m.Option("_control", "_page")
				} else {
					m.Option("cache.limit", "-1")
				}
				m.Option(mdb.FIELDS, kit.Select("time,id,text", mdb.DETAIL, len(arg) > 1))
				m.Cmdy(mdb.SELECT, SOURCE, kit.Keys(kit.MDB_HASH, arg[0]), mdb.LIST, kit.MDB_ID, arg[1:])
				m.Sort(kit.MDB_ID)
			}},
			TARGET: {Name: "target name", Help: "当前模块", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Search(arg[0], func(p *ice.Context, s *ice.Context, key string) {
					f := m.Target().Server().(*Frame)
					f.target = s
					f.prompt(m)
				})
			}},
			PROMPT: {Name: "prompt arg...", Help: "命令提示", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				f := m.Target().Server().(*Frame)
				f.ps1 = arg
				f.prompt(m)
			}},
			RETURN: {Name: "return", Help: "结束脚本", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch cb := m.Optionv("ssh.return").(type) {
				case func():
					cb()
				}
			}},
		},
	})
}
