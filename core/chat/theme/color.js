Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { msg.Show(can)
		msg.append[0] == mdb.TYPE? can.page.Select(can, can._output, html.TR, function(target, index) {
			can.page.Select(can, target, html.TD, function(target, index) { msg.Option("table.checkbox") && (index -= 1)
				msg.append[index] == mdb.NAME && can.page.style(can, target.parentNode, html.BACKGROUND_COLOR, target.innerText)
			})
		}): can.page.Select(can, can._output, html.TD, function(target, index) {
			can.page.style(can, target, html.BACKGROUND_COLOR, target.innerText.split(" ")[0])
		})
	},
})
