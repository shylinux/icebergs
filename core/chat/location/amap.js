Volcanos(chat.ONIMPORT, {
	_init: function(can, msg, cb) { window._AMapSecurityConfig = {serviceHost: location.origin+"/_AMapService"}, can.require([can.Conf(nfs.SCRIPT)], function() {
		can.require(["location.js"], function() { can.onimport._layout_init(can, msg, function() { cb && cb(msg)
			can.page.style(can, can.ui.content, html.WIDTH, can.ConfWidth(can._fields.offsetWidth)-can.ui.project.offsetWidth)
			can.onimport._content(can) })
		})
	}) },
	_content: function(can, item) {
		var map = new AMap.Map(can.ui.content, {viewMode: '3D', lang: can.getHeaderLanguage().split("-")[0], zoom: can.Action("zoom"), pitch: can.Action("pitch"), rotation: can.Action("rotation")}); can.ui.map = map
		can.onimport._style(can, can.getHeaderTheme()), can.onengine.listen(can, chat.ONTHEMECHANGE, function() { can.onimport._style(can, can.getHeaderTheme()) })
		map.on("moveend", function(event) { can.Action("pitch", map.getPitch().toFixed(2)), can.Action("rotation", map.getRotation()) })
		map.on("zoomend", function(event) { can.Action("zoom", map.getZoom().toFixed(2)) })
		map.on("click", function(event) { can.onexport.status(can, event.lnglat) })
		map.plugin(['AMap.Scale', 'AMap.ToolBar', 'AMap.MapType', 'AMap.Geolocation', 'AMap.Geocoder', 'AMap.Weather'], function() { 
			map.addControl(new AMap.Scale()), map.addControl(new AMap.ToolBar()), map.addControl(new AMap.MapType())
			map.addControl(can.ui.geolocation = new AMap.Geolocation({buttonPosition: 'RB', enableHighAccuracy: true})), can.ui.geocoder = new AMap.Geocoder({})
			AMap.event.addListener(can.ui.geolocation, 'complete', function() {
				can.ui.map.setZoom(can.Action("zoom", 16)), can.ui.map.setPitch(can.Action("pitch", 0)), can.ui.map.setRotation(can.Action("rotation", 0))
				can.user.toastSuccess(can), can.onmotion.delay(can, function() { can.onexport.status(can)
					var weather = new AMap.Weather(); weather.getLive(can.Status("city"), function(err, data) { can.Status(data) })
				}, 500)
			}), can.onmotion.delay(can, function() { can.onaction.current({}, can) })
			AMap.event.addListener(can.ui.geolocation, 'error', function(res) {
				can.user.toastFailure(can, res.message)
			})
		})
		map.add(can.ui.favor = new AMap.OverlayGroup())
		map.add(can.ui.marker = new AMap.Marker({position: [116.39, 39.9]}))
		map.add(can.ui.circle = new AMap.CircleMarker({
			center: new AMap.LngLat("116.403322", "39.920255"), radius: 100,
			strokeColor: "#F33", strokeWeight: 1, fillColor: "#ee2200", fillOpacity: 0.35,
		}))
	},
	_style: function(can, style) {
		style = {"light": "normal", "dark": "grey", "black": "blue", "white": "macaron", "silver": "grey", "blue": "graffiti", "red": "graffiti"}[style]||style
		can.ui.map.setMapStyle("amap://styles/"+can.Action("style", style))
		return style
	},
	_mark: function(can, item, target) {
		var p = new AMap.Marker({position: [parseFloat(item.longitude), parseFloat(item.latitude)]})
		p.on("click", function() { target.click() }), can.ui.favor.addOverlay(p)
	},
})
Volcanos(chat.ONACTION, {
	_trans: {
		current: "定位", favor: "收藏",
		input: {
			zoom: "缩放", pitch: "倾斜", rotation: "旋转",
			weather: "天气", temperature: "温度", humidity: "湿度", windPower: "风速",
		},
	},
	current: function(event, can) {
		can.user.toastProcess(can), can.ui.geolocation.getCurrentPosition()
	},
	feature: function(event, can, button, value) {
		if (value == "road") {
			can.ui.map.setFeatures(["bg", "road", "building"])
		} else {
			can.ui.map.setFeatures(["bg", "road", "building", "point"])
		}
	},
	center: function(can, item) { can.ui.marker.setTitle(item.name)
		can.ui.map.setCenter(new AMap.LngLat(parseFloat(item.longitude), parseFloat(item.latitude)))
		can.onmotion.delay(can, function() { can.onexport.status(can) }, 500)
	},
	direction: function(event, can, button, item) {
		can.user.isMobile && window.open(`https://uri.amap.com/marker?position=${item.longitude},${item.latitude}&name=${item.name||item.text}&callnative=1`)
	}
})
Volcanos(chat.ONEXPORT, {
	status: function(can, p) { p = p||can.ui.map.getCenter(), can.ui.marker.setPosition(p), can.ui.circle.setCenter(p)
		can.Status({longitude: p.getLng().toFixed(6), latitude: p.getLat().toFixed(6)}), can.ui.map.getCity(function(result) { can.Status(result) })
		can.ui.geocoder.getAddress(p, function(status, result) { can.Status(result.regeocode.addressComponent), can.Status({text: result.regeocode.formattedAddress}) })
	},
})
