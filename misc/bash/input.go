package bash

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	kit "github.com/shylinux/toolkits"

	"strings"
	"unicode"
)

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			"/qrcode": {Name: "/qrcode", Help: "二维码", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(cli.SYSTEM, cli.QRCODE, kit.MDB_TEXT, m.Option(kit.MDB_TEXT))
			}},
			"/input": {Name: "/input", Help: "补全", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				list := kit.Split(m.Option("line"), m.Option("break"))
				word := list[kit.Int(m.Option("index"))]
				switch arg[0] {
				case "shy":
					m.Cmd("web.code.input.find", word).Table(func(index int, value map[string]string, head []string) {
						m.Echo(value["text"]).Echo(" ")
					})

				case "line":
					if strings.HasPrefix(m.Option("line"), "ice ") {
						list := kit.Split(m.Option("line"))
						switch list[1] {
						case "add":
							m.Cmd("web.code.input.push", list[2:])
							m.Option("line", list[4])
							m.Option("point", 0)
						default:
							m.Cmdy(list[1:])
							break
						}
					}

					line := []rune(m.Option("line"))
					if begin := kit.Int(m.Option("point")); begin < len(line) {
						m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
							m.Echo(string(line[:begin]))
							for i := begin; i < len(line); i++ {
								if i-begin < 3 && i < len(line)-1 {
									continue
								}
								// 编码转换
								for j := 0; j < 4; j++ {
									code := string(line[begin : i+1-j])
									list := append(m.Cmd("web.code.input.find", code).Appendv("text"), code)
									if len(list) > 1 {
										m.Echo(kit.Select(code, list[0]))
										m.Info("input %s->%s", code, list[0])
										i = i - j
										break
									}
								}
								// 输出编码
								begin = i + 1
							}
						})
						break
					}
					fallthrough
				case "end":
					m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
						last_text := kit.Format(kit.Value(value, "last.text"))
						last_list := kit.Simple(kit.Value(value, "last.list"))
						last_index := kit.Int(kit.Value(value, "last.index"))

						if last_text != "" && strings.HasSuffix(m.Option("line"), last_text) {
							// 补全记录
							index := last_index + 1
							text := last_list[index%len(last_list)]
							kit.Value(value, "last.index", index)
							kit.Value(value, "last.text", text)
							m.Echo(strings.TrimSuffix(m.Option("line"), last_text) + text)
							m.Info("%d %v", index, last_list)
							return
						}

						line := []rune(m.Option("line"))
						for i := len(line); i >= 0; i-- {
							if i > 0 && len(line)-i < 4 && unicode.IsLower(line[i-1]) {
								continue
							}

							// 编码转换
							code := string(line[i:])
							list := append(m.Cmd("web.code.input.find", code).Appendv("text"), code)
							value["last"] = kit.Dict("code", code, "text", list[0], "list", list, "index", 0)

							// 输出编码
							m.Echo(string(line[:i]))
							m.Echo(kit.Select(code, list[0]))
							m.Info("input %s->%s", code, list[0])
							break
						}
					})
				}
				m.Info("trans: %v", m.Result())
			}},
		},
	})
}
