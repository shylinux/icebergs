package web

import (
	"github.com/gorilla/websocket"
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"text/template"
	"time"
)

const (
	MSG_MAPS = 1
)

type Frame struct {
	*http.Client
	*http.Server
	*http.ServeMux
	m    *ice.Message
	send map[string]*ice.Message
}

func Cookie(msg *ice.Message, sessid string) string {
	w := msg.Optionv("response").(http.ResponseWriter)
	expire := time.Now().Add(kit.Duration(msg.Conf(ice.AAA_SESS, ice.Meta("expire"))))
	msg.Log("cookie", "expire:%v sessid:%s", kit.Format(expire), sessid)
	http.SetCookie(w, &http.Cookie{Name: ice.WEB_SESS, Value: sessid, Path: "/", Expires: expire})
	return sessid
}
func (web *Frame) Login(msg *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	if msg.Options(ice.WEB_SESS) {
		// 会话认证
		sub := msg.Cmd(ice.AAA_SESS, "check", msg.Option(ice.WEB_SESS))
		msg.Info("role: %s user: %s", msg.Option(ice.MSG_USERROLE, sub.Append("userrole")),
			msg.Option(ice.MSG_USERNAME, sub.Append("username")))
	}

	if (!msg.Options(ice.MSG_SESSID) || !msg.Options(ice.MSG_USERNAME)) && msg.Option("user.ip") == "::1" {
		// 自动认证
		msg.Option(ice.MSG_USERNAME, msg.Conf(ice.CLI_RUNTIME, "boot.username"))
		msg.Option(ice.MSG_USERROLE, msg.Cmdx(ice.AAA_ROLE, "check", msg.Option(ice.MSG_USERNAME)))
		msg.Option(ice.MSG_SESSID, msg.Rich(ice.AAA_SESS, nil, kit.Dict(
			"username", msg.Option(ice.MSG_USERNAME), "userrole", msg.Option(ice.MSG_USERROLE),
		)))
		Cookie(msg, msg.Option(ice.MSG_SESSID))
		msg.Info("user: %s role: %s sess: %s", msg.Option(ice.MSG_USERNAME), msg.Option(ice.MSG_USERROLE), msg.Option(ice.MSG_SESSID))
	}

	if s, ok := msg.Target().Commands[ice.WEB_LOGIN]; ok {
		// 权限检查
		msg.Target().Run(msg, s, ice.WEB_LOGIN, kit.Simple(msg.Optionv("cmds"))...)
	}
	return msg.Option("url") != ""
}
func (web *Frame) HandleWSS(m *ice.Message, safe bool, c *websocket.Conn) bool {
	for {
		if t, b, e := c.ReadMessage(); m.Warn(e != nil, "space recv %d msg %v", t, e) {
			break
		} else {
			switch t {
			case MSG_MAPS:
				// 接收报文
				socket, msg := c, m.Spawn(b)
				source := kit.Simple(msg.Optionv(ice.MSG_SOURCE))
				target := kit.Simple(msg.Optionv(ice.MSG_TARGET))
				msg.Info("recv %v %v->%v %v", t, source, target, msg.Format("meta"))

				if len(target) == 0 {
					// 本地执行
					if msg.Optionv(ice.MSG_HANDLE, "true"); !msg.Warn(!safe, "no right") {
						msg = msg.Cmd()
					}
					kit.Revert(source)
					source, target = []string{source[0]}, source[1:]

				} else if s, ok := msg.Confv(ice.WEB_SPACE, kit.Keys("hash", target[0], "socket")).(*websocket.Conn); ok {
					// 转发报文
					msg.Info("space route")
					socket, source, target = s, append(source, target[0]), target[1:]

				} else if call, ok := web.send[msg.Option(ice.MSG_TARGET)]; len(target) == 1 && ok {
					// 接收响应
					msg.Info("space done")
					delete(web.send, msg.Option(ice.MSG_TARGET))
					call.Back(msg)
					break

				} else if msg.Warn(msg.Option("_handle") == "true", "space miss") {
					// 丢弃报文
					break

				} else {
					// 失败报文
					msg.Warn(true, "space error")
					source, target = []string{source[len(source)-1]}, kit.Revert(source)[1:]
				}

				// 发送报文
				msg.Optionv(ice.MSG_SOURCE, source)
				msg.Optionv(ice.MSG_TARGET, target)
				socket.WriteMessage(t, []byte(msg.Format("meta")))
				msg.Info("send %v %v->%v %v", t, source, target, msg.Format("meta"))
				msg.Log("cost", "%s: ", msg.Format("cost"))
				if msg.Detail() == "exit" {
					return true
				}
			}
		}
	}
	return false
}
func (web *Frame) HandleCGI(m *ice.Message, alias map[string]interface{}, which string) *template.Template {
	cgi := template.FuncMap{}

	tmpl := template.New(ice.WEB_TMPL)
	cb := func(k string, p []string, v *ice.Command) {
		cgi[k] = func(arg ...interface{}) (res interface{}) {
			m.TryCatch(m.Spawn(), true, func(msg *ice.Message) {
				msg.Target().Run(msg, v, k, kit.Simple(p, arg)...)

				buffer := bytes.NewBuffer([]byte{})
				m.Assert(tmpl.ExecuteTemplate(buffer, msg.Option(ice.WEB_TMPL), msg))
				res = string(buffer.Bytes())
			})
			return
		}
	}
	for k, v := range alias {
		list := kit.Simple(v)
		if v, ok := m.Target().Commands[list[0]]; ok {
			cb(k, list[1:], v)
		}
	}
	for k, v := range m.Target().Commands {
		if strings.HasPrefix(k, "/") || strings.HasPrefix(k, "_") {
			continue
		}
		cb(k, nil, v)
	}

	tmpl = tmpl.Funcs(cgi)
	// tmpl = template.Must(tmpl.ParseGlob(path.Join(m.Conf(ice.WEB_SERVE, ice.Meta("template", "path")), "/*.tmpl")))
	// tmpl = template.Must(tmpl.ParseGlob(path.Join(m.Conf(ice.WEB_SERVE, ice.Meta("template", "path")), m.Target().Name, "/*.tmpl")))
	tmpl = template.Must(tmpl.ParseFiles(which))
	m.Confm(ice.WEB_SERVE, ice.Meta("template", "list"), func(index int, value string) { tmpl = template.Must(tmpl.Parse(value)) })
	return tmpl
}
func (web *Frame) HandleCmd(m *ice.Message, key string, cmd *ice.Command) {
	web.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
		m.TryCatch(m.Spawns(), true, func(msg *ice.Message) {
			defer func() { msg.Log("cost", "%s: %s %v", msg.Format("cost"), r.URL.Path, msg.Optionv("cmds")) }()

			// 解析请求
			msg.Optionv("request", r)
			msg.Optionv("response", w)
			msg.Option("user.agent", r.Header.Get("User-Agent"))
			msg.Option("user.ip", r.Header.Get("user.ip"))
			msg.Option("referer", r.Header.Get("Referer"))
			msg.Option("accept", r.Header.Get("Accept"))
			msg.Option("method", r.Method)
			msg.Option("url", r.URL.Path)
			msg.Option(ice.WEB_SESS, "")

			// 请求环境
			for _, v := range r.Cookies() {
				if v.Value != "" {
					msg.Option(v.Name, v.Value)
				}
			}

			// 请求数据
			switch r.Header.Get("Content-Type") {
			case "application/json":
				var data interface{}
				if e := json.NewDecoder(r.Body).Decode(&data); !msg.Warn(e != nil, "%s", e) {
					msg.Optionv("content_data", data)
					msg.Info("%s", kit.Formats(data))
				}

				switch d := data.(type) {
				case map[string]interface{}:
					for k, v := range d {
						msg.Optionv(k, v)
					}
				}
			default:
				r.ParseMultipartForm(4096)
				if r.ParseForm(); len(r.PostForm) > 0 {
					for k, v := range r.PostForm {
						msg.Info("%s: %v", k, v)
					}
					msg.Info("")
				}
			}

			// 请求参数
			for k, v := range r.Form {
				msg.Optionv(k, v)
			}

			// 执行命令
			if web.Login(msg, w, r) && msg.Target().Run(msg, cmd, msg.Option("url"), kit.Simple(msg.Optionv("cmds"))...) != nil {
				// 输出响应
				switch msg.Append("_output") {
				case "void":
				case "file":
					msg.Info("_output: %s %s", msg.Append("_output"), msg.Append("file"))
					w.Header().Set("Content-Disposition", fmt.Sprintf("filename=%s", kit.Select("hi.txt", msg.Append("name"))))
					w.Header().Set("Content-Type", kit.Select("text/html", msg.Append("type")))
					http.ServeFile(w, r, msg.Append("file"))

				case "result":
					w.Header().Set("Content-Type", kit.Select("text/html", msg.Append("type")))
					fmt.Fprint(w, msg.Result())
				default:
					fmt.Fprint(w, msg.Formats("meta"))
				}
			}
		})
	})
}
func (web *Frame) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m := web.m

	index := r.Header.Get("index.module") == ""
	if index {
		// 解析地址
		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			r.Header.Set("user.ip", ip)
		} else if ip := r.Header.Get("X-Real-Ip"); ip != "" {
			r.Header.Set("user.ip", ip)
		} else if strings.HasPrefix(r.RemoteAddr, "[") {
			r.Header.Set("user.ip", strings.Split(r.RemoteAddr, "]")[0][1:])
		} else {
			r.Header.Set("user.ip", strings.Split(r.RemoteAddr, ":")[0])
		}
		m.Info("").Info("%s %s %s", r.Header.Get("user.ip"), r.Method, r.URL)

		// 解析地址
		r.Header.Set("index.module", "some")
		r.Header.Set("index.path", r.URL.Path)
		r.Header.Set("index.url", r.URL.String())
	}

	if index && kit.Right(m.Conf(ice.WEB_SERVE, "meta.logheaders")) {
		for k, v := range r.Header {
			m.Info("%s: %v", k, kit.Format(v))
		}
		m.Info(" ")
	}

	web.ServeMux.ServeHTTP(w, r)

	if index && kit.Right(m.Conf(ice.WEB_SERVE, "meta.logheaders")) {
		for k, v := range w.Header() {
			m.Info("%s: %v", k, kit.Format(v))
		}
		m.Info(" ")
	}
}

