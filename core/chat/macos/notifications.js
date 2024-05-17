Volcanos(chat.ONIMPORT, {_init: function(can, msg, cb) {
	can.page.Appends(can, can._output, msg.Table(function(item) { return {view: [[html.ITEM, item.status]], _init: function(target) {
		var ui = can.onappend.layout(can, [html.ICON, [[wiki.TITLE, mdb.TIME], wiki.CONTENT]], "", target)
		can.page.Append(can, ui.icon, [{img: can.misc.PathJoin(item.icon||can.page.drawText(can, item.name||item.index, 60))}])
		ui.title.innerHTML = item.name||"", ui.content.innerHTML = item.text||"", ui.time.innerHTML = item.time.split(lex.SP).pop().split(nfs.DF).slice(0, 2).join(nfs.DF)
		target.onclick = function(event) { can.sup.onexport.record(can.sup, item.index, ctx.INDEX, item)
			can.runAction(can.request(event, item), "read", [], function() { can.onappend.style(can, "read", target) })
		}
	}} })), can.onmotion.hidden(can, can._fields), can.onappend._action(can), can.page.style(can, can._action, html.DISPLAY, html.BLOCK)
}})
Volcanos(chat.ONACTION, {
	list: [cli.CLOSE, web.REFRESH, mdb.PRUNES], _trans: {refresh: "刷新", toggle: "隐藏"},
	close: function(event, can, button) { can.onmotion.hidden(can, can._fields) },
	refresh: function(event, can, button) { can.Update(event) },
})
