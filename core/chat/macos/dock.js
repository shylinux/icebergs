Volcanos(chat.ONIMPORT, {_init: function(can, msg) { can.onmotion.clear(can)
	can.onimport.icon(can, msg = msg||can._msg, can._output, function(target, item) { can.page.Modify(can, target, {
		onclick: function(event) { can.sup.onexport.record(can, item.name, mdb.NAME, item) },
		oncontextmenu: function(event) { var carte = can.user.carte(event, can, {
			remove: function() { item.name != "Finder" && can.runAction(event, mdb.REMOVE, [item.hash]) },
		}); can.page.style(can, carte._target, html.LEFT, event.x) },
	}) }), can.page.Append(can, can._output, [{view: "space"}])
	can.page.style(can, can._output, html.MAX_WIDTH, can.page.width())
	return
	var current = null, before, begin
	can.page.SelectChild(can, can._output, mdb.FOREACH, function(target) { target.draggable = true
		target.ondragstart = function() { current = target, can.page.style(can, target, "visibility", html.HIDDEN) }
		target.ondragenter = function(event) { before = target, begin = {x: event.x, y: event.y} }
		target.ondragover = function(event) { var offset = event.x - begin.x; can.page.style(can, target, {position: "relative", left: -offset}) }
		target.ondragend = function(event) { before && can.page.insertBefore(can, current, before)
			 can.page.SelectChild(can, can._output, mdb.FOREACH, function(target) { can.page.style(can, target, {position: "", left: "", visibility: html.VISIBLE}) })
		}
	})
}})
