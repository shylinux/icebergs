Volcanos(chat.ONIMPORT, {
	_init: function(can, msg, cb) {
		can.db.hash[0] = can.db.hash[0]||(can.user.info.nodetype == web.WORKER? ice.OPS: ice.DEV)
		can.ui = can.onappend.layout(can), can.onimport._project(can, msg, [])
		cb && cb(msg), can.onappend._filter(can)
	},
	_project: function(can, msg, dev, target) {
		msg.Table(function(value) { if (value.type == web.WORKER) { return }
			value.nick = [{text: value.name}, value.exists == "true" && {text: ["●", "", "exists"]}]
			value._hash = dev.concat([value.name]).join(":"), value._select = can.base.beginWith(can.db.hash.join(":"), value._hash)
			value.icons = can.misc.Resource(can, value.icons||"usr/icons/icebergs.png", "", value.origin)
			can.onimport.itemlist(can, [value], function(event, item, show, target) {
				can.onimport.tabsCache(can, value, target, function(event) {
					can.run(event, [value.origin], function(msg) {
						can.onimport._project(can, msg, dev.concat([value.name]), target)
						can.onimport._content(can, msg), can.onappend._status(can, msg)
					})
				})
			}, null, target)
		})
	},
	_content: function(can, msg) {
		can.onimport.card(can, msg, null, function(value) { if (value.type == web.SERVER) { return true }
			value.icons = can.misc.Resource(can, value.icons||"usr/icons/icebergs.png", "", value.origin)
		}), can.onappend.style(can, "output card", can.ui.content), can.onimport.layout(can)
	},
})
