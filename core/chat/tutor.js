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
		can.core.Next(can._msg.Table(), function(value, next, index) {
			can.onaction.show(can, value.type, value.text), can.onmotion.delay(can, next, 3000)
			can.onmotion.select(can, can.page.SelectOne(can, can._output, "tbody"), html.TR, index)
		}, function() {
			can.user.toastSuccess(can, "play done")
		})
	},
	show: function(can, type, text) {
		switch (type) {
			case "storm": var ls = text.split(",")
				can._root.River.onaction.action({}, can._root.River, ls[0], ls[1])
				break
			case "theme":
				can._root.Header.onimport.theme(can._root.Header, text, {})
				break
			case "remove":
				can.page.Select(can, document.body, text, function(target) {
					can.page.Remove(can, target)
				})
				break
			case "click":
				can.page.Select(can, document.body, text, function(target) {
					target.click()
					can.page.ClassList.add(can, target, "picker")
					can.onmotion.delay(can, function() {
						can.page.ClassList.del(can, target, "picker")
					}, 3000)
				})
				break
		}
	},
})