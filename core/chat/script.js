Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) {
		can.onappend.table(can, msg)
		can.onappend.board(can, msg)
	},
})
Volcanos(chat.ONACTION, {
	play: function(event, can) {
		can.core.Next(can._msg.Table(), function(value, next) {
			var done = false
			can.onappend.plugin(can, {index: value.index}, function(sub) {
				can.onmotion.delay(can, function() {
					if (!sub._auto) { sub.Update({}, [ctx.ACTION, value.auto], function() { next() }) }
				}, 300)
				sub.onexport.output = function() { done || sub.Update({}, [ctx.ACTION, value.auto], function() { next() }), done = true }
			})
		})
	},
})
