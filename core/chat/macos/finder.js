Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { can.ui = can.onappend.layout(can), msg.Table(function(value, index) {
		var item = can.onimport.item(can, value, function(event) { if (can.onmotion.cache(can, function() { return value.name }, can.ui.content)) { return }
			can.runActionCommand(event, value.index, [], function(msg) {
				switch (value.name) {
					case ".":
					case "applications": can.onimport.icons(can, msg, can.ui.content); break
					default: can.onappend.table(can, msg, null, can.ui.content)
				} can.onimport.layout(can)
			})
		}); index == 0 && item.click()
	}), can.onmotion.hidden(can, can.ui.project) },
	icons: function(can, msg, target) { can.onimport.icon(can, msg = msg||can._msg, target, function(target, item) { can.page.Modify(can, target, {
		onclick: function(event) { can.sup.onexport.record(can.sup, item.name, mdb.NAME, item) },
		oncontextmenu: function(event) { can.user.carteRight(event, can, {
			"add to desktop": function() { can.sup.onappend.desktop(item) },
			"add to dock": function() { can.sup.onappend.dock(item) },
		}, []) }, draggable: true, ondragstart: function(event) { window._drag_item = item },
	})}) },
	layout: function(can) { can.ui.layout(can.ConfHeight(), can.ConfWidth())
		var width = can.ConfWidth()-(can.ui? can.ui.project.offsetWidth: 0), margin = width%80/parseInt(width/80)/2
		can.page.SelectChild(can, can.ui.content, mdb.FOREACH, function(target) { can.page.style(can, target, html.MARGIN, margin) })
	},
})
