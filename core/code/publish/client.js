Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) {
		can.ui = can.onappend.layout(can)
		msg.Table(function(value) {
			can.onimport.item(can, {icons: value.icons, name: value["client.name"]}, function(event, item) {
				if (can.onmotion.cache(can, function() { return item.name }, can.ui.content)) { return }
				can.run(event, [item.name], function(msg) {
					can.onappend.table(can, msg, null, can.ui.content)
					
				})
			})
		})
	},
})