Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) {
		can.ui = can.onappend.layout(can), can.onimport._project(can, msg)
	},
	_project: function(can, msg) { var current = can.db.hash[0]||(can.user.info.nodetype == web.WORKER? ice.OPS: ice.DEV)
		msg.Table(function(value) { value._select = value.name == current
			can.onimport.item(can, value, function(event, item, target) {
				can.onimport.dream(event, can, item, target)
			})
		}), can.onmotion.delay(can, function() { can.onappend._filter(can) })
		can.onappend.style(can, "output card", can.ui.content), can.onmotion.delay(can, function() { can.onimport.layout(can) })
	},
	_content: function(can, msg, dev, target) { var list = []
		can.onimport.card(can, msg, null, function(value) {
			value.icons = can.misc.Resource(can, value.icons||"usr/icons/icebergs.png", "", value.origin)
			if (value.type == web.SERVER) { list.push(value); return true }
		})
		can.onimport.itemlist(can, list, function(event, value, target) { value.key = can.core.Keys(dev, value.name)
			can.onimport.dream(event, can, value, target)
		}, null, target)
	},
	dream: function(event, can, value, target) {
		if (can.onmotion.cache(can, function() { return value.key||value.name }, can.ui.content, can._status)) { return can.onimport.layout(can) }
		can.run(can.request(event, {_toast: ice.PROCESS}), [value.origin], function(msg) {
			can.onimport._content(can, msg, value.name, target), can.onappend._status(can, msg), can.onimport.layout(can)
		})
	},
}, [""])
