package tcp

import (
	"os"
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	kit "github.com/shylinux/toolkits"

	"bufio"
	"net"
	"net/url"
	"strings"
)

func _port_list(m *ice.Message) string {
	return ""
}
func _port_get(m *ice.Message, begin string) string {
	current := kit.Int(kit.Select(m.Conf(PORT, "meta.current"), begin))
	end := kit.Int(m.Conf(PORT, "meta.end"))
	if current >= end {
		current = kit.Int(m.Conf(PORT, "meta.begin"))
	}
	for i := current; i < end; i++ {
		if c, e := net.Dial("tcp", kit.Format(":%d", i)); e == nil {
			m.Info("port exists %v", i)
			defer c.Close()
			continue
		}
		m.Conf(PORT, "meta.current", i)
		m.Log_SELECT(PORT, i)
		return kit.Format("%d", i)
	}
	return ""
}

func _ip_list(m *ice.Message, ifname string) {
	if ifs, e := net.Interfaces(); m.Assert(e) {
		for _, v := range ifs {
			if ifname != "" && !strings.Contains(v.Name, ifname) {
				continue
			}
			if ips, e := v.Addrs(); m.Assert(e) {
				for _, x := range ips {
					ip := strings.Split(x.String(), "/")
					if strings.Contains(ip[0], ":") || len(ip) == 0 {
						continue
					}
					if len(v.HardwareAddr.String()) == 0 {
						continue
					}

					m.Push("index", v.Index)
					m.Push("name", v.Name)
					m.Push("ip", ip[0])
					m.Push("mask", ip[1])
					m.Push("hard", v.HardwareAddr.String())
				}
			}
		}
	}
}
func _ip_islocal(m *ice.Message, ip string) (ok bool) {
	if ip == "::1" || strings.HasPrefix(ip, "127.") {
		return true
	}

	if m.Richs(IP, kit.Keys("meta.white"), ip, nil) == nil {
		return false
	}
	m.Log_AUTH(aaa.White, ip)
	return true
}
func IPIsLocal(m *ice.Message, ip string) bool {
	return _ip_islocal(m, ip)
}

const (
	IP   = "ip"
	PORT = "port"
)

var Index = &ice.Context{Name: "tcp", Help: "通信模块",
	Configs: map[string]*ice.Config{
		PORT: {Name: "port", Help: "端口", Value: kit.Data(
			"begin", 10000, "current", 10000, "end", 20000,
		)},
		IP: {Name: "ip", Help: "地址", Value: kit.Data(
			"black", kit.Dict(),
			"white", kit.Data(kit.MDB_SHORT, kit.MDB_TEXT),
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Cmd(IP).Table(func(index int, value map[string]string, head []string) {
				m.Cmd(IP, aaa.White, value[IP])
			})

		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save(PORT) }},

		IP: {Name: "ip", Help: "地址", Action: map[string]*ice.Action{
			aaa.White: {Name: "show ip", Help: "白名单", Hand: func(m *ice.Message, arg ...string) {
				m.Rich(IP, kit.Keys("meta.white"), kit.Dict(
					kit.MDB_NAME, "",
					kit.MDB_TEXT, arg[0],
				))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_ip_list(m, "")
		}},
		PORT: {Name: "port", Help: "端口", Action: map[string]*ice.Action{
			"get": {Name: "get", Help: "分配端口", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(_port_get(m, ""))
			}},
			"select": {Name: "select [begin]", Help: "分配端口", Hand: func(m *ice.Message, arg ...string) {
				port, p := kit.Select("", arg, 0), ""
				for i := 0; i < 10; i++ {
					port = _port_get(m, port)
					p = path.Join(m.Conf(cli.DAEMON, kit.META_PATH), port)
					if _, e := os.Stat(p); e != nil && os.IsNotExist(e) {
						break
					}
					port = kit.Format(kit.Int(port) + 1)
				}
				os.MkdirAll(p, ice.MOD_DIR)
				m.Echo(port)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_port_list(m)
		}},

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
}

func init() { ice.Index.Register(Index, nil, IP, PORT) }
