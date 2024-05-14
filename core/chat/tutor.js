Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) {
		if (can.Option(mdb.ZONE) == "_current" && can._root.Footer.db.tutor) {
			can._root.Footer.db.tutor.Table(function(value) {
				msg.Push(value, [mdb.TIME, mdb.TYPE, mdb.TEXT])
				if (can.base.isIn(value.type, "storm", mdb.REMOVE)) {
					msg.PushButton(web.SHOW)
				} else {
					msg.PushButton(web.SHOW, "view", "data")
				}
			})
		}
		msg.Show(can)
	},
})
Volcanos(chat.ONACTION, {
	_trans: {
		icons: {
			view: "bi bi-code-slash", data: "bi bi-diagram-3",
		},
	},
	save: function(event, can) {
		can.user.input(event, can, ["zone"], function(data) {
			can.core.Next(can._msg.Table(), function(value, next, index, list) {
				can.user.toastProcess(can, `save ${data.zone} ${index}/${list.length}`)
				var args = [ctx.ACTION, mdb.INSERT, mdb.ZONE, data.zone]; can.core.Item(value, function(key, value) { args.push(key, value) })
				can.run(can.request(event, {_handle: ice.TRUE}), args, function() { next() })
			}, function() {
				can.user.toastSuccess(can)
			})
		})
	},
	play: function(can) {
		can.core.Next(can._msg.Table(), function(value, next, index, list) { var delay = 30
			if (list[index+1]) { delay = (Date.parse(list[index+1].time)-Date.parse(value.time)) }
			can.onaction.show(can, value.type, value.text, delay), can.onmotion.delay(can, next, delay)
			can.onmotion.select(can, can.page.SelectOne(can, can._output, "tbody"), html.TR, index)
			can.user.toastProcess(can, `show ${index}/${list.length} ${value.type} ${delay}ms`)
		}, function() {
			can.user.toastSuccess(can, "play done")
		})
	},
	show: function(can, type, text, delay) {
		switch (type) {
			case "theme":
				can._root.Header.onimport.theme(can._root.Header, text, {})
				break
			case "storm": var ls = text.split(",")
				can._root.River.onaction.action({}, can._root.River, ls[0], ls[1])
				break
			case "index":
				can.page.Select(can, document.body, "fieldset.panel.Header>div.output>div.Action>div._tabs>div.tabs."+text, function(target) {
					target.click()
				})
				break
			case "click":
				can.page.Select(can, document.body, text, function(target) { var count = 5
					can.core.Next(can.core.List(count), function(value, next, index) { can.page.ClassList.add(can, target, "picker")
						can.onmotion.delay(can, function() { can.page.ClassList.del(can, target, "picker"), can.onmotion.delay(can, function() { next() }, delay/5/4) }, delay/5/4)
					}, function() {
						target.click()
					})
				})
				break
			case "focus": var ls = text.split(",")
				can.page.Select(can, document.body, ls[0], function(target) {
					target.focus()
				})
				break
			case "blur": var ls = text.split(",")
				can.page.Select(can, document.body, ls[0], function(target) { target.value = ls[1]
					can.onmotion.delay(can, function() { target.blur() }, 300)
				})
				break
			case "item": var ls = text.split(",")
				can.page.Select(can, document.body, ls[0], function(target) {
					can.onmotion.delay(can, function() { target._can.sub.ui[ls[1]].click() })
				})
				break
			case "remove":
				can.page.Select(can, document.body, text, function(target) {
					can.page.Remove(can, target)
				})
				break
		}
	},
	view: function(can, type, text, delay) {
		can.onappend._float(can, "can.view", [], function(sub) {
			if (type == "click") {
				can.page.Select(can, document.body, can.core.Split(text, ">,")[0], function(target) {
					sub.Conf("_target", target)
				})
			} else if (type == "index") {
				can.page.Select(can, document.body, "fieldset.plugin."+text, function(target) {
					sub.Conf("_target", target)
				})
			}
		})
	},
	data: function(can, type, text, delay) {
		can.onappend._float(can, "can.data", [], function(sub) {
			if (type == "click") {
				can.page.Select(can, document.body, can.core.Split(text, ">,")[0], function(target) {
					sub.Conf("_target", target._can)
				})
			} else if (type == "index") {
				can.page.Select(can, document.body, "fieldset.plugin."+text, function(target) {
					sub.Conf("_target", target._can)
				})
			}
		})
	},
})