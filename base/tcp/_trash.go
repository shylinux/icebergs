package tcp

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"

	"bufio"
	"net"
	"net/url"
	"strings"
)

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			"server": {Name: "server [tcp4|tcp6|udp4|udp6] addr", Help: "server", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				proto := "tcp4"
				switch arg[0] {
				case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6", "ip", "ip4", "ip6":
					proto, arg = arg[0], arg[1:]
				}

				if l, e := net.Listen(proto, arg[0]); m.Assert(e) {
					m.Gos(m, func(m *ice.Message) {
						// 启动服务
						m.Logs(ice.LOG_LISTEN, "addr", l.Addr())
						for {
							if c, e := l.Accept(); m.Assert(e) {
								m.Gos(m.Spawns(), func(msg *ice.Message) {
									// 建立连接
									msg.Logs(ice.LOG_ACCEPT, "addr", c.RemoteAddr())
									msg.Option(ice.MSG_USERADDR, c.RemoteAddr())
									msg.Option(ice.MSG_USERNAME, "")
									msg.Option(ice.MSG_USERROLE, "")

									switch msg.Cmdx("check", c.RemoteAddr().String()) {
									case "local":
										// 本机用户
										msg.Option(ice.MSG_USERNAME, msg.Conf(cli.RUNTIME, "boot.username"))
										msg.Option(ice.MSG_USERROLE, msg.Cmdx(aaa.ROLE, "check", msg.Option(ice.MSG_USERNAME)))
										msg.Logs(ice.LOG_AUTH, "name", msg.Option(ice.MSG_USERNAME), "role", msg.Option(ice.MSG_USERROLE))
									}

									cmds := []string{}
									buf := bufio.NewWriter(c)
									for bio := bufio.NewScanner(c); bio.Scan(); {
										text := bio.Text()
										msg.Logs("scan", "len", len(text), "text", text)

										if len(text) == 0 {
											if len(cmds) > 0 {
												msg.Cmd(aaa.ROLE, "right")
												// 执行命令
												res := msg.Cmd(cmds)

												// 返回结果
												for _, str := range res.Resultv() {
													buf.WriteString("result:")
													buf.WriteString(url.QueryEscape(str))
													buf.WriteString("\n")
												}
												buf.WriteString("\n")
												buf.Flush()

												cmds = cmds[:0]
											}
											continue
										}

										// 解析请求
										line := strings.SplitN(bio.Text(), ":", 2)
										line[0], e = url.QueryUnescape(line[0])
										m.Assert(e)
										line[1], e = url.QueryUnescape(line[1])
										m.Assert(e)
										switch line[0] {
										case "cmds", ice.MSG_DETAIL:
											cmds = append(cmds, line[1])
										default:
											msg.Option(line[0], line[1])
										}
									}
									msg.Logs(ice.LOG_FINISH, "addr", c.RemoteAddr())
								})
							}
						}
						m.Logs(ice.LOG_FINISH, "addr", l.Addr())
					})
				}
			}},
		},
	}, nil)
}
