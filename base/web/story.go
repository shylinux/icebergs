package web

import (
	"os"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _story_list(m *ice.Message, name string, key string) {
	if name == "" {
		m.Richs(STORY, HEAD, mdb.FOREACH, func(key string, value ice.Map) {
			m.Push(key, value, []string{mdb.TIME, mdb.COUNT, STORY})
		})
		m.SortTimeR(mdb.TIME)
		return
	}
	if key == "" {
		_story_history(m, name)
		return
	}

	m.Richs(STORY, nil, key, func(key string, value ice.Map) {
		m.Push(mdb.DETAIL, value)
	})

}
func _story_index(m *ice.Message, name string, withdata bool) {
	m.Richs(STORY, HEAD, name, func(key string, value ice.Map) {
		// 查询索引
		m.Push(HEAD, key)
		name = kit.Format(value[LIST])
	})

	m.Richs(STORY, nil, name, func(key string, value ice.Map) {
		// 查询节点
		m.Push(LIST, key)
		m.Push(key, value, []string{SCENE, STORY})
		name = kit.Format(value[DATA])
	})

	m.Richs(CACHE, nil, name, func(key string, value ice.Map) {
		// 查询数据
		m.Push(DATA, key)
		m.Push(key, value, []string{mdb.TEXT, nfs.FILE, nfs.SIZE, mdb.TIME, mdb.NAME, mdb.TYPE})
		if withdata {
			if value[nfs.FILE] == "" {
				m.Echo("%s", kit.Format(value[mdb.TEXT]))
			} else {
				m.Echo("%s", m.Cmdx(nfs.CAT, value[nfs.FILE]))
			}
		}
	})
}
func _story_history(m *ice.Message, name string) {
	// 历史记录
	list := m.Cmd(STORY, INDEX, name).Append(LIST)
	for i := 0; i < kit.Int(kit.Select("30", m.Option(ice.CACHE_LIMIT))) && list != ""; i++ {
		m.Richs(STORY, nil, list, func(key string, value ice.Map) {
			// 直连节点
			m.Push(key, value, []string{mdb.TIME, mdb.KEY, mdb.COUNT, SCENE, STORY})
			m.Richs(CACHE, nil, value[DATA], func(key string, value ice.Map) {
				m.Push(DRAMA, value[mdb.TEXT])
				m.Push(DATA, key)
			})

			kit.Fetch(value[LIST], func(key string, val string) {
				m.Richs(STORY, nil, val, func(key string, value ice.Map) {
					// 复合节点
					m.Push(key, value, []string{mdb.TIME, mdb.KEY, mdb.COUNT, SCENE, STORY})
					m.Richs(CACHE, nil, value[DATA], func(key string, value ice.Map) {
						m.Push(DRAMA, value[mdb.TEXT])
						m.Push(DATA, key)
					})
				})
			})

			// 切换节点
			list = kit.Format(value[PREV])
		})
	}
}
func _story_write(m *ice.Message, scene, name, text string, arg ...string) {
	if len(arg) < 1 || text == "" || m.Richs(CACHE, nil, text, func(key string, value ice.Map) { text = key }) == nil {
		// 添加缓存
		m.Cmdy(CACHE, CATCH, scene, name, text, arg)
		scene, name, text = m.Append(mdb.TYPE), m.Append(mdb.NAME), m.Append(DATA)
	}

	// 查询索引
	head, prev, value, count := "", "", kit.Dict(), 0
	m.Richs(STORY, HEAD, name, func(key string, val ice.Map) {
		head, prev, value, count = key, kit.Format(val[LIST]), val, kit.Int(val[mdb.COUNT])
		m.Logs("info", HEAD, head, PREV, prev, mdb.COUNT, count)
	})

	if last := m.Richs(STORY, nil, prev, nil); prev != "" && last != nil && last[DATA] == text {
		// 重复提交
		m.Push(prev, last, []string{mdb.TIME, mdb.COUNT, mdb.KEY})
		m.Logs("info", "file", "exists")
		m.Echo(prev)
		return
	}

	// 添加节点
	list := m.Rich(STORY, nil, kit.Dict(
		SCENE, scene, STORY, name, mdb.COUNT, count+1, DATA, text, PREV, prev,
	))
	m.Log_CREATE(STORY, list, mdb.TYPE, scene, mdb.NAME, name)
	m.Push(mdb.COUNT, count+1)
	m.Push(mdb.KEY, list)

	if head == "" {
		// 添加索引
		m.Rich(STORY, HEAD, kit.Dict(SCENE, scene, STORY, name, mdb.COUNT, count+1, LIST, list))
	} else {
		// 更新索引
		value[mdb.COUNT] = count + 1
		value[mdb.TIME] = m.Time()
		value[LIST] = list
	}
	m.Echo(list)
}
func _story_catch(m *ice.Message, scene, name string, arg ...string) {
	if last := m.Richs(STORY, HEAD, name, nil); last != nil {
		if t, e := time.ParseInLocation(ice.MOD_TIME, kit.Format(last[mdb.TIME]), time.Local); e == nil {
			if s, e := os.Stat(name); e == nil && s.ModTime().Before(t) {
				m.Push(name, last, []string{mdb.TIME, mdb.COUNT, mdb.KEY})
				m.Logs("info", "file", "exists")
				m.Echo("%s", last[LIST])
				// 重复提交
				return
			}
		}
	}
	_story_write(m, scene, name, "", arg...)
}
func _story_watch(m *ice.Message, key, file string) {
	_story_index(m, key, false)
	_cache_watch(m, m.Append(DATA), file)
}

const (
	HEAD = "head"
	LIST = "list"
	PREV = "prev"
	DATA = "data"

	HISTORY = "history"

	PULL   = "pull"
	PUSH   = "push"
	COMMIT = "commit"
)
const (
	SCENE = "scene"
	DRAMA = "drama"
)
const STORY = "story"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		STORY: {Name: "story", Help: "故事会", Value: kit.Dict(
			mdb.META, kit.Dict(mdb.SHORT, DATA),
			HEAD, kit.Data(mdb.SHORT, STORY),
		)},
	}, Commands: map[string]*ice.Command{
		STORY: {Name: "story story auto", Help: "故事会", Action: map[string]*ice.Action{
			WRITE: {Name: "write type name text arg...", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_story_write(m, arg[0], arg[1], arg[2], arg[3:]...)
			}},
			CATCH: {Name: "catch type name arg...", Help: "捕捉", Hand: func(m *ice.Message, arg ...string) {
				_story_catch(m, arg[0], arg[1], arg[2:]...)
			}},
			WATCH: {Name: "watch key name", Help: "释放", Hand: func(m *ice.Message, arg ...string) {
				_story_watch(m, arg[0], arg[1])
			}},
			INDEX: {Name: "index key", Help: "索引", Hand: func(m *ice.Message, arg ...string) {
				_story_index(m, arg[0], false)
			}},
			HISTORY: {Name: "history name", Help: "历史", Hand: func(m *ice.Message, arg ...string) {
				_story_history(m, arg[0])
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			_story_list(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
		}},
		"/story/": {Name: "/story/", Help: "故事会", Hand: func(m *ice.Message, arg ...string) {
			switch arg[0] {
			case PULL:
				list := m.Cmd(STORY, INDEX, m.Option("begin")).Append("list")
				for i := 0; i < 10 && list != "" && list != m.Option("end"); i++ {
					if m.Richs(STORY, nil, list, func(key string, value ice.Map) {
						// 节点信息
						m.Push("list", key)
						m.Push("node", kit.Format(value))
						m.Push("data", value["data"])
						m.Push("save", kit.Format(m.Richs(CACHE, nil, value["data"], nil)))
						list = kit.Format(value["prev"])
					}) == nil {
						break
					}
				}
				m.Log(ice.LOG_EXPORT, "%s %s", m.Option("begin"), m.FormatSize())

			case PUSH:
				if m.Richs(CACHE, nil, m.Option("data"), nil) == nil {
					// 导入缓存
					m.Log(ice.LOG_IMPORT, "%v: %v", m.Option("data"), m.Option("save"))
					m.Conf(CACHE, kit.Keys("hash", m.Option("data")), kit.UnMarshal(m.Option("save")))
				}

				node := kit.UnMarshal(m.Option("node")).(ice.Map)
				if m.Richs(STORY, nil, m.Option("list"), nil) == nil {
					// 导入节点
					m.Log(ice.LOG_IMPORT, "%v: %v", m.Option("list"), m.Option("node"))
					m.Conf(STORY, kit.Keys("hash", m.Option("list")), node)
				}

				if head := m.Richs(STORY, "head", m.Option("story"), nil); head == nil {
					// 自动创建
					h := m.Rich(STORY, "head", kit.Dict(
						"scene", node["scene"], "story", m.Option("story"),
						"count", node["count"], "list", m.Option("list"),
					))
					m.Log(ice.LOG_CREATE, "%v: %v", h, m.Option("story"))
				} else if head["list"] == kit.Format(node["prev"]) || head["list"] == kit.Format(node["pull"]) {
					// 快速合并
					head["list"] = m.Option("list")
					head["count"] = node["count"]
					head["time"] = node["time"]
				} else {
					// 推送失败
				}

			case UPLOAD:
				// 上传数据
				m.Cmdy(CACHE, "upload")

			case DOWNLOAD:
				// 下载数据
				m.Cmdy(STORY, INDEX, arg[1])
				m.Render(kit.Select(ice.RENDER_DOWNLOAD, ice.RENDER_RESULT, m.Append("file") == ""), m.Append("text"))
			}
		}},
	}})
}

