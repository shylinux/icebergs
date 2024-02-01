Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { can.ui = can.onappend.layout(can), can.onimport.__project(can, msg) },
	__project: function(can, msg) { var select, current = can.sup.db._zone||can.db.hash[0]||ice.DEV
		msg.Table(function(value) {
			var _target = can.onimport.item(can, value, function(event) { can.isCmdMode()? can.misc.SearchHash(can, value.name): can.sup.db._zone = value.name
				if (can.onmotion.cache(can, function() { return value.name }, can.ui.content, can._status)) { return can.onimport.layout(can) }
				can.run(can.request(event, {_toast: ice.PROCESS}), [value.name], function(msg) {
					can.onimport.__content(can, msg), can.onappend._status(can, msg), can.onimport.layout(can)
				})
			}, function() {}, can.ui.project); select = (value.name == current? _target: select)||_target
		}), select && select.click(), can.onmotion.orderShow(can, can.ui.project)
		can.onappend.style(can, "output card", can.ui.content), can.onmotion.delay(can, function() { can.onimport.layout(can) })
	}, __content: function(can, msg) { can.onimport.card(can, msg) },
}, [""])
