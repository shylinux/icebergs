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
	if uri, e := url.Parse(address); !m.Warn(e != nil || address == "") {
		m.Logs(mdb.CREATE, SPIDE, name, ADDRESS, address)
		dir, file := path.Split(uri.EscapedPath())
		mdb.HashCreate(m, CLIENT_NAME, name)
		mdb.HashSelectUpdate(m, name, func(value ice.Map) {
			value[SPIDE_CLIENT] = kit.Dict(
				mdb.NAME, name, SPIDE_METHOD, SPIDE_POST, "url", address,
				tcp.PROTOCOL, uri.Scheme, tcp.HOSTNAME, uri.Host,
				nfs.PATH, dir, nfs.FILE, file, "query", uri.RawQuery,
				cli.TIMEOUT, "600s", LOGHEADERS, ice.FALSE,
			)
		})
	}
}
func _spide_list(m *ice.Message, arg ...string) {
	msg := mdb.HashSelects(m.Spawn(), arg[0])
	if len(arg) == 2 && msg.Append(arg[1]) != "" {
		m.Echo(msg.Append(arg[1]))
		return
	}

	cache, save := "", ""
	switch arg[1] { // 缓存方法
	case SPIDE_RAW:
		cache, arg = arg[1], arg[1:]
	case SPIDE_MSG:
		cache, arg = arg[1], arg[1:]
	case SPIDE_SAVE:
		cache, save, arg = arg[1], arg[2], arg[2:]
	case SPIDE_CACHE:
		cache, arg = arg[1], arg[1:]
	}

	method := kit.Select(SPIDE_POST, msg.Append(CLIENT_METHOD))
	switch arg = arg[1:]; arg[0] { // 请求方法
	case SPIDE_GET:
		method, arg = SPIDE_GET, arg[1:]
	case SPIDE_PUT:
		method, arg = SPIDE_PUT, arg[1:]
	case SPIDE_POST:
		method, arg = SPIDE_POST, arg[1:]
	case SPIDE_DELETE:
		method, arg = SPIDE_DELETE, arg[1:]
	}

	// 构造请求
	uri, arg := arg[0], arg[1:]
	body, head, arg := _spide_body(m, method, arg...)
	req, e := http.NewRequest(method, kit.MergeURL2(msg.Append(CLIENT_URL), uri, arg), body)
	if m.Warn(e, ice.ErrNotValid, uri) {
		return
	}

	// 请求变量
	mdb.HashSelectDetail(m, msg.Append(CLIENT_NAME), func(value ice.Map) {
		_spide_head(m, req, head, value)
	})

	// 发送请求
	res, e := _spide_send(m, msg.Append(CLIENT_NAME), req, kit.Format(msg.Append(CLIENT_TIMEOUT)))
	if m.Warn(e, ice.ErrNotFound, uri) {
		return
	}
	defer res.Body.Close()

	// 请求日志
	if m.Config(LOGHEADERS) == ice.TRUE {
		for k, v := range res.Header {
			m.Logs(mdb.IMPORT, k, v)
		}
	}
	m.Cost(cli.STATUS, res.Status, nfs.SIZE, res.Header.Get(ContentLength), mdb.TYPE, res.Header.Get(ContentType))

	// 响应变量
	mdb.HashSelectUpdate(m, msg.Append(CLIENT_NAME), func(value ice.Map) {
		for _, v := range res.Cookies() {
			kit.Value(value, kit.Keys(SPIDE_COOKIE, v.Name), v.Value)
			m.Logs(mdb.IMPORT, v.Name, v.Value)
		}
	})

	// 处理异常
	if m.Warn(res.StatusCode != http.StatusOK, ice.ErrNotValid, uri, cli.STATUS, res.Status) {
		switch m.SetResult(); res.StatusCode {
		case http.StatusNotFound:
			m.Warn(true, ice.ErrNotFound, uri)
			return
		case http.StatusUnauthorized:
			m.Warn(true, ice.ErrNotRight, uri)
			return
		}
	}

	// 解析结果
	_spide_save(m, cache, save, uri, res)
}
func _spide_body(m *ice.Message, method string, arg ...string) (io.Reader, ice.Maps, []string) {
	head := ice.Maps{}
	body, ok := m.Optionv(SPIDE_BODY).(io.Reader)
	if !ok && len(arg) > 0 && method != SPIDE_GET {
		if len(arg) == 1 {
			arg = []string{SPIDE_DATA, arg[0]}
		}
		switch arg[0] {
		case SPIDE_FORM:
			data := []string{}
			for i := 1; i < len(arg)-1; i += 2 {
				data = append(data, url.QueryEscape(arg[i])+"="+url.QueryEscape(arg[i+1]))
			}
			body = bytes.NewBufferString(strings.Join(data, "&"))
			head[ContentType] = ContentFORM

		case SPIDE_PART:
			body, head[ContentType] = _spide_part(m, arg...)

		case SPIDE_DATA:
			if len(arg) == 1 {
				arg = append(arg, "{}")
			}
			body, arg = bytes.NewBufferString(arg[1]), arg[2:]
			head[ContentType] = ContentJSON

		case SPIDE_FILE:
			if f, e := nfs.OpenFile(m, arg[1]); m.Assert(e) {
				defer f.Close()
				body, arg = f, arg[2:]
			}

		case SPIDE_JSON:
			arg = arg[1:]
			fallthrough
		default:
			data := ice.Map{}
			for i := 0; i < len(arg)-1; i += 2 {
				kit.Value(data, arg[i], arg[i+1])
			}
			if b, e := json.Marshal(data); m.Assert(e) {
				head[ContentType] = ContentJSON
				body = bytes.NewBuffer(b)
			}
			m.Logs(mdb.EXPORT, SPIDE_JSON, kit.Format(data))
		}
		arg = arg[:0]
	} else {
		body = bytes.NewBuffer([]byte{})
	}
	return body, head, arg
}
func _spide_part(m *ice.Message, arg ...string) (io.Reader, string) {
	buf := &bytes.Buffer{}
	mp := multipart.NewWriter(buf)
	defer mp.Close()

	cache := time.Now().Add(-time.Hour * 240000)
	var size int64
	for i := 1; i < len(arg)-1; i += 2 {
		if arg[i] == nfs.SIZE {
			size = kit.Int64(arg[i+1])
		}
		if arg[i] == SPIDE_CACHE {
			if t, e := time.ParseInLocation(ice.MOD_TIME, arg[i+1], time.Local); e == nil {
				cache = t
			}
		}
		if strings.HasPrefix(arg[i+1], "@") {
			if s, e := nfs.StatFile(m, arg[i+1][1:]); e == nil {
				m.Logs(mdb.IMPORT, "local", s.ModTime(), nfs.SIZE, s.Size(), CACHE, cache, nfs.SIZE, size)
				if s.Size() == size && s.ModTime().Before(cache) {
					// break
				}
			}
			if f, e := nfs.OpenFile(m, arg[i+1][1:]); m.Assert(e) {
				defer f.Close()
				if p, e := mp.CreateFormFile(arg[i], path.Base(arg[i+1][1:])); m.Assert(e) {
					if n, e := io.Copy(p, f); m.Assert(e) {
						m.Logs(mdb.EXPORT, nfs.FILE, arg[i+1], nfs.SIZE, n)
					}
				}
			}
		} else {
			mp.WriteField(arg[i], arg[i+1])
		}
	}
	return buf, mp.FormDataContentType()
}
func _spide_head(m *ice.Message, req *http.Request, head ice.Maps, value ice.Map) {
	m.Info("%s %s", req.Method, req.URL)
	kit.Fetch(value[SPIDE_HEADER], func(key string, value string) {
		req.Header.Set(key, value)
		m.Logs(key, value)
	})
	kit.Fetch(value[SPIDE_COOKIE], func(key string, value string) {
		req.AddCookie(&http.Cookie{Name: key, Value: value})
		m.Logs(key, value)
	})
	list := kit.Simple(m.Optionv(SPIDE_COOKIE))
	for i := 0; i < len(list)-1; i += 2 {
		req.AddCookie(&http.Cookie{Name: list[i], Value: list[i+1]})
		m.Logs(list[i], list[i+1])
	}

	list = kit.Simple(m.Optionv(SPIDE_HEADER))
	for i := 0; i < len(list)-1; i += 2 {
		req.Header.Set(list[i], list[i+1])
		m.Logs(list[i], list[i+1])
	}
	for k, v := range head {
		req.Header.Set(k, v)
	}
	if req.Method == SPIDE_POST {
		m.Logs(req.Header.Get(ContentLength), req.Header.Get(ContentType))
	}
}
func _spide_send(m *ice.Message, name string, req *http.Request, timeout string) (*http.Response, error) {
	client := mdb.HashTarget(m, name, func() ice.Any { return &http.Client{Timeout: kit.Duration(timeout)} }).(*http.Client)
	return client.Do(req)
}
func _spide_save(m *ice.Message, cache, save, uri string, res *http.Response) {
	m.Debug("what %v", m.OptionCB(""))
	switch cache {
	case SPIDE_RAW:
		b, _ := ioutil.ReadAll(res.Body)
		if strings.HasPrefix(res.Header.Get(ContentType), ContentJSON) {
			m.Echo(kit.Formats(kit.UnMarshal(string(b))))
		} else {
			m.Echo(string(b))
		}

	case SPIDE_MSG:
		var data map[string][]string
		m.Assert(json.NewDecoder(res.Body).Decode(&data))
		for _, k := range data[ice.MSG_APPEND] {
			for i := range data[k] {
				m.Push(k, data[k][i])
			}
		}
		m.Resultv(data[ice.MSG_RESULT])

	case SPIDE_SAVE:
		if f, p, e := nfs.CreateFile(m, save); m.Assert(e) {
			defer f.Close()

			total := kit.Int(res.Header.Get(ContentLength)) + 1
			m.Debug("what %v", m.OptionCB(""))
			switch cb := m.OptionCB("").(type) {
			case func(int, int, int):
				count := 0
				nfs.CopyFile(m, f, res.Body, func(n int) {
					count += n
					m.Debug("what %v %v", n, count)
					cb(count, total, count*100/total)
				})
			default:
				if n, e := io.Copy(f, res.Body); m.Assert(e) {
					m.Logs(mdb.EXPORT, nfs.SIZE, n, nfs.FILE, p)
					m.Echo(p)
				}
			}
		}

	case SPIDE_CACHE:
		m.Optionv(RESPONSE, res)
		m.Cmdy(CACHE, DOWNLOAD, res.Header.Get(ContentType), uri)
		m.Echo(m.Append(mdb.DATA))

	default:
		b, _ := ioutil.ReadAll(res.Body)

		var data ice.Any
		if e := json.Unmarshal(b, &data); e != nil {
			m.Echo(string(b))
			break
		}

		m.Optionv(SPIDE_RES, data)
		data = kit.KeyValue(ice.Map{}, "", data)
		m.Push("", data)
	}
}

