Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { can.ui = can.onappend.layout(can), can.onimport._project(can, msg)
		can.onappend.style(can, "output card", can.ui.content), can.onmotion.delay(can, function() { can.onimport.layout(can) })
		can.sup.onimport._field = function(sup, msg) { msg.Table(function(item) { can.onappend._plugin(can, item, {style: html.FLOAT}, function(sub) {}) }) }
	},
	_project: function(can, msg) { var select, current = can.sup.db._zone||can.db.hash[0]||ice.DEV
		msg.Table(function(value) {
			var _target = can.onimport.item(can, value, function(event) {
				can.isCmdMode()? can.misc.SearchHash(can, value.name): can.sup.db._zone = value.name
				if (can.onmotion.cache(can, function() { return value.name }, can.ui.content, can._status)) { return can.onimport.layout(can) }
				can.run(can.request(event, {_toast: ice.PROCESS}), [value.name], function(msg) { can.onappend._status(can, msg.Option(ice.MSG_STATUS)), can.onimport.card(can, msg), can.onimport.layout(can) })
			}, function() {}, can.ui.project); select = (value.name == current? _target: select)||_target
		}), select && select.click()
	},
	layout: function(can) { can.Action(html.FILTER) && can.onmotion.filter(can, can.Action(html.FILTER))
		can.ui.layout(can.ConfHeight(), can.ConfWidth()), can.onlayout.expand(can, can.ui.content)
		// can.user.isMobile && can.page.style(can, can.ui.content, html.HEIGHT, "")
	},
}, [""])