func (web *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (web *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	web.send = map[string]*ice.Message{}
	return web
}
func (web *Frame) Start(m *ice.Message, arg ...string) bool {
	m.Travel(func(p *ice.Context, s *ice.Context) {
		if w, ok := s.Server().(*Frame); ok {
			if w.ServeMux != nil {
				return
			}
			w.ServeMux = http.NewServeMux()

			// 静态路由
			msg := m.Spawns(s)
			m.Confm(ice.WEB_SERVE, ice.Meta("static"), func(key string, value string) {
				m.Log("route", "%s <- %s <- %s", s.Name, key, value)
				w.Handle(key, http.StripPrefix(key, http.FileServer(http.Dir(value))))
			})

			// 级联路由
			route := "/" + s.Name + "/"
			if n, ok := p.Server().(*Frame); ok && n.ServeMux != nil {
				msg.Log("route", "%s <= %s", p.Name, route)
				n.Handle(route, http.StripPrefix(path.Dir(route), w))
			}

			// 命令路由
			m.Travel(func(p *ice.Context, sub *ice.Context, k string, x *ice.Command) {
				if s == sub && k[0] == '/' {
					msg.Log("route", "%s <- %s", s.Name, k)
					w.HandleCmd(msg, k, x)
				}
			})
		}
	})

	// TODO simple
	m.Richs(ice.WEB_SPIDE, nil, arg[0], func(key string, value map[string]interface{}) {
		client := value["client"].(map[string]interface{})

		// 服务地址
		port := m.Cap(ice.CTX_STREAM, client["hostname"])
		m.Log("serve", "listen %s %s %v", arg[0], port, m.Conf(ice.CLI_RUNTIME, "node"))

		// 启动服务
		web.m, web.Server = m, &http.Server{Addr: port, Handler: web}
		m.Log("serve", "listen %s", web.Server.ListenAndServe())
	})
	return true
}
func (web *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

var Index = &ice.Context{Name: "web", Help: "网页模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ice.WEB_SPIDE: {Name: "spide", Help: "客户端", Value: kit.Data(kit.MDB_SHORT, "name")},
		ice.WEB_SERVE: {Name: "serve", Help: "服务器", Value: kit.Data(
			"static", map[string]interface{}{"/": "usr/volcanos/",
				"/static/volcanos/": "usr/volcanos/",
			},
			"template", map[string]interface{}{"path": "usr/template", "list": []interface{}{
				`{{define "raw"}}{{.Result}}{{end}}`,
			}},
			"logheaders", "false",
		)},
		ice.WEB_SPACE: {Name: "space", Help: "空间站", Value: kit.Data(kit.MDB_SHORT, "name",
			"redial.a", 3000, "redial.b", 1000, "redial.c", 10,
			"buffer.r", 4096, "buffer.w", 4096,
		)},
		ice.WEB_DREAM: {Name: "dream", Help: "梦想家", Value: kit.Data("path", "usr/local/work",
			"cmd", []interface{}{ice.CLI_SYSTEM, "sh", "ice.sh", "start", ice.WEB_SPACE, "connect"},
		)},
		ice.WEB_FAVOR: {Name: "favor", Help: "收藏夹", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
		ice.WEB_CACHE: {Name: "cache", Help: "缓存池", Value: kit.Data(kit.MDB_SHORT, "text", "path", "var/file", "store", "var/data", "limit", "30", "least", "10")},
		ice.WEB_STORY: {Name: "story", Help: "故事会", Value: kit.Dict(kit.MDB_META, kit.Dict(kit.MDB_SHORT, "data"), "head", kit.Data(kit.MDB_SHORT, "story"))},
		ice.WEB_SHARE: {Name: "share", Help: "共享链", Value: kit.Data("template", share_template)},
		ice.WEB_ROUTE: {Name: "route", Help: "路由表", Value: kit.Data()},
		ice.WEB_PROXY: {Name: "proxy", Help: "代理商", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.WEB_SPIDE, "add", "self", "http://:9020")
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Done() }},

		ice.WEB_SPIDE: {Name: "spide", Help: "客户端", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 爬虫列表
				m.Richs(ice.WEB_SPIDE, nil, "", func(key string, value map[string]interface{}) {
					m.Push("name", value["name"])
					m.Push(key, value["client"], []string{"method", "protocol", "hostname", "path", "file"})
				})
				return
			}
			switch arg[0] {
			case "add":
				// 添加爬虫
				if uri, e := url.Parse(arg[2]); e == nil && arg[2] != "" {
					dir, file := path.Split(uri.EscapedPath())
					m.Rich(ice.WEB_SPIDE, nil, kit.Dict(
						"name", arg[1],
						"cookie", kit.Dict(),
						"header", kit.Dict(),
						"client", kit.Dict(
							"logheaders", false,
							"timeout", "100s",
							"method", "GET",
							"protocol", uri.Scheme,
							"hostname", uri.Host,
							"path", dir,
							"file", file,
							"query", uri.RawQuery,
							"url", arg[2],
						),
					))
				}
			default:
				m.Richs(ice.WEB_SPIDE, nil, arg[0], func(key string, value map[string]interface{}) {
					client := value["client"].(map[string]interface{})
					method := kit.Select("GET", client["method"])
					uri := kit.MergeURL(kit.Format(client["url"]), arg[1:])

					body, ok := m.Optionv("body").(io.Reader)
					if !ok && method == "POST" {
						if ls := strings.Split(uri, "?"); len(ls) > 0 {
							uri, body = ls[0], bytes.NewReader([]byte(ls[1]))
						}
					}

					req, e := http.NewRequest(method, uri, body)
					m.Log("info", "%s %s", req.Method, req.URL)
					m.Assert(e)

					web := m.Target().Server().(*Frame)
					if web.Client == nil {
						web.Client = &http.Client{Timeout: kit.Duration(kit.Format(client["timeout"]))}
					}

					// 发送请求
					res, e := web.Client.Do(req)
					if m.Warn(e != nil, "%s", e) {
						return
					}

					if m.Warn(res.StatusCode != http.StatusOK, "%s", res.Status) {
						return
					}

					if buf, e := ioutil.ReadAll(res.Body); m.Assert(e) {
						m.Echo(string(buf))
					}
				})
			}
		}},
		ice.WEB_SERVE: {Name: "serve", Help: "服务器", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Conf(ice.CLI_RUNTIME, "node.name", m.Conf(ice.CLI_RUNTIME, "boot.hostname"))
			m.Conf(ice.CLI_RUNTIME, "node.type", kit.MIME_SERVER)

			m.Rich(ice.WEB_SPACE, nil, kit.Dict(
				"type", kit.MIME_MYSELF,
				"name", m.Conf(ice.CLI_RUNTIME, "boot.hostname"),
				"user", m.Conf(ice.CLI_RUNTIME, "boot.username"),
			))
			if _, e := os.Stat("usr/volcanos"); e != nil {
				m.Cmd("cli.system", "git", "clone", "https://github.com/shylinux/volcanos", "usr/volcanos")
			}
			m.Target().Start(m, kit.Select("self", arg, 0))
		}},
		ice.WEB_SPACE: {Name: "space", Help: "空间站", Meta: kit.Dict("exports", []string{"pod", "name"}), List: kit.List(
			kit.MDB_INPUT, "text", "name", "node",
			kit.MDB_INPUT, "button", "value", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "value", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 节点列表
				m.Richs(ice.WEB_SPACE, nil, "", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"time", "type", "name", "user"})
				})
				return
			}

			web := m.Target().Server().(*Frame)
			switch arg[0] {
			case "connect":
				// 基本信息
				node := m.Conf(ice.CLI_RUNTIME, "node.type")
				name := kit.Select(m.Conf(ice.CLI_RUNTIME, "boot.pathname"), m.Conf(ice.CLI_RUNTIME, "boot.hostname"), node == kit.MIME_SERVER)
				user := m.Conf(ice.CLI_RUNTIME, "boot.username")

				m.Richs(ice.WEB_SPIDE, nil, kit.Select("self", arg, 1), func(key string, value map[string]interface{}) {
					host := kit.Format(kit.Value(value, "client.hostname"))

					if u, e := url.Parse(kit.MergeURL("ws://"+host+"/space/", "node", node, "name", name, "user", user)); m.Assert(e) {
						for i := 0; i < m.Confi(ice.WEB_SPACE, "meta.redial.c"); i++ {
							if s, e := net.Dial("tcp", host); !m.Warn(e != nil, "%s", e) {
								if s, _, e := websocket.NewClient(s, u, nil, m.Confi(ice.WEB_SPACE, "meta.buffer.r"), m.Confi(ice.WEB_SPACE, "meta.buffer.w")); !m.Warn(e != nil, "%s", e) {
									// 连接成功
									m.Info("conn %d success %s", i, u)
									if i = 0; web.HandleWSS(m, true, s) {
										break
									}
								}
							}

							// 断线重连
							sleep := time.Duration(rand.Intn(m.Confi(ice.WEB_SPACE, "meta.redial.a"))*i+i*m.Confi(ice.WEB_SPACE, "meta.redial.b")) * time.Millisecond
							m.Info("%d sleep: %s reconnect: %s", i, sleep, u)
							time.Sleep(sleep)
						}
					}
				})

			default:
				// 本地命令
				if arg[0] == "" || arg[0] == m.Conf(ice.CLI_RUNTIME, "node.name") {
					m.Cmdy(arg[1:])
					break
				}

				target := strings.Split(arg[0], ".")
				m.Warn(m.Richs(ice.WEB_SPACE, nil, target[0], func(key string, value map[string]interface{}) {
					if socket, ok := value["socket"].(*websocket.Conn); ok {
						// 构造路由
						id := kit.Format(c.ID())
						m.Optionv(ice.MSG_SOURCE, []string{id, target[0]})
						m.Optionv(ice.MSG_TARGET, target[1:])
						m.Set(ice.MSG_DETAIL, arg[1:]...)
						m.Info("send %s %s", id, m.Format("meta"))

						// 下发命令
						m.Target().Server().(*Frame).send[id] = m
						socket.WriteMessage(MSG_MAPS, []byte(m.Format("meta")))
						m.Call(true, func(msg *ice.Message) *ice.Message {
							// 返回结果
							m.Copy(msg).Log("cost", "%s: %s %v", m.Format("cost"), arg[0], arg[1:])
							return nil
						})
					}
				}) == nil, "not found %s", arg[0])
			}
		}},
		ice.WEB_DREAM: {Name: "dream", Help: "梦想家", Meta: kit.Dict("exports", []string{"you", "name"},
			"detail", []interface{}{"启动", "停止"},
		), List: kit.List(
			kit.MDB_INPUT, "text", "value", "", "name", "name",
			kit.MDB_INPUT, "text", "value", "", "name", "type",
			kit.MDB_INPUT, "button", "value", "创建", "action", "auto",
		), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) > 1 {
				switch arg[1] {
				case "启动":
					arg = arg[:1]
				case "停止", "stop":
					m.Cmd(ice.WEB_SPACE, arg[0], "exit", "1")
					time.Sleep(time.Second * 3)
					m.Cmd(ice.GDB_EVENT, "action", ice.DREAM_CLOSE, arg[0])
					arg = arg[:0]
				}
			}

			if len(arg) > 0 {
				// 规范命名
				if !strings.Contains(arg[0], "-") {
					arg[0] = m.Time("20060102-") + arg[0]
				}

				// 创建目录
				p := path.Join(m.Conf(ice.WEB_DREAM, "meta.path"), arg[0])
				if _, e := os.Stat(p); e != nil {
					os.MkdirAll(p, 0777)
				}

				if m.Richs(ice.WEB_SPACE, nil, arg[0], nil) == nil {
					// 启动任务
					m.Option("cmd_dir", p)
					m.Option("cmd_type", "daemon")
					m.Cmd(m.Confv(ice.WEB_DREAM, "meta.cmd"), "self", arg[0])
					time.Sleep(time.Second * 3)
					m.Cmd(ice.GDB_EVENT, "action", ice.DREAM_START, arg)
				}
			}

			// 任务列表
			m.Cmdy("nfs.dir", m.Conf(ice.WEB_DREAM, "meta.path"), "", "time name")
			m.Table(func(index int, value map[string]string, head []string) {
				m.Push("status", kit.Select("start", "stop", m.Richs(ice.WEB_SPACE, nil, value["name"], nil) == nil))
			})
		}},
		ice.WEB_FAVOR: {Name: "favor", Help: "收藏夹", Meta: kit.Dict("remote", "you", "exports", []string{"hot", "favor"}, "detail", []string{"执行", "编辑", "收录", "下载"}), List: kit.List(
			kit.MDB_INPUT, "text", "name", "hot", "action", "auto",
			kit.MDB_INPUT, "text", "name", "id", "action", "auto",
			kit.MDB_INPUT, "button", "value", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "value", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 3 {
				id := kit.Select(arg[0], kit.Select(m.Option("id"), arg[3], arg[2] == "id"))
				switch arg[1] {
				case "modify":
					// 编辑收藏
					m.Richs(ice.WEB_FAVOR, nil, m.Option("hot"), func(key string, value map[string]interface{}) {
						m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "id", id, func(index int, value map[string]interface{}) {
							m.Info("modify favor: %s index: %d value: %v->%v", key, index, value[arg[2]], arg[3])
							kit.Value(value, arg[2], arg[3])
						})
					})
					arg = []string{m.Option("hot")}
				case "收录":
					m.Richs(ice.WEB_FAVOR, nil, m.Option("hot"), func(key string, value map[string]interface{}) {
						m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "id", id, func(index int, value map[string]interface{}) {
							m.Cmd(ice.WEB_STORY, "add", value["type"], value["name"], value["text"])
						})
					})
					arg = []string{m.Option("hot")}
				}
			}
			if len(arg) > 0 {
				switch arg[0] {
				case "import":
					if len(arg) == 2 {
						m.Cmdy(ice.MDB_IMPORT, ice.WEB_FAVOR, kit.MDB_HASH, kit.MDB_HASH, arg[1])
					} else {
						m.Richs(ice.WEB_FAVOR, nil, arg[2], func(key string, value map[string]interface{}) {
							m.Cmdy(ice.MDB_IMPORT, ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), kit.MDB_LIST, arg[1])
						})
					}
					return
				case "export":
					if len(arg) == 1 {
						m.Cmdy(ice.MDB_EXPORT, ice.WEB_FAVOR, kit.MDB_HASH, kit.MDB_HASH, "favor.json")
					} else {
						m.Richs(ice.WEB_FAVOR, nil, arg[1], func(key string, value map[string]interface{}) {
							m.Cmdy(ice.MDB_EXPORT, ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key, kit.MDB_LIST), kit.MDB_LIST, arg[1]+".csv")
						})
					}
					return
				}
			}

			if len(arg) == 0 {
				// 收藏门类
				m.Richs(ice.WEB_FAVOR, nil, "", func(key string, value map[string]interface{}) {
					m.Push("time", kit.Value(value, "meta.time"))
					m.Push("favor", kit.Value(value, "meta.name"))
					m.Push("count", kit.Value(value, "meta.count"))
				})
				return
			}

			// 创建收藏
			favor := ""
			if m.Richs(ice.WEB_FAVOR, nil, arg[0], func(key string, value map[string]interface{}) {
				favor = key
			}) == nil {
				favor = m.Rich(ice.WEB_FAVOR, nil, kit.Data(kit.MDB_NAME, arg[0]))
				m.Info("create favor: %s name: %s", favor, arg[0])
			}

			if len(arg) == 1 {
				// 收藏列表
				m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, favor), "", "", func(index int, value map[string]interface{}) {
					m.Push(kit.Format(index), value, []string{kit.MDB_TIME, kit.MDB_ID, kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT})
					m.Push("link", kit.Format(m.Conf(ice.WEB_SHARE, "meta.template.download"), value["data"], kit.Short(value["data"])))
				})
				return
			}

			if len(arg) == 2 {
				// 收藏详情
				m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, favor), "id", arg[1], func(index int, value map[string]interface{}) {
					for k, v := range value {
						m.Push("key", k).Push("value", v)
					}
					m.Sort("key")
				})
				return
			}
			if arg[1] == "file" {
				arg[1] = kit.MIME_FILE
			}

			// 添加收藏
			extra := kit.Dict()
			for i := 4; i < len(arg)-1; i += 2 {
				kit.Value(extra, arg[i], arg[i+1])
			}
			index := m.Grow(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, favor), kit.Dict(
				kit.MDB_TYPE, arg[1], kit.MDB_NAME, arg[2], kit.MDB_TEXT, arg[3],
				"extra", extra,
			))
			m.Info("create favor: %s index: %d name: %s", favor, index, arg[2])
			m.Echo("%d", index)
		}},
		ice.WEB_CACHE: {Name: "cache", Help: "缓存池", Meta: kit.Dict("remote", "you"), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 记录列表
				m.Grows(ice.WEB_CACHE, nil, "", "", func(index int, value map[string]interface{}) {
					m.Push(kit.Format(index), value, []string{kit.MDB_TIME, kit.MDB_SIZE, kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT})
					m.Push(kit.MDB_LINK, kit.Format(m.Conf(ice.WEB_SHARE, "meta.template.download"), value["data"], kit.Short(value["data"])))

				})
				return
			}

			switch arg[0] {
			case "upload":
				// 打开文件
				r := m.Optionv("request").(*http.Request)
				if f, h, e := r.FormFile(kit.Select("upload", arg, 1)); m.Assert(e) {
					defer f.Close()

					// 创建文件
					file := kit.Hashs(f)
					if o, p, e := kit.Create(path.Join(m.Conf(ice.WEB_CACHE, ice.Meta("path")), file[:2], file)); m.Assert(e) {
						defer o.Close()

						// 保存文件
						f.Seek(0, os.SEEK_SET)
						if n, e := io.Copy(o, f); m.Assert(e) {
							m.Info("upload: %s file: %s", kit.FmtSize(n), p)
							arg = kit.Simple(arg[0], h.Header.Get("Content-Type"), h.Filename, p, p, n)
						}
					}
				}
				fallthrough
			case "add":
				// 添加数据
				size := kit.Select(kit.Format(len(arg[3])), arg, 5)
				h := m.Rich(ice.WEB_CACHE, nil, kit.Dict(
					kit.MDB_TYPE, arg[1], kit.MDB_NAME, arg[2], kit.MDB_TEXT, arg[3],
					kit.MDB_FILE, kit.Select("", arg, 4),
					kit.MDB_SIZE, size, arg[1], arg[3],
				))
				m.Info("cache: %s type: %s name: %s", h, arg[1], arg[2])

				// 添加记录
				m.Grow(ice.WEB_CACHE, nil, kit.Dict(
					kit.MDB_TYPE, arg[1], kit.MDB_NAME, arg[2], kit.MDB_TEXT, arg[3],
					kit.MDB_SIZE, size, "data", h,
				))

				// 返回结果
				m.Push("time", m.Time())
				m.Push("type", arg[1])
				m.Push("name", arg[2])
				m.Push("text", arg[3])
				m.Push("size", size)
				m.Push("data", h)

				// 保存数据
				if arg[0] == "add" {
					m.Cmd("nfs.save", path.Join(m.Conf(ice.WEB_CACHE, ice.Meta("path")), h[:2], h), arg[3])
				}
			}
		}},
		ice.WEB_STORY: {Name: "story", Help: "故事会", Meta: kit.Dict("remote", "you", "exports", []string{"top", "story"}, "detail", []string{"归档", "共享", "下载"}), List: kit.List(
			kit.MDB_INPUT, "text", "name", "top", "action", "auto",
			kit.MDB_INPUT, "button", "value", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "value", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 {
				switch arg[1] {
				case "共享":
					switch arg[2] {
					case "story":
						m.Cmd(ice.WEB_SHARE, "add", "story", arg[3], arg[3])
						arg = []string{}
					case "list":
						m.Cmd(ice.WEB_SHARE, "add", "story", m.Option("top"), arg[3])
						arg = []string{m.Option("top")}
					}
				}
			}

			if len(arg) == 0 {
				// 故事列表
				m.Richs(ice.WEB_STORY, "head", "", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"time", "scene", "story", "count"})
				})
				return
			}

			// head list data time text file
			switch arg[0] {
			case "add", "upload":
				// 保存数据
				m.Cmdy(ice.WEB_CACHE, arg)
				arg = []string{arg[0], m.Append("type"), m.Append("name"), m.Append("data")}

				// 查询索引
				head, prev, count := "", "", 0
				m.Richs(ice.WEB_STORY, "head", arg[2], func(key string, value map[string]interface{}) {
					head, prev, count = key, kit.Format(value["list"]), kit.Int(value["count"])
					m.Log("info", "head: %v prev: %v count: %v", head, prev, count)
				})

				// 添加节点
				list := m.Rich(ice.WEB_STORY, nil, kit.Dict(
					"count", count+1, "scene", arg[1], "story", arg[2], "data", arg[3], "prev", prev,
				))
				m.Info("%s: %s story: %s", arg[1], arg[2], list)
				m.Push("list", list)

				// 添加索引
				m.Rich(ice.WEB_STORY, "head", kit.Dict(
					"count", count+1, "scene", arg[1], "story", arg[2], "list", list,
				))

				m.Echo(list)

			case "download":
				// 下载文件
				if m.Cmdy(ice.WEB_STORY, "index", arg[1]); m.Append("file") != "" {
					m.Push("_output", "file")
				} else {
					m.Push("_output", "result")
				}

			case "commit":
				// 查询索引
				head := kit.Hashs(arg[1])
				prev := m.Conf("story", ice.Meta("head", head, "list"))
				m.Log("info", "head: %v prev: %v", head, prev)

				// 查询节点
				menu := map[string]string{}
				for i := 2; i < len(arg); i++ {
					if i < len(arg)-1 && m.Confs("story", kit.Keys("hash", arg[i+1])) {
						menu[arg[i]] = arg[i+1]
						i++
					} else if head := kit.Hashs(arg[i]); m.Confs("story", kit.Keys("meta", "head", head)) {
						menu[arg[i]] = m.Conf(ice.WEB_STORY, ice.Meta("head", head, "list"))
					} else {
						m.Error(true, "not found %v", arg[i])
						return
					}
				}

				// 添加节点
				meta := map[string]interface{}{
					"time":  m.Time(),
					"scene": "commit",
					"story": arg[1],
					"list":  menu,
					"prev":  prev,
				}
				list := m.Rich("story", nil, meta)
				m.Log("info", "list: %v meta: %v", list, kit.Format(meta))

				// 添加索引
				m.Conf("story", ice.Meta("head", head), map[string]interface{}{
					"time": m.Time(), "scene": "commit", "story": arg[1], "list": list,
				})
				m.Echo(list)

			case "history":
				// 历史记录
				list := m.Cmd(ice.WEB_STORY, "index", arg[1]).Append("list")
				for i := 0; i < 10 && list != ""; i++ {
					m.Confm(ice.WEB_STORY, kit.Keys("hash", list), func(value map[string]interface{}) {
						// 直连节点
						m.Confm(ice.WEB_CACHE, kit.Keys("hash", value["data"]), func(val map[string]interface{}) {
							m.Push(list, value, []string{"key", "time", "count", "scene", "story"})

							m.Push("drama", val["text"])
							m.Push("data", value["data"])
						})

						// 复合节点
						kit.Fetch(value["list"], func(key string, val string) {
							m.Push(list, value, []string{"key", "time", "count"})

							node := m.Confm(ice.WEB_STORY, kit.Keys("hash", val))
							m.Push("scene", node["scene"])
							m.Push("story", kit.Keys(kit.Format(value["story"]), key))

							m.Push("drama", m.Conf(ice.WEB_CACHE, kit.Keys("hash", node["data"], "text")))
							m.Push("data", node["data"])
						})

						list = kit.Format(value["prev"])
					})
				}

			case "index":
				// 查询索引
				if m.Richs(ice.WEB_STORY, "head", arg[1], func(key string, value map[string]interface{}) {
					arg[1] = kit.Format(value["list"])
				}) == nil {
					arg[1] = kit.Select(arg[1], m.Conf(ice.WEB_STORY, kit.Keys("head.hash", arg[1], "list")))
				}

				// 查询节点
				if node := m.Confm(ice.WEB_STORY, kit.Keys("hash", arg[1])); node != nil {
					m.Push("list", arg[1])
					m.Push(arg[1], node, []string{"scene", "story"})
					arg[1] = kit.Format(node["data"])
				}

				// 查询数据
				if node := m.Confm(ice.WEB_CACHE, kit.Keys("hash", arg[1])); node != nil {
					m.Push("data", arg[1])
					m.Push(arg[1], node, []string{"text", "time", "size", "type", "name", "file"})
					m.Echo("%s", node["text"])
				}
			default:
				m.Cmd(ice.WEB_STORY, "history", arg).Table(func(index int, value map[string]string, head []string) {
					m.Push("time", value["time"])
					m.Push("list", value["key"])
					m.Push("scene", value["scene"])
					m.Push("story", value["story"])
					m.Push("drama", value["drama"])
					m.Push("link", kit.Format(m.Conf(ice.WEB_SHARE, "meta.template.download"), value["data"], kit.Short(value["data"])))
				})
			}
		}},
		ice.WEB_SHARE: {Name: "share", Help: "共享链", Meta: kit.Dict("remote", "you"), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 共享列表
				m.Grows(ice.WEB_SHARE, nil, "", "", func(key int, value map[string]interface{}) {
					m.Push(kit.Format(key), value, []string{kit.MDB_TIME, kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT})
					m.Push("link", fmt.Sprintf(m.Conf(ice.WEB_SHARE, "meta.template.share"), value["share"], value["share"]))
				})
				return
			}

			switch arg[0] {
			case "add":
				// 创建共享
				h := m.Rich(ice.WEB_SHARE, nil, kit.Dict(
					kit.MDB_TYPE, arg[1], kit.MDB_NAME, arg[2], kit.MDB_TEXT, arg[3],
				))
				m.Grow(ice.WEB_SHARE, nil, kit.Dict(
					kit.MDB_TYPE, arg[1], kit.MDB_NAME, arg[2], kit.MDB_TEXT, arg[3],
					"share", h,
				))
				m.Info("share: %s", h)
				m.Echo(h)
			}
		}},
		ice.WEB_ROUTE: {Name: "route", Help: "路由表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		ice.WEB_PROXY: {Name: "proxy", Help: "代理商", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},

		"/proxy/": {Name: "/proxy/", Help: "代理商", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"/share/": {Name: "/share/", Help: "共享链", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			key := kit.Select("", strings.Split(cmd, "/"), 2)
			m.Confm(ice.WEB_SHARE, kit.Keys("hash", key), func(value map[string]interface{}) {
				m.Info("share %s %v", key, kit.Format(value))
				switch value["type"] {
				case kit.MIME_STORY:
					if m.Cmdy(ice.WEB_STORY, "index", value["data"]); m.Append("file") != "" {
						m.Push("_output", "file")
					} else {
						m.Push("_output", "result")
					}
				default:
					if m.Cmdy(ice.WEB_STORY, "index", value["data"]); m.Append("file") != "" {
						m.Push("_output", "file")
					} else {
						m.Push("_output", "result")
					}
				}

			})
		}},
		"/space/": {Name: "/space/", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			r := m.Optionv("request").(*http.Request)
			w := m.Optionv("response").(http.ResponseWriter)

			if s, e := websocket.Upgrade(w, r, nil, m.Confi(ice.WEB_SPACE, "meta.buffer"), m.Confi(ice.WEB_SPACE, "meta.buffer")); m.Assert(e) {

				// 添加节点
				h := m.Rich(ice.WEB_SPACE, nil, kit.Dict(
					kit.MDB_TYPE, m.Option("node"),
					kit.MDB_NAME, m.Option("name"),
					kit.MDB_USER, m.Option("user"),
					"share", m.Cmdx(ice.WEB_SHARE, "add", m.Option("node"), m.Option("name"), m.Option("user")),
					"socket", s,
				))
				m.Info("space: %s", m.Option(kit.MDB_NAME))

				m.Gos(m, func(m *ice.Message) {
					// 监听消息
					web := m.Target().Server().(*Frame)
					web.HandleWSS(m, false, s)
					m.Log("close", "%s: %s", m.Option(kit.MDB_NAME), kit.Format(m.Confv(ice.WEB_SPACE, kit.Keys(kit.MDB_HASH, h))))
					m.Confv(ice.WEB_SPACE, kit.Keys(kit.MDB_HASH, h), "")
				})
			}
		}},
		"/static/volcanos/plugin/github.com/": {Name: "/space/", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			file := strings.TrimPrefix(cmd, "/static/volcanos/")
			if _, e := os.Stat(path.Join("usr/volcanos", file)); e != nil {
				m.Cmd("cli.system", "git", "clone", "https://"+strings.Join(strings.Split(cmd, "/")[4:7], "/"),
					path.Join("usr/volcanos", strings.Join(strings.Split(cmd, "/")[3:7], "/")))
			}

			m.Push("_output", "void")
			r := m.Optionv("request").(*http.Request)
			w := m.Optionv("response").(http.ResponseWriter)
			http.ServeFile(w, r, path.Join("usr/volcanos", file))
		}},
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
