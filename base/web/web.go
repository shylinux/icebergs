package web

import (
	"github.com/gorilla/websocket"
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"bytes"
	"encoding/json"
	"math/rand"
	"net"
	"net/http"
	"net/url"
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
	expire := time.Now().Add(kit.Duration(msg.Conf("aaa.sess", "meta.expire")))
	msg.Log("cookie", "expire %v sessid %s", kit.Format(expire), sessid)
	http.SetCookie(w, &http.Cookie{Name: ice.WEB_SESS, Value: sessid, Path: "/", Expires: expire})
	return sessid
}
func (web *Frame) Login(msg *ice.Message, w http.ResponseWriter, r *http.Request) bool {
	if msg.Options(ice.WEB_SESS) {
		sub := msg.Cmd("aaa.sess", "check", msg.Option(ice.WEB_SESS))
		msg.Log("info", "role: %s user: %s", msg.Option("userrole", sub.Append("userrole")),
			msg.Option("username", sub.Append("username")))
	}

	msg.Target().Runs(msg, msg.Option("url"), ice.WEB_LOGIN, kit.Simple(msg.Optionv("cmds"))...)
	return true
}
func (web *Frame) HandleWSS(m *ice.Message, safe bool, c *websocket.Conn) bool {
	for {
		if t, b, e := c.ReadMessage(); e != nil {
			m.Log("warn", "space recv %d msg %v", t, e)
			break
		} else {
			switch t {
			case MSG_MAPS:
				socket, msg := c, m.Spawn(b)
				source := kit.Simple(msg.Optionv("_source"))
				target := kit.Simple(msg.Optionv("_target"))
				msg.Log("space", "recv %v %v->%v %v", t, source, target, msg.Formats("meta"))

				if len(target) > 0 {
					if s, ok := msg.Confv("web.space", "hash."+target[0]+".socket").(*websocket.Conn); ok {
						msg.Log("space", "route")
						// 转发报文
						socket, source, target = s, append(source, target[0]), target[1:]
					} else if call, ok := web.send[msg.Option("_target")]; len(target) == 1 && ok {
						msg.Log("space", "done")
						// 接收响应
						delete(web.send, msg.Option("_target"))
						call.Back(msg)
						break
					} else if msg.Option("_handle") == "true" {
						msg.Log("space", "miss")
						// 丢弃报文
						break
					} else {
						// 失败报文
						msg.Log("space", "error")
						msg.Echo("error")
						source, target = []string{source[len(source)-1]}, kit.Revert(source)[1:]
					}
				} else {
					msg.Log("space", "run")
					// 本地执行
					if safe {
						msg = msg.Cmd()
						if msg.Detail() == "exit" {
							return true
						}
					} else {
						msg.Echo("no right")
					}
					msg.Optionv("_handle", "true")
					kit.Revert(source)
					source, target = []string{source[0]}, source[1:]
				}

				// 发送报文
				msg.Optionv("_source", source)
				msg.Optionv("_target", target)
				msg.Log("space", "send %v %v->%v %v", t, source, target, msg.Formats("meta"))
				socket.WriteMessage(t, []byte(msg.Format("meta")))
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
				m.Log("cmd", "%v %v %v", k, p, arg)
				v.Hand(msg, m.Target(), k, kit.Simple(p, arg)...)

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
			m.Log("info", "%v, %v", k, v.Name)
			cb(k, list[1:], v)
		}
	}
	for k, v := range m.Target().Commands {
		m.Log("info", "%v, %v", k, v.Name)
		if strings.HasPrefix(k, "/") || strings.HasPrefix(k, "_") {
			continue
		}
		cb(k, nil, v)
	}

	tmpl = tmpl.Funcs(cgi)
	// tmpl = template.Must(tmpl.ParseGlob(path.Join(m.Conf(ice.WEB_SERVE, "template.path"), "/*.tmpl")))
	// tmpl = template.Must(tmpl.ParseGlob(path.Join(m.Conf(ice.WEB_SERVE, "template.path"), m.Target().Name, "/*.tmpl")))
	tmpl = template.Must(tmpl.ParseFiles(which))
	m.Confm(ice.WEB_SERVE, "template.list", func(index int, value string) { tmpl = template.Must(tmpl.Parse(value)) })
	for i, v := range tmpl.Templates() {
		m.Log("info", "%v, %v", i, v.Name())
	}
	return tmpl
}
func (web *Frame) HandleCmd(m *ice.Message, key string, cmd *ice.Command) {
	web.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
		m.TryCatch(m.Spawns(), true, func(msg *ice.Message) {
			defer func() {
				msg.Log("cost", msg.Format("cost"))
			}()

			msg.Optionv("request", r)
			msg.Optionv("response", w)
			msg.Option("remote_ip", r.Header.Get("remote_ip"))
			msg.Option("agent", r.Header.Get("User-Agent"))
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

			// 请求参数
			r.ParseMultipartForm(4096)
			if r.ParseForm(); len(r.PostForm) > 0 {
				for k, v := range r.PostForm {
					msg.Log("info", "%s: %v", k, v)
				}
				msg.Log("info", "")
			}
			for k, v := range r.Form {
				for _, v := range v {
					msg.Add(ice.MSG_OPTION, k, v)
				}
			}

			// 请求数据
			switch r.Header.Get("Content-Type") {
			case "application/json":
				var data interface{}
				if e := json.NewDecoder(r.Body).Decode(&data); e != nil {
					msg.Log("warn", "%v", e)
				}
				msg.Optionv("content_data", data)
				msg.Log("info", "%v", kit.Formats(data))

				switch d := data.(type) {
				case map[string]interface{}:
					for k, v := range d {
						for _, v := range kit.Simple(v) {
							msg.Add(ice.MSG_OPTION, k, v)
						}
					}
				}
			}

			if web.Login(msg, w, r) {
				msg.Log("cmd", "%s %s", msg.Target().Name, key)
				cmd.Hand(msg, msg.Target(), msg.Option("url"), kit.Simple(msg.Optionv("cmds"))...)
				w.Write([]byte(msg.Formats("meta")))
			}
		})
	})
}
func (web *Frame) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m := web.m

	index := r.Header.Get("index.module") == ""
	if index {
		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			r.Header.Set("remote_ip", ip)
		} else if ip := r.Header.Get("X-Real-Ip"); ip != "" {
			r.Header.Set("remote_ip", ip)
		} else if strings.HasPrefix(r.RemoteAddr, "[") {
			r.Header.Set("remote_ip", strings.Split(r.RemoteAddr, "]")[0][1:])
		} else {
			r.Header.Set("remote_ip", strings.Split(r.RemoteAddr, ":")[0])
		}
		m.Log("info", "").Log("info", "%v %s %s", r.Header.Get("remote_ip"), r.Method, r.URL)
		r.Header.Set("index.module", "some")
		r.Header.Set("index.url", r.URL.String())
		r.Header.Set("index.path", r.URL.Path)
	}

	web.ServeMux.ServeHTTP(w, r)
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
			msg := m.Spawns(s)

			// 级联路由
			route := "/" + s.Name + "/"
			if n, ok := p.Server().(*Frame); ok && n.ServeMux != nil {
				msg.Log("route", "%s <= %s", p.Name, route)
				n.Handle(route, http.StripPrefix(path.Dir(route), w))
			}

			// 静态路由
			m.Confm("web.serve", "static", func(key string, value string) {
				msg.Log("route", "%s <- %s <- %s", s.Name, key, value)
				w.Handle(key, http.StripPrefix(key, http.FileServer(http.Dir(value))))
			})

			// 命令路由
			m.Travel(func(p *ice.Context, sub *ice.Context, k string, x *ice.Command) {
				if s == sub && k[0] == '/' {
					msg.Log("route", "%s <- %s", s.Name, k)
					w.HandleCmd(msg, k, x)
				}
			})
		}
	})

	port := m.Cap(ice.CTX_STREAM, kit.Select(m.Conf(ice.WEB_SPIDE, ice.Meta("self", "port")), arg, 0))
	m.Log("serve", "listen %s %v", port, m.Conf("cli.runtime", "node"))
	web.m, web.Server = m, &http.Server{Addr: port, Handler: web}
	m.Log("serve", "listen %s", web.Server.ListenAndServe())
	return true
}
func (web *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

var Index = &ice.Context{Name: "web", Help: "网页模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ice.WEB_SPIDE: {Name: "spide", Help: "客户端", Value: ice.Data("self.port", ice.WEB_PORT)},
		ice.WEB_SERVE: {Name: "serve", Help: "服务器", Value: map[string]interface{}{
			"static": map[string]interface{}{"/": "usr/volcanos/",
				"/static/volcanos/": "usr/volcanos/",
			},
			"template": map[string]interface{}{"path": "usr/template", "list": []interface{}{
				`{{define "raw"}}{{.Result}}{{end}}`,
			}},
		}},
		ice.WEB_SPACE: {Name: "space", Help: "空间站", Value: ice.Meta("buffer", 4096, "redial", 3000)},
		ice.WEB_STORY: {Name: "story", Help: "故事会", Value: ice.Data("short", "data")},
		ice.WEB_CACHE: {Name: "cache", Help: "缓存", Value: ice.Data(
			"short", "text", "path", "var/file",
			"store", "var/data", "limit", "30", "least", "10",
		)},
		ice.WEB_ROUTE: {Name: "route", Help: "路由", Value: ice.Data()},
		ice.WEB_PROXY: {Name: "proxy", Help: "代理", Value: ice.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Done() }},
		ice.WEB_SPIDE: {Name: "spide", Help: "客户端", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		ice.WEB_SERVE: {Name: "serve", Help: "服务器", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Conf("cli.runtime", "node.name", m.Conf("cli.runtime", "boot.hostname"))
			m.Conf("cli.runtime", "node.type", "server")
			m.Target().Start(m, arg...)
		}},
		ice.WEB_SPACE: {Name: "space", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Conf(ice.WEB_SPACE, ice.MDB_HASH, func(key string, value map[string]interface{}) {
					m.Push(key, value)
				})
				return
			}

			web := m.Target().Server().(*Frame)
			switch arg[0] {
			case "connect":
				node, name := m.Conf("cli.runtime", "node.type"), m.Conf("cli.runtime", "boot.hostname")
				if node == "worker" {
					name = m.Conf("cli.runtime", "boot.pathname")
				}
				host := kit.Select(m.Conf("web.spide", "self.port"), arg, 1)
				p := "ws://" + host + kit.Select("/space", arg, 2) + "?node=" + node + "&name=" + name

				if u, e := url.Parse(p); m.Assert(e) {
					m.TryCatch(m, true, func(m *ice.Message) {
						for {
							if s, e := net.Dial("tcp", host); e == nil {
								if s, _, e := websocket.NewClient(s, u, nil, m.Confi("web.space", "meta.buffer"), m.Confi("web.space", "meta.buffer")); e == nil {
									id := m.Option("_source", []string{kit.Format(c.ID()), "some"})
									web.send[id] = m
									s.WriteMessage(MSG_MAPS, []byte(m.Format("meta")))

									if web.HandleWSS(m, true, s) {
										break
									}
								} else {
									m.Log("warn", "wss %s", e)
								}
							} else {
								m.Log("warn", "dial %s", e)
							}
							time.Sleep(time.Duration(rand.Intn(m.Confi("web.space", "meta.redial"))) * time.Millisecond)
							m.Log("info", "reconnect %v", u)
						}
					})
				}
			default:
				if arg[0] == "" {
					m.Cmdy(arg[1:])
					break
				}

				target := strings.Split(arg[0], ".")
				if socket, ok := m.Confv(ice.WEB_SPACE, "hash."+target[0]+".socket").(*websocket.Conn); !ok {
					m.Echo("error").Echo("not found")
				} else {
					id := kit.Format(c.ID())
					m.Optionv("_source", []string{id, target[0]})
					m.Optionv("_target", target[1:])
					m.Set(ice.MSG_DETAIL, arg[1:]...)

					web := m.Target().Server().(*Frame)
					web.send[id] = m

					now := time.Now()
					socket.WriteMessage(MSG_MAPS, []byte(m.Format("meta")))
					m.Call(true, func(msg *ice.Message) *ice.Message {
						m.Copy(msg)
						m.Log("info", "cost %s", time.Now().Sub(now))
						return nil
					})
				}
			}
		}},
		ice.WEB_STORY: {Name: "story", Help: "故事会", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Confm("story", ice.Meta("head"), func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"key", "type", "time", "story"})
				})
				return
			}

			// head list data time text file
			switch arg[0] {
			case "add":
				// 查询索引
				head := kit.Hashs(arg[1])
				prev := m.Conf("story", ice.Meta("head", head, "list"))
				m.Log("info", "head: %v prev: %v", head, prev)

				// 添加节点
				meta := map[string]interface{}{
					"time":  m.Time(),
					"story": arg[1],
					"scene": arg[2],
					"data":  m.Cmdx(ice.WEB_CACHE, "add", "text", arg[2]),
					"prev":  prev,
				}
				list := m.Rich("story", nil, meta)
				m.Log("info", "list: %v meta: %v", list, kit.Format(meta))

				// 添加索引
				m.Conf("story", ice.Meta("head", head), map[string]interface{}{
					"time": m.Time(), "type": "text",
					"story": arg[1], "list": list,
				})
				m.Echo(list)

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
						menu[arg[i]] = head
					} else {
						m.Error("not found %v", arg[i])
						return
					}
				}

				// 添加节点
				meta := map[string]interface{}{
					"time":  m.Time(),
					"story": arg[1],
					"list":  menu,
					"prev":  prev,
				}
				list := m.Rich("story", nil, meta)
				m.Log("info", "list: %v meta: %v", list, kit.Format(meta))

				// 添加索引
				m.Conf("story", ice.Meta("head", head), map[string]interface{}{
					"time": m.Time(), "type": "list",
					"story": arg[1], "list": list,
				})
				m.Echo(list)
			}
		}},
		ice.WEB_CACHE: {Name: "cache", Help: "缓存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Confm("cache", "hash", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"time", "text"})
				})
				return
			}

			switch arg[0] {
			case "add":
				// 添加数据
				data := m.Rich("cache", nil, map[string]interface{}{
					"time": m.Time(), "type": arg[1], arg[1]: arg[2],
				})
				m.Info("data: %v type: %v text: %v", data, arg[1], arg[2])
				m.Echo(data)
				m.Cmd("nfs.save", path.Join(m.Conf("cache", ice.Meta("path")), data[:2], data), arg[2])
			}
		}},
		ice.WEB_ROUTE: {Name: "route", Help: "路由", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		ice.WEB_PROXY: {Name: "proxy", Help: "代理", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},

		"/space": &ice.Command{Name: "/space", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			r := m.Optionv("request").(*http.Request)
			w := m.Optionv("response").(http.ResponseWriter)
			if s, e := websocket.Upgrade(w, r, nil, m.Confi("web.space", "meta.buffer"), m.Confi("web.space", "meta.buffer")); m.Assert(e) {
				h := m.Option("name")

				meta := map[string]interface{}{
					"create_time": m.Time(),
					"socket":      s,
					"type":        m.Option("node"),
					"name":        m.Option("name"),
				}
				m.Confv(ice.WEB_SPACE, []string{ice.MDB_HASH, h}, meta)
				m.Log("space", "conn %v %v", h, kit.Formats(m.Confv(ice.WEB_SPACE)))

				web := m.Target().Server().(*Frame)
				m.Gos(m, func(m *ice.Message) {
					web.HandleWSS(m, false, s)
					m.Log("space", "close %v %v", h, kit.Formats(m.Confv(ice.WEB_SPACE)))
					m.Confv(ice.WEB_SPACE, []string{ice.MDB_HASH, h}, "")
				})
			}
		}},
	},
}

func init() { ice.Index.Register(Index, &Frame{}) }
