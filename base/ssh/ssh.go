package ssh

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"bufio"
	"fmt"
	"io"
	"os"
)

type Frame struct {
	in  io.ReadCloser
	out io.WriteCloser
}

func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	return f
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	f.in = os.Stdin
	f.out = os.Stdout
	m.Cap("stream", "stdio")

	prompt := "%d[15:04:05]%s> "
	target := m.Target()
	count := 0

	bio := bufio.NewScanner(f.in)
	fmt.Fprintf(f.out, m.Time(prompt, count, target.Name))
	for bio.Scan() {
		ls := kit.Split(bio.Text())
		m.Log("info", "input %v", ls)

		msg := m.Spawn(target)
		if msg.Cmdy(ls); !msg.Hand {
			msg = msg.Cmdy("cli.system", ls)
		}
		res := msg.Result()
		if res == "" {
			msg.Table()
			res = msg.Result()
		}

		fmt.Fprintf(f.out, res)
		fmt.Fprintf(f.out, m.Time(prompt, count, target.Name))
		count++
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

var Index = &ice.Context{Name: "ssh", Help: "终端模块",
	Caches:  map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{},
	Commands: map[string]*ice.Command{
		"_init": {Name: "_init", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo("hello %s world", c.Name)
		}},
		"_exit": {Name: "_exit", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			f := m.Target().Server().(*Frame)
			f.in.Close()
			m.Target().Done()
		}},
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
