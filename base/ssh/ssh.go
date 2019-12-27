package ssh

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
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
	m.Cap(ice.CTX_STREAM, "stdio")

	prompt := "%d[15:04:05]%s> "
	target := m.Target()
	count := 0

	bio := bufio.NewScanner(f.in)
	fmt.Fprintf(f.out, m.Time(prompt, count, target.Name))
	for bio.Scan() {
		ls := kit.Split(bio.Text())
		m.Log("info", "stdin input %v", ls)

		msg := m.Spawns(target)
		if msg.Cmdy(ls); !msg.Hand {
			msg = msg.Set("result").Cmdy(ice.CLI_SYSTEM, ls)
		}
		res := msg.Result()
		if res == "" {
			msg.Table()
			res = msg.Result()
		}

		fmt.Fprint(f.out, res)
		if !strings.HasSuffix(res, "\n") {
			fmt.Fprint(f.out, "\n")
		}
		fmt.Fprintf(f.out, m.Time(prompt, count, target.Name))
		msg.Cost("stdin %v", ls)
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
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo("hello %s world", c.Name)
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			f := m.Target().Server().(*Frame)
			f.in.Close()
			m.Done()
		}},
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
