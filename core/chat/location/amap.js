Volcanos(chat.ONIMPORT, {
	_init: function(can, msg, cb) { window._AMapSecurityConfig = {serviceHost: location.origin+"/_AMapService"}, can.require([can.Conf(nfs.SCRIPT)], function() {
		can.require(["location.js"], function() { can.onimport._layout_init(can, msg, function() { cb && cb(msg)
			can.page.style(can, can.ui.content, html.WIDTH, can.ConfWidth(can._fields.offsetWidth)-can.ui.project.offsetWidth)
			can.onimport._content(can)
		}) })
	}) },
	_content: function(can, item) {
		var map = new AMap.Map(can.ui.content, {viewMode: '3D', lang: can.getHeaderLanguage().split("-")[0], zoom: can.Action("zoom"), pitch: can.Action("pitch"), rotation: can.Action("rotation")})
		can.ui.map = map, can.onimport._style(can, can.getHeaderTheme()), can.onengine.listen(can, chat.ONTHEMECHANGE, function() { can.onimport._style(can, can.getHeaderTheme()) })
		map.on("moveend", function(event) { can.Action("pitch", map.getPitch().toFixed(2)), can.Action("rotation", map.getRotation()) })
		map.on("zoomend", function(event) { can.Action("zoom", map.getZoom().toFixed(2)) })
		map.on("click", function(event) { can.onexport.status(can, event.lnglat) })
		map.plugin([
			'AMap.Scale', 'AMap.ToolBar', 'AMap.MapType', 'AMap.Geolocation', 'AMap.Geocoder', 'AMap.Weather',
			'AMap.Autocomplete',
			'AMap.DistrictLayer',
		], function() {
			map.addControl(new AMap.Scale()), map.addControl(new AMap.ToolBar()), map.addControl(new AMap.MapType())
			map.addControl(can.ui.geolocation = new AMap.Geolocation({buttonPosition: 'RB', enableHighAccuracy: true})), can.ui.geocoder = new AMap.Geocoder({})
			AMap.event.addListener(can.ui.geolocation, 'error', function(res) { can.user.toastFailure(can, res.message) })
			AMap.event.addListener(can.ui.geolocation, 'complete', function() {
				can.ui.map.setZoom(can.Action("zoom", 16)), can.ui.map.setPitch(can.Action("pitch", 0)), can.ui.map.setRotation(can.Action("rotation", 0))
				can.user.toastSuccess(can), can.onmotion.delay(can, function() { can.onexport.status(can)
					var weather = new AMap.Weather(); weather.getLive(can.Status("city"), function(err, data) { can.Status(data) })
				}, 500)
			}), can.onmotion.delay(can, function() { can.onaction.current({}, can) })
		})
		can.ui.layer = {}
		map.add(can.ui.layer.favor = new AMap.OverlayGroup())
		map.add(can.ui.layer.search = new AMap.OverlayGroup())
		map.add(can.ui.marker = new AMap.Marker({position: [116.39, 39.9]}))
		map.add(can.ui.circle = new AMap.CircleMarker({
			center: new AMap.LngLat("116.403322", "39.920255"), radius: 100,
			strokeColor: "#F33", strokeWeight: 1, fillColor: "#ee2200", fillOpacity: 0.35,
		}))
	},
	_district: function(can, city) { can.ui._district = can.ui._district||{}; if (can.ui._district[city]) { return }
		can.ui.map.add(can.ui._district[city] = new AMap.DistrictLayer.Province({
			zIndex: 12, depth: 2, adcode: [city],
			styles:{'fill': "transparent", 'city-stroke': 'blue', 'province-stroke': 'red'}
		}))
		if (can.ui._district.length > 1) { return }
		can.ui.map.add(new AMap.DistrictLayer.Country({
			zIndex: 10, depth: 2, SOC: 'CHN',
			styles:{'fill': 'transparent', 'city-stroke': 'blue', 'province-stroke': 'red'}
		}))
	},
	_style: function(can, style) {
		style = {"light": "normal", "dark": "grey", "black": "blue", "white": "macaron", "silver": "grey", "blue": "graffiti", "red": "graffiti"}[style]||style
		can.ui.map.setMapStyle("amap://styles/"+can.Action("style", style))
		return style
	},
	_mark: function(can, item, target, layer) { layer = layer||can.ui.layer.favor
		var p = new AMap.Marker({position: [parseFloat(item.longitude), parseFloat(item.latitude)], label: item.label && {content: item.label, direction: "bottom"}, title: item.name})
		p.on("click", function() { target.click() }), layer.addOverlay(p)
	},
})
Volcanos(chat.ONACTION, {
	_trans: {
		favor: "收藏",
		explore: "周边",
		direction: "导航",
		district: "行政",
		current: "定位",
		input: {
			district: "区域", street: "街道",
			zoom: "缩放", pitch: "倾斜", rotation: "旋转",
			weather: "天气", temperature: "温度", humidity: "湿度", windPower: "风速",
		},
		icons: {
			current: "bi bi-pin-map",
		},
	},
	feature: function(event, can, button, value) {
		if (value == "road") {
			can.ui.map.setFeatures(["bg", "road", "building"])
		} else {
			can.ui.map.setFeatures(["bg", "road", "building", "point"])
		}
	},
	current: function(event, can) {
		can.user.toastProcess(can), can.ui.geolocation.getCurrentPosition()
	},
	search: function(event, can) {
		can.user.input(event, can, ["keyword"], function(data) {
			can.ui.autoComplete = new AMap.Autocomplete({city: can.ui.city.citycode})
			can.ui.autoComplete.search(data.keyword, function(status, result) { var _select
				can.core.List(result.tips, function(value) { value = {name: value.name, label: value.name, longitude: value.location.lng, latitude: value.location.lat}
					var item = can.onimport._item(can, value, can.ui.zone.search._target, can.ui.layer.search)
					_select = _select||item._target
				}), can.ui.zone.favor.toggle(false), can.ui.zone.search.toggle(true), _select.click()
				// can.ui.map.setFitView([can.ui.layer.search], true, [10, 10, 10, 10], 15)
			})
		})
	},
	direction: function(event, can, button, item) {
		can.user.isMobile && window.open(`https://uri.amap.com/marker?position=${item.longitude},${item.latitude}&name=${item.name||item.text}&callnative=1`)
	},
	center: function(can, item) { can.ui.marker.setTitle(item.name)
		can.ui.map.setCenter(new AMap.LngLat(parseFloat(item.longitude), parseFloat(item.latitude)))
		can.onmotion.delay(can, function() { can.onexport.status(can) }, 500)
	},
})
Volcanos(chat.ONEXPORT, {
	status: function(can, p) { p = p||can.ui.map.getCenter(), can.ui.marker.setPosition(p), can.ui.circle.setCenter(p)
		can.Status({longitude: p.getLng().toFixed(6), latitude: p.getLat().toFixed(6)}), can.ui.map.getCity(function(result) {
			can.ui.city = result, can.Status(result)
			can.onimport._district(can, can.ui.map.getAdcode())
		})
		can.ui.geocoder.getAddress(p, function(status, result) {
			var info = result.regeocode.addressComponent, text = result.regeocode.formattedAddress
			text = can.base.trimPrefix(text, info.province, info.city, info.district, info.township)
			can.Status(info), can.Status({text: text})
		})
	},
})
