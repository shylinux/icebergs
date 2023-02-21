package web

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _spide_create(m *ice.Message, name, address string) {
	if uri, e := url.Parse(address); !m.Warn(e != nil || address == "", ice.ErrNotValid, address) {
		m.Logs(mdb.CREATE, SPIDE, name, ADDRESS, address)
		mdb.HashSelectUpdate(m, mdb.HashCreate(m, CLIENT_NAME, name), func(value ice.Map) {
			dir, file := path.Split(uri.EscapedPath())
			value[SPIDE_CLIENT] = kit.Dict(mdb.NAME, name, SPIDE_METHOD, http.MethodPost, "url", address, "origin", uri.Scheme+"://"+uri.Host,
				tcp.PROTOCOL, uri.Scheme, tcp.HOSTNAME, uri.Hostname(), tcp.HOST, uri.Host, nfs.PATH, dir, nfs.FILE, file, "query", uri.RawQuery,
				cli.TIMEOUT, "30s", LOGHEADERS, ice.FALSE,
			)
		})
	}
}
func _spide_args(m *ice.Message, arg []string, val ...string) (string, []string) {
	if kit.IndexOf(val, arg[0]) > -1 {
		return arg[0], arg[1:]
	}
	return "", arg
}
func _spide_show(m *ice.Message, name string, arg ...string) {
	msg := mdb.HashSelects(m.Spawn(), name)
	if len(arg) == 1 && msg.Append(arg[0]) != "" {
		m.Echo(msg.Append(arg[0]))
		return
	}
	format, arg := _spide_args(m, arg, SPIDE_RAW, SPIDE_MSG, SPIDE_CACHE, SPIDE_SAVE)
	file := ""
	if format == SPIDE_SAVE {
		file, arg = arg[0], arg[1:]
	}
	method, arg := _spide_args(m, arg, http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete)
	method = kit.Select(http.MethodPost, kit.Select(msg.Append(CLIENT_METHOD), method))
	uri, arg := arg[0], arg[1:]
	body, head, arg := _spide_body(m, method, arg...)
	if c, ok := body.(io.Closer); ok {
		defer c.Close()
	}
	req, e := http.NewRequest(method, kit.MergeURL2(msg.Append(CLIENT_URL), uri, arg), body)
	if m.Warn(e, ice.ErrNotValid, uri) {
		return
	}
	mdb.HashSelectDetail(m, name, func(value ice.Map) { _spide_head(m, req, head, value) })
	res, e := _spide_send(m, name, req, kit.Format(msg.Append(CLIENT_TIMEOUT)))
	if m.Warn(e, ice.ErrNotFound, uri) {
		return
	}
	defer res.Body.Close()
	if m.Config(LOGHEADERS) == ice.TRUE {
		for k, v := range res.Header {
			m.Logs(mdb.IMPORT, k, v)
		}
	}
	m.Cost(cli.STATUS, res.Status, nfs.SIZE, res.Header.Get(ContentLength), mdb.TYPE, res.Header.Get(ContentType))
	mdb.HashSelectUpdate(m, name, func(value ice.Map) {
		for _, v := range res.Cookies() {
			kit.Value(value, kit.Keys(SPIDE_COOKIE, v.Name), v.Value)
			m.Logs(mdb.IMPORT, v.Name, v.Value)
		}
	})
	if m.Warn(res.StatusCode != http.StatusOK, ice.ErrNotValid, uri, cli.STATUS, res.Status) {
		switch res.StatusCode {
		case http.StatusNotFound, http.StatusUnauthorized:
			return
		}
	}
	_spide_save(m, format, file, uri, res)
}
func _spide_body(m *ice.Message, method string, arg ...string) (io.Reader, ice.Maps, []string) {
	head := ice.Maps{}
	body, ok := m.Optionv(SPIDE_BODY).(io.Reader)
	if !ok && len(arg) > 0 && method != http.MethodGet {
		if len(arg) == 1 {
			arg = []string{SPIDE_DATA, arg[0]}
		}
		switch arg[0] {
		case SPIDE_FORM:
			arg = kit.Simple(arg, func(v string) string { return url.QueryEscape(v) })
			head[ContentType], body = ContentFORM, bytes.NewBufferString(kit.JoinKV("=", "&", arg[1:]...))
		case SPIDE_PART:
			head[ContentType], body = _spide_part(m, arg...)
		case SPIDE_DATA:
			head[ContentType], body = ContentJSON, bytes.NewBufferString(kit.Select("{}", arg, 1))
		case SPIDE_FILE:
			if f, e := nfs.OpenFile(m, arg[1]); m.Assert(e) {
				m.Logs(mdb.IMPORT, nfs.FILE, arg[1])
				body = f
			}
		case SPIDE_JSON:
			arg = arg[1:]
			fallthrough
		default:
			data := ice.Map{}
			kit.Fetch(arg, func(k, v string) { kit.Value(data, k, v) })
			head[ContentType], body = ContentJSON, bytes.NewBufferString(kit.Format(data))
		}
		arg = arg[:0]
	} else {
		body = bytes.NewBuffer([]byte{})
	}
	return body, head, arg
}
func _spide_part(m *ice.Message, arg ...string) (string, io.Reader) {
	buf := &bytes.Buffer{}
	mp := multipart.NewWriter(buf)
	defer mp.Close()
	size, cache := int64(0), time.Now().Add(-time.Hour*240000)
	for i := 1; i < len(arg)-1; i += 2 {
		if arg[i] == nfs.SIZE {
			size = kit.Int64(arg[i+1])
		} else if arg[i] == SPIDE_CACHE {
			if t, e := time.ParseInLocation(ice.MOD_TIME, arg[i+1], time.Local); !m.Warn(e, ice.ErrNotValid) {
				cache = t
			}
		} else if strings.HasPrefix(arg[i+1], ice.AT) {
			if s, e := nfs.StatFile(m, arg[i+1][1:]); !m.Warn(e, ice.ErrNotValid) {
				if s.Size() == size && s.ModTime().Before(cache) {
					continue
				}
				m.Logs(mdb.IMPORT, "local", s.ModTime(), nfs.SIZE, s.Size(), CACHE, cache, nfs.SIZE, size)
			}
			if f, e := nfs.OpenFile(m, arg[i+1][1:]); !m.Warn(e, ice.ErrNotValid, arg[i+1]) {
				defer f.Close()
				if p, e := mp.CreateFormFile(arg[i], path.Base(arg[i+1][1:])); !m.Warn(e, ice.ErrNotValid, arg[i+1]) {
					if n, e := io.Copy(p, f); !m.Warn(e, ice.ErrNotValid, arg[i+1]) {
						m.Logs(mdb.EXPORT, nfs.FILE, arg[i+1], nfs.SIZE, n)
					}
				}
			}
		} else {
			mp.WriteField(arg[i], arg[i+1])
		}
	}
	return mp.FormDataContentType(), buf
}
func _spide_head(m *ice.Message, req *http.Request, head ice.Maps, value ice.Map) {
	m.Logs(req.Method, req.URL.String())
	kit.Fetch(value[SPIDE_HEADER], func(k string, v string) {
		req.Header.Set(k, v)
		m.Logs("Header", k, v)
	})
	kit.Fetch(value[SPIDE_COOKIE], func(k string, v string) {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
		m.Logs("Cookie", k, v)
	})
	kit.Fetch(kit.Simple(m.Optionv(SPIDE_COOKIE)), func(k, v string) {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
		m.Logs("Cookie", k, v)
	})
	kit.Fetch(kit.Simple(m.Optionv(SPIDE_HEADER)), func(k, v string) {
		req.Header.Set(k, v)
		m.Logs("Header", k, v)
	})
	kit.Fetch(head, func(k, v string) {
		req.Header.Set(k, v)
		m.Logs("Header", k, v)
	})
	if req.Method == http.MethodPost {
		m.Logs(kit.Select(ice.AUTO, req.Header.Get(ContentLength)), req.Header.Get(ContentType))
	}
}
func _spide_send(m *ice.Message, name string, req *http.Request, timeout string) (*http.Response, error) {
	client := mdb.HashSelectTarget(m, name, func() ice.Any { return &http.Client{Timeout: kit.Duration(timeout)} }).(*http.Client)
	return client.Do(req)
}
func _spide_save(m *ice.Message, format, file, uri string, res *http.Response) {
	switch format {
	case SPIDE_RAW:
		if b, _ := ioutil.ReadAll(res.Body); strings.HasPrefix(res.Header.Get(ContentType), ContentJSON) {
			m.Echo(kit.Formats(kit.UnMarshal(string(b))))
		} else {
			m.Echo(string(b))
		}
	case SPIDE_MSG:
		var data map[string][]string
		m.Assert(json.NewDecoder(res.Body).Decode(&data))
		kit.Fetch(data[ice.MSG_APPEND], func(k string) { kit.Fetch(data[k], func(v string) { m.Push(k, v) }) })
		m.Resultv(data[ice.MSG_RESULT])
	case SPIDE_SAVE:
		_cache_download(m, res, file, m.OptionCB(SPIDE))
	case SPIDE_CACHE:
		m.Cmdy(CACHE, DOWNLOAD, res.Header.Get(ContentType), uri, kit.Dict(RESPONSE, res), m.OptionCB(SPIDE))
		m.Echo(m.Append(mdb.HASH))
	default:
		var data ice.Any
		if b, e := ioutil.ReadAll(res.Body); !m.Warn(e) {
			if json.Unmarshal(b, &data) == nil {
				m.Push("", kit.KeyValue(ice.Map{}, "", m.Optionv(SPIDE_RES, data)))
			} else {
				m.Echo(string(b))
			}
		}
	}
}

