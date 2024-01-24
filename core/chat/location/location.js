Volcanos(chat.ONIMPORT, {_init: function(can, msg) {},
	_layout_init: function(can, msg, cb) {
		can.db.list = {}, can.ui = can.onappend.layout(can), can.onimport._project(can)
		can.onmotion.hidden(can, can.ui.project)
		can.onmotion.hidden(can, can._action)
		// can.core.Item(can.ui.zone, function(key, item) { key == "favor" || item._legend.click() })
		if (can.user.isMobile) {
			can.page.style(can, can.ui.project, "z-index", 10, "position", "absolute", html.MAX_HEIGHT, can.ConfHeight()-120)
			can.page.style(can, can.ui.content, html.HEIGHT, can.ConfHeight(), html.WIDTH, can.ConfWidth())
			can.onmotion.hidden(can, can._action), can.onmotion.hidden(can, can._status)
		} else {
			can.ui.layout(can.ConfHeight(), can.ConfWidth())
		}
		msg.Option(ice.MSG_ACTION, ""), cb && cb(msg)
		if (msg.IsDetail()) { can.onaction.center(can, can.onimport._item(can, msg.TableDetail())) } else {
			msg.Table(function(item) { can.onimport._item(can, item) }), can.ui.zone.favor._total(msg.Length()), can.ui.zone.favor.toggle(true)
			var item = can.db.list[can.db.hash[0]]; item &&  item.click()
		}
	},
	_project: function(can) { can.onmotion.clear(can, can.ui.project), can.onimport.zone(can, [
		{name: "search"}, {name: "explore"}, {name: "direction"},
		{name: "favor", _menu: shy({"play": function(event, can, button) {
			can.core.Next(can.page.Select(can, can.ui.zone.favor._target, html.DIV_ITEM), function(item, next) {
				item.click(), can.onmotion.delay(can, next, 3000)
			}, function() { can.user.toastSuccess(can) })
		}})},
		{name: "district", _delay_init: function(target, zone) { can.onimport._province(can, target) }},
	], can.ui.project) },
	_item: function(can, item, target) { if (!item.latitude || !item.longitude) { return item }
		var _target = can.onimport.item(can, item, function(event) {
			can.onaction.center(can, item), can.ui.map.setZoom(can.Action("zoom", 16)), can.misc.SearchHash(can, item.hash)
			can.onimport.plugin(can, item) && can.onmotion.delay(can, function() {
				var ls = can.ui.map.getBounds().path
				ls && can.ui.map.setCenter([ls[1][0]-(ls[1][0]-ls[3][0])*3/8, ls[1][1]+(ls[3][1]-ls[1][1])*3/8])
			}, 500)
		}, function(event) {
			can.onaction.center(can, item), can.user.carteRight(event, can, {
				direction: function(event, button) { can.onaction[button](event, can, button, item) },
				favor: function(event, button) { can.onaction.create(can.request(event, item), can, button) },
				plugin: function(event, button) { can.user.input(can.request(event, item), can, [ctx.INDEX, ctx.ARGS], function(data) {
					item.extra = can.base.Copy(item.extra||{}, data), can.onimport.plugin(can, item)
					can.runAction(event, mdb.MODIFY, ["extra.index", data.index, "extra.args", data.args], function() {})
				}) },
				remove: function(event, button) { can.runAction(event, mdb.REMOVE, [mdb.HASH, item.hash], function() { can.page.Remove(can, _target) }) },
			})
		}, target||can.ui.zone.favor._target); can.db.list[item.hash] = _target, can.ui.zone.favor._total()
		return can.onimport._mark(can, item, _target), item
	}, _mark: function(can, item) {}, _style: function(can, style) {},
	plugin: function(can, item) { var extra = can.base.Obj(item.extra, {})
		if (!extra.index) { return can.onmotion.toggle(can, can.ui.profile, false) } can.onmotion.toggle(can, can.ui.profile, true)
		if (can.onmotion.cache(can, function() { return item.hash }, can.ui.profile)) { return true }
		can.onappend.plugin(can, {space: item.space, index: extra.index, args: extra.args}, function(sub) { item._plugin = sub
			sub.onaction._close = function() { can.onmotion.hidden(can, can.ui.profile) }
			sub.onexport.output = function() {
				var width = (can.user.isMobile? can.ConfWidth()-120: can.ConfWidth()/2)
				sub.onimport.size(sub, can.ConfHeight()/2, width, true)
				can.page.style(can, can.ui.profile, html.HEIGHT, can.ConfHeight()/2, html.WIDTH, width)
			}
		}, can.ui.profile)
		return true
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
	list: [
		["style", "normal", "light", "whitesmoke", "fresh", "macaron", "graffiti", "darkblue", "blue", "grey", "dark"],
		["feature", "point", "road"],
		{type: html.TEXT, name: "zoom", value: 16, range: [3, 21]},
		{type: html.TEXT, name: "pitch", value: 30, range: [0, 80, 5]},
		{type: html.TEXT, name: "rotation", value: 0, range: [0, 360, 10]},
		"current:button", "search", "explore", "direction", mdb.CREATE,
	], _trans: {current: "定位", favor: "收藏"},
	style: function(event, can, button, value) { can.onimport._style(can, value) },
	zoom: function(event, can) { can.ui.map.setZoom(can.Action("zoom")) },
	pitch: function(event, can) { can.ui.map.setPitch(can.Action("pitch")) },
	rotation: function(event, can) { can.ui.map.setRotation(can.Action("rotation")) },
	center: function(can, item) {},
})
Volcanos(chat.ONEXPORT, {list: ["province", "city", "district", "street", aaa.LONGITUDE, aaa.LATITUDE, "type", "name", "text", "space", "weather", "temperature", "humidity", "windPower"]})