const (
	// 缓存方法
	SPIDE_RAW   = "raw"
	SPIDE_MSG   = "msg"
	SPIDE_SAVE  = "save"
	SPIDE_CACHE = "cache"

	// 请求方法
	SPIDE_GET    = "GET"
	SPIDE_PUT    = "PUT"
	SPIDE_POST   = "POST"
	SPIDE_DELETE = "DELETE"

	// 请求参数
	SPIDE_BODY = "body"
	SPIDE_FORM = "form"
	SPIDE_PART = "part"
	SPIDE_JSON = "json"
	SPIDE_DATA = "data"
	SPIDE_FILE = "file"

	// 响应数据
	SPIDE_RES = "content_data"

	// 请求头
	Bearer        = "Bearer"
	Authorization = "Authorization"
	ContentType   = "Content-Type"
	ContentLength = "Content-Length"
	UserAgent     = "User-Agent"
	Referer       = "Referer"
	Accept        = "Accept"

	// 数据格式
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

	CLIENT_NAME    = "client.name"
	CLIENT_METHOD  = "client.method"
	CLIENT_TIMEOUT = "client.timeout"
	CLIENT_URL     = "client.url"
	LOGHEADERS     = "logheaders"

	FORM     = "form"
	ADDRESS  = "address"
	REQUEST  = "request"
	RESPONSE = "response"

	MERGE  = "merge"
	SUBMIT = "submit"
)
const SPIDE = "spide"

