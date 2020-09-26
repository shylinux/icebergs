package vim

/*
	m.Conf(web.FAVOR, "meta.render.vimrc", m.AddCmd(&ice.Command{Name: "render favor id", Help: "渲染引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		value := m.Optionv("value").(map[string]interface{})
		switch value["name"] {
		case "read", "write", "exec":
			p := path.Join(kit.Format(kit.Value(value, "extra.pwd")), kit.Format(kit.Value(value, "extra.buf")))
			if strings.HasPrefix(kit.Format(kit.Value(value, "extra.buf")), "/") {
				p = path.Join(kit.Format(kit.Value(value, "extra.buf")))
			}

			f, e := os.Open(p)
			m.Assert(e)
			defer f.Close()
			b, e := ioutil.ReadAll(f)
			m.Assert(e)
			m.Echo(string(b))
		default:
			m.Cmdy(cli.SYSTEM, "sed", "-n", fmt.Sprintf("/%s/,/^}$/p", value["text"]), kit.Value(value, "extra.buf"))
		}
	}}))

*/
