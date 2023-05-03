Volcanos(chat.ONIMPORT, {_init: function(can, msg) {
	function show(msg) { can.onmotion.clear(can)
		msg.Table(function(item) { can.page.Append(can, can._output, [{view: html.ITEM, list: [{view: html.ICON, list: [{img: can.misc.PathJoin(item.icon)}] }],
			onclick: function(event) { can.sup.onexport.record(can, item.name, mdb.NAME, item) },
			oncontextmenu: function(event) { var carte = can.user.carte(event, can, {
				remove: function() { item.name != "Finder" && can.runAction(event, mdb.REMOVE, [item.hash]) },
			}); can.page.style(can, carte._target, html.LEFT, event.x) },
		}]) }), can.page.Append(can, can._output, [{view: "space"}])
	} show(msg)
	return
	var current = null, before, begin
	can.page.SelectChild(can, can._output, mdb.FOREACH, function(target) { target.draggable = true
		target.ondragstart = function() { current = target, can.page.style(can, target, "visibility", html.HIDDEN) }
		target.ondragenter = function(event) { before = target, begin = {x: event.x, y: event.y} }
		target.ondragover = function(event) { var offset = event.x - begin.x; can.page.style(can, target, {position: "relative", left: -offset}) }
		target.ondragleave = function(event) { }
		target.ondragend = function(event) { before && can.page.insertBefore(can, current, before)
			 can.page.SelectChild(can, can._output, mdb.FOREACH, function(target) { can.page.style(can, target, {position: "", left: "", visibility: html.VISIBLE}) })
		}
	})
}})