func init() {
	Index.MergeCommands(ice.Commands{
		SPIDE: {Name: "spide client.name action=raw,msg,save,cache method=GET,PUT,POST,DELETE url format=form,part,json,data,file arg run create", Help: "蜘蛛侠", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				conf := m.Confm(cli.RUNTIME, cli.CONF)
				m.Cmd(SPIDE, mdb.CREATE, ice.OPS, kit.Select("http://127.0.0.1:9020", conf[cli.CTX_OPS]))
				m.Cmd(SPIDE, mdb.CREATE, ice.DEV, kit.Select("http://contexts.woa.com:80", conf[cli.CTX_DEV]))
				m.Cmd(SPIDE, mdb.CREATE, ice.SHY, kit.Select("https://shylinux.com:443", conf[cli.CTX_SHY]))
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, SPIDE, SUBMIT)
			}},
			mdb.CREATE: {Name: "create name address", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_spide_create(m, m.Option(mdb.NAME), m.Option(ADDRESS))
			}},
			MERGE: {Name: "merge name path", Help: "拼接", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(kit.MergeURL2(m.CmdAppend(SPIDE, arg[0], CLIENT_URL), arg[1], arg[2:]))
			}},
			SUBMIT: {Name: "submit dev pod path size cache", Help: "发布", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SPIDE, ice.DEV, SPIDE_RAW, m.Option(ice.DEV), SPIDE_PART, m.OptionSimple(ice.POD), nfs.PATH, ice.BIN_ICE_BIN, UPLOAD, "@"+ice.BIN_ICE_BIN)
			}},
		}, mdb.HashAction(mdb.SHORT, CLIENT_NAME, mdb.FIELD, "time,client.name,client.url", LOGHEADERS, ice.FALSE)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 2 || arg[0] == "" || (len(arg) > 3 && arg[3] == "") {
				mdb.HashSelect(m, kit.Slice(arg, 0, 1)...).Sort(CLIENT_NAME)
			} else {
				_spide_list(m, arg...)
			}
		}},
		SPIDE_GET: {Name: "GET url key value run", Help: "蜘蛛侠", Hand: func(m *ice.Message, arg ...string) {
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, SPIDE_GET, arg[0], arg[1:]))))
		}},
		SPIDE_PUT: {Name: "PUT url key value run", Help: "蜘蛛侠", Hand: func(m *ice.Message, arg ...string) {
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, SPIDE_PUT, arg[0], arg[1:]))))
		}},
		SPIDE_POST: {Name: "POST url key value run", Help: "蜘蛛侠", Hand: func(m *ice.Message, arg ...string) {
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, SPIDE_POST, arg[0], arg[1:]))))
		}},
		SPIDE_DELETE: {Name: "DELETE url key value run", Help: "蜘蛛侠", Hand: func(m *ice.Message, arg ...string) {
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, SPIDE_DELETE, arg[0], arg[1:]))))
		}},
	})
}

func SpideGet(m *ice.Message, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(SPIDE_GET, arg))
}
func SpidePut(m *ice.Message, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(SPIDE_PUT, arg))
}
func SpidePost(m *ice.Message, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(SPIDE_POST, arg))
}
func SpideDelete(m *ice.Message, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(SPIDE_DELETE, arg))
}
