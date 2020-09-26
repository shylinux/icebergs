package ssh

func (f *Frame) history(m *ice.Message, line string) string {
	favor := m.Conf(SOURCE, kit.Keys(kit.MDB_META, web.FAVOR))
	if strings.HasPrefix(strings.TrimSpace(line), "!!") {
		if len(line) == 2 {
			line = m.Cmd(web.FAVOR, favor).Append(kit.MDB_TEXT)
		}
	} else if strings.HasPrefix(strings.TrimSpace(line), "!") {
		if len(line) == 1 {
			// 历史记录
			msg := m.Cmd(web.FAVOR, favor)
			msg.Sort(kit.MDB_ID)
			msg.Appendv(ice.MSG_APPEND, kit.MDB_TIME, kit.MDB_ID, kit.MDB_TEXT)
			f.printf(m, msg.Table().Result())
			return ""
		}
		if i, e := strconv.Atoi(line[1:]); e == nil {
			// 历史命令
			line = kit.Format(kit.Value(m.Cmd(web.FAVOR, favor, i).Optionv("value"), kit.MDB_TEXT))
		} else {
			f.printf(m, m.Cmd("history", "search", line[1:]).Table().Result())
			return ""
		}
	} else if strings.TrimSpace(line) != "" && f.source == STDIO {
		// 记录历史
		m.Cmd(web.FAVOR, favor, "cmd", f.source, line)
	}
	return line
}

const (
	REMOTE = "remote"
	QRCODE = "qrcode"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			REMOTE: {Name: "remote", Help: "远程连接", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{

			QRCODE: {Name: "qrcode arg...", Help: "命令提示", Action: map[string]*ice.Action{
				"json": {Name: "json [key val]...", Help: "json", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.PYTHON, QRCODE, kit.Format(kit.Parse(nil, "", arg...)))
					m.Render(ice.RENDER_RESULT)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(cli.PYTHON, QRCODE, strings.Join(arg, ""))
				m.Render(ice.RENDER_RESULT)
			}},
			REMOTE: {Name: "remote user remote port local", Help: "远程连接", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				key := m.Rich(REMOTE, nil, kit.Dict(
					"user", arg[0], "remote", arg[1], "port", arg[2], "local", arg[3],
				))
				m.Echo(key)
				m.Info(key)

				m.Gos(m, func(m *ice.Message) {
					for {
						m.Cmd(cli.SYSTEM, "ssh", "-CNR", kit.Format("%s:%s:22", arg[2], kit.Select("localhost", arg, 3)),
							kit.Format("%s@%s", arg[0], arg[1]))
						m.Info("reconnect after 10s")
						time.Sleep(time.Second * 10)
					}
				})
			}},
		},
	}, nil)
}