const (
	SPIDE_RAW   = "raw"
	SPIDE_MSG   = "msg"
	SPIDE_CACHE = "cache"
	SPIDE_SAVE  = "save"

	SPIDE_BODY = "body"
	SPIDE_FORM = "form"
	SPIDE_PART = "part"
	SPIDE_JSON = "json"
	SPIDE_DATA = "data"
	SPIDE_FILE = "file"
	SPIDE_RES  = "content_data"

	Bearer        = "Bearer"
	Authorization = "Authorization"
	ContentType   = "Content-Type"
	ContentLength = "Content-Length"
	UserAgent     = "User-Agent"
	Referer       = "Referer"
	Accept        = "Accept"

	ContentFORM = "application/x-www-form-urlencoded"
	ContentJSON = "application/json"
	ContentPNG  = "image/png"
	ContentHTML = "text/html"
	ContentCSS  = "text/css"
)
const (
	SPIDE_CLIENT = "client"
	SPIDE_METHOD = "method"
	SPIDE_HEADER = "header"
	SPIDE_COOKIE = "cookie"

	CLIENT_PROTOCOL = "client.protocol"
	CLIENT_HOSTNAME = "client.hostname"
	CLIENT_TIMEOUT  = "client.timeout"

	CLIENT_NAME   = "client.name"
	CLIENT_METHOD = "client.method"
	CLIENT_ORIGIN = "client.origin"
	CLIENT_URL    = "client.url"

	OPEN       = "open"
	FULL       = "full"
	LINK       = "link"
	HTTP       = "http"
	FORM       = "form"
	MERGE      = "merge"
	ADDRESS    = "address"
	REQUEST    = "request"
	RESPONSE   = "response"
	LOGHEADERS = "logheaders"
)
const SPIDE = "spide"

