Volcanos(chat.ONIMPORT, {_init: function(can, msg) { can.page.style(can, can._output, html.MAX_WIDTH, "")
	can.page.Append(can, can._output, can.user.header(can)), can.page.Append(can, can._output, [
		{view: [html.ITEM], list: [{icon: icon.notifications}], onclick: function(event) { can.sup.onexport.record(can, "notifications") }},
		{view: [html.ITEM], list: [{icon: icon.search}], onclick: function(event) { can.sup.onexport.record(can, "searchs") }},
		{view: [html.ITEM], list: [{icon: icon.qrcode}], onclick: function(event) { can.sup.onexport.record(can, "qrcode") }},
	].concat(msg.Table(function(item) {
		return {view: [html.ITEM], list: [{img: can.page.drawText(can, item.name||item.index, 25, 0, 20)}], onclick: function(event) { can.sup.onexport.record(can, item) }}
	}), [
		{view: [[html.MENU, "title"]], list: [{img: can.misc.ResourceFavicon(can)}, {text: can.ConfSpace()||can.misc.Search(can, ice.POD)||location.host}],
			onclick: function(event) { can.sup.onexport.record(can, html.DESKTOP) }},
		{view: [[html.MENU, mdb.ICON, web.REFRESH], "", can.page.unicode.refresh], onclick: function(event) { can.user.reload(true) }},
		{view: [[html.MENU, mdb.ICON, mdb.CREATE], "", can.page.unicode.create], onclick: function(event) { can.sup.onexport.record(can, mdb.CREATE) }},
	]))
}})
