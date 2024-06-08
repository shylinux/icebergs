package wx

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const (
	UNLIMIT            = "unlimit"
	QR_STR_SCENE       = "QR_STR_SCENE"
	QR_LIMIT_STR_SCENE = "QR_LIMIT_STR_SCENE"
	EXPIRE_SECONDS     = "expire_seconds"
)
const SCAN = "scan"

func init() {
	Index.MergeCommands(ice.Commands{
		SCAN: {Name: "scan access hash auto", Help: "桌牌", Meta: kit.Merge(Meta(), kit.Dict(ice.CTX_TRANS, kit.Dict(html.VALUE, kit.Dict(
			QR_LIMIT_STR_SCENE, "永久码", QR_STR_SCENE, "临时码", mdb.VALID, "有效", mdb.EXPIRED, "失效",
			"develop", "开发版", "trial", "体验版", "release", "发布版",
		)))), Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case SCENE:
					m.Cmdy(IDE).Cut(mdb.HASH, mdb.NAME, PAGES, ctx.INDEX).RenameAppend(mdb.HASH, arg[0])
				}
			}},
			mdb.CREATE: {Name: "create type*=QR_STR_SCENE,QR_LIMIT_STR_SCENE name*=1 text icons expire_seconds=3600 space index* args", Hand: func(m *ice.Message, arg ...string) {
				h := mdb.HashCreate(m.Spawn(), arg)
				res := SpidePost(m, QRCODE_CREATE, "action_name", m.Option(mdb.TYPE), "action_info.scene.scene_str", h, m.OptionSimple(EXPIRE_SECONDS))
				mdb.HashModify(m, mdb.HASH, h, mdb.LINK, kit.Value(res, web.URL), mdb.TIME, m.Time(kit.Format("%ss", kit.Select("60", m.Option(EXPIRE_SECONDS)))))
				m.EchoQRCode(kit.Format(kit.Value(res, web.URL)))
			}},
			UNLIMIT: {Name: "unlimit env*=develop,release,trial,develop scene* name", Help: "小程序码", Hand: func(m *ice.Message, arg ...string) {
				defer m.ProcessInner()
				scene := m.Option(SCENE)
				meta, info := "", m.Cmd(IDE, scene)
				if u := web.UserWeb(m); u.Scheme == ice.HTTP {
					if info.Append(tcp.WIFI) != "" {
						wifi := m.Cmd(tcp.WIFI, info.Append(tcp.WIFI))
						ls := kit.Split(tcp.PublishLocalhost(m, u.Hostname()), nfs.PT)
						meta = path.Join("w", kit.Format("%x", kit.Int(ls[3])), scene, wifi.Append(tcp.SSID), wifi.Append(aaa.PASSWORD))
					} else {
						meta = path.Join("h", tcp.PublishLocalhost(m, u.Host), scene)
					}
				} else {
					meta = path.Join("s", u.Host, scene)
				}
				msg := spidePost(m, WXACODE_UNLIMIT, web.SPIDE_DATA, kit.Format(kit.Dict(
					"env_version", m.Option(ENV), SCENE, meta, "page", info.Append(PAGES),
					html.WIDTH, kit.Int(kit.Select("360", "280", m.IsMobileUA())),
				)))
				switch kit.Select("", kit.Split(msg.Option(html.ContentType), "; "), 0) {
				case web.IMAGE_JPEG:
					image := m.Cmd(web.CACHE, web.WRITE, mdb.TYPE, web.IMAGE_JPEG, mdb.NAME, scene, kit.Dict(mdb.TEXT, msg.Result())).Append(mdb.HASH)
					mdb.HashSelects(m, mdb.HashCreate(m.Spawn(), m.OptionSimple(mdb.NAME), mdb.TEXT, meta, nfs.IMAGE, image, ctx.INDEX, m.Prefix(IDE), ctx.ARGS, scene, mdb.TYPE, m.Option(ENV)))
					m.EchoImages(web.SHARE_CACHE + m.Append(nfs.IMAGE))
				default:
					m.Echo(msg.Result())
				}
			}},
		}, mdb.ExportHashAction(mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,name,text,icons,space,index,args,type,image,link")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(ACCESS).PushAction().Action()
			} else if mdb.HashSelect(m, arg[1:]...); len(arg) == 1 {
				m.Table(func(value ice.Maps) {
					m.Push(mdb.STATUS, kit.Select(mdb.VALID, mdb.EXPIRED, value[mdb.TYPE] == QR_STR_SCENE && value[mdb.TIME] < m.Time()))
				}).Action(mdb.CREATE, UNLIMIT)
			} else {
				kit.If(m.Time() < m.Append(mdb.TIME), func() { m.PushQRCode(cli.QRCODE, m.Append(mdb.LINK)) })
			}
		}},
	})
}
