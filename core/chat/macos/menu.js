Volcanos(chat.ONIMPORT, {_init: function(can, msg) { can.page.Append(can, can._output, can.user.header(can)), can.page.Append(can, can._output, [
	{view: [html.ITEM, "", "notifications"], onclick: function(event) { can.sup.onexport.record(can, "notifications") }},
	{view: [html.ITEM, "", "searchs"], onclick: function(event) { can.sup.onexport.record(can, "searchs") }},
	{view: [html.MENU, "", location.hostname], onclick: function(event) { can.sup.onexport.record(can, html.DESKTOP) }},
	{view: [html.MENU, "", "+"], onclick: function(event) { can.sup.onexport.record(can, mdb.CREATE) }},
]) }})
