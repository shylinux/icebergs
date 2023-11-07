Volcanos(chat.ONIMPORT, {_init: function(can, msg) {},
	_layout_init: function(can, msg, cb) {
		can.ui = can.onappend.layout(can); if (can.user.isMobile) {
			can.page.style(can, can.ui.project, "z-index", 10, "position", "absolute")
			can.page.style(can, can.ui.content, html.HEIGHT, can.ConfHeight(), html.WIDTH, can.ConfWidth())
			can.page.Select(can, can._action, "div.item.text", function(target) { can.onmotion.hidden(can, target) })
			can.onmotion.hidden(can, can._status)
		} else {
			can.ui.layout(can.ConfHeight(), can.ConfWidth())
		}
		can.onimport._project(can), can.db.list = {}
		can.user.isMobile && can.core.Item(can.ui.zone, function(key, item) { key == "favor" || item._legend.click() })
		can.user.isMobile && can.onmotion.hidden(can, can._action)
		msg.Option(ice.MSG_ACTION, ""), cb && cb(msg)
		if (msg.IsDetail()) {
			can.onaction.center(can, can._current = can.onimport._item(can, msg.TableDetail()))
		} else {
			msg.Table(function(item) { can.onimport._item(can, item) }), can.ui.zone.favor._total(msg.Length())
			var item = can.db.list[can.db.hash[0]]; item? item.click(): can.user.agent.getLocation(can, function(res) { res.type = "current", can.onaction.center(can, can._current = res) })
		}
	},
	_project: function(can) { can.onmotion.clear(can, can.ui.project), can.onimport.zone(can, [
		{name: "explore"}, {name: "search"}, {name: "direction"},
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
			can.onimport.plugin(can, item) && can.onmotion.delay(can, function() { var ls = can.ui.map.getBounds().path
				can.ui.map.setCenter([ls[1][0]-(ls[1][0]-ls[3][0])*3/8, ls[1][1]+(ls[3][1]-ls[1][1])*3/8])
			}, 500)
		}, function(event) {
			can.onaction.center(can, item), can.user.carteRight(event, can, {
				plugin: function(event, button) { can.user.input(can.request(event, item), can, [ctx.INDEX, ctx.ARGS], function(data) {
					item.extra = can.base.Copy(item.extra||{}, data), can.onimport.plugin(can, item)
					can.runAction(event, mdb.MODIFY, ["extra.index", data.index, "extra.args", data.args], function() {})
				}) },
				favor: function(event) { can.request(event, item), can.onaction.create(event, can) },
				direction: function(event, button) { can.onaction.center(can, item), can.onaction[button](event, can, button) },
				remove: function(event, button) { can.runAction(event, mdb.REMOVE, [mdb.HASH, item.hash], function() { can.page.Remove(can, _target) }) },
			})
		}, target||can.ui.zone.favor._target); can.db.list[item.hash] = _target, can.ui.zone.favor._total()
		return can.onimport._mark(can, item, _target), item
	},
	_mark: function(can, item) {},
	_style: function(can, style) {},
	plugin: function(can, item) { var extra = can.base.Obj(item.extra, {}); can.onmotion.toggle(can, can.ui.profile, true)
		if (can.onmotion.cache(can, function() { return item.hash }, can.ui.profile)) { return true }
		if (!extra.index) { return can.onmotion.toggle(can, can.ui.profile, false) }
		can.onappend.plugin(can, {space: item.space, index: extra.index, args: extra.args}, function(sub) { item._plugin = sub
			sub.onaction._close = function() { can.onmotion.hidden(can, can.ui.profile) }
			sub.onexport.output = function() { sub.onimport.size(sub, can.ConfHeight()/2, can.ConfWidth()/2, true)
				can.page.style(can, can.ui.profile, html.HEIGHT, can.ConfHeight()/2, html.WIDTH, can.ConfWidth()/2)
			}
		}, can.ui.profile)
		return true
	},
})
Volcanos(chat.ONACTION, {
	list: [
		["style", "normal", "light", "whitesmoke", "fresh", "macaron", "graffiti", "darkblue", "blue", "grey", "dark"],
		["feature", "point", "road"],
		{type: html.TEXT, name: "zoom", value: 16, range: [3, 21]},
		{type: html.TEXT, name: "pitch", value: 30, range: [0, 80, 5]},
		{type: html.TEXT, name: "rotation", value: 0, range: [0, 360, 10]},
		"current:button", "explore", "search", "direction", mdb.CREATE,
	], _trans: {current: "定位", favor: "收藏"},
	style: function(event, can, button, value) { can.onimport._style(can, value) },
	zoom: function(event, can) { can.ui.map.setZoom(can.Action("zoom")) },
	pitch: function(event, can) { can.ui.map.setPitch(can.Action("pitch")) },
	rotation: function(event, can) { can.ui.map.setRotation(can.Action("rotation")) },
})
Volcanos(chat.ONEXPORT, {list: ["province", "city", "district", "street", "longitude", "latitude", "type", "name", "text", "space", "weather", "temperature", "humidity", "windPower"]})
