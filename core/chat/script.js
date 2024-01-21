(function() { SCRIPT_ZONE = "web.chat.script:zone"
Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { can.onappend.table(can, msg), can.onappend.board(can, msg)
		var zone = can.misc.sessionStorage(can, SCRIPT_ZONE), tr = can.page.Select(can, can._output, html.TR)[1]
		msg.Table(function(value, index) { zone && value.zone == zone && can.onmotion.select(can, tr.parentNode, html.TR, index, function(target) {
			can.onappend.style(can, html.DANGER, target)
		}) })
	},
})
Volcanos(chat.ONACTION, {
	record: function(event, can, msg) { can.misc.sessionStorage(can, SCRIPT_ZONE, msg.Option(mdb.ZONE)), can.user.toastSuccess(can, msg.Option(mdb.ZONE)), can.Update(event) },
	enable: function(event, can, msg) { can.runAction(event, mdb.MODIFY, [mdb.STATUS, mdb.ENABLE]) },
	disable: function(event, can, msg) { can.runAction(event, mdb.MODIFY, [mdb.STATUS, mdb.DISABLE]) },
	stop: function(event, can, msg) { can.misc.sessionStorage(can, SCRIPT_ZONE, ""), can.Update(event) },
	play: function(event, can) { can.core.Next(can._msg.Table(), function(value, next, index) {
		can.user.toastProcess(can, `${value.index} ${value.play} ${index} / ${can._msg.Length()}`)
		can.Status(cli.STEP, value.index)
		var tr = can.page.Select(can, can._output, html.TR)[1]; can.onmotion.select(can, tr.parentNode, html.TR, index)
		value.status == mdb.DISABLE? next(): can.onaction.preview({}, can, can.request({}, value), next)
	}, function() { can.user.toastSuccess(can) }) },
	preview: function(event, can, msg, next) {
		can.onappend.plugin(can, {space: msg.Option(web.SPACE), index: msg.Option(ctx.INDEX)}, function(sub) { var done = false
			function action(skip) { sub.Update(can.request({}, {_handle: ice.TRUE}), [ctx.ACTION, msg.Option(cli.PLAY)], function(msg) {
				sub.onimport._process(sub, msg) || msg.Length() == 0 && msg.Result() == "" || can.onappend._output(sub, msg), next && next() }) }
			can.onmotion.delay(can, function() { if (done || sub._auto) { return } done = true, action() }, 300)
			sub.onexport.output = function() { if (done) { return } done = true, action(true)
				can.page.style(can, sub._output, html.HEIGHT, "", html.MAX_HEIGHT, "")
			}, can.onmotion.scrollIntoView(can, sub._target)
		})
	},
})
})()
