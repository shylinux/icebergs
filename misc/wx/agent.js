Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) {
		if (!can.user.info.username && can.user.info._cmd != "web.chat.oauth.client" && msg.Option("oauth")) {
			return can.user.jumps(msg.Option("oauth"))
		}
		msg.Option(ice.MSG_ACTION, ""), can.require([msg.Option(nfs.SCRIPT)], function(can) {
			var debug = msg.isDebug() && can.user.info.userrole == aaa.TECH; debug && can.onmotion.toggle(can, can._fields, true)
			// debug = false
			can.onmotion.hidden(can, can._fields)
			wx.config({debug: debug, signature: msg.Option("signature"), timestamp: msg.Option("timestamp"), nonceStr: msg.Option("noncestr"), appId: msg.Option("appid"),
				openTagList: ["wx-open-subscribe"],
				jsApiList: can.core.Item({
					getLocation: function(can, cb) { wx.getLocation({type: "gcj02", success: function (res) {
						can.base.isFunc(cb) && cb({type: "gcj02", name: "当前位置", text: "当前位置", latitude: parseInt(res.latitude*100000), longitude: parseInt(res.longitude*100000) })
					} }) },
					openLocation: function(can, msg) { wx.openLocation({
						name: msg.Option(mdb.NAME), address: msg.Option(mdb.TEXT), infoUrl: msg.Option(web.LINK),
						longitude: parseFloat(msg.Option(aaa.LONGITUDE)), latitude: parseFloat(msg.Option(aaa.LATITUDE)), scale: msg.Option("scale")||14,
					}) },
					scanQRCode: function(can, cb) { wx.scanQRCode({needResult: cb? 1: 0, scanType: ["qrCode", "barCode"], success: function (res) {
						can.base.isFunc(cb) && cb(can.base.ParseJSON(res.resultStr))
					} }) },
					chooseImage: function(can, cb, count) { wx.chooseImage({count: count||9, sourceType: ["camera", "album"], sizeType: ["original", "compressed"], success: function (res) {
						can.base.isFunc(cb) && cb(res.localIds)
					} }) },
					uploadImage: function(can, id, cb) { wx.uploadImage({ localId: id, isShowProgressTips: 1, success: function (res) {
						can.base.isFunc(cb) && cb(res.serverId)
					} }) },
				}, function(key, value) { return can.user.agent[key] = value, key }).concat([
					"updateAppMessageShareData", "updateTimelineShareData",
				]),
			})
			wx.ready(function () {
				var p = can.misc.Resource(can, can.user.info.favicon); can.base.beginWith(p, "/") && (p = location.origin + p)
				wx.updateAppMessageShareData({title: can.user.info.titles, desc: "车管所业务代办", link: location.href, imgUrl: p})
				wx.updateTimelineShareData({title: can.user.info.titles, link: location.href, imgUrl: p})
			})
		})
	},
})
Volcanos(chat.ONACTION, {
	list: [
		"getLocation", "openLocation", "openAddress",
		"scanQRCode", "scanQRCode1",
		"chooseImage", "uploadImage", "previewImage",
		"updateAppMessageShareData", "updateTimelineShareData",
	],
	getLocation: function(event, can, button) {
		wx.getLocation({type: "gcj02", success: function (res) {
			can.run(event, [ctx.ACTION, button, mdb.NAME, "current", aaa.LONGITUDE, res.longitude.toFixed(6), aaa.LATITUDE, res.latitude.toFixed(6)], function() {})
			can._output.innerHTML = JSON.stringify(res)
		} })
	},
	openLocation: function(event, can, button) {
		wx.getLocation({type: "gcj02", success: function (res) { wx.openLocation(res) }})
	},
	openAddress: function(event, can, button) {
		wx.openAddress({success: function (res) {
			can._output.innerHTML = JSON.stringify(res)
			var userName = res.userName; // 收货人姓名
			var cityName = res.cityName; // 国标收货地址第二级地址（市）
			var provinceName = res.provinceName; // 国标收货地址第一级地址（省）
			var countryName = res.countryName; // 国标收货地址第三级地址（国家）
			var detailInfo = res.detailInfo; // 详细收货地址信息
			var nationalCode = res.nationalCode; // 收货地址国家码
			var postalCode = res.postalCode; // 邮编
			var telNumber = res.telNumber; // 收货人手机号码
		}})
	},
	scanQRCode: function(event, can, button) {
		wx.scanQRCode({needResult: 0, scanType: ["qrCode","barCode"]})
	},
	scanQRCode1: function(event, can, button) {
		wx.scanQRCode({needResult: 1, scanType: ["qrCode","barCode"], success: function (res) {
			can.run(event, [ctx.ACTION, button, mdb.TEXT, res.resultStr], function() {})
			can._output.innerHTML = res.resultStr
		} })
	},
	uploadImage: function(event, can, button) {
		wx.chooseImage({success: function (res) { can.core.List(res.localIds, function(item) {
			wx.uploadImage({localId: item, isShowProgressTips: 1, success: function (res) {
				var serverId = res.serverId;
				can._output.innerHTML = serverId
			} })
		}) }})
	},
	chooseImage: function(event, can, button) {
		wx.chooseImage({count: 9, sourceType: ["camera", "album"], sizeType: ["original", "compressed"], success: function (res) {
			can.page.Append(can, can._output, can.core.List(res.localIds, function(item) {
				return {img: item, style: {"max-width": can.ConfWidth()}}
			}))
		} })
	},
	previewImage: function(event, can, button) {
		wx.previewImage({urls: [
			"https://2021.shylinux.com/share/local/usr/icons/timg.png",
			"https://2021.shylinux.com/share/local/usr/icons/mall.png",
		]})
	},
	updateAppMessageShareData: function(event, can, button) {
		wx.updateAppMessageShareData({
			title: document.title, desc: "工具系统", link: location.href,
			imgUrl: "https://2021.shylinux.com/share/local/usr/icons/timg.png",
			success: function (res) { can._output.innerHTML = JSON.stringify(res) },
		})
	},
	updateTimelineShareData: function(event, can, button) {
		wx.updateTimelineShareData({
			title: document.title, desc: "工具系统", link: location.href,
			imgUrl: "https://2021.shylinux.com/share/local/usr/icons/timg.png",
			success: function (res) { can._output.innerHTML = JSON.stringify(res) },
		})
	},
})
