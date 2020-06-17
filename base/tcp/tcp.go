package tcp

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/toolkits"

	"bufio"
	"net"
	"net/url"
	"strings"
)

type Frame struct{}

const (
	GETPORT = "getport"
)

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

	msg := m.Spawn()
	_ip_list(msg, "")
	msg.Table(func(index int, value map[string]string, head []string) {
		if value["ip"] == ip {
			ok = true
		}
	})
	return ok
}
func _tcp_port(m *ice.Message) {
	current := kit.Int(m.Conf(GETPORT, "meta.current"))
	end := kit.Int(m.Conf(GETPORT, "meta.end"))
	if current >= end {
		current = kit.Int(m.Conf(GETPORT, "meta.begin"))
	}
	for i := current; i < end; i++ {
		if m.Cmd(cli.SYSTEM, "lsof", "-i", kit.Format(":%d", i)).Append(cli.CMD_CODE) != "0" {
			m.Conf(GETPORT, "meta.current", i)
			m.Log_CREATE(GETPORT, i)
			m.Echo("%d", i)
			break
		}
	}
}

func IPIsLocal(m *ice.Message, ip string) bool {
	return _ip_islocal(m, ip)
}

var Index = &ice.Context{Name: "tcp", Help: "通信模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		GETPORT: &ice.Config{Name: "getport", Help: "分配端口", Value: kit.Data(
			"begin", 10000, "current", 10000, "end", 20000,
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save(GETPORT) }},

		"ifconfig": {Name: "ifconfig [name]", Help: "网络配置", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_ip_list(m, "")
		}},
		GETPORT: {Name: "getport", Help: "分配端口", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_tcp_port(m)
		}},

		"ip": {Name: "ifconfig [name]", Help: "网络配置", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if addr, e := net.InterfaceAddrs(); m.Assert(e) {
				for _, v := range addr {
					m.Info("%v", v)
				}
			}
		}},
		"netstat": {Name: "netstat [name]", Help: "网络配置", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(ice.CLI_SYSTEM, "netstat", "-lanp")
		}},

		"check": {Name: "check addr", Help: "server", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if strings.Contains(arg[0], ".") {
				switch list := strings.Split(arg[0], ":"); list[0] {
				case "127.0.0.1":
					m.Echo("local")
				}
			} else {
				switch list := strings.Split(arg[0], "]:"); strings.TrimPrefix(list[0], "[") {
				case "::1":
					m.Echo("local")
				}
			}
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
									msg.Option(ice.MSG_USERNAME, msg.Conf(ice.CLI_RUNTIME, "boot.username"))
									msg.Option(ice.MSG_USERROLE, msg.Cmdx(ice.AAA_ROLE, "check", msg.Option(ice.MSG_USERNAME)))
									msg.Logs(ice.LOG_AUTH, "name", msg.Option(ice.MSG_USERNAME), "role", msg.Option(ice.MSG_USERROLE))
								}

								cmds := []string{}
								buf := bufio.NewWriter(c)
								for bio := bufio.NewScanner(c); bio.Scan(); {
									text := bio.Text()
									msg.Logs("scan", "len", len(text), "text", text)

									if len(text) == 0 {
										if len(cmds) > 0 {
											msg.Cmd(ice.AAA_ROLE, "right")
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

func init() { ice.Index.Register(Index, nil) }
