package code

func _pprof_show(m *ice.Message, zone string, id string) {
	favor := m.Conf(PPROF, kit.Keys(kit.MDB_META, web.FAVOR))

	m.Richs(PPROF, nil, zone, func(key string, val map[string]interface{}) {
		val = val[kit.MDB_META].(map[string]interface{})

		list := []string{}
		task.Put(val, func(task *task.Task) error {
			m.Sleep("1s")
			m.Grows(PPROF, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
				// 压测命令
				m.Log_EXPORT(kit.MDB_META, PPROF, kit.MDB_ZONE, zone, kit.MDB_VALUE, kit.Format(value))
				cmd := kit.Format(value[kit.MDB_TYPE])
				arg := kit.Format(value[kit.MDB_TEXT])
				res := m.Cmd(mdb.ENGINE, value[kit.MDB_TYPE], value[kit.MDB_NAME], value[kit.MDB_TEXT], value[kit.MDB_EXTRA]).Result()
				m.Cmd(web.FAVOR, favor, cmd, arg, res)
				list = append(list, cmd+": "+arg, res)
			})
			return nil
		})

		// 收藏程序
		msg := m.Cmd(web.CACHE, web.CATCH, kit.MIME_FILE, kit.Format(val[BINNARY]))
		bin := msg.Append(kit.MDB_TEXT)
		m.Cmd(web.FAVOR, favor, kit.MIME_FILE, bin, val[BINNARY])

		// 性能分析
		msg = m.Cmd(web.SPIDE, "self", web.CACHE, http.MethodGet, kit.Select("/code/pprof/profile", val[SERVICE]), "seconds", kit.Select("5", kit.Format(val[SECONDS])))
		m.Cmd(web.FAVOR, favor, PPROF, msg.Append(kit.MDB_TEXT), kit.Keys(zone, "pd.gz"))

		// 结果摘要
		cmd := kit.Simple(m.Confv(PPROF, "meta.pprof"), "-text", val[BINNARY], msg.Append(kit.MDB_TEXT))
		res := strings.Split(m.Cmdx(cli.SYSTEM, cmd), "\n")
		if len(res) > 20 {
			res = res[:20]
		}
		m.Cmd(web.FAVOR, favor, web.TYPE_SHELL, strings.Join(cmd, " "), strings.Join(res, "\n"))
		list = append(list, web.TYPE_SHELL+": "+strings.Join(cmd, " "), strings.Join(res, "\n"))

		// 结果展示
		u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
		p := kit.Format("%s:%s", u.Hostname(), m.Cmdx(tcp.PORT, aaa.Right))
		m.Option(cli.CMD_STDOUT, "var/daemon/stdout")
		m.Option(cli.CMD_STDERR, "var/daemon/stderr")
		m.Cmd(cli.DAEMON, m.Confv(PPROF, "meta.pprof"), "-http="+p, val[BINNARY], msg.Append(kit.MDB_TEXT))

		url := u.Scheme + "://" + p + "/ui/top"
		m.Cmd(web.FAVOR, favor, web.SPIDE, url, msg.Append(kit.MDB_TEXT))
		m.Set(ice.MSG_RESULT).Echo(url).Echo(" \n").Echo("\n")
		m.Echo(strings.Join(list, "\n")).Echo("\n")

		m.Push("url", url)
		m.Push(PPROF, msg.Append(kit.MDB_TEXT))
		m.Push(SERVICE, strings.Replace(kit.Format(val[SERVICE]), "profile", "", -1))
		m.Push("bin", bin)
	})
}
