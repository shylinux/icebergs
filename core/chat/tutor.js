Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) {
		if (can.Option(mdb.ZONE) == "_current" && can._root.Footer.db.tutor) {
			can._root.Footer.db.tutor.Table(function(value) { msg.Push(value, [mdb.TIME, mdb.TYPE, mdb.TEXT])
				if (can.base.isIn(value.type, chat.STORM, mdb.REMOVE)) {
					msg.PushButton(web.SHOW)
				} else {
					msg.PushButton(web.SHOW, mdb.VIEW, mdb.DATA)
				}
			})
		} msg.Show(can)
	},
})
Volcanos(chat.ONACTION, {_trans: {icons: {view: "bi bi-code-slash", data: "bi bi-diagram-3"}},
	save: function(event, can) {
		can.user.input(event, can, [mdb.ZONE], function(data) {
			can.core.Next(can._msg.Table(), function(value, next, index, list) { can.user.toastProcess(can, `save ${data.zone} ${index}/${list.length}`)
				var args = [ctx.ACTION, mdb.INSERT, mdb.ZONE, data.zone]; can.core.Item(value, function(k, v) { args.push(k, v) })
				can.run(can.request(event, {_handle: ice.TRUE}), args, function() { next() })
			}, function() { can.user.toastSuccess(can) })
		})
	},
	play: function(can) {
		can.core.Next(can._msg.Table(), function(value, next, index, list) { var delay = 30
			if (list[index+1]) { delay = (Date.parse(list[index+1].time)-Date.parse(value.time)) }
			can.user.toastProcess(can, `show ${index}/${list.length} ${value.type} ${delay}ms`)
			can.onmotion.select(can, can.page.SelectOne(can, can._output, html.TBODY), html.TR, index)
			can.onaction.show(can, value.type, value.text, delay, next)
		}, function() { can.user.toastSuccess(can) })
	},
	show: function(can, type, text, delay, next) { var ls = text.split(","), target = can.page.SelectOne(can, document.body, ls[0])
		switch (type) {
			case chat.THEME: can._root.Header.onimport.theme(can._root.Header, text, {}); break
			case chat.STORM: can._root.River.onaction.action({}, can._root.River, ls[0], ls[1]); break
			case ctx.INDEX: can.page.Select(can, document.body, "fieldset.panel.Header>div.output>div.Action>div._tabs>div.tabs."+text, function(target) { target.click() }); break
			case html.LEGEND:
			case html.OPTION:
			case html.ACTION:
			case html.BUTTON:
			case html.STATUS:
			case html.SUBMIT:
			case html.CANCEL:
			case html.CLOSE:
			case html.ITEM:
			case html.CLICK: var count = 5; delay = can.base.Min(delay, 3000)
				can.core.Next(can.core.List(count), function(value, next, index) { can.page.ClassList.add(can, target, html.PICKER)
					can.onmotion.delay(can, function() { can.page.ClassList.del(can, target, html.PICKER), can.onmotion.delay(can, function() { next() }, delay/count/4) }, delay/count/4)
				}, function() { target.click() }); break
			case html.FOCUS: target.focus(); break
			case html.BLUR: target.value = ls[1], can.onmotion.delay(can, function() { target.blur() }, 300); break
			case mdb.REMOVE: can.page.Remove(can, target); break
		} next && can.onmotion.delay(can, function() { next() }, delay)
	},
	view: function(can, type, text) { can.onappend._float(can, "can.view", [], function(sub) {
		can.page.Select(can, document.body, type == ctx.INDEX? "fieldset.plugin."+text: can.core.Split(text, ">,")[0], function(target) { sub.Conf("_target", target) })
	}) },
	data: function(can, type, text) { can.onappend._float(can, "can.data", [], function(sub) {
		can.page.Select(can, document.body, type == ctx.INDEX? "fieldset.plugin."+text: can.core.Split(text, ">,")[0], function(target) { sub.Conf("_target", target._can) })
	}) },
})