func _story_pull(m *ice.Message, arg ...string) {
	// 起止节点
	prev, begin, end := "", arg[3], ""
	repos := kit.Keys("remote", arg[2], arg[3])
	m.Richs(STORY, "head", arg[1], func(key string, val ice.Map) {
		end = kit.Format(kit.Value(val, kit.Keys(repos, "pull", "list")))
		prev = kit.Format(val["list"])
	})

	pull := end
	var first ice.Map
	for begin != "" && begin != end {
		if m.Cmd(SPIDE, arg[2], "msg", "/story/pull", "begin", begin, "end", end).Table(func(index int, value map[string]string, head []string) {
			if m.Richs(CACHE, nil, value["data"], nil) == nil {
				m.Log(ice.LOG_IMPORT, "%v: %v", value["data"], value["save"])
				if node := kit.UnMarshal(value["save"]); kit.Format(kit.Value(node, "file")) != "" {
					// 下载文件
					m.Cmd(SPIDE, arg[2], "cache", "GET", "/story/download/"+value["data"])
				} else {
					// 导入缓存
					m.Conf(CACHE, kit.Keys("hash", value["data"]), node)
				}
			}

			node := kit.UnMarshal(value["node"]).(ice.Map)
			if m.Richs(STORY, nil, value["list"], nil) == nil {
				// 导入节点
				m.Log(ice.LOG_IMPORT, "%v: %v", value["list"], value["node"])
				m.Conf(STORY, kit.Keys("hash", value["list"]), node)
			}

			if first == nil {
				if m.Richs(STORY, "head", arg[1], nil) == nil {
					// 自动创建
					h := m.Rich(STORY, "head", kit.Dict(
						"scene", node["scene"], "story", arg[1],
						"count", node["count"], "list", value["list"],
					))
					m.Log(ice.LOG_CREATE, "%v: %v", h, node["story"])
				}

				pull, first = kit.Format(value["list"]), node
				m.Richs(STORY, "head", arg[1], func(key string, val ice.Map) {
					prev = kit.Format(val["list"])
					if kit.Int(node["count"]) > kit.Int(kit.Value(val, kit.Keys(repos, "pull", "count"))) {
						// 更新分支
						m.Log(ice.LOG_IMPORT, "%v: %v", arg[2], pull)
						kit.Value(val, kit.Keys(repos, "pull"), kit.Dict(
							"head", arg[3], "time", node["time"], "list", pull, "count", node["count"],
						))
					}
				})
			}

			if prev == kit.Format(node["prev"]) || prev == kit.Format(node["push"]) {
				// 快速合并
				m.Log(ice.LOG_IMPORT, "%v: %v", pull, arg[2])
				m.Richs(STORY, "head", arg[1], func(key string, val ice.Map) {
					val["count"] = first["count"]
					val["time"] = first["time"]
					val["list"] = pull
				})
				prev = pull
			}

			begin = kit.Format(node["prev"])
		}).Appendv("list") == nil {
			break
		}
	}

}
func _story_push(m *ice.Message, arg ...string) {
	// 更新分支
	m.Cmdx(STORY, "pull", arg[1:])

	repos := kit.Keys("remote", arg[2], arg[3])
	// 查询索引
	prev, pull, some, list := "", "", "", ""
	m.Richs(STORY, "head", arg[1], func(key string, val ice.Map) {
		prev = kit.Format(val["list"])
		pull = kit.Format(kit.Value(val, kit.Keys(repos, "pull", "list")))
		for some = pull; prev != some && some != ""; {
			local := m.Richs(STORY, nil, prev, nil)
			remote := m.Richs(STORY, nil, some, nil)
			if diff := kit.Time(kit.Format(remote["time"])) - kit.Time(kit.Format(local["time"])); diff > 0 {
				some = kit.Format(remote["prev"])
			} else if diff < 0 {
				prev = kit.Format(local["prev"])
			}
		}

		if prev = kit.Format(val["list"]); prev == pull {
			// 相同节点
			return
		}

		if some != pull {
			// 合并节点
			local := m.Richs(STORY, nil, prev, nil)
			remote := m.Richs(STORY, nil, pull, nil)
			list = m.Rich(STORY, nil, kit.Dict(
				"scene", val["scene"], "story", val["story"], "count", kit.Int(remote["count"])+1,
				"data", local["data"], "prev", pull, "push", prev,
			))
			m.Log(ice.LOG_CREATE, "merge: %s %s->%s", list, prev, pull)
			val["list"] = list
			prev = list
			val["count"] = kit.Int(remote["count"]) + 1
		}

		// 查询节点
		nodes := []string{}
		for list = prev; list != some; {
			m.Richs(STORY, nil, list, func(key string, value ice.Map) {
				nodes, list = append(nodes, list), kit.Format(value["prev"])
			})
		}

		for _, v := range kit.Revert(nodes) {
			m.Richs(STORY, nil, v, func(list string, node ice.Map) {
				m.Richs(CACHE, nil, node["data"], func(data string, save ice.Map) {
					if kit.Format(save["file"]) != "" {
						// 推送缓存
						m.Cmd(SPIDE, arg[2], "/story/upload",
							"part", "upload", "@"+kit.Format(save["file"]),
						)
					}

					// 推送节点
					m.Log(ice.LOG_EXPORT, "%s: %s", v, kit.Format(node))
					m.Cmd(SPIDE, arg[2], "/story/push",
						"story", arg[3], "list", v, "node", kit.Format(node),
						"data", node["data"], "save", kit.Format(save),
					)
				})
			})
		}
	})

	// 更新分支
	m.Cmd(STORY, "pull", arg[1:])

}
func _story_commit(m *ice.Message, arg ...string) {
	// 查询索引
	head, prev, value, count := "", "", ice.Map{}, 0
	m.Richs(STORY, "head", arg[1], func(key string, val ice.Map) {
		head, prev, value, count = key, kit.Format(val["list"]), val, kit.Int(val["count"])
		m.Log("info", "head: %v prev: %v count: %v", head, prev, count)
	})

	// 提交信息
	arg[2] = m.Cmdx(STORY, "add", "submit", arg[2], "hostname,username")

	// 节点信息
	menu := map[string]string{}
	for i := 3; i < len(arg); i++ {
		menu[arg[i]] = m.Cmdx(STORY, INDEX, arg[i])
	}

	// 添加节点
	list := m.Rich(STORY, nil, kit.Dict(
		"scene", "commit", "story", arg[1], "count", count+1, "data", arg[2], "list", menu, "prev", prev,
	))
	m.Log(ice.LOG_CREATE, "commit: %s %s: %s", list, arg[1], arg[2])
	m.Push("list", list)

	if head == "" {
		// 添加索引
		m.Rich(STORY, "head", kit.Dict("scene", "commit", "story", arg[1], "count", count+1, "list", list))
	} else {
		// 更新索引
		value["count"] = count + 1
		value["time"] = m.Time()
		value["list"] = list
	}
	m.Echo(list)

}
