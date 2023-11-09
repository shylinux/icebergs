Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { msg.Option(ice.MSG_ACTION, ""), can.require(["https://res.wx.qq.com/open/js/jweixin-1.6.0.js"], function(can) {
		wx.config({debug: msg.Option("debug") == ice.TRUE, signature: msg.Option("signature"), timestamp: msg.Option("timestamp"), nonceStr: msg.Option("noncestr"), appId: msg.Option("appid"),
			jsApiList: can.core.Item({
				scanQRCode: function(can, cb) { wx.scanQRCode({needResult: cb? 1: 0, scanType: ["qrCode","barCode"], success: function (res) {
					can.base.isFunc(cb) && cb(can.base.ParseJSON(res.resultStr))
				} }) },
				getLocation: function(can, cb) { wx.getLocation({type: "gcj02", success: function (res) {
					can.base.isFunc(cb) && cb({type: "gcj02", name: "当前位置", text: "当前位置", latitude: parseInt(res.latitude*100000), longitude: parseInt(res.longitude*100000) })
				} }) },
				openLocation: function(can, msg) { wx.openLocation({
					name: msg.Option(mdb.NAME), address: msg.Option(mdb.TEXT), infoUrl: msg.Option(mdb.LINK),
					longitude: parseFloat(msg.Option("longitude")), latitude: parseFloat(msg.Option("latitude")), scale: msg.Option("scale")||14,
				}) },
				chooseImage: function(can, cb, count) { wx.chooseImage({count: count||9, sizeType: ['original', 'compressed'], sourceType: ['album', 'camera'], success: function (res) {
					can.base.isFunc(cb) && cb(res.localIds)
				} }) },
			}, function(key, value) { return can.user.agent[key] = value, key }).concat([
				// "uploadImage", "previewImage",
				// "updateAppMessageShareData", "updateTimelineShareData",
			]),
		})
	}) },
})
Volcanos(chat.ONACTION, {list: [
	"scanQRCode", "scanQRCode1", "getLocation", "openLocation",
	"uploadImage", "chooseImage", "previewImage",
	"updateAppMessageShareData", "updateTimelineShareData",
	"openAddress",
],
	scanQRCode: function(event, can, button) {
		wx.scanQRCode({needResult: 0, scanType: ["qrCode","barCode"]})
	},
	scanQRCode1: function(event, can, button) {
		wx.scanQRCode({needResult: 1, scanType: ["qrCode","barCode"], success: function (res) {
			can.run(event, [ctx.ACTION, button, mdb.TEXT, res.resultStr], function() {})
			can._output.innerHTML = res.resultStr
		} })
	},
	getLocation: function(event, can, button) {
		wx.getLocation({type: "gcj02", success: function (res) {
			can.run(event, [ctx.ACTION, button, mdb.NAME, "current", "longitude", res.longitude.toFixed(6), "latitude", res.latitude.toFixed(6)], function() {})
			can._output.innerHTML = JSON.stringify(res)
		} })
	},
	openLocation: function(event, can, button) {
		wx.getLocation({type: "gcj02", success: function (res) { wx.openLocation(res) }})
	},
	uploadImage: function(event, can, button) {
		wx.chooseImage({success: function (res) {
			can.core.List(res.localIds, function(item) {
				wx.uploadImage({
					localId: item, isShowProgressTips: 1,
					success: function (res) {
						var serverId = res.serverId;
						can._output.innerHTML = serverId
					}
				})

			})
		}})
	},
	chooseImage: function(event, can, button) {
		wx.chooseImage({
			// count: 9, sourceType: ['album', 'camera'], sizeType: ['original', 'compressed'],
			success: function (res) {
				can.page.Append(can, can._output, can.core.List(res.localIds, function(item) {
					return {img: item, style: {"max-width": can.ConfWidth()}}
				}))
			}
		})
	},
	previewImage: function(event, can, button) {
		wx.previewImage({
			urls: [
				'https://2021.shylinux.com/share/local/usr/icons/timg.png',
				"https://2021.shylinux.com/share/local/usr/icons/mall.png",
			],
		})
	},
	updateAppMessageShareData: function(event, can, button) {
		wx.updateAppMessageShareData({ 
			title: document.title, desc: "工具系统", link: location.href,
			imgUrl: 'https://2021.shylinux.com/share/local/usr/icons/timg.png',
			success: function (res) { can._output.innerHTML = JSON.stringify(res) },
		})
	},
	updateTimelineShareData: function(event, can, button) {
		wx.updateTimelineShareData({ 
			title: document.title, desc: "工具系统", link: location.href,
			imgUrl: 'https://2021.shylinux.com/share/local/usr/icons/timg.png',
			success: function (res) { can._output.innerHTML = JSON.stringify(res) },
		})
	},
	openAddress: function(event, can, button) {
		wx.openAddress({
			success: function (res) {
				can._output.innerHTML = JSON.stringify(res)

				var userName = res.userName; // 收货人姓名
				var postalCode = res.postalCode; // 邮编
				var provinceName = res.provinceName; // 国标收货地址第一级地址（省）
				var cityName = res.cityName; // 国标收货地址第二级地址（市）
				var countryName = res.countryName; // 国标收货地址第三级地址（国家）
				var detailInfo = res.detailInfo; // 详细收货地址信息
				var nationalCode = res.nationalCode; // 收货地址国家码
				var telNumber = res.telNumber; // 收货人手机号码
			}
		})
	},
})