func init() {
	Index.MergeCommands(ice.Commands{
		SPIDE: {Name: "spide client.name action=raw,msg,save,cache method=GET,PUT,POST,DELETE url format=form,part,json,data,file arg run create", Help: "蜘蛛侠", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				conf := m.Confm(cli.RUNTIME, cli.CONF)
				m.Cmd("", mdb.CREATE, ice.OPS, kit.Select("http://127.0.0.1:9020", conf[cli.CTX_OPS]))
				m.Cmd("", mdb.CREATE, ice.DEV, kit.Select(kit.Select("https://contexts.com.cn", ice.Info.Make.Domain), conf[cli.CTX_DEV]))
				m.Cmd("", mdb.CREATE, ice.COM, kit.Select("https://contexts.com.cn", conf[cli.CTX_COM]))
				m.Cmd("", mdb.CREATE, ice.SHY, kit.Select(kit.Select("https://shylinux.com", ice.Info.Make.Remote), conf[cli.CTX_SHY]))
			}},
			mdb.CREATE: {Name: "create name address", Hand: func(m *ice.Message, arg ...string) { _spide_create(m, m.Option(mdb.NAME), m.Option(ADDRESS)) }},
			tcp.CLIENT: {Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd("", kit.Select(ice.DEV, arg, 0))
				ls := kit.Split(msg.Append(kit.Keys(SPIDE_CLIENT, tcp.HOST)), ice.DF)
				m.Push(tcp.HOST, ls[0]).Push(tcp.PORT, kit.Select(kit.Select("443", "80", msg.Append(CLIENT_PROTOCOL) == ice.HTTP), ls, 1))
				m.Push(DOMAIN, msg.Append(CLIENT_PROTOCOL)+"://"+msg.Append(CLIENT_HOSTNAME)+kit.Select("", arg, 1))
				m.Push(tcp.PROTOCOL, msg.Append(CLIENT_PROTOCOL)).Push(tcp.HOSTNAME, msg.Append(CLIENT_HOSTNAME))
			}},
			MERGE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(kit.MergeURL2(m.CmdAppend("", arg[0], CLIENT_URL), arg[1], arg[2:]))
			}},
		}, mdb.HashAction(mdb.SHORT, CLIENT_NAME, mdb.FIELD, "time,client.name,client.url", LOGHEADERS, ice.FALSE), mdb.ClearHashOnExitAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 2 || arg[0] == "" || (len(arg) > 3 && arg[3] == "") {
				mdb.HashSelect(m, kit.Slice(arg, 0, 1)...).Sort(CLIENT_NAME)
			} else {
				_spide_show(m, arg[0], arg[1:]...)
			}
		}},
		"/spide-demo/": {Actions: aaa.WhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			m.Push("hi", "hello")
			m.Echo("hello world")
		}},
		http.MethodGet: {Name: "GET url key value run", Help: "蜘蛛侠", Hand: func(m *ice.Message, arg ...string) {
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, http.MethodGet, arg[0], arg[1:]))))
		}},
		http.MethodPut: {Name: "PUT url key value run", Help: "蜘蛛侠", Hand: func(m *ice.Message, arg ...string) {
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, http.MethodPut, arg[0], arg[1:]))))
		}},
		http.MethodPost: {Name: "POST url key value run", Help: "蜘蛛侠", Hand: func(m *ice.Message, arg ...string) {
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, http.MethodPost, arg[0], arg[1:]))))
		}},
		http.MethodDelete: {Name: "DELETE url key value run", Help: "蜘蛛侠", Hand: func(m *ice.Message, arg ...string) {
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, http.MethodDelete, arg[0], arg[1:]))))
		}},
	})
}

func SpideGet(m *ice.Message, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(http.MethodGet, arg))
}
func SpidePut(m *ice.Message, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(http.MethodPut, arg))
}
func SpidePost(m *ice.Message, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(http.MethodPost, arg))
}
func SpideDelete(m *ice.Message, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(http.MethodDelete, arg))
}
func SpideSave(m *ice.Message, file, link string, cb func(int, int, int)) *ice.Message {
	return m.Cmd("web.spide", ice.DEV, SPIDE_SAVE, file, http.MethodGet, link, cb)
}
