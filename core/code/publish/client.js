Volcanos(chat.ONIMPORT, {
	_init: function(can, msg) { can.ui = can.onappend.layout(can)
		msg.Table(function(value) {
			can.onimport.item(can, {icons: value.icons, name: value["client.name"]}, function(event, item, show, target) { can.db.client_name = item.name
				can.onimport.tabsCache(can, item, target, function(event) {
					can.run(event, [item.name], function(msg) {
						can.onappend.table(can, msg, null, can.ui.content)
						can.onappend._status(can, msg)
					})
				})
			})
		})
	},
})
Volcanos(chat.ONACTION, {
	download: function(event, can) {
		var msg = can.request(event); msg.Option("client.name", can.db.client_name)
		can.runAction(event, web.DOWNLOAD)
	},
})