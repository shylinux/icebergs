Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { can.page.style(can, can._output, html.MAX_WIDTH, can.page.width())
		can.onimport.icon(can, msg = msg||can._msg, can._output, function(target, item) {
			can.page.Modify(can, target, {
				style: {"max-width": (can.page.width()-15)/msg.Length()},
				onclick: function(event) { can.sup.onexport.record(can, item.name, mdb.NAME, item) },
				oncontextmenu: function(event) { var carte = can.user.carte(event, can, {
					remove: function() { item.name != "Finder" && can.runAction(event, mdb.REMOVE, [item.hash]) },
				}); can.page.style(can, carte._target, html.LEFT, event.x) },
			})
		}), can.page.Append(can, can._output, [{view: "space"}])
	},
})
