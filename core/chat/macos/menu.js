Volcanos(chat.ONIMPORT, {_init: function(can, msg) { can.page.style(can, can._output, html.MAX_WIDTH, "")
	can.page.Append(can, can._output, can.user.header(can)), can.page.Append(can, can._output, [
		{view: [html.ITEM], list: [{img: can.page.drawText(can, "n", 25, 0, 20)}], onclick: function(event) { can.sup.onexport.record(can, "notifications") }},
		{view: [html.ITEM], list: [{img: can.page.drawText(can, "s", 25, 0, 20)}], onclick: function(event) { can.sup.onexport.record(can, "searchs") }},
	].concat(msg.Table(function(item) {
		return {view: [html.ITEM], list: [{img: can.page.drawText(can, item.name||item.index, 25, 0, 20)}], onclick: function(event) { can.sup.onexport.record(can, item) }}
	}), [
		{view: [html.MENU, "", can.ConfSpace()||can.misc.Search(can, ice.POD)||location.host], onclick: function(event) { can.sup.onexport.record(can, html.DESKTOP) }},
		{view: [[html.MENU, mdb.CREATE], "", can.page.unicode.create], onclick: function(event) { can.sup.onexport.record(can, mdb.CREATE) }},
	]))
}})
