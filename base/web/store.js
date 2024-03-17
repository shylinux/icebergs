Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { can.ui = can.onappend.layout(can), can.onimport._project(can, msg) },
	_project: function(can, msg) { var select, current = can.sup.db._zone||can.db.hash[0]||(can.user.info.nodetype == web.WORKER? ice.OPS: ice.DEV)
		msg.Table(function(value) {
			var _target = can.onimport.item(can, value, function(event, value) {
				can.onimport.dream(event, can, value, _target)
			}, null, can.ui.project); select = (value.name == current? _target: select)||_target
		}), select && select.click(), can.onmotion.orderShow(can, can.ui.project)
		can.onappend.style(can, "output card", can.ui.content), can.onmotion.delay(can, function() { can.onimport.layout(can) })
	},
	_content: function(can, msg, dev, target) { var list = []
		can.onimport.card(can, msg, null, function(value) { value.icons = can.misc.Resource(can, value.icons, "", value.origin); if (value.type == web.SERVER) { list.push(value); return true } })
		can.onimport.itemlist(can, list, function(event, value) {
			value.key = can.core.Keys(dev, value.name)
			can.onimport.dream(event, can, value, event.currentTarget)
		}, null, target)
	},
	dream: function(event, can, value, target) { can.isCmdMode()? can.misc.SearchHash(can, value.name): can.sup.db._zone = value.name
		can.page.Select(can, can.ui.project, html.DIV_ITEM, function(_target) { can.page.ClassList.set(can, _target, html.SELECT, _target == target) })
		if (can.onmotion.cache(can, function() { return value.key||value.name }, can.ui.content, can._status)) { return can.onimport.layout(can) }
		can.run(can.request(event, {_toast: ice.PROCESS}), [value.origin], function(msg) {
			can.onimport._content(can, msg, value.name, target), can.onappend._status(can, msg), can.onimport.layout(can)
		})
	},
}, [""])
