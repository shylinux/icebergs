Volcanos(chat.ONIMPORT, {_init: function(can, msg) { can.page.Append(can, can._output, can.user.header(can)), can.page.Append(can, can._output, [
	{view: [html.ITEM], list: [{img: can.page.drawText(can, "n", 25, 0, 20)}], onclick: function(event) { can.sup.onexport.record(can, "notifications") }},
	{view: [html.ITEM], list: [{img: can.page.drawText(can, "s", 25, 0, 20)}], onclick: function(event) { can.sup.onexport.record(can, "searchs") }},
].concat(msg.Table(function(item) {
	return {view: [html.ITEM], list: [{img: can.page.drawText(can, item.name||item.index, 25, 0, 20)}], onclick: function(event) { can.sup.onexport.record(can, item) }}
}), [
	{view: [html.MENU, "", can.Conf("_space")? can.Conf("_space"): can.user.mod.isPod? can.misc.ParseURL(can)[ice.POD]: location.host], onclick: function(event) { can.sup.onexport.record(can, html.DESKTOP) }},
	{view: [html.MENU, "", "+"], onclick: function(event) { can.sup.onexport.record(can, mdb.CREATE) }},
])) }})
