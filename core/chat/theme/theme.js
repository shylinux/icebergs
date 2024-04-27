Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { msg.Show(can)
		can.page.Select(can, can._output, "tr>td:not(:first-child):not(:last-child)", function(target) {
			can.page.style(can, target, html.BACKGROUND_COLOR, target.innerText)
		})
	},
})
