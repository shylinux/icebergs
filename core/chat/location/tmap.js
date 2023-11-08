Volcanos(chat.ONIMPORT, {
	_init: function(can, msg, cb) { can.require([can.Conf(nfs.SCRIPT)], function() {
		can.require(["location.js"], function() { can.onimport._layout_init(can, msg, function() { cb && cb(msg)
			var res = {type: "unknown", longitude: 116.307480, latitude: 39.984120}, current = can.base.Obj(msg.Option(chat.LOCATION), {})
			res.province = current.province||current.regionName, res.city = current.city
			res.longitude = current.longitude||current.lon||res.longitude, res.latitude = current.latitude||current.lat||res.latitude
			res.name = current.name||"当前位置", res.text = current.text||"某某大街", res.ip = current.ip||current.query
			can.onimport._content(can, res), can.Status(res)
		}) })
	}) },
	_content: function(can, item) {
		can.ui.map = new TMap.Map(can.ui.content, {center: can.onimport.point(can, item), zoom: can.Action("zoom"), pitch: can.Action("pitch"), rotation: can.Action("rotation")})
		can.ui.map.on("zoom", function(event) { can.Action("zoom", can.ui.map.getZoom().toFixed(2)) })
		can.ui.map.on("pitch", function(event) { can.Action("pitch", can.ui.map.getPitch().toFixed(2)) })
		can.ui.map.on("rotate", function(event) { can.Action("rotation", can.ui.map.getRotation().toFixed(2)) })
		can.ui.map.on("click", function(event) { var point = can.onaction._point(event, can, {name: event.poi? event.poi.name: ""}); can.onaction.center(can, point) })
		can.ui.current = {item: item, info: can.onfigure.info(can, item), hover: can.onfigure.info(can, item),
			label: new TMap.MultiLabel({map: can.ui.map, geometries: [{id: "current", position: can.onimport.point(can, item), content: item.name}]}),
			marker: new TMap.MultiMarker({map: can.ui.map, geometries: [{id: "current", position: can.onimport.point(can, item)}]}),
			circle: can.onfigure.circle(can, item, 100),
		}
		can.onfigure._mark(can), can.onimport._mark(can, item)
		can.page.Select(can, can._target, "div.content>div", function(item) { can.page.style(can, item, {"z-index": 9}) })
	},
	_mark: function(can, item) {
		can.mark && can.mark.add({position: can.onimport.point(can, item), properties: item})
	},

	_search: function(can, keyword, i) { var p = can.onimport.point(can, can.ui.current.item)
		can.runAction(can.request({}, {_method: http.GET, "boundary": "region("+can.base.join([can.Status("city"), p.lat, p.lng], mdb.FS)+")", "page_index": i||1, "keyword": keyword}), "search", [], function(msg) {
			var res = can.base.Obj(msg.Result()); can.core.List(res.data, function(item) {
				can.onimport._item(can, can.onexport.point(can, item.location, {type: item.category, name: item.title, text: item.address}), can.ui.zone.search._target)
			})
		})
	},
	_explore: function(can, keyword, i) { var p = can.onimport.point(can, can.ui.current.item)
		can.runAction(can.request({}, {_method: http.GET, "boundary": "nearby("+can.base.join([p.lat, p.lng, "500"], mdb.FS)+")", "page_index": i||1, "keyword": keyword}), "search", [], function(msg) {
			var res = can.base.Obj(msg.Result()); can.core.List(res.data, function(item) {
				can.onimport._item(can, can.onexport.point(can, item.location, {type: item.category, name: item.title, text: item.address}), can.ui.zone.explore._target)
			})
		})
	},
	_list_result: function(can, msg, cb) { var res = can.base.Obj(msg.Result()); if (res.status) { can.user.toastFailure(can, res.message); return }
		return res && res.result && can.core.List(res.result[0], function(item) { item.name = item.name||item.fullname; return can.base.isFunc(cb)? cb(item): item })
	},
	_district: function(can, id, cb) { can.runAction(can.request({}, {_method: http.GET, id: id}), "district", [], cb) },
	_province: function(can, target) { can.onimport._district(can, "", function(msg) {
		can.onimport._list_result(can, msg, function(province) {
			can.onimport.item(can, province, function(event, _, show) { if (show === false) { return }
				can.onaction.center(can, can.onexport.point(can, province.location, {type: "province", name: province.name, text: province.fullname}))
				can.ui.map.setZoom(can.Action("zoom", 8)), can.Status({nation: "中国", province: province.fullname})
				show === true || can.onimport._city(can, province, event.target)
			}, function() {}, target)
		})
	}) },
	_city: function(can, province, target) { can.onimport._district(can, province.id, function(msg) {
		can.onimport.itemlist(can, can.onimport._list_result(can, msg), function(event, city, show) { if (show === false) { return }
			can.onaction.center(can, can.onexport.point(can, city.location, {type: "city", name: city.name, text: city.fullname}))
			can.ui.map.setZoom(can.Action("zoom", 12)), can.Status({nation: "中国", province: province.fullname, city: city.fullname})
			show === true || can.onimport._county(can, province, city, event.target)
		}, function() {}, target)
	}) },
	_county: function(can, province, city, target) { can.onimport._district(can, city.id, function(msg) {
		can.onimport.itemlist(can, can.onimport._list_result(can, msg), function(event, county) {
			can.onaction.center(can, can.onexport.point(can, county.location, {type: "county", name: city.name, text: county.fullname}))
			can.ui.map.setZoom(can.Action("zoom", 15)), can.Status({nation: "中国", province: province.fullname, city: city.fullname})
		}, function() {}, target)
	}) },
	point: function(can, item) { return new TMap.LatLng(item.latitude, item.longitude) },
})
Volcanos(chat.ONACTION, {
	_trans: {
		current: "定位", favor: "收藏",
		input: {
			zoom: "缩放", pitch: "倾斜", rotation: "旋转",
		},
	},
	_point: function(event, can, item) { var rect = can.ui.content.getBoundingClientRect()
		return can.base.Copy({left: rect.left+event.point.x, bottom: rect.top+event.point.y, latitude: event.latLng.lat.toFixed(6), longitude: event.latLng.lng.toFixed(6)}, item, true)
	},
	current: function(event, can) { can.user.toastProcess(can)
		can.user.agent.getLocation(can, function(res) { can.user.toastSuccess(can)
			res.type = "current", can.onaction.center(can, res)
			can.ui.map.setZoom(can.Action("zoom", 16)), can.ui.map.setPitch(can.Action("pitch", 0)), can.ui.map.setRotation(can.Action("rotation", 0))
		})
	},
	center: function(can, item) {
		var point = can.onimport.point(can, item); can.ui.map.setCenter(point); if (!item.name) { return }
		can.ui.current.item = item, can.Status(mdb.NAME, ""), can.Status(mdb.TEXT, ""), can.Status(item), can.Status({latitude: point.lat, longitude: point.lng})
		can.ui.current.info.setPosition(point), can.ui.current.info.setContent((item.name||"")+"<br/>"+(item.text||""))
		can.ui.current.label.updateGeometries([{id: "current", position: point, content: item.name}])
		can.ui.current.marker.updateGeometries([{id: "current", position: point}])
		can.ui.current.circle.setGeometries([{center: point, radius: 300}])
	},

	search: function(event, can, button) { can.onmotion.clear(can, can.ui.zone.search._target)
		can.user.input(event, can, ["keyword"], function(list) {
			for (var i = 1; i < 6; i++) {
				can.onimport._search(can, list[0], i)
			}
		})
	},
	explore: function(event, can, button) { can.onmotion.clear(can, can.ui.zone.explore._target)
		can.user.input(event, can, ["keyword"], function(list) {
			for (var i = 1; i < 6; i++) {
				can.onimport._explore(can, list[0], i)
			}
		})
	},
	direction: function(event, can, button) { var p = can.ui.map.getCenter(); can.onmotion.clear(can, can.ui.zone.direction._target)
		can.user.input(event, can, [["type", "driving", "walking", "bicycling", "transit"]], function(list) {
			var from = can.onimport.point(can, can._current), to = can.onimport.point(can, can.ui.current.item)
			var msg = can.request({}, {_method: http.GET, type: list[0], "from": can.base.join([from.lat, from.lng], mdb.FS), "to": can.base.join([to.lat, to.lng], mdb.FS)})
			can.runAction(msg._event, button, [], function(msg) { var res = can.base.Obj(msg.Result()), route = res.result.routes[0]
				var coors = route.polyline, pl = [], kr = 1000000
				for (var i = 2; i < coors.length; i++) { coors[i] = Number(coors[i - 2]) + Number(coors[i]) / kr }
				for (var i = 0; i < coors.length; i += 2) { pl.push(new TMap.LatLng(coors[i], coors[i+1])) }
				can.onfigure._polyline(can, pl)
				can.core.List(route.steps, function(item) { var i = item.polyline_idx[0]
					can.onimport._item(can, can.onexport.point(can, {lat: coors[i], lng: coors[i+1]}, {type: item.category, name: item.instruction, text: item.act_desc}), can.ui.zone.direction._target)
				}), can.user.toastProcess(can, "distance: "+route.distance+" duration: "+route.duration)
			})
		})
	},
	create: function(event, can) { can.request(event, can.ui.current.item)
		can.user.input(event, can, can.core.Split("type,name,text"), function(args) { var p = can.onexport.center(can)
			can.runAction(event, mdb.CREATE, args.concat("latitude", p.latitude, "longitude", p.longitude), function(msg) {
				can.onimport._item(can, can.base.Copy(p, {name: msg.Option(mdb.NAME), text: msg.Option(mdb.TEXT)}))
			}, true)
		})
	},
})
Volcanos(chat.ONEXPORT, {
	point: function(can, point, item) { return can.base.Copy({latitude: point.lat, longitude: point.lng}, item, true) },
	center: function(can) { return can.onexport.point(can, can.ui.map.getCenter()) },
	current: function(can) { var p = can.onexport.center(can); p.latitude, p.longitude; can.Status(p)
		can.ui.current.marker.updateGeometries([{id: "current", position: can.ui.map.getCenter()}])
		can.ui.current.label.updateGeometries([{id: "current", position: can.ui.map.getCenter()}])
		can.ui.current.circle.setGeometries([{center: can.ui.map.getCenter(), radius: 300}])
	},
	hover: function(can, item) {
		can.ui.current.hover.setPosition(can.onimport.point(can, item))
		can.ui.current.hover.setContent(item.name+"<br/>"+item.text)
	},
})
Volcanos(chat.ONFIGURE, {
	circle: function(can, item, radius) {
		return new TMap.MultiCircle({
			map: can.ui.map, styles: {circle: new TMap.CircleStyle({color: 'rgba(41,91,255,0.16)', borderColor: 'rgba(41,91,255,1)', borderWidth: 2, showBorder: true})},
			geometries: [{styleId: 'circle', center: can.onimport.point(can, item), radius: radius||300}],
		})
	},
	info: function(can, item) {
		return new TMap.InfoWindow({map: can.ui.map,
			position: can.onimport.point(can, item), offset: {x: 0, y: -32},
			content: (item.name||"")+"<br/>"+(item.text||""),
		})
	},
	_mark: function(can, msg) { can.mark = new TMap.MultiMarker({map: can.ui.map})
		can.mark.on("click", function(event) { if (!event.geometry) { return }
			var item = event.geometry.properties; can.db.list[item.hash].click()
		})
		can.mark.on("hover", function(event) { if (!event.geometry) { return }
			var item = event.geometry.properties; can.onexport.hover(can, item)
		})
	},
	_polyline: function(can, path) { return new TMap.MultiPolyline({
		map: can.ui.map, styles: {
			'style_blue': new TMap.PolylineStyle({
				'color': '#3777FF', //线填充色
				'width': 6, //折线宽度
				'borderWidth': 5, //边线宽度
				'borderColor': '#FFF', //边线颜色
				'lineCap': 'butt' //线端头方式
			}),
			'style_red': new TMap.PolylineStyle({
				'color': '#CC0000', //线填充色
				'width': 6, //折线宽度
				'borderWidth': 5, //边线宽度
				'borderColor': '#CCC', //边线颜色
				'lineCap': 'round' //线端头方式
			})
		}, geometries: [{'styleId': 'style_blue', 'paths': path}]
	}) },
	_move: function(can) {
		can.mark.add({id: 'car', styleId: 'car-down', position: new TMap.LatLng(39.98481500648338, 116.30571126937866)})
		can.mark.moveAlong({"car": {path: [
			new TMap.LatLng(39.98481500648338, 116.30571126937866),
			new TMap.LatLng(39.982266575222155, 116.30596876144409),
			new TMap.LatLng(39.982348784165886, 116.3111400604248),
			new TMap.LatLng(39.978813710266024, 116.3111400604248),
			new TMap.LatLng(39.978813710266024, 116.31699800491333)
		], speed: 70}}, {autoRotation:true})
	},
})
