Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { can.ui = can.onappend.layout(can), can.onimport.__project(can, msg) },
	_layout: function(can) {
		can.page.style(can, can.ui.content, html.HEIGHT, can._output.style[html.HEIGHT], html.MAX_HEIGHT, can._output.style[html.MAX_HEIGHT])
		can.page.style(can, can.ui.project, html.HEIGHT, can.ui.content.offsetHeight+can.ui.display.offsetHeight)
		can.onlayout.expand(can, can.ui.content)
	},
}, [""])
