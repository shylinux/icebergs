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
	stop: function(event, can, msg) { can.misc.sessionStorage(can, SCRIPT_ZONE, ""), can.Update(event) },
	play: function(event, can) { var last = 0, count = 0, begin = new Date().getTime(); can.core.Next(can._msg.Table(), function(value, next, index, list, data) {
		var ls = can.core.Split(value.style||""); data = data||{}, data.list = data.list||[]; var fork
		if (data.skip > 0) { return next({skip: data.skip-1}) }
		if (data.done === 0) { return } if (data.done > 0) { data.done -= 1 } data.list.push(value)
		if (ls && ls.length > 0 && ls[0] == "fork") { data.done = parseInt(ls[1]), fork = {skip: parseInt(ls[1])} }
		if (index >= last) { last = index, can.user.toastProcess(can, `${can.core.Keys(value.space, value.index)} ${value.play} ${index}/${can._msg.Length()}`, "", index*100/list.length) }
		var tr = can.page.Select(can, can._output, html.TR)[index+1]; tr.onclick = function(event) { tr._sub && can.onmotion.scrollIntoView(can, tr._sub._target) }
		can.page.Select(can, can._output, html.TR, function(tr, i) { i-1 == index && can.onappend.style(can, value.status == mdb.DISABLE? "done": "select", tr) })
		can.Status(cli.STEP, ++count), can.Status(cli.COST, can.base.Duration(new Date().getTime()-begin))
		value.status == mdb.DISABLE? next(data): can.onaction.preview({}, can, can.request({}, value), function(data) {
			can.page.Select(can, can._output, html.TR, function(tr, i) { i-1 == index && can.onappend.style(can, "done", tr) }), next(data)
		}, data, tr)
		if (fork) { next(fork) }
	}, function(list) { can.Status(cli.STEP, list.length), can.Status(cli.COST, can.base.Duration(new Date().getTime()-begin)), can.user.toastSuccess(can) }) },
	preview: function(event, can, msg, next, data, tr) {
		can.onappend.plugin(can, {space: msg.Option(web.SPACE), index: msg.Option(ctx.INDEX), style: msg.Option(ctx.STYLE)}, function(sub) { var done = false
			function action(skip) { sub.Update(sub.request({}, {"space.timeout": "300s",_handle: ice.TRUE}), [ctx.ACTION, msg.Option(cli.PLAY)], function(msg) {
				sub.onimport._process(sub, msg) || msg.Length() == 0 && msg.Result() == "" || can.onappend._output(sub, msg), next && next(data)
			}) }
			if (msg.Option(ctx.STYLE) == "async") {
				done = true, sub.Update(sub.request({}, {"space.timeout": "300s", _handle: ice.TRUE}), [ctx.ACTION, msg.Option(cli.PLAY)]), next && next(data)
			} else {
				can.onmotion.delay(can, function() { if (done || sub._auto) { return } done = true, action() }, 300)
			}
			sub.onexport.output = function() { can.page.style(can, sub._output, html.HEIGHT, "", html.MAX_HEIGHT, "")
				if (done) { return } done = true, action(true)
			}
			tr && (tr._sub = sub)
			// msg.Option(ctx.STYLE) == html.HIDE || can.onmotion.delay(can, function() { can.onmotion.scrollIntoView(can, sub._target) }, 300)
		})
	},
})
})()
