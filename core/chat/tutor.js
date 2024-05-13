Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) {
		if (can.Option(mdb.ZONE) == "_current") {
			can._root.Footer.db.tutor.Table(function(value) { msg.Push(value) }), msg.PushAction("show")
		}
		msg.Dump(can)
	},
})
Volcanos(chat.ONACTION, {
	play: function(can) {
		can.core.Next(can._msg.Table(), function(value, next, index, list) { var delay = 3000
			if (list[index+1]) { delay = (Date.parse(list[index+1].time)-Date.parse(value.time)) }
			can.onaction.show(can, value.type, value.text, delay), can.onmotion.delay(can, next, delay)
			can.onmotion.select(can, can.page.SelectOne(can, can._output, "tbody"), html.TR, index)
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
			case "click":
				can.page.Select(can, document.body, text, function(target) {
					can.core.Next([1, 2, 3, 4, 5], function(value, next, index) { can.page.ClassList.add(can, target, "picker")
						can.onmotion.delay(can, function() { can.page.ClassList.del(can, target, "picker"), can.onmotion.delay(can, function() { next() }, delay/20) }, delay/20)
					}, function() {
						target.click()
					})
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
})